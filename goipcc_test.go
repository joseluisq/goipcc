package goipcc

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"testing"
	"time"
)

func Test_socketReader(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name           string
		unixSocketPath string
		args           args
		want           []byte
	}{
		{
			name:           "valid socket connection and data response",
			unixSocketPath: "/tmp/mysocket",
			args: args{
				data: []byte("sample data"),
			},
			want: []byte("sample data"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Remove socket file
			exec.Command("rm", "-rf", tt.unixSocketPath).Run()

			// 2. Create a socket listening unix socket via the `socat` tool
			var out bytes.Buffer
			cmd := exec.Command("socat", "UNIX-LISTEN:"+tt.unixSocketPath+",fork", "exec:'/bin/cat'")
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Stdout = &out

			var wg sync.WaitGroup

			if err := cmd.Start(); err != nil {
				t.Errorf("%v", err)
			}

			wg.Add(1)
			go cmd.Wait()

			time.Sleep(300 * time.Millisecond)

			// 3. Create a socket client
			conn, err := net.Dial("unix", tt.unixSocketPath)
			if err != nil {
				t.Errorf("%v", err)
			}

			chanResp := make(chan ipcSockResp)
			go socketReader(conn, chanResp)

			// 4. Write testing data to current socket
			_, err = conn.Write(tt.args.data)
			if err != nil {
				t.Errorf("%v", err)
				return
			}

			defer conn.Close()

			select {
			case r := <-chanResp:
				if r.err != nil {
					t.Errorf("%v", err)
					return
				}

				// 5. Perform socket data assertions
				if !reflect.DeepEqual(r.data, tt.want) {
					t.Errorf("%v", err)
				}
				wg.Done()
			}

			wg.Wait()

			// 6. Send signal to `socat` listening process
			err = cmd.Process.Signal(os.Interrupt)
			if err != nil {
				t.Errorf("%v", err)
				return
			}
		})
	}
}

func TestConnect(t *testing.T) {
	const unixSocketFilePath = "/tmp/mysocket"

	// 1. Remove socket file
	exec.Command("rm", "-rf", unixSocketFilePath).Run()

	// 2. Create a socket listening unix socket via the `socat` tool
	var out bytes.Buffer
	cmd := exec.Command("socat", "UNIX-LISTEN:"+unixSocketFilePath+",fork", "exec:'/bin/cat'")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out

	if err := cmd.Start(); err != nil {
		t.Errorf("%v", err)
	}
	go cmd.Wait()
	time.Sleep(300 * time.Millisecond)

	type args struct {
		unixSocketFilePath string
	}
	tests := []struct {
		name        string
		args        args
		closeSocket bool
		want        *IPCSockClient
		wantErr     bool
	}{
		{
			name: "invalid socket path",
			args: args{
				unixSocketFilePath: "/tmp/mysocket-xyz",
			},
			wantErr: true,
		},
		{
			name: "valid socket path connection",
			args: args{
				unixSocketFilePath: unixSocketFilePath,
			},
			closeSocket: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Connect(tt.args.unixSocketFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.zSock == nil {
					t.Errorf("Connect() = zSock: %v, want not nil", got.zSock)
				}
				if got.zSockResp == nil {
					t.Errorf("Connect() = zSockResp: %v, want not nil", got.zSockResp)
				}
			}

			if tt.closeSocket {
				err = cmd.Process.Signal(os.Interrupt)
				if err != nil {
					t.Errorf("%v", err)
					return
				}
			}
		})
	}
}

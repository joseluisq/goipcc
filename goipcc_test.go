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
			if _, err := conn.Write(tt.args.data); err != nil {
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
			if err := cmd.Process.Signal(os.Interrupt); err != nil {
				t.Errorf("%v", err)
				return
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		unixSocketFilePath string
	}
	tests := []struct {
		name string
		args args
		want *IPCSockClient
	}{
		{
			name: "valid unix socket client instance",
			args: args{
				unixSocketFilePath: "/tmp/mysocket",
			},
			want: &IPCSockClient{
				zSocketFilePath: "/tmp/mysocket",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.unixSocketFilePath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPCSockClient_Connect(t *testing.T) {
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

	tests := []struct {
		name               string
		unixSocketFilePath string
		closeSocket        bool
		wantErr            bool
	}{
		{
			name:               "invalid unix socket connection",
			unixSocketFilePath: unixSocketFilePath + "xyz",
			wantErr:            true,
		},
		{
			name:               "valid unix socket connection",
			unixSocketFilePath: unixSocketFilePath,
			closeSocket:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.unixSocketFilePath)
			if err := c.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("IPCSockClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if c.zSock == nil {
					t.Errorf("Connect() = zSock: %v, want not nil", c.zSock)
				}
				if c.zSockResp == nil {
					t.Errorf("Connect() = zSockResp: %v, want not nil", c.zSockResp)
				}
			}

			if tt.closeSocket {
				if err := cmd.Process.Signal(os.Interrupt); err != nil {
					t.Errorf("%v", err)
					return
				}
			}
		})
	}
}

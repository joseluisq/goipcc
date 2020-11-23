package goipcc

import (
	"bytes"
	"io"
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
		r    io.Reader
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
				r:    nil,
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

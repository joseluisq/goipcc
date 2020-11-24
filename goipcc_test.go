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

// unixSocketPath is a default Unix socket path used on tests.
const unixSocketPath = "/tmp/mysocket"

// unixSocketDelay defines milliseconds pause in order to wait until
// the listening server (`socat`) is ready to accept connections.
const unixSocketDelay = 150

// listeningSocket defines a listening unix socket.
type listeningSocket struct {
	cmd *exec.Cmd
	wg  *sync.WaitGroup
}

// createListeningSocket creates a new listening unix socket using `socat` tool.
func createListeningSocket() (*listeningSocket, error) {
	exec.Command("rm", "-rf", unixSocketPath).Run()

	var out bytes.Buffer
	cmd := exec.Command("socat", "UNIX-LISTEN:"+unixSocketPath+",fork", "exec:'/bin/cat'")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go cmd.Wait()
	time.Sleep(unixSocketDelay * time.Millisecond)

	return &listeningSocket{
		wg:  &wg,
		cmd: cmd,
	}, nil
}

// close method closes current socket connection signaling it to finish.
func (s *listeningSocket) close() error {
	return s.cmd.Process.Signal(os.Interrupt)
}

func Test_socketReader(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

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
				lsock.wg.Done()
			}

			lsock.wg.Wait()
		})
	}

	// 6. Send signal to `socat` listening process
	if err := lsock.close(); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestNew(t *testing.T) {
	type args struct {
		unixSocketPath string
	}
	tests := []struct {
		name string
		args args
		want *IPCSockClient
	}{
		{
			name: "valid unix socket client instance",
			args: args{
				unixSocketPath: "/tmp/mysocket",
			},
			want: &IPCSockClient{
				zSocketFilePath: "/tmp/mysocket",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.unixSocketPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPCSockClient_Connect(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	tests := []struct {
		name           string
		unixSocketPath string
		wantErr        bool
	}{
		{
			name:           "invalid unix socket connection",
			unixSocketPath: unixSocketPath + "xyz",
			wantErr:        true,
		},
		{
			name:           "valid unix socket connection",
			unixSocketPath: unixSocketPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.unixSocketPath)
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
		})
	}

	if err := lsock.close(); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestIPCSockClient_Write(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	type args struct {
		data        []byte
		respHandler func(data []byte, err error)
	}
	tests := []struct {
		name           string
		unixSocketPath string
		args           args
		socketNil      bool
		wantN          int
		wantErr        bool
	}{
		{
			name:           "valid unix socket write without handler",
			unixSocketPath: unixSocketPath,
			args: args{
				data:        []byte(nil),
				respHandler: nil,
			},
			wantN: 0,
		},
		{
			name:           "valid unix socket write with handler",
			unixSocketPath: unixSocketPath,
			args: args{
				data:        []byte("ñ"),
				respHandler: func(data []byte, err error) {},
			},
			wantN: 2,
		},
		{
			name:           "nil socket connection reference",
			unixSocketPath: unixSocketPath,
			socketNil:      true,
			args: args{
				data: []byte(nil),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.unixSocketPath)
			// Check nil socket references on demand
			if !tt.socketNil {
				if err := c.Connect(); (err != nil) != tt.wantErr {
					t.Errorf("IPCSockClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			gotN, err := c.Write(tt.args.data, tt.args.respHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("IPCSockClient.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("IPCSockClient.Write() = %v, want %v", gotN, tt.wantN)
			}
		})
	}

	if err := lsock.close(); err != nil {
		t.Errorf("%v", err)
		return
	}
}

func TestIPCSockClient_Close(t *testing.T) {
	lsock, err := createListeningSocket()
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "close current socket connection",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(unixSocketPath)
			if err := c.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("IPCSockClient.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
			c.Close()
		})
	}

	if err := lsock.close(); err != nil {
		t.Errorf("%v", err)
		return
	}
}

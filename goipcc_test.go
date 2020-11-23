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

			chanResult := make(chan ipcSockResult)
			go socketReader(conn, chanResult)

			// 4. Write testing data to current socket
			_, err = conn.Write(tt.args.data)
			if err != nil {
				t.Errorf("%v", err)
				return
			}

			defer conn.Close()

			select {
			case r := <-chanResult:
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

func TestNew(t *testing.T) {
	type args struct {
		unixSocketFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    *IPCSockClient
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.unixSocketFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPCSockClient_Write(t *testing.T) {
	type fields struct {
		UnixSocketFilePath string
		WriteDelayMs       time.Duration
		conn               net.Conn
		chanResult         chan ipcSockResult
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &IPCSockClient{
				UnixSocketFilePath: tt.fields.UnixSocketFilePath,
				WriteDelayMs:       tt.fields.WriteDelayMs,
				zConn:              tt.fields.conn,
				zChanResult:        tt.fields.chanResult,
			}
			gotN, err := c.Write(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("IPCSockClient.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("IPCSockClient.Write() = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}

func TestIPCSockClient_Listen(t *testing.T) {
	type fields struct {
		UnixSocketFilePath string
		WriteDelayMs       time.Duration
		conn               net.Conn
		chanResult         chan ipcSockResult
	}
	type args struct {
		handler func([]byte, error)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &IPCSockClient{
				UnixSocketFilePath: tt.fields.UnixSocketFilePath,
				WriteDelayMs:       tt.fields.WriteDelayMs,
				zConn:              tt.fields.conn,
				zChanResult:        tt.fields.chanResult,
			}
			c.Listen(tt.args.handler)
		})
	}
}

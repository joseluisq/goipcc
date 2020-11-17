package goipcc

import (
	"io"
	"net"
	"reflect"
	"testing"
	"time"
)

func Test_socketReader(t *testing.T) {
	type args struct {
		r          io.Reader
		chanResult chan<- ipcSockResult
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			socketReader(tt.args.r, tt.args.chanResult)
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
		ConnTimeoutMs      time.Duration
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
				ConnTimeoutMs:      tt.fields.ConnTimeoutMs,
				conn:               tt.fields.conn,
				chanResult:         tt.fields.chanResult,
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
		ConnTimeoutMs      time.Duration
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
				ConnTimeoutMs:      tt.fields.ConnTimeoutMs,
				conn:               tt.fields.conn,
				chanResult:         tt.fields.chanResult,
			}
			c.Listen(tt.args.handler)
		})
	}
}

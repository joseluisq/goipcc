package goipcc

import (
	"io"
	"net"
	"time"
)

// IPCSockClient defines a IPC socket client
type IPCSockClient struct {
	UnixSocketFilePath string
	ConnTimeoutMs      time.Duration

	conn       net.Conn
	chanResult chan ipcSockResult
}

// ipcSockResult defines a IPC socket client result pair
type ipcSockResult struct {
	data []byte
	err  error
}

// socketReader reads socket response data
func socketReader(r io.Reader, chanResult chan<- ipcSockResult) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			chanResult <- ipcSockResult{
				data: make([]byte, 0),
				err:  err,
			}
			return
		}
		chanResult <- ipcSockResult{
			data: buf[0:n],
			err:  err,
		}
	}
}

// New creates a new IPC socket client instance configuring the communication.
func New(unixSocketFilePath string) (*IPCSockClient, error) {
	// Unix IPC socket communication
	conn, err := net.Dial("unix", unixSocketFilePath)
	if err != nil {
		return nil, err
	}
	chanResult := make(chan ipcSockResult)
	go socketReader(conn, chanResult)
	return &IPCSockClient{
		UnixSocketFilePath: unixSocketFilePath,
		ConnTimeoutMs:      500,
		conn:               conn,
		chanResult:         chanResult,
	}, nil
}

// Write writes bytes to current socket.
func (c *IPCSockClient) Write(data []byte) (n int, err error) {
	n, err = c.conn.Write(data)
	if err != nil {
		return n, err
	}
	time.Sleep(c.ConnTimeoutMs * time.Millisecond)
	return n, nil
}

// Listen listens for data socket responses.
func (c *IPCSockClient) Listen(handler func([]byte, error)) {
	defer c.conn.Close()
	for {
		r := <-c.chanResult
		handler(r.data, r.err)
	}
}

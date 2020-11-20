package goipcc

import (
	"io"
	"net"
	"time"
)

// IPCSockClient defines a IPC socket client.
type IPCSockClient struct {
	// Unix socket file path.
	UnixSocketFilePath string
	// Delay for every socket write in milliseconds (default 500ms).
	WriteDelayMs time.Duration

	zConn       net.Conn
	zChanResult chan ipcSockResult
}

// ipcSockResult defines a IPC socket client result pair.
type ipcSockResult struct {
	data []byte
	err  error
}

// socketReader reads socket response data.
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
		WriteDelayMs:       500,
		zConn:              conn,
		zChanResult:        chanResult,
	}, nil
}

// Write writes bytes to current socket.
func (c *IPCSockClient) Write(data []byte) (n int, err error) {
	n, err = c.zConn.Write(data)
	if err != nil {
		return n, err
	}
	time.Sleep(c.WriteDelayMs * time.Millisecond)
	return n, nil
}

// Listen listens for data socket responses.
func (c *IPCSockClient) Listen(handler func([]byte, error)) {
	defer c.zConn.Close()
	for {
		r := <-c.zChanResult
		handler(r.data, r.err)
	}
}

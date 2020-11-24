package goipcc

import (
	"io"
	"net"
	"sync"
)

// IPCSockClient defines a Unix IPC socket client.
type IPCSockClient struct {
	zSocketFilePath string
	zSock           net.Conn
	zSockResp       chan ipcSockResp
}

// ipcSockResp defines a Unix IPC socket client response pair.
type ipcSockResp struct {
	data []byte
	err  error
}

// socketReader reads socket response data.
func socketReader(r io.Reader, sockResp chan<- ipcSockResp) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			sockResp <- ipcSockResp{
				data: make([]byte, 0),
				err:  err,
			}
			return
		}
		sockResp <- ipcSockResp{
			data: buf[0:n],
			err:  err,
		}
	}
}

// New creates a new Unix IPC socket client instance.
func New(unixSocketFilePath string) *IPCSockClient {
	return &IPCSockClient{
		zSocketFilePath: unixSocketFilePath,
	}
}

// Connect establishes a new Unix IPC socket connection.
func (c *IPCSockClient) Connect() error {
	conn, err := net.Dial("unix", c.zSocketFilePath)
	if err != nil {
		return err
	}
	sockResp := make(chan ipcSockResp)
	go socketReader(conn, sockResp)
	c.zSock = conn
	c.zSockResp = sockResp
	return nil
}

// Write writes bytes to current socket and provides an optional response handler.
func (c *IPCSockClient) Write(data []byte, respHandler func(data []byte, err error)) (n int, err error) {
	n, err = c.zSock.Write(data)
	if err == nil && respHandler != nil {
		var res ipcSockResp
		wg := new(sync.WaitGroup)
		wg.Add(1)
		select {
		case res = <-c.zSockResp:
			respHandler(res.data, res.err)
			wg.Done()
		}
		wg.Wait()
	}
	return n, err
}

// Close closes current socket client connection.
func (c *IPCSockClient) Close() {
	c.zSock.Close()
}

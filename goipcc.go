package goipcc

import (
	"fmt"
	"io"
	"net"
)

// IPCSockClient defines an Unix IPC socket client.
type IPCSockClient struct {
	socketFilePath string
	sock           net.Conn
	sockResp       chan ipcSockResp
}

// ipcSockResp defines an Unix IPC socket client response pair.
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
		socketFilePath: unixSocketFilePath,
	}
}

// Connect establishes a new Unix IPC socket connection.
func (c *IPCSockClient) Connect() error {
	conn, err := net.Dial("unix", c.socketFilePath)
	if err != nil {
		return err
	}
	sockResp := make(chan ipcSockResp)
	go socketReader(conn, sockResp)
	c.sock = conn
	c.sockResp = sockResp
	return nil
}

// Write writes bytes to current socket. It also provides an optional data response handler.
// When a `respHandler` function is provided then three params are provided: `data []byte`, `err error`, `done func()`.
// The `done()` function param acts as a callback completion in order to finish the current write execution.
func (c *IPCSockClient) Write(data []byte, respHandler func(data []byte, err error, done func())) (n int, err error) {
	if c.sock == nil {
		return 0, fmt.Errorf("no available unix socket connection to write")
	}
	n, err = c.sock.Write(data)
	if err == nil && respHandler != nil {
		var res ipcSockResp
		quitCh := make(chan struct{})
	loop:
		for {
			select {
			case <-quitCh:
				break loop
			case res = <-c.sockResp:
				respHandler(res.data, res.err, func() {
					close(quitCh)
				})
			}
		}
	}
	return n, err
}

// Close closes current socket client connection.
func (c *IPCSockClient) Close() {
	if c.sock != nil {
		c.sock.Close()
	}
}

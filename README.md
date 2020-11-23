# goipcc [![Build Status](https://travis-ci.com/joseluisq/goipcc.svg?branch=master)](https://travis-ci.com/joseluisq/goipcc) [![codecov](https://codecov.io/gh/joseluisq/goipcc/branch/master/graph/badge.svg)](https://codecov.io/gh/joseluisq/goipcc) [![Go Report Card](https://goreportcard.com/badge/github.com/joseluisq/goipcc)](https://goreportcard.com/report/github.com/joseluisq/goipcc) [![GoDoc](https://godoc.org/github.com/joseluisq/goipcc?status.svg)](https://pkg.go.dev/github.com/joseluisq/goipcc)

> A simple [Unix IPC Socket](https://en.wikipedia.org/wiki/Unix_domain_socket) client for [Go](https://golang.org/pkg/net/).

**Status:** WIP

## Usage

```go
package main

import (
	"log"
	"strings"

	"github.com/joseluisq/goipcc"
)

func main() {
	// Code for example purposes only

    // 1. Create a simple listening Unix socket with echo functionality
    // using the `socat` tool -> http://www.dest-unreach.org/socat/
	// Then execute the following commands on your terminal:
	//  rm -f /tmp/mysocket && socat UNIX-LISTEN:/tmp/mysocket,fork exec:'/bin/cat'

    // 2. Now just run this client code example in order to exchange data with current socket.
    //  go run examples/main.go

    // 2.1 Connect to the listening socket
	sock, err := goipcc.Connect("/tmp/mysocket")
	if err != nil {
		log.Fatalln("unable to communicate with socket:", err)
	}

	// 2.2 Send some sequential data to current socket (example only)
	pangram := strings.Split("The quick brown fox jumps over the lazy dog", " ")
	for _, word := range pangram {
		log.Println("client data sent:", word)
		_, err := sock.Write([]byte(word), func(resp []byte, err error) {
			log.Println("client data received:", string(resp))
		})
		if err != nil {
			log.Fatalln("unable to write to socket:", err)
		}
	}

	sock.Close()

    // 3. Finally after running the client you'll see a similar output like:
    //
    // 2020/11/24 00:39:27 client data sent: The
    // 2020/11/24 00:39:27 client data received: The
    // 2020/11/24 00:39:28 client data sent: quick
    // 2020/11/24 00:39:28 client data received: quick
    // 2020/11/24 00:39:29 client data sent: brown
    // 2020/11/24 00:39:29 client data received: brown
    // 2020/11/24 00:39:30 client data sent: fox
    // 2020/11/24 00:39:30 client data received: fox
    // 2020/11/24 00:39:31 client data sent: jumps
    // 2020/11/24 00:39:31 client data received: jumps
    // 2020/11/24 00:39:32 client data sent: over
    // 2020/11/24 00:39:32 client data received: over
    // 2020/11/24 00:39:33 client data sent: the
    // 2020/11/24 00:39:33 client data received: the
    // 2020/11/24 00:39:34 client data sent: lazy
    // 2020/11/24 00:39:34 client data received: lazy
    // 2020/11/24 00:39:35 client data sent: dog
    // 2020/11/24 00:39:35 client data received: dog
}
```

## Contributions

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in current work by you, as defined in the Apache-2.0 license, shall be dual licensed as described below, without any additional terms or conditions.

Feel free to send some [Pull request](https://github.com/joseluisq/goipcc/pulls) or [issue](https://github.com/joseluisq/goipcc/issues).

## License

This work is primarily distributed under the terms of both the [MIT license](LICENSE-MIT) and the [Apache License (Version 2.0)](LICENSE-APACHE).

© 2020-present [Jose Quintana](https://git.io/joseluisq)

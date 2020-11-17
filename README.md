# goipcc [![Build Status](https://travis-ci.com/joseluisq/goipcc.svg?branch=master)](https://travis-ci.com/joseluisq/goipcc) [![codecov](https://codecov.io/gh/joseluisq/goipcc/branch/master/graph/badge.svg)](https://codecov.io/gh/joseluisq/goipcc) [![Go Report Card](https://goreportcard.com/badge/github.com/joseluisq/goipcc)](https://goreportcard.com/report/github.com/joseluisq/goipcc) [![GoDoc](https://godoc.org/github.com/joseluisq/goipcc?status.svg)](https://pkg.go.dev/github.com/joseluisq/goipcc)

> A simple [Unix IPC Socket](https://en.wikipedia.org/wiki/Unix_domain_socket) client for [Go](https://golang.org/pkg/net/).

**Status:** WIP

## Usage

```go
package main

import (
    "log"
    "os"
    "strings"

    "github.com/joseluisq/goipcc"
)

func main() {
    // Code for example purposes only

    // 1. Create a listening unix socket via the `socat` tool (example purposes only)
    // On your terminal execute:
    // 	rm -f /tmp/mysocket && socat UNIX-LISTEN:/tmp/mysocket -

    // 2. Then just run the client example in order to exchange data with current socket
    ipc, err := goipcc.New("/tmp/mysocket")
    if err != nil {
        log.Println("unable to communicate with socket:", err)
        os.Exit(1)
    }

    // 3. Send many data requests (example purposes only)
    pangrama := strings.Split("The quick brown fox jumps over the lazy dog", " ")
    for _, word := range pangrama {
        _, err := ipc.Write([]byte(word + "\n"))
        if err != nil {
            log.Fatalln("unable to write to socket:", err)
            break
        }
        log.Println("client data sent:", word)
    }

    // 4. Listen for socket data responses
    ipc.Listen(func(data []byte, err error) {
        if err != nil {
            log.Fatalln("unable to get data:", err)
        }
        log.Println("client data got:", string(data))
    })

    // 5. Finally after running the client you'll see the output on both sides
    // with a slight delay on purpose (see `ConnTimeoutMs` prop):
    //
    // 	The
    // 	quick
    // 	brown
    // 	fox
    // 	jumps
    // 	over
    // 	the
    // 	lazy
    // 	dog
}
```

## Contributions

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in current work by you, as defined in the Apache-2.0 license, shall be dual licensed as described below, without any additional terms or conditions.

Feel free to send some [Pull request](https://github.com/joseluisq/goipcc/pulls) or [issue](https://github.com/joseluisq/goipcc/issues).

## License

This work is primarily distributed under the terms of both the [MIT license](LICENSE-MIT) and the [Apache License (Version 2.0)](LICENSE-APACHE).

© 2020-present [Jose Quintana](https://git.io/joseluisq)

package main

import (
	"log"
	"strings"

	"github.com/joseluisq/goipcc"
)

func main() {
	// Code for example purposes only

	// 1. Create a listening unix socket via the `socat` tool
	// On your terminal execute:
	// rm -f /tmp/mysocket && socat UNIX-LISTEN:/tmp/mysocket -

	// 2. Then just run the client example in order to exchange data with current socket
	ipc, err := goipcc.New("/tmp/mysocket")
	if err != nil {
		log.Fatalln("unable to communicate with socket:", err)
	}

	// Send some sequential data to current socket
	pangram := strings.Split("The quick brown fox jumps over the lazy dog", " ")
	for _, word := range pangram {
		_, err := ipc.Write([]byte(word + "\n"))
		if err != nil {
			log.Fatalln("unable to write to socket:", err)
		}
		log.Println("client data sent:", word)
	}

	// Listen for socket responses
	ipc.Listen(func(data []byte, err error) {
		if err != nil {
			log.Fatalln("unable to get data:", err)
		}
		log.Println("client data got:", string(data))
	})
}

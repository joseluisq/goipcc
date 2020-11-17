package main

import (
	"log"
	"os"
	"strings"

	"github.com/joseluisq/goipcc"
)

func main() {
	// 1. Create a listening unix socket via the `socat` tool (for example purposes only)
	// On your terminal execute:
	// rm -f /tmp/mysocket && socat UNIX-LISTEN:/tmp/mysocket -

	// 2. Then just run the client example in order to exchange data with current socket
	ipc, err := goipcc.New("/tmp/mysocket")
	if err != nil {
		log.Println("unable to communicate with socket:", err)
		os.Exit(1)
	}

	// Send many requests (example purposes only)
	pangrama := strings.Split("The quick brown fox jumps over the lazy dog", " ")
	for _, word := range pangrama {
		_, err := ipc.Write([]byte(word + "\n"))
		if err != nil {
			log.Fatalln("unable to write to socket:", err)
			break
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

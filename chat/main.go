package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Uso: go run . [porta] [apelido] [ips]")
		os.Exit(1)
	}

	port := os.Args[1]
	name := os.Args[2]
	knownPeers := os.Args[3:]

	node := NewNode(name)

	go node.StartListening(port)

	node.ConnectToPeers(knownPeers)

	StartSession(name, port, node)
}

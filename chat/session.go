package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func StartSession(name string, port string, node *Node) {
	fmt.Printf("%s rodando na porta %s. ctrl + c para sair.\n", name, port)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		if strings.EqualFold(line, "/quit") {
			fmt.Println("saindo do chat...")
			os.Exit(0)
		}

		if strings.EqualFold(line, "/list") {
			node.ListConnections()
			continue
		}

		node.Broadcast(line)
	}
}

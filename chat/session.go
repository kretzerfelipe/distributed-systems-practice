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

		if strings.EqualFold(line, QuitCommand) {
			fmt.Println("saindo do chat...")
			os.Exit(0)
		}

		if strings.EqualFold(line, ListCommand) {
			node.ListConnections()
			continue
		}

		if strings.HasPrefix(line, PrivateMsgCommand) {
			messageGroup := strings.SplitN(line, " ", 3)

			if len(messageGroup) < 3 {
				fmt.Println("[-] mensagem incorreta, /msg [apelido] [texto]")
			}

			node.SendPrivate(strings.TrimSpace(messageGroup[1]), messageGroup[2])
			continue
		}

		node.Broadcast(line)
	}
}

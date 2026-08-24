package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type Node struct {
	Name        string
	mu          sync.Mutex
	connections map[net.Conn]struct{}
}

func NewNode(name string) *Node {
	return &Node{
		Name:        name,
		connections: make(map[net.Conn]struct{}),
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Uso: go run . [porta] [apelido] [ip:porta conhecidos...]")
		os.Exit(1)
	}

	port := os.Args[1]
	name := os.Args[2]
	knownPeers := os.Args[3:]

	node := NewNode(name)

	go node.StartListening(port)

	node.ConnectToPeers(knownPeers)

	fmt.Printf("Nó '%s' rodando na porta %s. Pressione Ctrl+C para sair.\n", name, port)
	select {}
}

func (n *Node) StartListening(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Erro ao abrir porta %s: %v\n", port, err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}

		fmt.Printf("[+] Nova conexão recebida de: %s\n", conn.RemoteAddr())
		n.AddConnection(conn)
	}
}

func (n *Node) ConnectToPeers(peers []string) {
	for _, peerAddr := range peers {
		conn, err := net.DialTimeout("tcp", peerAddr, 3*time.Second)
		if err != nil {
			fmt.Printf("[-] Não foi possível conectar ao par %s: %v\n", peerAddr, err)
			continue
		}

		fmt.Printf("[+] Conectado com sucesso ao par: %s\n", peerAddr)
		n.AddConnection(conn)
	}
}

func (n *Node) AddConnection(conn net.Conn) {
	n.mu.Lock()
	n.connections[conn] = struct{}{}
	n.mu.Unlock()

	go n.handleConnection(conn)
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		frame, err := ReadFrame(conn)
		if err != nil {
			fmt.Printf("[-] Um participante desconectou: %s\n", conn.RemoteAddr())

			n.RemoveConnection(conn)
			return
		}

		fmt.Printf("%s\n", string(frame))
	}
}

func (n *Node) RemoveConnection(conn net.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.connections, conn)
}

func (n *Node) Broadcast(texto string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	msg := fmt.Sprintf("%s: %s", n.Name, texto)
	payload := []byte(msg)

	for conn := range n.connections {

		err := WriteFrame(conn, payload)

		if err != nil {
			fmt.Printf("[-] Falha ao enviar para %s\n", conn.RemoteAddr())
		}
	}
}

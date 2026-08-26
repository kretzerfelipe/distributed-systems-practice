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
	connections map[net.Conn]chan []byte
	mu          sync.Mutex
}

func NewNode(name string) *Node {
	return &Node{
		Name:        name,
		connections: make(map[net.Conn]chan []byte),
	}
}

func (n *Node) StartListening(port string) {
	// (anotação de entendimento) quando chega aqui, trava até achar alguém se conectar
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		msg := "erro ao abrir porta: " + port
		SendError(msg, err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		// (anotação de entendimento) destrava quando acha alguem
		conn, err := listener.Accept()
		if err != nil {
			SendError("[-] erro ao aceitar conexão:", err)
			continue
		}

		fmt.Printf("[+] nova conexão: %s\n", conn.RemoteAddr())
		n.AddConnection(conn)
	}
}

func (n *Node) ConnectToPeers(peers []string) {
	for _, peerAddr := range peers {
		conn, err := net.DialTimeout("tcp", peerAddr, 3*time.Second)
		if err != nil {
			fmt.Printf("[-] não foi possível conectar ao par %s: %v\n", peerAddr, err)
			continue
		}

		fmt.Printf("[+] conectado com sucesso ao par: %s\n", peerAddr)
		n.AddConnection(conn)
	}
}

func (n *Node) AddConnection(conn net.Conn) {
	n.mu.Lock()
	msgCh := make(chan []byte, 100)
	n.connections[conn] = msgCh
	n.mu.Unlock()

	go n.handleConnection(conn)
	go n.writeLoop(conn, msgCh)
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		frame, err := ReadFrame(conn)
		if err != nil {
			fmt.Printf("[-] %s desconectou: \n", conn.RemoteAddr())

			n.RemoveConnection(conn)
			return
		}

		fmt.Println(string(frame))
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

	message := fmt.Sprintf("%s: %s", n.Name, texto)
	payload := []byte(message)

	for conn, msgCh := range n.connections {
		select {
		case msgCh <- payload:
		default:
			fmt.Printf("[-] %s está travado. Desconectando...\n", conn.RemoteAddr())
			conn.Close()
		}
	}
}

func (n *Node) writeLoop(conn net.Conn, msgCh chan []byte) {
	for msg := range msgCh {
		err := WriteFrame(conn, msg)
		if err != nil {
			SendError("erro ao escrever mensagem", err)
			return
		}
	}
}

func (n *Node) ListConnections() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if len(n.connections) == 0 {
		fmt.Println("[-] você não está conectado a nenhum participante no momento.")
		return
	}

	fmt.Printf("[i] você está conectado a %s participante(s):\n", len(n.connections))

	for conn := range n.connections {
		fmt.Printf("    - %s\n", conn.RemoteAddr().String())
	}
}

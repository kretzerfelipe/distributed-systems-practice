package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Node struct {
	Name        string
	connections map[net.Conn]*Peer
	mu          sync.Mutex
}

func NewNode(name string) *Node {
	return &Node{
		Name:        name,
		connections: make(map[net.Conn]*Peer),
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
		// doc: tenta estabelecer conexão em até 3 segundos, caso contrário não conecta
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
	// doc: cria canal com limite de 100 mensagens, se o usuário não recebe-las a tempo
	// a esteira vai travar e o usuário vai ser descontado por conta da lentidão
	msgCh := make(chan []byte, 100)
	newPeer := Peer{
		Name:  "",
		MsgCh: msgCh,
	}
	n.connections[conn] = &newPeer
	n.mu.Unlock()

	go n.handleConnection(conn)
	go n.writeLoop(conn, msgCh)
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()
	handshakeMsg := fmt.Sprintf("/start-connection %s", n.Name)
	WriteFrame(conn, []byte(handshakeMsg))

	for {
		frame, err := ReadFrame(conn)
		if err != nil {
			fmt.Printf("[-] %s desconectou: \n", conn.RemoteAddr())

			n.RemoveConnection(conn)
			return
		}

		text := string(frame)

		if strings.HasPrefix(text, StartConnection) {
			peerName := strings.TrimPrefix(text, StartConnection)

			n.mu.Lock()
			peer, exists := n.connections[conn]
			if exists {
				peer.Name = strings.TrimSpace(peerName)
			}
			n.mu.Unlock()

			continue
		}

		fmt.Println(text)
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

	for conn, peer := range n.connections {
		select {
		case peer.MsgCh <- payload:
		default:
			// doc: cai aqui no dafault quando a esteira está travada com 100 mensagens
			// desconecta usuário devido a lentidão
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

	fmt.Printf("[i] você está conectado a %d participante(s):\n", len(n.connections))

	for conn, peer := range n.connections {
		fmt.Printf("Nome: %s - %s\n", peer.Name, conn.RemoteAddr().String())
	}
}

func (n *Node) SendPrivate(targetName string, texto string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	message := fmt.Sprintf("[Privado de %s]: %s", n.Name, texto)
	payload := []byte(message)
	found := false

	for conn, peer := range n.connections {
		if strings.EqualFold(peer.Name, targetName) {
			found = true
			select {
			case peer.MsgCh <- payload:
			default:
				// doc: cai aqui no dafault quando a esteira está travada com 100 mensagens
				// desconecta usuário devido a lentidão
				fmt.Printf("[-] %s está travado. Desconectando...\n", peer.Name)
				conn.Close()
			}
			break
		}
	}

	if !found {
		fmt.Printf("[-] '%s' não encontrado.\n", targetName)
	}
}

package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
)

const QuitCommand = "/quit"

func RunSession(ctx context.Context, conn net.Conn, name string) {
	fmt.Printf("Connected to %s.\n", conn.RemoteAddr())
	fmt.Printf("You are '%s'. Type a message and press Enter. Type %s to leave.\n\n", name, QuitCommand)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	done := make(chan struct{}, 2)

	go func() {
		receiveLoop(conn)
		done <- struct{}{}
	}()

	go func() {
		sendLoop(conn, name)
		done <- struct{}{}
	}()

	<-done
	cancel()

	fmt.Println("Session closed.")
}

func receiveLoop(conn net.Conn) {
	for {
		frame, err := ReadFrame(conn)
		if err != nil {
			fmt.Println("[peer left the conversation]")
			return
		}
		fmt.Println(string(frame))
	}
}

func sendLoop(conn net.Conn, name string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.EqualFold(line, QuitCommand) {
			return
		}

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		payload := fmt.Sprintf("%s: %s", name, line)
		if err := WriteFrame(conn, []byte(payload)); err != nil {
			fmt.Fprintf(os.Stderr, "[send failed: %v]\n", err)
			return
		}
	}
}

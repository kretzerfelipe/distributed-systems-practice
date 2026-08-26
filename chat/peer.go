package main

type Peer struct {
	Name  string
	MsgCh chan []byte
}

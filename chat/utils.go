package main

import "fmt"

func SendError(msg string, err error) {
	fmt.Printf("[ERROR]: Msg: %s, Err: %v", msg, err)
}

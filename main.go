package main

import (
	"log"
	"ssh-go/sshsrv"
	"ssh-go/utils"

	"golang.org/x/crypto/ssh"
)

func runShell(ch ssh.Channel) {
	defer ch.Close()

	buf := make([]byte, 1024)

	for {
		n, err := ch.Read(buf)
		if err != nil {
			return
		}

		ch.Write(buf[:n]) // echo back
	}
}

func handleSession(ch ssh.Channel, reqs <-chan *ssh.Request) {
	for req := range reqs {
		switch req.Type {
		case "pty-req":
			utils.HandlePty(*req)
		case "window-change":
			utils.HandleWindowChange(*req)
		case "shell":
			req.Reply(true, nil)

			go runShell(ch)

		default:
			log.Printf("Request type %v not implemented!", req.Type)
			req.Reply(false, nil)
		}
	}
}

func main() {
	sshsrv.RunServer(2222, handleSession)
}

package main

import (
	"bufio"
	"log"
	"ssh-go/sshsrv"

	"golang.org/x/crypto/ssh"
)

func runShell(ch ssh.Channel) {
	defer ch.Close()

	scanner := bufio.NewScanner(ch)

	for {
		ch.Write([]byte("> "))

		if !scanner.Scan() {
			log.Println("scanner ended")
			return
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}

		log.Printf("received: %q", scanner.Text())
	}
}

func handleSession(ch ssh.Channel, reqs <-chan *ssh.Request) {
	for req := range reqs {
		switch req.Type {

		case "pty-req":
			req.Reply(true, nil)

		case "shell":
			req.Reply(true, nil)

			go runShell(ch)

		default:
			req.Reply(false, nil)
			log.Printf("Request type %v not implemented!", req.Type)
		}
	}
}

func main() {
	sshsrv.RunServer(2222, handleSession)
}

package main

import (
	"ssh-go/sshsrv"

	"golang.org/x/crypto/ssh"
)

func runShell(state *sshsrv.SessionState, req ssh.Request) {
	go func() {
		defer state.Close()

		buf := make([]byte, 1024)

		for {
			n, err := state.Channel.Read(buf)
			if err != nil {
				return
			}

			state.Channel.Write(buf[:n]) // echo back
		}
	}()
}

func main() {
	reqHandlers := sshsrv.RequestHandlers{
		"shell": runShell,
	}

	sshsrv.RunServer(2222, reqHandlers)
}

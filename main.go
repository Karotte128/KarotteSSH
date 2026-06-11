package main

import (
	"ssh-go/sshsrv"
)

func main() {
	reqHandlers := sshsrv.RequestHandlers{}

	sshsrv.RunServer(2222, reqHandlers)
}

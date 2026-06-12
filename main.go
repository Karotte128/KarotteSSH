package main

import (
	"ssh-go/sshsrv"
)

func main() {
	config := sshsrv.NewConfig()
	sshsrv.RunServer(config)
}

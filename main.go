package main

import (
	"ssh-go/sshsrv"
)

func main() {
	config := sshsrv.Config{}
	config.Authentication.EnablePublicKeyAuth = true

	sshsrv.RunServer(config)
}

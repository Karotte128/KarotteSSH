package main

import (
	"path"
	"ssh-go/sshsrv"
)

func main() {
	reqHandlers := sshsrv.RequestHandlers{}
	config := sshsrv.Config{
		Port:               2222,
		PrivateKeyFile:     path.Join(".ssh", "key"),
		AuthorizedKeysFile: path.Join(".ssh", "authorized_keys"),
	}

	sshsrv.RunServer(config, reqHandlers)
}

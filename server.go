package karottessh

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
)

func handleServerConn(channels <-chan ssh.NewChannel) {
	for newChannel := range channels {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		ch, reqs, err := newChannel.Accept()
		if err != nil {
			log.Printf("Error accepting channel: %v", err)
			continue
		}

		go handleSession(ch, reqs)
	}
}

func listen(config *ssh.ServerConfig, port int) {
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}

		_, chans, reqs, err := ssh.NewServerConn(conn, config)
		if err != nil {
			log.Printf("Handshake error: %v", err)
			continue
		}

		go ssh.DiscardRequests(reqs)
		go handleServerConn(chans)
	}
}

func getHostKey(keyPath string) (ssh.Signer, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func RunServer(config Config) error {
	sshConfig := &ssh.ServerConfig{}

	sshConfig.NoClientAuth = config.Authentication.EnableNoAuth

	if config.Authentication.EnablePublicKeyAuth {
		sshConfig.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return config.Authentication.PublicKeyHandler(conn, key, config.Authentication.Attributes)
		}
	}

	if config.Authentication.EnablePasswordAuth {
		sshConfig.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			return config.Authentication.PasswordHandler(conn, password, config.Authentication.Attributes)
		}
	}

	key, err := getHostKey(config.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("Error loading host key: %v", err)
	}

	sshConfig.AddHostKey(key)

	setRequestHandlers(config.RequestHandlers)

	listen(sshConfig, config.Port)

	return nil
}

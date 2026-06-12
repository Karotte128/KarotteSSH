package sshsrv

import (
	"bytes"
	"errors"
	"log"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	Port               int
	PrivateKeyFile     string
	AuthorizedKeysFile string
}

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

func authorizePublicKey(key ssh.PublicKey, keyFile string) (bool, error) {
	authorizedKeysBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return false, err
	}

	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			return false, err
		}

		if bytes.Equal(pubKey.Marshal(), key.Marshal()) {
			return true, nil
		}

		authorizedKeysBytes = rest
	}

	return false, nil
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

func RunServer(config Config, reqHandlers RequestHandlers) {
	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			authorized, err := authorizePublicKey(key, config.AuthorizedKeysFile)
			if err != nil {
				return nil, err
			}

			if authorized {
				return &ssh.Permissions{}, nil
			} else {
				return nil, errors.New("unauthorized")
			}

		},
	}

	key, err := getHostKey(config.PrivateKeyFile)
	if err != nil {
		log.Fatalf("Error loading host key: %v", err)
	}
	serverConfig.AddHostKey(key)

	setRequestHandlers(reqHandlers)

	listen(serverConfig, config.Port)
}

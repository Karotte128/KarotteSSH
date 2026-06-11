package main

import (
	"errors"
	"log"
	"net"
	"os"
	"path/filepath"
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

		go func(in <-chan *ssh.Request) {
			defer func() {
				_ = ch.Close()
			}()

			for req := range in {
				log.Printf("request: %s wantReply=%v payload=%x", req.Type, req.WantReply, req.Payload)
				req.Reply(true, nil)
				// do something?
			}
		}(reqs)
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

func authorizePublicKey(key ssh.PublicKey) (bool, error) {
	fp := ssh.FingerprintSHA256(key)
	log.Println(fp)
	return true, nil // TODO: check if public key is authorized
}

func getHostKeys() ([]ssh.Signer, error) {
	var hostKeys []ssh.Signer

	keyPath := filepath.Join(".", "ssh", "key")

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, err
	}

	hostKeys = append(hostKeys, signer)
	return hostKeys, nil
}

func RunServer(port int) {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			authorized, err := authorizePublicKey(key)
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

	keys, err := getHostKeys()
	if err != nil {
		log.Fatalf("Error loading host keys: %v", err)
	}
	for _, key := range keys {
		config.AddHostKey(key)
	}

	listen(config, port)
}

func main() {
	RunServer(2222)
}

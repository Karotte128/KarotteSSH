package sshsrv

import (
	"bytes"
	"errors"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
)

type PublicKeyAuth func(conn ssh.ConnMetadata, key ssh.PublicKey, attributes map[string]string) (*ssh.Permissions, error)

type Authentication struct {
	Attributes          map[string]string
	PublicKeyHandler    PublicKeyAuth
	EnablePublicKeyAuth bool
	EnableNoAuth        bool
}

func defaultPublicKeyAuth(conn ssh.ConnMetadata, key ssh.PublicKey, attributes map[string]string) (*ssh.Permissions, error) {
	if attributes["authorized_keys_file"] == "" {
		attributes["authorized_keys_file"] = path.Join(".ssh", "authorized_keys")
	}

	authorizedKeysBytes, err := os.ReadFile(attributes["authorized_keys_file"])
	if err != nil {
		return nil, err
	}

	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(pubKey.Marshal(), key.Marshal()) {
			return &ssh.Permissions{}, nil
		}

		authorizedKeysBytes = rest
	}

	return nil, errors.New("unauthorized")
}

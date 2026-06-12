package sshsrv

import (
	"bytes"
	"errors"
	"os"

	"golang.org/x/crypto/ssh"
)

type Authentication struct {
	Attributes          map[string]string
	PublicKeyHandler    func(conn ssh.ConnMetadata, key ssh.PublicKey, attributes map[string]string) (*ssh.Permissions, error)
	EnablePublicKeyAuth bool
	EnableNoAuth        bool
}

func defaultPublicKeyAuth(conn ssh.ConnMetadata, key ssh.PublicKey, attributes map[string]string) (*ssh.Permissions, error) {
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

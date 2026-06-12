package karottessh

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type Authentication struct {
	Attributes          map[string]string
	PublicKeyHandler    func(conn ssh.ConnMetadata, key ssh.PublicKey, attributes map[string]string) (*ssh.Permissions, error)
	EnablePublicKeyAuth bool
	PasswordHandler     func(conn ssh.ConnMetadata, password []byte, attributes map[string]string) (*ssh.Permissions, error)
	EnablePasswordAuth  bool
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

type passwordDB map[string][]byte

func loadPasswordDB(path string) (passwordDB, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	db := make(passwordDB)

	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		user, hash, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid password file format at line %d", lineNum)
		}

		user = strings.TrimSpace(user)
		hash = strings.TrimSpace(hash)

		if user == "" || hash == "" {
			return nil, fmt.Errorf("invalid password file format at line %d", lineNum)
		}

		db[user] = []byte(hash)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return db, nil
}

func defaultPasswordAuth(conn ssh.ConnMetadata, password []byte, attributes map[string]string) (*ssh.Permissions, error) {
	db, err := loadPasswordDB(attributes["password_file"])
	if err != nil {
		return nil, fmt.Errorf("Error reading password file: %d", err)
	}

	hash, ok := db[conn.User()]
	if !ok {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword(hash, password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return &ssh.Permissions{}, nil
}

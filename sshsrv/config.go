package sshsrv

import "path"

type Config struct {
	Port            int
	PrivateKeyFile  string
	RequestHandlers RequestHandlers
	Authentication  Authentication
}

func NewConfig() Config {
	config := Config{
		Port:           2222,
		PrivateKeyFile: path.Join(".ssh", "key"),
		RequestHandlers: RequestHandlers{
			"pty-req":       defaultHandlePty,
			"window-change": defaultHandleWindowChange,
			"exec":          defaultHandleExec,
			"shell":         defaultHandleShell,
		},
		Authentication: Authentication{
			Attributes: map[string]string{
				"authorized_keys_file": path.Join(".ssh", "authorized_keys"),
				"password_file":        path.Join(".ssh", "passwords"),
			},
			PublicKeyHandler:    defaultPublicKeyAuth,
			EnablePublicKeyAuth: true,
			PasswordHandler:     defaultPasswordAuth,
			EnablePasswordAuth:  false,
			EnableNoAuth:        false,
		},
	}

	return config
}

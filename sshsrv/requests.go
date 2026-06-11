package sshsrv

import (
	"log"

	"golang.org/x/crypto/ssh"
)

type PtyRequest struct {
	Term          string
	WidthChars    uint32
	HeightRows    uint32
	WidthPixels   uint32
	HeightPixels  uint32
	TerminalModes string
}

type WindowChange struct {
	WidthChars   uint32
	HeightRows   uint32
	WidthPixels  uint32
	HeightPixels uint32
}

func defaultHandlePty(_ *SessionState, req ssh.Request) {
	var pty PtyRequest

	if err := ssh.Unmarshal(req.Payload, &pty); err != nil {
		log.Printf("failed to parse pty request: %v", err)

		if req.WantReply {
			req.Reply(false, nil)
		}
		return
	}

	log.Printf(
		"PTY: term=%s size=%dx%d",
		pty.Term,
		pty.WidthChars,
		pty.HeightRows,
	)

	req.Reply(true, nil)
}

func defaultHandleWindowChange(_ *SessionState, req ssh.Request) {
	var wc WindowChange

	if err := ssh.Unmarshal(req.Payload, &wc); err == nil {
		log.Printf(
			"resize: %dx%d",
			wc.WidthChars,
			wc.HeightRows,
		)
	}

	req.Reply(false, nil)
}

func defaultHandleShell(state *SessionState, req ssh.Request) {
	req.Reply(true, nil)

	go func() {
		defer state.Close()

		buf := make([]byte, 1024)

		for {
			n, err := state.Channel.Read(buf)
			if err != nil {
				return
			}

			state.Channel.Write(buf[:n]) // echo back
		}
	}()
}

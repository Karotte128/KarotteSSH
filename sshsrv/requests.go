package sshsrv

import (
	"fmt"
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

type ExecRequest struct {
	Command string
}

func sendExitStatus(ch ssh.Channel, status uint32) {
	_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{Status: status}))
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

func defaultHandleExec(state *SessionState, req ssh.Request) {
	defer state.Close()

	var payload ExecRequest

	if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
		req.Reply(false, nil)
		return
	}

	req.Reply(true, nil)

	resp := fmt.Sprintf("exec command: %q\n", payload.Command)

	_, err := state.Channel.Write([]byte(resp))
	if err != nil {
		log.Printf("error writing to channel: %v", err)
	}

	sendExitStatus(state.Channel, 0)
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

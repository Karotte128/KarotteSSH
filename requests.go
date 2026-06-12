package karottessh

import (
	"fmt"
	"log"

	"golang.org/x/crypto/ssh"
)

type ptyRequest struct {
	Term          string
	WidthChars    uint32
	HeightRows    uint32
	WidthPixels   uint32
	HeightPixels  uint32
	TerminalModes string
}

type windowChange struct {
	WidthChars   uint32
	HeightRows   uint32
	WidthPixels  uint32
	HeightPixels uint32
}

type execRequest struct {
	Command string
}

func defaultHandlePty(state *Session, req ssh.Request) {
	var pty ptyRequest

	if err := ssh.Unmarshal(req.Payload, &pty); err != nil {
		log.Printf("failed to parse pty request: %v", err)

		if req.WantReply {
			req.Reply(false, nil)
		}
		return
	}

	state.SetStorage("term", pty.Term)
	state.SetStorage("width", pty.WidthChars)
	state.SetStorage("height", pty.HeightRows)

	req.Reply(true, nil)
}

func defaultHandleWindowChange(state *Session, req ssh.Request) {
	var wc windowChange

	if err := ssh.Unmarshal(req.Payload, &wc); err != nil {
		log.Printf("failed to parse pty request: %v", err)

		if req.WantReply {
			req.Reply(false, nil)
		}
		return
	}

	state.SetStorage("width", wc.WidthChars)
	state.SetStorage("height", wc.HeightRows)

	req.Reply(false, nil)
}

func defaultHandleExec(state *Session, req ssh.Request) {
	var payload execRequest

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

	state.Close(0)
}

func defaultHandleShell(state *Session, req ssh.Request) {
	req.Reply(true, nil)

	go func() {
		defer state.Close(0)

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

package sshsrv

import (
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

type SessionState struct {
	Channel ssh.Channel
	Storage map[string]any

	Mu     sync.RWMutex
	closed bool
}

func (s *SessionState) Close() {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true

	err := s.Channel.Close()
	if err != nil {
		log.Printf("Error closing session: %v", err)
	}
}

func (s *SessionState) Closed() bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	return s.closed
}

type RequestHandler func(state *SessionState, req ssh.Request)

type RequestHandlers map[string]RequestHandler

var handlers RequestHandlers

func setRequestHandlers(reqHandlers RequestHandlers) {
	handlers = reqHandlers

	if reqHandlers["pty-req"] == nil {
		reqHandlers["pty-req"] = defaultHandlePty
	}

	if reqHandlers["window-change"] == nil {
		reqHandlers["window-change"] = defaultHandleWindowChange
	}

	if reqHandlers["exec"] == nil {
		reqHandlers["exec"] = defaultHandleExec
	}

	if reqHandlers["shell"] == nil {
		reqHandlers["shell"] = defaultHandleShell
	}
}

func handleSession(ch ssh.Channel, reqs <-chan *ssh.Request) {
	state := SessionState{
		Channel: ch,
		Storage: make(map[string]any),
	}

	for req := range reqs {
		reqType := req.Type
		if handlerFunction, ok := handlers[reqType]; ok {
			handlerFunction(&state, *req)
		} else {
			log.Printf("Request type %v not implemented!", reqType)
			req.Reply(false, nil)
		}

		if state.Closed() {
			return
		}
	}
}

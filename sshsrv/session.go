package sshsrv

import (
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	Channel ssh.Channel
	Storage map[string]any

	Mu     sync.RWMutex
	closed bool
}

func (s *Session) Close() {
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

func (s *Session) Closed() bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	return s.closed
}

type RequestHandlers map[string]func(state *Session, req ssh.Request)

var handlers RequestHandlers

func setRequestHandlers(reqHandlers RequestHandlers) {
	handlers = reqHandlers
}

func handleSession(ch ssh.Channel, reqs <-chan *ssh.Request) {
	state := Session{
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

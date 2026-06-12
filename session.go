package karottessh

import (
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	Channel ssh.Channel
	Conn    *ssh.Conn

	storage map[string]any

	mu     sync.RWMutex
	closed bool
}

func (s *Session) SetStorage(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = value
}

func (s *Session) GetStorage(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.storage[key]
	return value, ok
}

func (s *Session) DeleteStorage(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.storage, key)
}

func (s *Session) HasStorage(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.storage[key]
	return ok
}

func (s *Session) Close(status uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true

	_, err := s.Channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{Status: status}))
	if err != nil {
		log.Printf("Error sending exit status: %v", err)
	}

	closeErr := s.Channel.Close()
	if closeErr != nil {
		log.Printf("Error closing session: %v", closeErr)
	}
}

func (s *Session) Closed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.closed
}

type RequestHandlers map[string]func(state *Session, req ssh.Request)

var handlers RequestHandlers

func setRequestHandlers(reqHandlers RequestHandlers) {
	handlers = reqHandlers
}

func handleSession(ch ssh.Channel, reqs <-chan *ssh.Request, conn *ssh.Conn) {
	state := Session{
		Channel: ch,
		storage: make(map[string]any),
		Conn:    conn,
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

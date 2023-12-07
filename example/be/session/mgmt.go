package session

import (
	"sync"
)

type ClientRequest struct {
	Id      string
	Message []byte
}

type Management struct {
	mu sync.RWMutex

	ReadCh chan ClientRequest

	Sessions map[string]Session
}

func (s *Management) Add(id string, session Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Sessions[id] = session
}

func (s *Management) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Sessions, id)
}

func (s *Management) SendAll(data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, session := range s.Sessions {
		session.Send(data)
	}
}

func (s *Management) Send(id string, values [][]byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.Sessions[id]
	for _, value := range values {
		session.Send(value)
	}
}

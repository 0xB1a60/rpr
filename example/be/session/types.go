package session

type Session struct {
	WriteCh chan []byte
}

func (s *Session) Send(data []byte) {
	s.WriteCh <- data
}

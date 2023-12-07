package transport

import (
	"context"
	"github.com/0xB1a60/rpr/example/basic/session"
	"github.com/gorilla/websocket"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (s *Server) serveWs(procCtx context.Context, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Logger.Error("upgrade ws err",
			zap.Error(err))
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.Logger.Error("close ws err",
				zap.Error(err))
		}
	}()

	id, err := gonanoid.New()
	if err != nil {
		s.Logger.Error("ws Id generate err",
			zap.Error(err))
		return
	}

	writeCh := make(chan []byte, 10_000)

	defer s.SessionMGMT.Remove(id)
	s.SessionMGMT.Add(id, session.Session{
		WriteCh: writeCh,
	})

	s.Logger.Debug("Client connect", zap.String("Id", id))

	ctx := r.Context()

	conn.SetReadLimit(maxMessageSize)
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		s.Logger.Error("set read deadline err",
			zap.Error(err))
		return
	}
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			s.Logger.Error("set pong read deadline err",
				zap.Error(err))
		}
		return nil
	})

	// read
	hasError := make(chan error, 1)
	go func() {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				hasError <- err
				return
			}

			if messageType == websocket.TextMessage {
				s.SessionMGMT.ReadCh <- session.ClientRequest{
					Id:      id,
					Message: message,
				}
			}
		}
	}()

	// write
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case data := <-writeCh:
				if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					s.Logger.Error("set write deadline ticker err",
						zap.Error(err))
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					s.Logger.Error("write err",
						zap.Error(err))
					hasError <- err
				}
			case <-ticker.C:
				if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					s.Logger.Error("set write deadline ticker err",
						zap.Error(err))
					return
				}
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					s.Logger.Error("ping write err",
						zap.Error(err))
					hasError <- err
				}
			}
		}
	}()

	select {
	case err := <-hasError:
		if websocket.IsCloseError(err, websocket.CloseNoStatusReceived) || websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			s.Logger.Debug("Client disconnects", zap.String("Id", id))
			break
		}
		s.Logger.Debug("Client disconnect err", zap.String("Id", id), zap.Error(err))
		break
	case <-ctx.Done():
		s.Logger.Debug("Client disconnect client", zap.String("Id", id))
		break
	case <-procCtx.Done():
		s.Logger.Debug("Client disconnect process", zap.String("Id", id))
		break
	}
}

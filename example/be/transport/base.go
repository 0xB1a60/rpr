package transport

import (
	"context"
	"errors"
	"fmt"
	"github.com/0xB1a60/rpr/example/basic/session"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Database interface {
	AddKV(ctx context.Context, id string, value string) error
	EditKV(ctx context.Context, id string, newValue string) (*int64, error)
	RemoveKV(ctx context.Context, id string) error
	RemoveKVAccess(ctx context.Context, id string) error
}

type Server struct {
	Logger *zap.Logger

	DB Database

	SessionMGMT *session.Management
}

func (s *Server) Start(procCtx context.Context, port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.serveWs(procCtx, w, r)
	})
	mux.HandleFunc("/add", s.processAdd)
	mux.HandleFunc("/edit", s.processEdit)
	mux.HandleFunc("/remove", s.processRemove)
	mux.HandleFunc("/remove-access", s.processRemoveAccess)

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Logger.Error("shutting server down",
				zap.Error(err))
		}
	}()

	select {
	case <-procCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			s.Logger.Error("shutdown server err",
				zap.Error(err))
		}
	}
}

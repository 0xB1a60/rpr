package rpr

import (
	"context"
	"github.com/0xB1a60/rpr/example/basic/db"
	"github.com/0xB1a60/rpr/example/basic/session"
	"go.uber.org/zap"
)

type Service struct {
	Logger      *zap.Logger
	SessionMGMT *session.Management
	DB          *db.Database
}

func (s *Service) Start(procCtx context.Context) {
	go func() {
		for {
			select {
			case <-procCtx.Done():
				return
			case read := <-s.SessionMGMT.ReadCh:
				s.onWSMessage(read)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-procCtx.Done():
				return
			case change := <-s.DB.ChangeCh:
				s.onDBChange(change)
			}
		}
	}()
}

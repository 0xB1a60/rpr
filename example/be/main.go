package main

import (
	"context"
	"github.com/0xB1a60/rpr/example/basic/db"
	"github.com/0xB1a60/rpr/example/basic/rpr"
	"github.com/0xB1a60/rpr/example/basic/session"
	"github.com/0xB1a60/rpr/example/basic/transport"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("error creating logger", err)
	}
	defer logger.Sync()

	if err := os.MkdirAll("./data", os.ModePerm); err != nil {
		log.Fatal("error creating data directory", err)
	}

	dbInstance, cleanUpFunc, err := db.Open("./data/db.sqlite")
	if err != nil {
		logger.Fatal("error creating db", zap.Error(err))
	}
	defer cleanUpFunc()

	if err := dbInstance.ApplySqlScripts(); err != nil {
		logger.Fatal("err while applying scripts",
			zap.Error(err))
	}

	if err := dbInstance.Cleanup(); err != nil {
		logger.Fatal("err while cleaning database",
			zap.Error(err))
	}

	procCtx, cleanupProc := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cleanupProc()

	sessionMGMT := &session.Management{
		Sessions: map[string]session.Session{},
		ReadCh:   make(chan session.ClientRequest, 10_000),
	}

	rprSvc := &rpr.Service{
		Logger:      logger,
		DB:          dbInstance,
		SessionMGMT: sessionMGMT,
	}
	go rprSvc.Start(procCtx)

	ts := &transport.Server{
		Logger:      logger,
		DB:          dbInstance,
		SessionMGMT: sessionMGMT,
	}
	go ts.Start(procCtx, 9999)
	logger.Info("Server started @ 0.0.0.0:9999")

	select {
	case <-procCtx.Done():
	}
}

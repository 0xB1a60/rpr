package main_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/0xB1a60/rpr/example/basic/db"
	"github.com/0xB1a60/rpr/example/basic/rpr"
	"github.com/0xB1a60/rpr/example/basic/session"
	"github.com/0xB1a60/rpr/example/basic/transport"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"net"
	"net/http"
	"testing"
	"time"
)

type TestSuite struct {
	HttpUrl string
	WsUrl   string

	Cleanup func()
}

// each test is a separate database file with a separate http(ws) server
func setupTestSuite(t *testing.T) TestSuite {
	dbInstance, cleanUpFunc, err := db.Open(":memory:")
	if err != nil {
		require.NoError(t, err)
	}

	require.NoError(t, dbInstance.ApplySqlScripts())

	sessionMGMT := &session.Management{
		Sessions: map[string]session.Session{},
		ReadCh:   make(chan session.ClientRequest, 1_000),
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	logger := zaptest.NewLogger(t)
	rprSvc := &rpr.Service{
		Logger:      logger,
		DB:          dbInstance,
		SessionMGMT: sessionMGMT,
	}
	go rprSvc.Start(ctx)

	port, err := findAvailablePort()
	if err != nil {
		require.NoError(t, err)
	}

	ts := &transport.Server{
		Logger:      logger,
		DB:          dbInstance,
		SessionMGMT: sessionMGMT,
	}
	go ts.Start(ctx, *port)

	// wait a bit until ws starts
	time.Sleep(1 * time.Second)

	return TestSuite{
		WsUrl:   fmt.Sprintf("ws://0.0.0.0:%d/ws", *port),
		HttpUrl: fmt.Sprintf("http://0.0.0.0:%d", *port),
		Cleanup: func() {
			assert.NoError(t, cleanUpFunc())
			cancelFunc()
			// wait a bit until everything finishes
			time.Sleep(3 * time.Second)
		},
	}
}

func findAvailablePort() (*int, error) {
	for i := 8000; i < 50_000; i++ {
		conn, err := net.Listen("tcp", fmt.Sprintf(":%d", i))
		if err == nil {
			conn.Close()
			i := i
			return &i, nil
		}
	}
	return nil, errors.New("no free port available")
}

func readMessages(t *testing.T, c *websocket.Conn) func() []string {
	messages := make([]string, 0)

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if errors.Is(err, websocket.ErrCloseSent) {
					return
				}
				assert.NoError(t, err)
				return
			}
			messages = append(messages, string(message))
		}
	}()

	return func() []string {
		return messages
	}
}

func cleanupWS(t *testing.T, c *websocket.Conn) {
	assert.NoError(t, c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")))
	time.Sleep(1 * time.Second)

	assert.NoError(t, c.Close())
}

func addItem(t *testing.T, httpUrl string, value string) string {
	req := transport.AddReq{
		Value: value,
	}
	b, err := jsoniter.Marshal(req)
	require.NoError(t, err)

	post, err := http.Post(fmt.Sprintf("%s/add", httpUrl), "application/json", bytes.NewBuffer(b))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, post.StatusCode)

	var res transport.AddRes
	require.NoError(t, jsoniter.NewDecoder(post.Body).Decode(&res))
	require.NoError(t, post.Body.Close())

	return res.Id
}

func editItem(t *testing.T, httpUrl string, id string, newValue string) {
	req := transport.EditReq{
		Id:    id,
		Value: newValue,
	}
	b, err := jsoniter.Marshal(req)
	require.NoError(t, err)

	post, err := http.Post(fmt.Sprintf("%s/edit", httpUrl), "application/json", bytes.NewBuffer(b))
	require.NoError(t, err)
	require.NoError(t, post.Body.Close())
	require.Equal(t, http.StatusNoContent, post.StatusCode)
}

func removeItem(t *testing.T, httpUrl string, id string) {
	req := transport.RemoveReq{
		Id: id,
	}
	b, err := jsoniter.Marshal(req)
	require.NoError(t, err)

	post, err := http.Post(fmt.Sprintf("%s/remove", httpUrl), "application/json", bytes.NewBuffer(b))
	require.NoError(t, err)
	require.NoError(t, post.Body.Close())
	require.Equal(t, http.StatusNoContent, post.StatusCode)
}

func removeItemAccess(t *testing.T, httpUrl string, id string) {
	req := transport.RemoveAccessReq{
		Id: id,
	}
	b, err := jsoniter.Marshal(req)
	require.NoError(t, err)

	post, err := http.Post(fmt.Sprintf("%s/remove", httpUrl), "application/json", bytes.NewBuffer(b))
	require.NoError(t, err)
	require.NoError(t, post.Body.Close())
	require.Equal(t, http.StatusNoContent, post.StatusCode)
}

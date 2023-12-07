package main_test

import (
	"github.com/0xB1a60/rpr/example/basic/rpr"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestChange_Add(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	addedId := addItem(t, suite.HttpUrl, t.Name())
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 1)

	var res rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &res))

	assert.Equal(t, rpr.ChangeType, res.Type)
	assert.Equal(t, "kv", res.CollectionName)
	assert.True(t, res.UpdatedAt > currentTime)
	assert.Equal(t, rpr.CreateChangeType, res.ChangeType)
	assert.Equal(t, addedId, res.Id)
	assert.Nil(t, res.Before)
	assert.NotNil(t, res.After)
	assert.Equal(t, t.Name(), res.After.Other["value"])
}

func TestChange_Add_Edit(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	addedId := addItem(t, suite.HttpUrl, t.Name())
	time.Sleep(1 * time.Second)

	editItem(t, suite.HttpUrl, addedId, t.Name()+t.Name())
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 2)

	var fRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &fRes))

	assert.Equal(t, rpr.ChangeType, fRes.Type)
	assert.Equal(t, "kv", fRes.CollectionName)
	assert.True(t, fRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.CreateChangeType, fRes.ChangeType)
	assert.Equal(t, addedId, fRes.Id)
	assert.Nil(t, fRes.Before)
	assert.NotNil(t, fRes.After)
	assert.Equal(t, t.Name(), fRes.After.Other["value"])

	var sRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[1]), &sRes))

	assert.Equal(t, rpr.ChangeType, sRes.Type)
	assert.Equal(t, "kv", sRes.CollectionName)
	assert.True(t, sRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.UpdateChangeType, sRes.ChangeType)
	assert.Equal(t, addedId, sRes.Id)
	assert.NotNil(t, sRes.Before)
	assert.Equal(t, t.Name(), sRes.Before.Other["value"])
	assert.NotNil(t, sRes.After)
	assert.Equal(t, t.Name()+t.Name(), sRes.After.Other["value"])

	assert.True(t, sRes.UpdatedAt > fRes.UpdatedAt)
}

func TestChange_Add_Remove(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	addedId := addItem(t, suite.HttpUrl, t.Name())
	time.Sleep(1 * time.Second)

	removeItem(t, suite.HttpUrl, addedId)
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 2)

	var fRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &fRes))

	assert.Equal(t, rpr.ChangeType, fRes.Type)
	assert.Equal(t, "kv", fRes.CollectionName)
	assert.True(t, fRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.CreateChangeType, fRes.ChangeType)
	assert.Equal(t, addedId, fRes.Id)
	assert.Nil(t, fRes.Before)
	assert.NotNil(t, fRes.After)
	assert.Equal(t, t.Name(), fRes.After.Other["value"])

	var sRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[1]), &sRes))

	assert.Equal(t, rpr.ChangeType, sRes.Type)
	assert.Equal(t, "kv", sRes.CollectionName)
	assert.True(t, sRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.RemoveChangeType, sRes.ChangeType)
	assert.Equal(t, addedId, sRes.Id)
	assert.NotNil(t, sRes.Before)
	assert.Equal(t, t.Name(), sRes.Before.Other["value"])
	assert.Nil(t, sRes.After)

	assert.True(t, sRes.UpdatedAt > fRes.UpdatedAt)
}

func TestChange_Add_RemoveAccess(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	addedId := addItem(t, suite.HttpUrl, t.Name())
	time.Sleep(1 * time.Second)

	removeItemAccess(t, suite.HttpUrl, addedId)
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 2)

	var fRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &fRes))

	assert.Equal(t, rpr.ChangeType, fRes.Type)
	assert.Equal(t, "kv", fRes.CollectionName)
	assert.True(t, fRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.CreateChangeType, fRes.ChangeType)
	assert.Equal(t, addedId, fRes.Id)
	assert.Nil(t, fRes.Before)
	assert.NotNil(t, fRes.After)
	assert.Equal(t, t.Name(), fRes.After.Other["value"])

	var sRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[1]), &sRes))

	assert.Equal(t, rpr.ChangeType, sRes.Type)
	assert.Equal(t, "kv", sRes.CollectionName)
	assert.True(t, sRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.RemoveChangeType, sRes.ChangeType)
	assert.Equal(t, addedId, sRes.Id)
	assert.NotNil(t, sRes.Before)
	assert.Equal(t, t.Name(), sRes.Before.Other["value"])
	assert.Nil(t, sRes.After)

	assert.True(t, sRes.UpdatedAt > fRes.UpdatedAt)
}

func TestChange_Add_Edit_Remove(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	addedId := addItem(t, suite.HttpUrl, t.Name())
	time.Sleep(1 * time.Second)

	editItem(t, suite.HttpUrl, addedId, t.Name()+t.Name())
	time.Sleep(1 * time.Second)

	removeItem(t, suite.HttpUrl, addedId)
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 3)

	var fRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &fRes))

	assert.Equal(t, rpr.ChangeType, fRes.Type)
	assert.Equal(t, "kv", fRes.CollectionName)
	assert.True(t, fRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.CreateChangeType, fRes.ChangeType)
	assert.Equal(t, addedId, fRes.Id)
	assert.Nil(t, fRes.Before)
	assert.NotNil(t, fRes.After)
	assert.Equal(t, t.Name(), fRes.After.Other["value"])

	var sRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[1]), &sRes))

	assert.Equal(t, rpr.ChangeType, sRes.Type)
	assert.Equal(t, "kv", sRes.CollectionName)
	assert.True(t, sRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.UpdateChangeType, sRes.ChangeType)
	assert.Equal(t, addedId, sRes.Id)
	assert.NotNil(t, sRes.Before)
	assert.Equal(t, t.Name(), sRes.Before.Other["value"])
	assert.NotNil(t, sRes.After)
	assert.Equal(t, t.Name()+t.Name(), sRes.After.Other["value"])

	assert.True(t, sRes.UpdatedAt > fRes.UpdatedAt)

	var tRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[2]), &tRes))

	assert.Equal(t, rpr.ChangeType, tRes.Type)
	assert.Equal(t, "kv", tRes.CollectionName)
	assert.True(t, tRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.RemoveChangeType, tRes.ChangeType)
	assert.Equal(t, addedId, tRes.Id)
	assert.NotNil(t, tRes.Before)
	assert.Equal(t, t.Name()+t.Name(), tRes.Before.Other["value"])
	assert.Nil(t, tRes.After)

	assert.True(t, tRes.UpdatedAt > sRes.UpdatedAt)
}

func TestChange_Add_Edit_RemoveAccess(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	addedId := addItem(t, suite.HttpUrl, t.Name())
	time.Sleep(1 * time.Second)

	editItem(t, suite.HttpUrl, addedId, t.Name()+t.Name())
	time.Sleep(1 * time.Second)

	removeItemAccess(t, suite.HttpUrl, addedId)
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 3)

	var fRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &fRes))

	assert.Equal(t, rpr.ChangeType, fRes.Type)
	assert.Equal(t, "kv", fRes.CollectionName)
	assert.True(t, fRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.CreateChangeType, fRes.ChangeType)
	assert.Equal(t, addedId, fRes.Id)
	assert.Nil(t, fRes.Before)
	assert.NotNil(t, fRes.After)
	assert.Equal(t, t.Name(), fRes.After.Other["value"])

	var sRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[1]), &sRes))

	assert.Equal(t, rpr.ChangeType, sRes.Type)
	assert.Equal(t, "kv", sRes.CollectionName)
	assert.True(t, sRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.UpdateChangeType, sRes.ChangeType)
	assert.Equal(t, addedId, sRes.Id)
	assert.NotNil(t, sRes.Before)
	assert.Equal(t, t.Name(), sRes.Before.Other["value"])
	assert.NotNil(t, sRes.After)
	assert.Equal(t, t.Name()+t.Name(), sRes.After.Other["value"])

	assert.True(t, sRes.UpdatedAt > fRes.UpdatedAt)

	var tRes rpr.ChangeResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[2]), &tRes))

	assert.Equal(t, rpr.ChangeType, tRes.Type)
	assert.Equal(t, "kv", tRes.CollectionName)
	assert.True(t, tRes.UpdatedAt > currentTime)
	assert.Equal(t, rpr.RemoveChangeType, tRes.ChangeType)
	assert.Equal(t, addedId, tRes.Id)
	assert.NotNil(t, tRes.Before)
	assert.Equal(t, t.Name()+t.Name(), tRes.Before.Other["value"])
	assert.Nil(t, tRes.After)

	assert.True(t, tRes.UpdatedAt > sRes.UpdatedAt)
}

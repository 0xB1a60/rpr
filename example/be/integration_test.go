package main_test

import (
	"github.com/0xB1a60/rpr/example/basic/rpr"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"strconv"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)
}

func TestSendBadRequests(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	cases := map[string]string{
		"non-json": "random message",
		"json":     `{"test": 123}`,
	}

	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(value)))
			time.Sleep(1 * time.Second)
		})
	}

	assert.Empty(t, messagesFunc())
}

func TestSyncRequest_Empty(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	b, err := jsoniter.Marshal(rpr.Request{
		Type: rpr.RequestType,
	})
	require.NoError(t, err)

	require.NoError(t, c.WriteMessage(websocket.TextMessage, b))
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 1)

	var res rpr.FullSyncResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &res))

	assert.Equal(t, rpr.FullSyncType, res.Type)
	assert.Equal(t, "kv", res.CollectionName)
	assert.True(t, res.Version > currentTime)
	assert.Nil(t, res.RemovedIds)
	assert.Nil(t, res.Values)
}

func TestSyncRequest_NonExistingCollection(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	b, err := jsoniter.Marshal(rpr.Request{
		Type: rpr.RequestType,
		Versions: map[string]int64{
			"non_kv": 1,
		},
	})
	require.NoError(t, err)

	require.NoError(t, c.WriteMessage(websocket.TextMessage, b))
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 1)

	var res rpr.FullSyncResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &res))

	assert.Equal(t, rpr.FullSyncType, res.Type)
	assert.Equal(t, "kv", res.CollectionName)
	assert.True(t, res.Version > currentTime)
	assert.Nil(t, res.RemovedIds)
	assert.Nil(t, res.Values)
}

func TestSyncRequest_DeletedCollection(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	b, err := jsoniter.Marshal(rpr.Request{
		Type: rpr.RequestType,
		Versions: map[string]int64{
			"deleted_test_collection": currentTime,
		},
	})
	require.NoError(t, err)

	require.NoError(t, c.WriteMessage(websocket.TextMessage, b))
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 2)

	var firstRes rpr.FullSyncResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &firstRes))

	assert.Equal(t, rpr.FullSyncType, firstRes.Type)
	assert.Equal(t, "kv", firstRes.CollectionName)
	assert.True(t, firstRes.Version > currentTime)
	assert.Nil(t, firstRes.RemovedIds)
	assert.Nil(t, firstRes.Values)

	var secondRes rpr.DeleteResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[1]), &secondRes))

	assert.Equal(t, rpr.RemoveCollectionType, secondRes.Type)
	assert.Equal(t, "deleted_test_collection", secondRes.CollectionName)
}

func TestSyncRequest_NoVersions(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	const addCount = 50

	addedIds := make(map[string]struct{}, addCount)

	for i := 0; i < addCount; i++ {
		id := addItem(t, suite.HttpUrl, strconv.Itoa(i))

		if _, ok := addedIds[id]; ok {
			assert.Fail(t, "Id already exist")
		}
		addedIds[id] = struct{}{}
	}

	time.Sleep(5 * time.Second)

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	b, err := jsoniter.Marshal(rpr.Request{
		Type: rpr.RequestType,
	})
	require.NoError(t, err)

	require.NoError(t, c.WriteMessage(websocket.TextMessage, b))
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 1)

	var fsRes rpr.FullSyncResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &fsRes))

	assert.Equal(t, rpr.FullSyncType, fsRes.Type)
	assert.Equal(t, "kv", fsRes.CollectionName)
	assert.True(t, fsRes.Version > currentTime)
	assert.Nil(t, fsRes.RemovedIds)
	assert.NotNil(t, fsRes.Values)
	assert.Len(t, fsRes.Values, addCount)

	for _, item := range fsRes.Values {
		_, ok := addedIds[item.Id]
		assert.True(t, ok)
	}
}

func TestSyncRequest_Versions(t *testing.T) {
	var currentTime int64

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	const addCount = 50

	addedIds := make(map[string]struct{}, addCount)

	for i := 0; i < addCount; i++ {
		id := addItem(t, suite.HttpUrl, strconv.Itoa(i))

		if i == 24 {
			currentTime = time.Now().UnixMilli()
			maps.Clear(addedIds)
			time.Sleep(5 * time.Second)
		}

		if _, ok := addedIds[id]; ok {
			assert.Fail(t, "Id already exist")
		}
		addedIds[id] = struct{}{}
	}

	time.Sleep(5 * time.Second)

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	b, err := jsoniter.Marshal(rpr.Request{
		Type: rpr.RequestType,
		Versions: map[string]int64{
			"kv": currentTime,
		},
	})
	require.NoError(t, err)

	require.NoError(t, c.WriteMessage(websocket.TextMessage, b))
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 1)

	var fsRes rpr.FullSyncResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &fsRes))

	assert.Equal(t, rpr.FullSyncType, fsRes.Type)
	assert.Equal(t, "kv", fsRes.CollectionName)
	assert.True(t, fsRes.Version > currentTime)
	assert.Nil(t, fsRes.RemovedIds)
	assert.NotNil(t, fsRes.Values)
	assert.Len(t, fsRes.Values, 25)

	for _, item := range fsRes.Values {
		_, ok := addedIds[item.Id]
		assert.True(t, ok)
	}
}

func TestSyncRequest_RemovedIds(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	const addCount = 50

	addedIds := make(map[string]struct{}, addCount)

	for i := 0; i < addCount; i++ {
		id := addItem(t, suite.HttpUrl, strconv.Itoa(i))

		if _, ok := addedIds[id]; ok {
			assert.Fail(t, "Id already exist")
		}
		addedIds[id] = struct{}{}
	}

	const removeCount = 5
	removedIds := make(map[string]struct{}, removeCount)
	var i int
	for id := range addedIds {
		if i == 5 {
			break
		}
		i++

		removedIds[id] = struct{}{}
		delete(addedIds, id)

		removeItem(t, suite.HttpUrl, id)
	}

	time.Sleep(5 * time.Second)

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	b, err := jsoniter.Marshal(rpr.Request{
		Type: rpr.RequestType,
		Versions: map[string]int64{
			"kv": currentTime,
		},
	})
	require.NoError(t, err)

	require.NoError(t, c.WriteMessage(websocket.TextMessage, b))
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.Len(t, messages, 1)

	var fsRes rpr.FullSyncResponse
	require.NoError(t, jsoniter.Unmarshal([]byte(messages[0]), &fsRes))

	assert.Equal(t, rpr.FullSyncType, fsRes.Type)
	assert.Equal(t, "kv", fsRes.CollectionName)
	assert.True(t, fsRes.Version > currentTime)
	assert.NotNil(t, fsRes.RemovedIds)
	assert.Len(t, fsRes.RemovedIds, removeCount)
	assert.NotNil(t, fsRes.Values)
	assert.Len(t, fsRes.Values, addCount-removeCount)

	for _, item := range fsRes.Values {
		_, ok := addedIds[item.Id]
		assert.True(t, ok)

		_, ok = removedIds[item.Id]
		assert.False(t, ok)
	}

	for removedId := range fsRes.RemovedIds {
		_, ok := removedIds[removedId]
		assert.True(t, ok)
	}
}

func TestSyncRequest_AndPartials(t *testing.T) {
	currentTime := time.Now().UnixMilli()

	suite := setupTestSuite(t)
	defer suite.Cleanup()

	const addCount = 2500

	addedIds := make(map[string]struct{}, addCount)

	for i := 0; i < addCount; i++ {
		id := addItem(t, suite.HttpUrl, strconv.Itoa(i))

		if _, ok := addedIds[id]; ok {
			assert.Fail(t, "Id already exist")
		}
		addedIds[id] = struct{}{}
	}

	c, _, err := websocket.DefaultDialer.Dial(suite.WsUrl, nil)
	require.NoError(t, err)
	defer cleanupWS(t, c)

	messagesFunc := readMessages(t, c)

	b, err := jsoniter.Marshal(rpr.Request{
		Type: rpr.RequestType,
	})
	require.NoError(t, err)

	require.NoError(t, c.WriteMessage(websocket.TextMessage, b))
	time.Sleep(1 * time.Second)

	messages := messagesFunc()
	require.True(t, len(messages) >= 3)

	type response struct {
		Type           string `json:"type"`
		CollectionName string `json:"collection_name"`
	}

	fullResponses := make([]string, 0, 1)
	partialResponses := make([]string, 0, 2)
	for _, message := range messages {
		var res response
		require.NoError(t, jsoniter.Unmarshal([]byte(message), &res))

		assert.Equal(t, "kv", res.CollectionName)

		if res.Type == rpr.FullSyncType {
			fullResponses = append(fullResponses, message)
			continue
		}

		if res.Type == rpr.PartialSyncType {
			partialResponses = append(partialResponses, message)
			continue
		}
		assert.Equal(t, rpr.ChangeType, res.Type)
	}

	require.Len(t, fullResponses, 1)
	require.Len(t, partialResponses, 2)

	for _, fullResponse := range fullResponses {
		var fsRes rpr.FullSyncResponse
		require.NoError(t, jsoniter.Unmarshal([]byte(fullResponse), &fsRes))

		assert.Equal(t, rpr.FullSyncType, fsRes.Type)
		assert.Equal(t, "kv", fsRes.CollectionName)
		assert.True(t, fsRes.Version > currentTime)
		assert.Nil(t, fsRes.RemovedIds)
		assert.NotNil(t, fsRes.Values)
		assert.Len(t, fsRes.Values, 500)

		for _, item := range fsRes.Values {
			_, ok := addedIds[item.Id]
			assert.True(t, ok)
		}
	}

	for _, partialResponse := range partialResponses {
		var psRes rpr.PartialSyncResponse
		require.NoError(t, jsoniter.Unmarshal([]byte(partialResponse), &psRes))

		assert.Equal(t, rpr.PartialSyncType, psRes.Type)
		assert.Equal(t, "kv", psRes.CollectionName)
		assert.NotNil(t, psRes.Values)
		assert.Len(t, psRes.Values, 1_000)

		for _, item := range psRes.Values {
			_, ok := addedIds[item.Id]
			assert.True(t, ok)
		}
	}
}

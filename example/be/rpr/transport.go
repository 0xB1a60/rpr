package rpr

import (
	"context"
	"github.com/0xB1a60/rpr/example/basic/db"
	"github.com/0xB1a60/rpr/example/basic/session"
	jsoniter "github.com/json-iterator/go"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

const (
	batchSize = 1_000
)

var currentCollections = map[string]struct{}{
	"kv": {},
}

var deletedCollections = map[string]struct{}{
	"deleted_test_collection": {},
}

func (s *Service) onWSMessage(request session.ClientRequest) {
	var req Request
	if err := jsoniter.Unmarshal(request.Message, &req); err != nil {
		s.Logger.Error("client message unmarshal err",
			zap.String("input", string(request.Message)),
			zap.Error(err))
		return
	}

	if req.Type != RequestType {
		s.Logger.Error("request type unsupported",
			zap.String("type", req.Type),
			zap.String("id", request.Id))
		return
	}

	requested := make(map[string]*int64, len(currentCollections))
	for name := range currentCollections {
		requested[name] = nil
	}

	deleted := make([]string, 0, len(req.Versions))
	for name, version := range req.Versions {
		if _, ok := deletedCollections[name]; ok {
			deleted = append(deleted, name)
			continue
		}

		version := version
		requested[name] = &version
	}

	mu := sync.Mutex{}
	responses := make([]any, 0, len(requested)+len(deleted))

	current := time.Now().UnixMilli()

	g, gCtx := errgroup.WithContext(context.Background())
	for collectionName, version := range requested {
		if _, ok := deletedCollections[collectionName]; ok {
			continue
		}

		if _, ok := currentCollections[collectionName]; !ok {
			s.Logger.Info("collection is not supported", zap.String("name", collectionName))
			continue
		}

		version := version
		collectionName := collectionName

		g.Go(func() error {
			readRes, err := s.DB.Read(gCtx, collectionName, version)
			if err != nil {
				return err
			}

			batches := chunkSlice(readRes.Value, batchSize)
			if len(batches) == 0 {
				mu.Lock()
				responses = append(responses, FullSyncResponse{
					Type:           FullSyncType,
					CollectionName: collectionName,
					RemovedIds:     readRes.RemovedIds,
					Version:        current,
				})
				mu.Unlock()
				return nil
			}

			for i, batch := range batches {
				values := lo.Map[db.ReadVal, Item](batch, func(item db.ReadVal, index int) Item {
					return Item{
						Id:        item.Id,
						CreatedAt: item.CreatedAt.UnixMilli(),
						UpdatedAt: item.UpdatedAt.UnixMilli(),
						Other: map[string]any{
							"value": item.Value,
						},
					}
				})

				if isFull := i == len(batches)-1; isFull {
					mu.Lock()
					responses = append(responses, FullSyncResponse{
						Type:           FullSyncType,
						CollectionName: collectionName,
						Values:         values,
						RemovedIds:     readRes.RemovedIds,
						Version:        current,
					})
					mu.Unlock()
					continue
				}
				mu.Lock()
				responses = append(responses, PartialSyncResponse{
					Type:           PartialSyncType,
					CollectionName: collectionName,
					Values:         values,
				})
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		s.Logger.Error("err while reading collection",
			zap.Error(err))
		return
	}

	for _, collection := range deleted {
		responses = append(responses, DeleteResponse{
			Type:           RemoveCollectionType,
			CollectionName: collection,
		})
	}

	res := make([][]byte, 0, len(responses))
	for _, response := range responses {
		bytes, err := jsoniter.Marshal(response)
		if err != nil {
			s.Logger.Error("err while marshaling",
				zap.String("id", request.Id))
			return
		}
		res = append(res, bytes)
	}

	s.SessionMGMT.Send(request.Id, res)
}

func chunkSlice[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T
	for {
		if len(slice) == 0 {
			break
		}

		// necessary check to avoid slicing beyond
		// slice capacity
		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}

	return chunks
}

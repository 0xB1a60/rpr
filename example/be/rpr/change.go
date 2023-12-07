package rpr

import (
	"database/sql"
	"errors"
	"github.com/0xB1a60/rpr/example/basic/db"
	"github.com/0xB1a60/rpr/example/basic/util"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

func (s *Service) onDBChange(change db.Change) {
	defer func() {
		if err := s.DB.RemoveChange(change.Table, change.RowId); err != nil {
			s.Logger.Error("err while removing change record",
				zap.Error(err))
		}
	}()

	var res *ChangeResponse
	switch change.Table {
	case db.KVAccessChangesTable:
		res = s.onKVAccessChange(change)
	case db.KVChangesTable:
		res = s.onKVChange(change)
	default:
		s.Logger.Error("table is not supported", zap.Any("change", change))
		return
	}

	if res == nil {
		return
	}

	bytes, err := jsoniter.Marshal(res)
	if err != nil {
		s.Logger.Error("err while marshaling",
			zap.Any("res", res))
		return
	}

	s.SessionMGMT.SendAll(bytes)
}

type accessEntity struct {
	Access string `json:"access"`
}

func (s *Service) onKVAccessChange(change db.Change) *ChangeResponse {
	changeEntry, err := s.DB.ReadKVAccessChange(change.RowId)
	if err != nil {
		s.Logger.Error("err while reading kv access change",
			zap.Error(err))
		return nil
	}

	if change.Table == db.KVAccessChangesTable && changeEntry.Operation == "INSERT" {
		s.Logger.Debug("ignoring access changes - INSERT as kv_changes event will be used")
		return nil
	}
	// changeEntry.Operation will be always 'UPDATE' as DELETE is not permitted on DB Level

	var after accessEntity
	if err := jsoniter.Unmarshal([]byte(*changeEntry.After), &after); err != nil {
		s.Logger.Error("err while unmarshalling before",
			zap.Error(err))
		return nil
	}

	if after.Access == db.AccessDeleted {
		return nil
	}

	val, err := s.DB.ReadKV(changeEntry.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Logger.Debug("KV not found",
				zap.String("id", changeEntry.Id),
				zap.Error(err))
			return nil
		}
		s.Logger.Error("err while reading KV",
			zap.String("id", changeEntry.Id),
			zap.Error(err))
		return nil
	}

	item := Item{
		Id:        val.Key,
		CreatedAt: val.CreatedAt.UnixMilli(),
		UpdatedAt: val.UpdatedAt.UnixMilli(),
		Other: map[string]any{
			"value": val.Value,
		},
	}

	var beforeItem *Item
	var afterItem *Item
	var changeType string
	if after.Access == db.AccessBlocked {
		changeType = RemoveChangeType
		beforeItem = &item
	} else {
		changeType = CreateChangeType
		afterItem = &item
	}

	return &ChangeResponse{
		Id:             changeEntry.Id,
		Type:           ChangeType,
		CollectionName: db.KVTable,
		ChangeType:     changeType,
		UpdatedAt:      changeEntry.OriginatedAt.UnixMilli(),
		Before:         beforeItem,
		After:          afterItem,
	}
}

func (s *Service) onKVChange(change db.Change) *ChangeResponse {
	changeEntry, err := s.DB.ReadKVChange(change.RowId)
	if err != nil {
		s.Logger.Error("err while reading kv change",
			zap.Error(err))
		return nil
	}

	var changeType string
	switch changeEntry.Operation {
	case "INSERT":
		changeType = CreateChangeType
	case "UPDATE":
		changeType = UpdateChangeType
	case "DELETE":
		changeType = RemoveChangeType
	}

	return &ChangeResponse{
		Id:             changeEntry.Id,
		Type:           ChangeType,
		CollectionName: db.KVTable,
		ChangeType:     changeType,
		UpdatedAt:      changeEntry.OriginatedAt.UnixMilli(),
		Before:         s.convertToMap(changeEntry.Before, changeEntry.Id),
		After:          s.convertToMap(changeEntry.After, changeEntry.Id),
	}
}

func (s *Service) convertToMap(raw *string, id string) *Item {
	if raw == nil {
		return nil
	}

	var val map[string]any
	if err := jsoniter.Unmarshal([]byte(*raw), &val); err != nil {
		s.Logger.Error("error while unmarshalling",
			zap.Error(err))
		return nil
	}

	createdAt, err := util.ConvertFromMapTo[float64](val, CreatedField)
	if err != nil {
		s.Logger.Error("error while converting createdAt",
			zap.Error(err))
		return nil
	}
	delete(val, CreatedField)

	updatedAt, err := util.ConvertFromMapTo[float64](val, UpdatedField)
	if err != nil {
		s.Logger.Error("error while converting updatedAt",
			zap.Error(err))
		return nil
	}
	delete(val, UpdatedField)

	return &Item{
		Id:        id,
		CreatedAt: int64(*createdAt),
		UpdatedAt: int64(*updatedAt),
		Other:     val,
	}
}

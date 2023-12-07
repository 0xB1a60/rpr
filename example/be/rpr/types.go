package rpr

import (
	"encoding/json"
	"fmt"
	"github.com/0xB1a60/rpr/example/basic/util"
)

const (
	RequestType = "sync"

	RemoveCollectionType = "remove_collection"
	FullSyncType         = "full_sync"
	PartialSyncType      = "partial_sync"
	ChangeType           = "change"
)

const (
	CreateChangeType = "create"
	UpdateChangeType = "update"
	RemoveChangeType = "remove"
)

const (
	IdField      = "id"
	CreatedField = "created_at"
	UpdatedField = "updated_at"
)

type Request struct {
	Type     string           `json:"type"`
	Versions map[string]int64 `json:"collection_versions,omitempty"`
}

type DeleteResponse struct {
	Type           string `json:"type"`
	CollectionName string `json:"collection_name"`
}

type Item struct {
	Id string

	CreatedAt int64
	UpdatedAt int64

	Other map[string]any
}

func (i *Item) MarshalJSON() ([]byte, error) {
	res := make(map[string]any, 3)
	res[IdField] = i.Id
	res[CreatedField] = i.CreatedAt
	res[UpdatedField] = i.UpdatedAt

	for name, val := range i.Other {
		if _, ok := res[name]; ok {
			return nil, fmt.Errorf("val: %v already exist", val)
		}

		res[name] = val
	}
	return json.Marshal(res)
}

func (i *Item) UnmarshalJSON(p []byte) error {
	var vals map[string]any
	if err := json.Unmarshal(p, &vals); err != nil {
		return err
	}

	id, err := util.ConvertFromMapTo[string](vals, IdField)
	if err != nil {
		return err
	}
	delete(vals, IdField)

	createdAt, err := util.ConvertFromMapTo[float64](vals, CreatedField)
	if err != nil {
		return err
	}
	delete(vals, CreatedField)

	updatedAt, err := util.ConvertFromMapTo[float64](vals, UpdatedField)
	if err != nil {
		return err
	}
	delete(vals, UpdatedField)

	*i = Item{
		Id:        *id,
		CreatedAt: int64(*createdAt),
		UpdatedAt: int64(*updatedAt),
		Other:     vals,
	}
	return nil
}

type ChangeResponse struct {
	Type           string `json:"type"`
	CollectionName string `json:"collection_name"`

	ChangeType string `json:"change_type"`

	Id string `json:"id"`

	UpdatedAt int64 `json:"updated_at"`

	Before *Item `json:"before,omitempty"`
	After  *Item `json:"after,omitempty"`
}

type PartialSyncResponse struct {
	Type           string `json:"type"`
	CollectionName string `json:"collection_name"`

	Values []Item `json:"values,omitempty"`
}

type FullSyncResponse struct {
	Type           string `json:"type"`
	CollectionName string `json:"collection_name"`

	Values []Item `json:"values,omitempty"`

	RemovedIds map[string]int64 `json:"removed_ids,omitempty"`

	Version int64 `json:"version"`
}

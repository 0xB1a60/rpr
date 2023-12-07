package db

import "time"

const (
	KVTable       = "kv"
	kvAccessTable = KVTable + "_access"

	KVChangesTable       = KVTable + "_changes"
	KVAccessChangesTable = kvAccessTable + "_changes"

	AccessAllowed = "ALLOWED"
	AccessBlocked = "BLOCKED"
	AccessDeleted = "DELETED"
)

type ReadVal struct {
	Id string

	CreatedAt time.Time
	UpdatedAt time.Time

	Value any
}

type ReadRes struct {
	Value      []ReadVal
	RemovedIds map[string]int64
}

func (rr *ReadRes) setRemovedIds(val map[string]int64) {
	rr.RemovedIds = val
}

func (rr *ReadRes) setValue(val []ReadVal) {
	rr.Value = val
}

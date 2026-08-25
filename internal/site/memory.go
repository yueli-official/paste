package site

import (
	"context"
	"sync"
	"time"
)

type MemoryStore struct {
	mu    sync.RWMutex
	value Settings
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{value: Settings{
		Name: DefaultName, Description: DefaultDescription, Revision: 1,
		UpdatedAt: time.Unix(0, 0).UTC(),
	}}
}

func (store *MemoryStore) ReadSettings(ctx context.Context) (Settings, error) {
	if err := ctx.Err(); err != nil {
		return Settings{}, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.value, nil
}

func (store *MemoryStore) UpdateSettings(ctx context.Context, value Settings, expectedRevision int64) (Settings, error) {
	if err := ctx.Err(); err != nil {
		return Settings{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.value.Revision != expectedRevision {
		return Settings{}, ErrConflict
	}
	store.value = value
	return store.value, nil
}

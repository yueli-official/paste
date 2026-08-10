package paste

import (
	"context"
	"sort"
	"sync"
)

type MemoryStore struct {
	mu     sync.RWMutex
	values map[string]Paste
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{values: make(map[string]Paste)}
}

func (store *MemoryStore) Insert(ctx context.Context, value Paste) (Paste, error) {
	if err := ctx.Err(); err != nil {
		return Paste{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.values[value.Code]; exists {
		return Paste{}, ErrCodeCollision
	}
	store.values[value.Code] = clonePaste(value)
	return clonePaste(value), nil
}

func (store *MemoryStore) GetByCode(ctx context.Context, code string) (Paste, error) {
	if err := ctx.Err(); err != nil {
		return Paste{}, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, exists := store.values[code]
	if !exists {
		return Paste{}, ErrNotFound
	}
	return clonePaste(value), nil
}

func (store *MemoryStore) ListByOwner(ctx context.Context, userKey string) ([]Paste, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]Paste, 0)
	for _, value := range store.values {
		if value.OwnerUserKey == userKey {
			result = append(result, clonePaste(value))
		}
	}
	sort.Slice(result, func(left, right int) bool { return result[left].CreatedAt.Before(result[right].CreatedAt) })
	return result, nil
}

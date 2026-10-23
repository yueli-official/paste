package paste

import (
	"context"
	"time"
)

func (store *MemoryStore) ClaimGuest(ctx context.Context, guestID, userKey string, now time.Time) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if previous := store.guestClaims[guestID]; previous != "" && previous != userKey {
		return 0, ErrForbidden
	}
	store.guestClaims[guestID] = userKey
	var count int64
	for code, value := range store.values {
		if value.OwnerGuestID != guestID {
			continue
		}
		value.OwnerGuestID = ""
		value.OwnerUserKey = userKey
		value.Revision++
		value.UpdatedAt = now
		store.values[code] = value
		count++
	}
	return count, nil
}

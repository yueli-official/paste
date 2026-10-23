package paste

import (
	"context"
	"strings"
)

// Owner keeps Identity users and guest sessions in separate namespaces.
type Owner struct {
	UserKey string
	GuestID string
}

func (owner Owner) Valid() bool {
	return (strings.TrimSpace(owner.UserKey) != "") != (strings.TrimSpace(owner.GuestID) != "")
}

func (owner Owner) owns(value Paste) bool {
	return owner.Valid() && ((owner.UserKey != "" && owner.UserKey == value.OwnerUserKey) ||
		(owner.GuestID != "" && owner.GuestID == value.OwnerGuestID))
}

func (service *Service) ClaimGuest(ctx context.Context, guestID, userKey string) (int64, error) {
	if strings.TrimSpace(guestID) == "" || strings.TrimSpace(userKey) == "" {
		return 0, ErrForbidden
	}
	return service.store.ClaimGuest(ctx, guestID, userKey, service.now().UTC())
}

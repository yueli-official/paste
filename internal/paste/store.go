package paste

import (
	"context"
	"time"
)

type Store interface {
	Insert(context.Context, Paste) (Paste, error)
	GetByCode(context.Context, string) (Paste, error)
	ListByOwner(context.Context, string) ([]Paste, error)
	ListForAdministration(context.Context, AdministrationQuery) (AdministrationPage, error)
	Update(context.Context, Paste, int64) (Paste, error)
	ClaimGuest(context.Context, string, string, time.Time) (int64, error)
}

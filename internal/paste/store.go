package paste

import "context"

type Store interface {
	Insert(context.Context, Paste) (Paste, error)
	GetByCode(context.Context, string) (Paste, error)
	ListByOwner(context.Context, string) ([]Paste, error)
	Update(context.Context, Paste, int64) (Paste, error)
}

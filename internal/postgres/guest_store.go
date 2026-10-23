package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/yueli-official/paste/internal/paste"
)

// Both guest insertion and claim use this lock: a still-valid guest JWT must
// never create new unclaimed rows after the account transfer has committed.
func lockGuest(ctx context.Context, tx *sql.Tx, guestID string) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "paste:guest:"+guestID)
	return err
}

func checkGuestCreation(ctx context.Context, tx *sql.Tx, guestID string) error {
	if guestID == "" {
		return nil
	}
	if err := lockGuest(ctx, tx, guestID); err != nil {
		return err
	}
	var claimed bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM paste_guest_claims WHERE guest_id = $1)`, guestID).Scan(&claimed); err != nil {
		return err
	}
	if claimed {
		return paste.ErrForbidden
	}
	return nil
}

func (store *Store) ClaimGuest(ctx context.Context, guestID, userKey string, now time.Time) (int64, error) {
	tx, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := lockGuest(ctx, tx, guestID); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO paste_guest_claims (guest_id, user_key, claimed_at) VALUES ($1, $2, $3) ON CONFLICT (guest_id) DO NOTHING`, guestID, userKey, now); err != nil {
		return 0, err
	}
	var owner string
	if err := tx.QueryRowContext(ctx, `SELECT user_key FROM paste_guest_claims WHERE guest_id = $1`, guestID).Scan(&owner); err != nil {
		return 0, err
	}
	if owner != userKey {
		return 0, paste.ErrForbidden
	}
	result, err := tx.ExecContext(ctx, `UPDATE pastes SET owner_user_key = $1, owner_guest_id = NULL, revision = revision + 1, updated_at = $2 WHERE owner_guest_id = $3`, userKey, now, guestID)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

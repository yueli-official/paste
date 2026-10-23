ALTER TABLE pastes ADD COLUMN IF NOT EXISTS owner_guest_id text;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'pastes_owner_exclusive' AND conrelid = 'pastes'::regclass) THEN
        ALTER TABLE pastes ADD CONSTRAINT pastes_owner_exclusive CHECK (owner_user_key IS NULL OR owner_guest_id IS NULL);
    END IF;
END $$;
CREATE INDEX IF NOT EXISTS pastes_guest_owner_created ON pastes (owner_guest_id, created_at DESC) WHERE owner_guest_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS paste_guest_claims (
    guest_id text PRIMARY KEY,
    user_key text NOT NULL,
    claimed_at timestamptz NOT NULL
);

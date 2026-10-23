DROP TABLE paste_guest_claims;
DROP INDEX pastes_guest_owner_created;
ALTER TABLE pastes DROP CONSTRAINT pastes_owner_exclusive;
ALTER TABLE pastes DROP COLUMN owner_guest_id;

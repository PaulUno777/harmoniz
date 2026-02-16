-- When a track was first added to our DB (Unix timestamp).
-- Used to detect stale library: if most recent added_at is older than e.g. 24h,
-- we delete all tracks for that root and re-sync from disk.
ALTER TABLE tracks ADD COLUMN added_at INTEGER DEFAULT 0;

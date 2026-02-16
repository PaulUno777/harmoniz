-- Consolidated migration: Tracks table with all necessary columns and indexes
-- This migration includes everything needed for Phase 1 (scanning) and future phases

-- Main tracks table
CREATE TABLE tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE NOT NULL,      -- Chemin absolu
    filename TEXT NOT NULL,
    size INTEGER NOT NULL,          -- Pour pré-filtrage doublons
    mod_time INTEGER NOT NULL,      -- Pour cache invalidation

    -- Métadonnées (Raw = Brut, Norm = Pour recherche/matching)
    artist_raw TEXT,
    artist_norm TEXT,               -- lowercase, no accents, no spaces
    album_raw TEXT,
    album_norm TEXT,
    title TEXT,
    year INTEGER,
    track_num INTEGER,
    bitrate INTEGER,

    -- Identification Unique
    hash_partial TEXT,              -- SHA256 (1MB Head + 1MB Tail)
    hash_full TEXT,                 -- SHA256 Complet (Calculé à la demande)
    fingerprint TEXT,               -- AcoustID (Optionnel)

    -- État
    is_deleted BOOLEAN DEFAULT 0,   -- Soft Delete
    deleted_at INTEGER,             -- Unix timestamp (for future use)
    delete_reason TEXT,             -- e.g., "PRUNE", "USER" (for future use)
    status TEXT DEFAULT 'clean'     -- 'clean', 'modified', 'error'
);

-- Indexation pour la performance
CREATE INDEX idx_size ON tracks(size);
CREATE INDEX idx_artist_norm ON tracks(artist_norm);
CREATE INDEX idx_hash_partial ON tracks(hash_partial);
CREATE INDEX idx_tracks_deleted ON tracks(is_deleted) WHERE is_deleted = 1;

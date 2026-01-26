-- Migration: Add type column to tournaments table
-- Created: 2026-01-26
-- Purpose: Differentiate between IN_PERSON tournaments (with rounds) and ONLINE tournaments (round-robin)

ALTER TABLE tournaments ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'IN_PERSON' 
  CHECK (type IN ('IN_PERSON', 'ONLINE'));

-- Add index for filtering by type
CREATE INDEX IF NOT EXISTS idx_tournaments_type ON tournaments(type);

-- Add comment for clarity
COMMENT ON COLUMN tournaments.type IS 'Tournament type: IN_PERSON (uses rounds) or ONLINE (round-robin)';

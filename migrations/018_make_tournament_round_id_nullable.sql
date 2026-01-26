-- Migration: Make tournament_round_id nullable in tournament_matches
-- Created: 2026-01-26
-- Purpose: Support archiving of online (round-less) tournaments

ALTER TABLE tournament_matches ALTER COLUMN tournament_round_id DROP NOT NULL;

-- Add comment for clarity
COMMENT ON COLUMN tournament_matches.tournament_round_id IS 'NULL for archived online tournaments (no rounds), non-null for archived in-person tournaments';

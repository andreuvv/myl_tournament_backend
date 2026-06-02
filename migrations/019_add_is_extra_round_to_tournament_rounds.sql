-- Migration: Add is_extra_round to tournament_rounds
-- Created: 2026-06-02

ALTER TABLE tournament_rounds
ADD COLUMN IF NOT EXISTS is_extra_round BOOLEAN NOT NULL DEFAULT FALSE;

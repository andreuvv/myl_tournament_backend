-- Migration: Add format column to tournaments table
-- Created: 2026-01-26
-- Purpose: Store tournament format (PB or BF) for online tournaments

ALTER TABLE tournaments ADD COLUMN format VARCHAR(10) CHECK (format IN ('PB', 'BF'));

-- Add comment for clarity
COMMENT ON COLUMN tournaments.format IS 'Tournament format: PB or BF. Required for ONLINE tournaments, optional for IN_PERSON tournaments.';

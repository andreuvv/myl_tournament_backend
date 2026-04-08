-- Migration 022: Add subformat column to tournament_rounds table
-- Supports subformats for Primer Bloque (PBRL, PBRE) and Bloque Furia (BFRL, BFVCR)

ALTER TABLE tournament_rounds ADD COLUMN subformat VARCHAR(20);

-- Add comment explaining subformat values
COMMENT ON COLUMN tournament_rounds.subformat IS 'Subformat of the round (PBRL, PBRE for PB format; BFRL, BFVCR for BF format)';

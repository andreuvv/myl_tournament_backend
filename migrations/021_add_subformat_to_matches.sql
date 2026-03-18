-- Migration 021: Add subformat column to matches table
-- Supports subformats for Primer Bloque (PBRL, PBRE) and Bloque Furia (BFRL, BFVCR)

ALTER TABLE matches ADD COLUMN subformat VARCHAR(20);

-- Add comment explaining subformat values
COMMENT ON COLUMN matches.subformat IS 'Subformat of the match (PBRL, PBRE for PB format; BFRL, BFVCR for BF format)';

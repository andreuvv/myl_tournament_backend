-- Migration 023: Add race_libre and race_edition_vcr columns to tournament_player_races table
-- Adds new race formats to support Libre and VCR editions

ALTER TABLE tournament_player_races ADD COLUMN race_libre VARCHAR(100);
ALTER TABLE tournament_player_races ADD COLUMN race_edition_vcr VARCHAR(100);

-- Add comments explaining the new race format columns
COMMENT ON COLUMN tournament_player_races.race_libre IS 'Race chosen for Libre format tournaments';
COMMENT ON COLUMN tournament_player_races.race_edition_vcr IS 'Race chosen for Edition VCR format tournaments';

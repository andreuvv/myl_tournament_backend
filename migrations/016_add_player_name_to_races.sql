-- Migration: Add player_name to tournament_player_races for player matching
-- Created: 2026-01-06

ALTER TABLE tournament_player_races ADD COLUMN IF NOT EXISTS player_name VARCHAR(100);

-- Update existing records with player names from tournament_standings
UPDATE tournament_player_races tpr
SET player_name = ts.player_name
FROM tournament_standings ts
WHERE tpr.tournament_id = ts.tournament_id 
  AND tpr.player_id = ts.player_id
  AND tpr.player_name IS NULL;

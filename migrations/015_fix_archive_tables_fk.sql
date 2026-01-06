-- Migration: Remove foreign key constraints from archive tables
-- Reason: Archive tables should be independent of active players table
-- Created: 2026-01-05

-- Drop foreign key constraints on tournament_standings
ALTER TABLE tournament_standings DROP CONSTRAINT IF EXISTS tournament_standings_player_id_fkey;

-- Drop foreign key constraints on tournament_matches
ALTER TABLE tournament_matches DROP CONSTRAINT IF EXISTS tournament_matches_player1_id_fkey;
ALTER TABLE tournament_matches DROP CONSTRAINT IF EXISTS tournament_matches_player2_id_fkey;

-- Drop foreign key constraints on tournament_player_races
ALTER TABLE tournament_player_races DROP CONSTRAINT IF EXISTS tournament_player_races_player_id_fkey;

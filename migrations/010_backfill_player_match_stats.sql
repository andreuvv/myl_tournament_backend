-- Migration: Backfill player_match_stats from existing completed matches
-- Created: 2025-12-08

-- Insert stats for player1 from all completed matches
INSERT INTO player_match_stats (player_id, match_id, games_played, games_won, created_at, updated_at)
SELECT 
    m.player1_id,
    m.id,
    m.score1 + m.score2 as games_played,
    m.score1 as games_won,
    m.updated_at as created_at,
    m.updated_at
FROM matches m
WHERE m.completed = true
ON CONFLICT (player_id, match_id) 
DO UPDATE SET 
    games_played = EXCLUDED.games_played, 
    games_won = EXCLUDED.games_won,
    updated_at = EXCLUDED.updated_at;

-- Insert stats for player2 from all completed matches
INSERT INTO player_match_stats (player_id, match_id, games_played, games_won, created_at, updated_at)
SELECT 
    m.player2_id,
    m.id,
    m.score1 + m.score2 as games_played,
    m.score2 as games_won,
    m.updated_at as created_at,
    m.updated_at
FROM matches m
WHERE m.completed = true
ON CONFLICT (player_id, match_id) 
DO UPDATE SET 
    games_played = EXCLUDED.games_played, 
    games_won = EXCLUDED.games_won,
    updated_at = EXCLUDED.updated_at;

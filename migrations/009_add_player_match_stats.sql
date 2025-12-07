-- Migration: Add player_match_stats table to track individual games per match
-- Created: 2025-12-07

CREATE TABLE IF NOT EXISTS player_match_stats (
    id SERIAL PRIMARY KEY,
    player_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    match_id INTEGER NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    games_played INTEGER NOT NULL DEFAULT 0,
    games_won INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, match_id)
);

CREATE INDEX idx_player_match_stats_player ON player_match_stats(player_id);
CREATE INDEX idx_player_match_stats_match ON player_match_stats(match_id);

-- Recreate standings view to use player_match_stats
DROP VIEW IF EXISTS standings;

CREATE VIEW standings AS
SELECT 
    p.id,
    p.name,
    COUNT(CASE WHEN m.completed = true THEN 1 END) as matches_played,
    SUM(CASE 
        WHEN m.completed AND (
            (m.player1_id = p.id AND m.score1 > m.score2) OR 
            (m.player2_id = p.id AND m.score2 > m.score1)
        ) THEN 1 ELSE 0 
    END) as wins,
    SUM(CASE 
        WHEN m.completed AND (
            (m.player1_id = p.id AND m.score1 = m.score2) OR 
            (m.player2_id = p.id AND m.score2 = m.score1)
        ) THEN 1 ELSE 0 
    END) as ties,
    SUM(CASE 
        WHEN m.completed AND (
            (m.player1_id = p.id AND m.score1 < m.score2) OR 
            (m.player2_id = p.id AND m.score2 < m.score1)
        ) THEN 1 ELSE 0 
    END) as losses,
    SUM(CASE 
        WHEN m.completed AND (
            (m.player1_id = p.id AND m.score1 > m.score2) OR 
            (m.player2_id = p.id AND m.score2 > m.score1)
        ) THEN 3
        WHEN m.completed AND (
            (m.player1_id = p.id AND m.score1 = m.score2) OR 
            (m.player2_id = p.id AND m.score2 = m.score1)
        ) THEN 1
        ELSE 0 
    END) as points,
    COALESCE(SUM(CASE 
        WHEN m.completed AND m.player1_id = p.id THEN m.score1
        WHEN m.completed AND m.player2_id = p.id THEN m.score2
        ELSE 0
    END), 0) as total_points_scored,
    COALESCE((
        SELECT SUM(pms.games_played)
        FROM player_match_stats pms
        JOIN matches m2 ON pms.match_id = m2.id
        WHERE pms.player_id = p.id AND m2.completed = true
    ), 0) as total_matches
FROM players p
LEFT JOIN matches m ON (m.player1_id = p.id OR m.player2_id = p.id)
WHERE p.confirmed = true
GROUP BY p.id, p.name
ORDER BY points DESC, total_points_scored DESC;

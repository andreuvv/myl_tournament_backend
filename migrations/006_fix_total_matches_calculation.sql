-- Migration: Fix total_matches calculation in standings view
-- Created: 2025-12-07

DROP VIEW IF EXISTS standings;

CREATE OR REPLACE VIEW standings AS
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
    -- Match points: 3 for win, 1 for tie, 0 for loss
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
    -- Total goals/points scored in all matches
    COALESCE(SUM(CASE 
        WHEN m.completed AND m.player1_id = p.id THEN m.score1
        WHEN m.completed AND m.player2_id = p.id THEN m.score2
        ELSE 0
    END), 0) as total_points_scored,
    -- Total individual matches played (sum of both scores for each match the player participated in)
    COALESCE(SUM(CASE 
        WHEN m.completed THEN (COALESCE(m.score1, 0) + COALESCE(m.score2, 0))
        ELSE 0
    END), 0) as total_matches
FROM players p
LEFT JOIN matches m ON (m.player1_id = p.id OR m.player2_id = p.id)
WHERE p.confirmed = true
GROUP BY p.id, p.name
ORDER BY points DESC, total_points_scored DESC;

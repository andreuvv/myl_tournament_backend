-- Migration: Add ties column to standings view
-- Created: 2025-12-05

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
    SUM(CASE 
        WHEN m.completed AND m.player1_id = p.id THEN m.score1
        WHEN m.completed AND m.player2_id = p.id THEN m.score2
        ELSE 0
    END) as total_points_scored
FROM players p
LEFT JOIN matches m ON (m.player1_id = p.id OR m.player2_id = p.id)
WHERE p.confirmed = true
GROUP BY p.id, p.name
ORDER BY wins DESC, total_points_scored DESC;

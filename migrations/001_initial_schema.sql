-- Migration: Initial schema for tournament management
-- Created: 2025-12-05

-- Players table
CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    confirmed BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Rounds table
CREATE TABLE IF NOT EXISTS rounds (
    id SERIAL PRIMARY KEY,
    round_number INTEGER NOT NULL UNIQUE,
    format VARCHAR(10) NOT NULL CHECK (format IN ('PB', 'BF')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Matches table
CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    round_id INTEGER NOT NULL REFERENCES rounds(id) ON DELETE CASCADE,
    player1_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    player2_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    score1 INTEGER,
    score2 INTEGER,
    completed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT different_players CHECK (player1_id != player2_id)
);

-- Standings view (calculated from match results)
CREATE OR REPLACE VIEW standings AS
SELECT 
    p.id,
    p.name,
    COUNT(DISTINCT m.round_id) as matches_played,
    SUM(CASE 
        WHEN m.completed AND (
            (m.player1_id = p.id AND m.score1 > m.score2) OR 
            (m.player2_id = p.id AND m.score2 > m.score1)
        ) THEN 1 ELSE 0 
    END) as wins,
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
    END) as total_points_scored,
    SUM(CASE 
        WHEN m.completed AND m.player1_id = p.id THEN m.score2
        WHEN m.completed AND m.player2_id = p.id THEN m.score1
        ELSE 0
    END) as total_points_against
FROM players p
LEFT JOIN matches m ON (m.player1_id = p.id OR m.player2_id = p.id)
WHERE p.confirmed = true
GROUP BY p.id, p.name
ORDER BY wins DESC, total_points_scored DESC;

-- Indexes for better performance
CREATE INDEX idx_matches_round ON matches(round_id);
CREATE INDEX idx_matches_player1 ON matches(player1_id);
CREATE INDEX idx_matches_player2 ON matches(player2_id);
CREATE INDEX idx_players_confirmed ON players(confirmed);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to auto-update updated_at
CREATE TRIGGER update_players_updated_at BEFORE UPDATE ON players
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rounds_updated_at BEFORE UPDATE ON rounds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_matches_updated_at BEFORE UPDATE ON matches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

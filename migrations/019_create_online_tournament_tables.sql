-- Migration: Create online tournament tables for round-robin tournaments
-- Created: 2026-01-26
-- Purpose: Support active online tournaments that play round-robin style (no rounds)

-- Online tournament players (active)
CREATE TABLE IF NOT EXISTS online_tournament_players (
    id SERIAL PRIMARY KEY,
    tournament_id INTEGER NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    player_id INTEGER NOT NULL REFERENCES premier_players(id) ON DELETE CASCADE,
    player_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_id, player_id)
);

-- Online tournament matches (active, no rounds)
CREATE TABLE IF NOT EXISTS online_tournament_matches (
    id SERIAL PRIMARY KEY,
    tournament_id INTEGER NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    player1_id INTEGER NOT NULL REFERENCES premier_players(id) ON DELETE CASCADE,
    player2_id INTEGER NOT NULL REFERENCES premier_players(id) ON DELETE CASCADE,
    player1_name VARCHAR(100) NOT NULL,
    player2_name VARCHAR(100) NOT NULL,
    score1 INTEGER,
    score2 INTEGER,
    completed BOOLEAN DEFAULT false,
    match_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT different_players CHECK (player1_id != player2_id),
    UNIQUE(tournament_id, player1_id, player2_id)
);

-- Online tournament standings (active, calculated from matches)
CREATE OR REPLACE VIEW online_tournament_standings AS
SELECT 
    otp.tournament_id,
    otp.player_id,
    otp.player_name,
    COUNT(CASE WHEN otm.completed THEN 1 END) as matches_played,
    SUM(CASE 
        WHEN otm.completed AND (
            (otm.player1_id = otp.player_id AND otm.score1 > otm.score2) OR 
            (otm.player2_id = otp.player_id AND otm.score2 > otm.score1)
        ) THEN 1 ELSE 0 
    END) as wins,
    SUM(CASE 
        WHEN otm.completed AND otm.score1 = otm.score2 THEN 1 ELSE 0 
    END) as ties,
    SUM(CASE 
        WHEN otm.completed AND (
            (otm.player1_id = otp.player_id AND otm.score1 < otm.score2) OR 
            (otm.player2_id = otp.player_id AND otm.score2 < otm.score1)
        ) THEN 1 ELSE 0 
    END) as losses,
    SUM(CASE 
        WHEN otm.completed AND otm.player1_id = otp.player_id THEN 
            CASE WHEN otm.score1 > otm.score2 THEN 3 
                 WHEN otm.score1 = otm.score2 THEN 1 
                 ELSE 0 END
        WHEN otm.completed AND otm.player2_id = otp.player_id THEN 
            CASE WHEN otm.score2 > otm.score1 THEN 3 
                 WHEN otm.score2 = otm.score1 THEN 1 
                 ELSE 0 END
        ELSE 0
    END) as points
FROM online_tournament_players otp
LEFT JOIN online_tournament_matches otm ON otm.tournament_id = otp.tournament_id 
    AND (otm.player1_id = otp.player_id OR otm.player2_id = otp.player_id)
GROUP BY otp.tournament_id, otp.player_id, otp.player_name
ORDER BY points DESC, wins DESC;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_online_tournament_players_tournament ON online_tournament_players(tournament_id);
CREATE INDEX IF NOT EXISTS idx_online_tournament_players_player ON online_tournament_players(player_id);
CREATE INDEX IF NOT EXISTS idx_online_tournament_matches_tournament ON online_tournament_matches(tournament_id);
CREATE INDEX IF NOT EXISTS idx_online_tournament_matches_player1 ON online_tournament_matches(player1_id);
CREATE INDEX IF NOT EXISTS idx_online_tournament_matches_player2 ON online_tournament_matches(player2_id);
CREATE INDEX IF NOT EXISTS idx_online_tournament_matches_completed ON online_tournament_matches(completed);

-- Create triggers to auto-update updated_at
CREATE TRIGGER update_online_tournament_players_updated_at BEFORE UPDATE ON online_tournament_players
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_online_tournament_matches_updated_at BEFORE UPDATE ON online_tournament_matches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

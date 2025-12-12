-- Migration: Insert initial players
-- Created: 2025-12-11

-- Insert all players with confirmed status
INSERT INTO players (name, confirmed) VALUES
('Troke', true),
('Timmy', true),
('Piter', true),
('Folo', true),
('Wesh', true),
('Guari', true),
('Vinny', true),
('Chisco', true),
('Clanso', true),
('Traukolin', true),
('Chester', true),
('David', true)
ON CONFLICT (name) DO UPDATE SET confirmed = EXCLUDED.confirmed;

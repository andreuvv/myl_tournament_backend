-- Create table to store player race information per tournament
CREATE TABLE tournament_player_races (
  id SERIAL PRIMARY KEY,
  tournament_id INT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
  player_id INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  race_pb VARCHAR(100),
  race_bf VARCHAR(100),
  notes TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  -- Ensure one record per player per tournament
  UNIQUE(tournament_id, player_id)
);

-- Create index for faster lookups
CREATE INDEX idx_tournament_player_races_tournament_id ON tournament_player_races(tournament_id);
CREATE INDEX idx_tournament_player_races_player_id ON tournament_player_races(player_id);

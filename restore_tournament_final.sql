-- Tournament Data Restoration Script
-- Copa K&T Dic 2025
-- Using actual player IDs

BEGIN;

-- Insert rounds
INSERT INTO rounds (round_number, format) VALUES
(1, 'PB'), (2, 'BF'), (3, 'PB'), (4, 'BF'), (5, 'PB'), (6, 'BF'), (7, 'PB'), (8, 'BF');

-- Round 1 (PB)
INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES
(1, 150, 152, 2, 0, true),  -- Piter 2-0 Traukolin
(1, 145, 147, 1, 2, true),  -- Chester 1-2 Clanso
(1, 148, 154, 1, 1, true),  -- David 1-1 Wesh
(1, 149, 151, 2, 0, true),  -- Folo 2-0 Timmy
(1, 153, 146, 2, 0, true);  -- Troke 2-0 Chisco

-- Round 2 (BF)
INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES
(2, 153, 152, 0, 2, true),  -- Troke 0-2 Traukolin
(2, 149, 154, 1, 2, true),  -- Folo 1-2 Wesh
(2, 146, 151, 2, 0, true),  -- Chisco 2-0 Timmy
(2, 150, 147, 1, 2, true),  -- Piter 1-2 Clanso
(2, 145, 148, 0, 2, true);  -- Chester 0-2 David

-- Round 3 (PB)
INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES
(3, 153, 151, 2, 0, true),  -- Troke 2-0 Timmy
(3, 152, 147, 1, 2, true),  -- Traukolin 1-2 Clanso
(3, 146, 154, 1, 0, true),  -- Chisco 1-0 Wesh
(3, 150, 148, 2, 0, true),  -- Piter 2-0 David
(3, 149, 145, 2, 0, true);  -- Folo 2-0 Chester

-- Round 4 (BF)
INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES
(4, 153, 147, 0, 2, true),  -- Troke 0-2 Clanso
(4, 152, 148, 2, 0, true),  -- Traukolin 2-0 David
(4, 151, 154, 0, 2, true),  -- Timmy 0-2 Wesh
(4, 150, 149, 2, 1, true),  -- Piter 2-1 Folo
(4, 146, 145, 2, 1, true);  -- Chisco 2-1 Chester

-- Round 5 (PB)
INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES
(5, 153, 154, 1, 1, true),  -- Troke 1-1 Wesh
(5, 147, 148, 2, 0, true),  -- Clanso 2-0 David
(5, 151, 145, 0, 2, true),  -- Timmy 0-2 Chester
(5, 146, 150, 1, 1, true),  -- Chisco 1-1 Piter
(5, 152, 149, 0, 1, true);  -- Traukolin 0-1 Folo

-- Round 6 (BF)
INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES
(6, 153, 148, 2, 0, true),  -- Troke 2-0 David
(6, 154, 145, 2, 0, true),  -- Wesh 2-0 Chester
(6, 147, 149, 0, 2, true),  -- Clanso 0-2 Folo
(6, 151, 150, 0, 2, true),  -- Timmy 0-2 Piter
(6, 152, 146, 1, 1, true);  -- Traukolin 1-1 Chisco

-- Round 7 (PB)
INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES
(7, 153, 145, 1, 1, true),  -- Troke 1-1 Chester
(7, 148, 149, 0, 2, true),  -- David 0-2 Folo
(7, 154, 150, 1, 1, true),  -- Wesh 1-1 Piter
(7, 147, 146, 2, 1, true),  -- Clanso 2-1 Chisco
(7, 151, 152, 0, 2, true);  -- Timmy 0-2 Traukolin

-- Round 8 (BF)
INSERT INTO matches (round_id, player1_id, player2_id, score1, score2, completed) VALUES
(8, 153, 149, 1, 1, true),  -- Troke 1-1 Folo
(8, 145, 150, 2, 1, true),  -- Chester 2-1 Piter
(8, 148, 146, 0, 2, true),  -- David 0-2 Chisco
(8, 154, 152, 0, 2, true),  -- Wesh 0-2 Traukolin
(8, 147, 151, 2, 0, true);  -- Clanso 2-0 Timmy

COMMIT;

-- Verify final standings:
-- SELECT * FROM standings ORDER BY points DESC, total_points_scored DESC;
-- Expected:
-- 1. Clanso: 21 pts (7W 0T 1L)
-- 2. Folo: 16 pts (5W 1T 2L)
-- 3. Piter: 14 pts (4W 2T 2L)
-- 4. Chisco: 14 pts (4W 2T 2L)
-- 5. Traukolin: 13 pts (4W 1T 3L)
-- 6. Wesh: 12 pts (3W 3T 2L)
-- 7. Troke: 12 pts (3W 3T 2L)
-- 8. Chester: 7 pts (2W 1T 5L)
-- 9. David: 4 pts (1W 1T 6L)
-- 10. Timmy: 0 pts (0W 0T 8L)

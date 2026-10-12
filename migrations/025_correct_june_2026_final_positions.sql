-- Furia del Ragnarok (June 2026, tournament 17).
-- Before the final, Wesh and Folo were tied on points, game score, and wins.
-- Wesh had the head-to-head, so he was 2nd and Folo 3rd. The final was
-- Traukolin vs Wesh (Wesh won: 1st and 2nd) and Folo vs Piter (Folo won: 3rd).
-- Folo had been stored as 2nd because the tie used player id instead of that result.

UPDATE tournament_standings
SET final_position = CASE player_name
    WHEN 'Traukolin' THEN 2
    WHEN 'Folo' THEN 3
END
WHERE tournament_id = 17
  AND (
    (player_name = 'Traukolin' AND final_position = 3)
    OR (player_name = 'Folo' AND final_position = 2)
  );

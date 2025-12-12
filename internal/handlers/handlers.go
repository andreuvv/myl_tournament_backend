package handlers

import (
	"database/sql"
	"net/http"

	"github.com/andreuvv/premier_mitologico/backend/internal/database"
	"github.com/andreuvv/premier_mitologico/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// GetFixture returns all rounds with their matches
func GetFixture(c *gin.Context) {
	query := `
		SELECT 
			r.round_number,
			r.format,
			m.id as match_id,
			p1.name as player1_name,
			p2.name as player2_name,
			m.score1,
			m.score2,
			m.completed,
			m.updated_at
		FROM rounds r
		LEFT JOIN matches m ON m.round_id = r.id
		LEFT JOIN players p1 ON m.player1_id = p1.id
		LEFT JOIN players p2 ON m.player2_id = p2.id
		ORDER BY r.round_number, m.id
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch fixture"})
		return
	}
	defer rows.Close()

	roundsMap := make(map[int]*models.FixtureRound)

	for rows.Next() {
		var roundNum int
		var format string
		var match models.MatchDetail

		err := rows.Scan(
			&roundNum,
			&format,
			&match.ID,
			&match.Player1Name,
			&match.Player2Name,
			&match.Score1,
			&match.Score2,
			&match.Completed,
			&match.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if _, exists := roundsMap[roundNum]; !exists {
			roundsMap[roundNum] = &models.FixtureRound{
				Number:  roundNum,
				Format:  format,
				Matches: []models.MatchDetail{},
			}
		}

		match.RoundNumber = roundNum
		match.Format = format
		roundsMap[roundNum].Matches = append(roundsMap[roundNum].Matches, match)
	}

	// Convert map to sorted slice
	rounds := []models.FixtureRound{}
	for i := 1; i <= len(roundsMap); i++ {
		if round, exists := roundsMap[i]; exists {
			rounds = append(rounds, *round)
		}
	}

	c.JSON(http.StatusOK, models.FixtureResponse{Rounds: rounds})
}

// GetStandings returns current tournament standings
func GetStandings(c *gin.Context) {
	query := `
		SELECT 
			id,
			name,
			matches_played,
			wins,
			ties,
			losses,
			points,
			total_points_scored,
			total_matches
		FROM standings
		ORDER BY points DESC, total_points_scored DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch standings"})
		return
	}
	defer rows.Close()

	standings := []models.Standing{}
	for rows.Next() {
		var s models.Standing
		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.MatchesPlayed,
			&s.Wins,
			&s.Ties,
			&s.Losses,
			&s.Points,
			&s.TotalPointsScored,
			&s.TotalMatches,
		)
		if err != nil {
			continue
		}
		standings = append(standings, s)
	}

	c.JSON(http.StatusOK, standings)
}

// UpdateMatchScore updates the score for a specific match
func UpdateMatchScore(c *gin.Context) {
	matchID := c.Param("id")

	var req models.UpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get player IDs for this match
	var player1ID, player2ID int
	queryPlayers := `SELECT player1_id, player2_id FROM matches WHERE id = $1`
	err = tx.QueryRow(queryPlayers, matchID).Scan(&player1ID, &player2ID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch match"})
		return
	}

	// Update match score
	query := `
		UPDATE matches 
		SET score1 = $1, score2 = $2, completed = true, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id
	`

	var id int
	err = tx.QueryRow(query, req.Score1, req.Score2, matchID).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update score"})
		return
	}

	// Update player_match_stats for both players
	totalGames := req.Score1 + req.Score2

	// Player 1 stats
	upsertStats := `
		INSERT INTO player_match_stats (player_id, match_id, games_played, games_won, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (player_id, match_id) 
		DO UPDATE SET games_played = $3, games_won = $4, updated_at = CURRENT_TIMESTAMP
	`
	_, err = tx.Exec(upsertStats, player1ID, matchID, totalGames, req.Score1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player1 stats"})
		return
	}

	// Player 2 stats
	_, err = tx.Exec(upsertStats, player2ID, matchID, totalGames, req.Score2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player2 stats"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Score updated successfully", "match_id": id})
}

// GetPlayers returns all players
func GetPlayers(c *gin.Context) {
	query := `SELECT id, name, confirmed, created_at, updated_at FROM players ORDER BY name`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}
	defer rows.Close()

	players := []models.Player{}
	for rows.Next() {
		var p models.Player
		err := rows.Scan(&p.ID, &p.Name, &p.Confirmed, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		players = append(players, p)
	}

	c.JSON(http.StatusOK, players)
}

// CreatePlayer creates a new player
func CreatePlayer(c *gin.Context) {
	var req models.CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		INSERT INTO players (name, confirmed) 
		VALUES ($1, $2) 
		RETURNING id, name, confirmed, created_at, updated_at
	`

	var player models.Player
	err := database.DB.QueryRow(query, req.Name, req.Confirmed).Scan(
		&player.ID,
		&player.Name,
		&player.Confirmed,
		&player.CreatedAt,
		&player.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player"})
		return
	}

	c.JSON(http.StatusCreated, player)
}

// CreateFixture creates the complete fixture (players, rounds, and matches)
func CreateFixture(c *gin.Context) {
	var req models.CreateFixtureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Clear existing data
	if _, err := tx.Exec("DELETE FROM matches"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear matches"})
		return
	}
	if _, err := tx.Exec("DELETE FROM rounds"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear rounds"})
		return
	}
	if _, err := tx.Exec("DELETE FROM players"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear players"})
		return
	}

	// Create players and build name-to-id map
	playerMap := make(map[string]int)
	for _, p := range req.Players {
		var playerID int
		err := tx.QueryRow(
			"INSERT INTO players (name, confirmed) VALUES ($1, $2) RETURNING id",
			p.Name, p.Confirmed,
		).Scan(&playerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player: " + p.Name})
			return
		}
		playerMap[p.Name] = playerID
	}

	// Create rounds and matches
	for _, r := range req.Rounds {
		var roundID int
		err := tx.QueryRow(
			"INSERT INTO rounds (round_number, format) VALUES ($1, $2) RETURNING id",
			r.RoundNumber, r.Format,
		).Scan(&roundID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create round"})
			return
		}

		// Create matches for this round
		for _, m := range r.Matches {
			player1ID, ok1 := playerMap[m.Player1Name]
			player2ID, ok2 := playerMap[m.Player2Name]

			if !ok1 || !ok2 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Player not found: " + m.Player1Name + " or " + m.Player2Name})
				return
			}

			_, err := tx.Exec(
				"INSERT INTO matches (round_id, player1_id, player2_id) VALUES ($1, $2, $3)",
				roundID, player1ID, player2ID,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match"})
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Fixture created successfully",
		"players_created": len(req.Players),
		"rounds_created":  len(req.Rounds),
	})
}

// ClearTournament deletes all matches and rounds, optionally players too
func ClearTournament(c *gin.Context) {
	// Check if we should also clear players
	clearPlayers := c.Query("clear_players") == "true"

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Delete player_match_stats first (foreign key constraint)
	if _, err := tx.Exec("DELETE FROM player_match_stats"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete player stats"})
		return
	}

	// Delete matches (foreign key constraint)
	if _, err := tx.Exec("DELETE FROM matches"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete matches"})
		return
	}

	// Delete rounds
	if _, err := tx.Exec("DELETE FROM rounds"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rounds"})
		return
	}

	// Optionally delete players
	if clearPlayers {
		if _, err := tx.Exec("DELETE FROM players"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete players"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	message := "Tournament cleared: matches and rounds deleted"
	if clearPlayers {
		message = "Tournament cleared: matches, rounds, and players deleted"
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

// TogglePlayerConfirmed toggles the confirmed status of a player
func TogglePlayerConfirmed(c *gin.Context) {
	playerID := c.Param("id")

	query := `
		UPDATE players 
		SET confirmed = NOT confirmed, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 
		RETURNING id, name, confirmed, created_at, updated_at
	`

	var player models.Player
	err := database.DB.QueryRow(query, playerID).Scan(
		&player.ID,
		&player.Name,
		&player.Confirmed,
		&player.CreatedAt,
		&player.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player"})
		return
	}

	c.JSON(http.StatusOK, player)
}

// GetConfirmedPlayers returns only confirmed players
func GetConfirmedPlayers(c *gin.Context) {
	query := `SELECT id, name, confirmed, created_at, updated_at FROM players WHERE confirmed = true ORDER BY name`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch confirmed players"})
		return
	}
	defer rows.Close()

	players := []models.Player{}
	for rows.Next() {
		var p models.Player
		err := rows.Scan(&p.ID, &p.Name, &p.Confirmed, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		players = append(players, p)
	}

	c.JSON(http.StatusOK, players)
}

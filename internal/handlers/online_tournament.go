package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/andreuvv/premier_mitologico/backend/internal/database"
	"github.com/andreuvv/premier_mitologico/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// CreateOnlineTournament creates a new online tournament with auto-generated match pairings
func CreateOnlineTournament(c *gin.Context) {
	var req models.CreateOnlineTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate at least 2 players
	if len(req.PlayerIDs) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least 2 players are required"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Create tournament record
	var tournamentID int
	err = tx.QueryRow(`
		INSERT INTO tournaments (name, month, year, type, format, start_date, end_date, created_at, archived_at)
		VALUES ($1, $2, $3, 'ONLINE', $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`, req.Name, req.Month, req.Year, req.Format, req.StartDate, req.EndDate).Scan(&tournamentID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament"})
		return
	}

	// Get player names from IDs (from premier_players table)
	playerMap := make(map[int]string)
	for _, playerID := range req.PlayerIDs {
		var name string
		err := tx.QueryRow("SELECT name FROM premier_players WHERE id = $1", playerID).Scan(&name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Player with ID %d not found in premier players", playerID)})
			return
		}
		playerMap[playerID] = name
	}

	// Insert tournament players
	for _, playerID := range req.PlayerIDs {
		_, err := tx.Exec(`
			INSERT INTO online_tournament_players (tournament_id, player_id, player_name)
			VALUES ($1, $2, $3)
		`, tournamentID, playerID, playerMap[playerID])

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add player to tournament"})
			return
		}
	}

	// Generate all match pairings (round-robin)
	// Each player plays each other player once
	matchCount := 0
	for i := 0; i < len(req.PlayerIDs); i++ {
		for j := i + 1; j < len(req.PlayerIDs); j++ {
			player1ID := req.PlayerIDs[i]
			player2ID := req.PlayerIDs[j]
			player1Name := playerMap[player1ID]
			player2Name := playerMap[player2ID]

			_, err := tx.Exec(`
				INSERT INTO online_tournament_matches 
				(tournament_id, player1_id, player2_id, player1_name, player2_name, completed)
				VALUES ($1, $2, $3, $4, $5, false)
			`, tournamentID, player1ID, player2ID, player1Name, player2Name)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match"})
				return
			}
			matchCount++
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":           "Online tournament created successfully",
		"tournament_id":     tournamentID,
		"tournament_name":   req.Name,
		"format":            req.Format,
		"players_added":     len(req.PlayerIDs),
		"matches_generated": matchCount,
	})
}

// GetOnlineTournamentMatches returns all matches for an online tournament
func GetOnlineTournamentMatches(c *gin.Context) {
	tournamentID := c.Param("id")

	query := `
		SELECT 
			id,
			tournament_id,
			player1_id,
			player2_id,
			player1_name,
			player2_name,
			score1,
			score2,
			completed,
			match_date,
			created_at,
			updated_at
		FROM online_tournament_matches
		WHERE tournament_id = $1
		ORDER BY 
			CASE WHEN completed = false THEN 0 ELSE 1 END ASC,
			player1_name ASC,
			player2_name ASC
	`

	rows, err := database.DB.Query(query, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch matches"})
		return
	}
	defer rows.Close()

	matches := []models.OnlineTournamentMatch{}
	for rows.Next() {
		var match models.OnlineTournamentMatch
		err := rows.Scan(
			&match.ID,
			&match.TournamentID,
			&match.Player1ID,
			&match.Player2ID,
			&match.Player1Name,
			&match.Player2Name,
			&match.Score1,
			&match.Score2,
			&match.Completed,
			&match.MatchDate,
			&match.CreatedAt,
			&match.UpdatedAt,
		)
		if err != nil {
			continue
		}
		matches = append(matches, match)
	}

	c.JSON(http.StatusOK, matches)
}

// GetOnlineTournamentStandings returns standings for an online tournament
func GetOnlineTournamentStandings(c *gin.Context) {
	tournamentID := c.Param("id")

	query := `
		SELECT 
			tournament_id,
			player_id,
			player_name,
			matches_played,
			wins,
			ties,
			losses,
			points
		FROM online_tournament_standings
		WHERE tournament_id = $1
		ORDER BY points DESC, wins DESC
	`

	rows, err := database.DB.Query(query, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch standings"})
		return
	}
	defer rows.Close()

	standings := []models.OnlineTournamentStanding{}
	for rows.Next() {
		var standing models.OnlineTournamentStanding
		err := rows.Scan(
			&standing.TournamentID,
			&standing.PlayerID,
			&standing.PlayerName,
			&standing.MatchesPlayed,
			&standing.Wins,
			&standing.Ties,
			&standing.Losses,
			&standing.Points,
		)
		if err != nil {
			continue
		}
		standings = append(standings, standing)
	}

	c.JSON(http.StatusOK, standings)
}

// UpdateOnlineMatchScore updates the score for a match in an online tournament
func UpdateOnlineMatchScore(c *gin.Context) {
	matchID := c.Param("matchId")
	var req models.UpdateOnlineMatchScoreRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE online_tournament_matches
		SET score1 = $1, score2 = $2, completed = true, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, tournament_id, player1_name, player2_name, score1, score2, completed
	`

	var match models.OnlineTournamentMatch
	err := database.DB.QueryRow(query, req.Score1, req.Score2, matchID).Scan(
		&match.ID,
		&match.TournamentID,
		&match.Player1Name,
		&match.Player2Name,
		&match.Score1,
		&match.Score2,
		&match.Completed,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Match score updated successfully",
		"match_id": match.ID,
		"score":    fmt.Sprintf("%s %d-%d %s", match.Player1Name, *match.Score1, *match.Score2, match.Player2Name),
	})
}

// GetOnlinePendingMatches returns only pending matches (not completed) for an online tournament
func GetOnlinePendingMatches(c *gin.Context) {
	tournamentID := c.Param("id")

	query := `
		SELECT 
			id,
			tournament_id,
			player1_id,
			player2_id,
			player1_name,
			player2_name,
			score1,
			score2,
			completed,
			match_date,
			created_at,
			updated_at
		FROM online_tournament_matches
		WHERE tournament_id = $1 AND completed = false
		ORDER BY player1_name ASC, player2_name ASC
	`

	rows, err := database.DB.Query(query, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending matches"})
		return
	}
	defer rows.Close()

	matches := []models.OnlineTournamentMatch{}
	for rows.Next() {
		var match models.OnlineTournamentMatch
		err := rows.Scan(
			&match.ID,
			&match.TournamentID,
			&match.Player1ID,
			&match.Player2ID,
			&match.Player1Name,
			&match.Player2Name,
			&match.Score1,
			&match.Score2,
			&match.Completed,
			&match.MatchDate,
			&match.CreatedAt,
			&match.UpdatedAt,
		)
		if err != nil {
			continue
		}
		matches = append(matches, match)
	}

	c.JSON(http.StatusOK, matches)
}

// GetOnlineCompletedMatches returns only completed matches for an online tournament
func GetOnlineCompletedMatches(c *gin.Context) {
	tournamentID := c.Param("id")

	query := `
		SELECT 
			id,
			tournament_id,
			player1_id,
			player2_id,
			player1_name,
			player2_name,
			score1,
			score2,
			completed,
			match_date,
			created_at,
			updated_at
		FROM online_tournament_matches
		WHERE tournament_id = $1 AND completed = true
		ORDER BY updated_at DESC, player1_name ASC, player2_name ASC
	`

	rows, err := database.DB.Query(query, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch completed matches"})
		return
	}
	defer rows.Close()

	matches := []models.OnlineTournamentMatch{}
	for rows.Next() {
		var match models.OnlineTournamentMatch
		err := rows.Scan(
			&match.ID,
			&match.TournamentID,
			&match.Player1ID,
			&match.Player2ID,
			&match.Player1Name,
			&match.Player2Name,
			&match.Score1,
			&match.Score2,
			&match.Completed,
			&match.MatchDate,
			&match.CreatedAt,
			&match.UpdatedAt,
		)
		if err != nil {
			continue
		}
		matches = append(matches, match)
	}

	c.JSON(http.StatusOK, matches)
}

// DeleteOnlineTournament deletes an online tournament and all its data
func DeleteOnlineTournament(c *gin.Context) {
	tournamentID := c.Param("id")

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Delete tournament matches
	_, err = tx.Exec("DELETE FROM online_tournament_matches WHERE tournament_id = $1", tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete matches"})
		return
	}

	// Delete tournament players
	_, err = tx.Exec("DELETE FROM online_tournament_players WHERE tournament_id = $1", tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete players"})
		return
	}

	// Delete tournament record
	result, err := tx.Exec("DELETE FROM tournaments WHERE id = $1 AND type = 'ONLINE'", tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tournament"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found or is not an online tournament"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Online tournament deleted successfully"})
}

// GetOnlineTournamentInfo returns tournament info (metadata)
func GetOnlineTournamentInfo(c *gin.Context) {
	tournamentID := c.Param("id")

	query := `
		SELECT 
			id,
			name,
			month,
			year,
			type,
			format,
			start_date,
			end_date,
			created_at,
			archived_at
		FROM tournaments
		WHERE id = $1 AND type = 'ONLINE'
	`

	var tournament models.Tournament
	var format sql.NullString
	err := database.DB.QueryRow(query, tournamentID).Scan(
		&tournament.ID,
		&tournament.Name,
		&tournament.Month,
		&tournament.Year,
		nil, // type
		&format,
		&tournament.StartDate,
		&tournament.EndDate,
		&tournament.CreatedAt,
		&tournament.ArchivedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tournament"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         tournament.ID,
		"name":       tournament.Name,
		"month":      tournament.Month,
		"year":       tournament.Year,
		"format":     format.String,
		"type":       "ONLINE",
		"start_date": tournament.StartDate,
		"end_date":   tournament.EndDate,
		"created_at": tournament.CreatedAt,
	})
}

// GetAllActiveTournaments returns all active tournaments (in-person and online)
func GetAllActiveTournaments(c *gin.Context) {
	query := `
		SELECT 
			id,
			name,
			month,
			year,
			type,
			format,
			start_date,
			end_date,
			created_at
		FROM tournaments
		WHERE archived_at IS NOT NULL OR archived_at = CURRENT_TIMESTAMP
		ORDER BY created_at DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tournaments"})
		return
	}
	defer rows.Close()

	type TournamentInfo struct {
		ID        int     `json:"id"`
		Name      string  `json:"name"`
		Month     string  `json:"month"`
		Year      int     `json:"year"`
		Type      string  `json:"type"`
		Format    *string `json:"format"`
		StartDate *string `json:"start_date"`
		EndDate   *string `json:"end_date"`
		CreatedAt string  `json:"created_at"`
	}

	tournaments := []TournamentInfo{}
	for rows.Next() {
		var t TournamentInfo
		var format sql.NullString
		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.Month,
			&t.Year,
			&t.Type,
			&format,
			&t.StartDate,
			&t.EndDate,
			&t.CreatedAt,
		)
		if err != nil {
			continue
		}
		if format.Valid {
			t.Format = &format.String
		}
		tournaments = append(tournaments, t)
	}

	c.JSON(http.StatusOK, tournaments)
}

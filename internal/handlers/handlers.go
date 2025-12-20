package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

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

// ArchiveTournament archives the current tournament data
func ArchiveTournament(c *gin.Context) {
	var req models.ArchiveTournamentRequest
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

	// Create tournament record
	var tournamentID int
	err = tx.QueryRow(`
		INSERT INTO tournaments (name, month, year, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, req.Name, req.Month, req.Year, req.StartDate, req.EndDate).Scan(&tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament: " + err.Error()})
		return
	}

	// Archive current standings with position
	standingsQuery := `
		INSERT INTO tournament_standings (
			tournament_id, player_id, player_name, matches_played, wins, ties, losses,
			points, total_points_scored, total_matches, final_position
		)
		SELECT 
			$1, id, name, matches_played, wins, ties, losses,
			points, total_points_scored, total_matches,
			ROW_NUMBER() OVER (ORDER BY points DESC, total_points_scored DESC) as position
		FROM standings
	`
	_, err = tx.Exec(standingsQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive standings: " + err.Error()})
		return
	}

	// Archive rounds and matches
	// First, fetch all rounds into memory
	type roundData struct {
		ID     int
		Number int
		Format string
	}
	var rounds []roundData

	roundsRows, err := tx.Query(`SELECT id, round_number, format FROM rounds ORDER BY round_number`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds: " + err.Error()})
		return
	}

	for roundsRows.Next() {
		var r roundData
		if err := roundsRows.Scan(&r.ID, &r.Number, &r.Format); err != nil {
			roundsRows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan round: " + err.Error()})
			return
		}
		rounds = append(rounds, r)
	}
	roundsRows.Close()

	// Now process each round
	for _, round := range rounds {
		// Create tournament round
		var tournamentRoundID int
		err = tx.QueryRow(`
			INSERT INTO tournament_rounds (tournament_id, round_number, format)
			VALUES ($1, $2, $3)
			RETURNING id
		`, tournamentID, round.Number, round.Format).Scan(&tournamentRoundID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament round: " + err.Error()})
			return
		}

		// Archive matches for this round
		_, err = tx.Exec(`
			INSERT INTO tournament_matches (
				tournament_round_id, player1_id, player2_id, player1_name, player2_name,
				score1, score2, completed
			)
			SELECT 
				$1, m.player1_id, m.player2_id, 
				COALESCE(p1.name, 'Unknown'), COALESCE(p2.name, 'Unknown'),
				m.score1, m.score2, m.completed
			FROM matches m
			LEFT JOIN players p1 ON m.player1_id = p1.id
			LEFT JOIN players p2 ON m.player2_id = p2.id
			WHERE m.round_id = $2
		`, tournamentRoundID, round.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive matches: " + err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit tournament archive"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Tournament archived successfully",
		"tournament_id": tournamentID,
	})
}

// GetTournaments returns all archived tournaments
func GetTournaments(c *gin.Context) {
	query := `
		SELECT id, name, month, year, start_date, end_date, created_at, archived_at
		FROM tournaments
		ORDER BY year DESC, 
			CASE month
				WHEN 'Enero' THEN 1 WHEN 'Febrero' THEN 2 WHEN 'Marzo' THEN 3
				WHEN 'Abril' THEN 4 WHEN 'Mayo' THEN 5 WHEN 'Junio' THEN 6
				WHEN 'Julio' THEN 7 WHEN 'Agosto' THEN 8 WHEN 'Septiembre' THEN 9
				WHEN 'Octubre' THEN 10 WHEN 'Noviembre' THEN 11 WHEN 'Diciembre' THEN 12
			END DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tournaments"})
		return
	}
	defer rows.Close()

	tournaments := make([]models.Tournament, 0)
	for rows.Next() {
		var t models.Tournament
		err := rows.Scan(&t.ID, &t.Name, &t.Month, &t.Year, &t.StartDate, &t.EndDate, &t.CreatedAt, &t.ArchivedAt)
		if err != nil {
			continue
		}
		tournaments = append(tournaments, t)
	}

	c.JSON(http.StatusOK, tournaments)
}

// GetTournamentStandings returns standings for a specific tournament
func GetTournamentStandings(c *gin.Context) {
	tournamentID := c.Param("id")

	query := `
		SELECT 
			ts.id, ts.tournament_id, ts.player_id, ts.player_name, ts.matches_played, ts.wins, ts.ties, ts.losses,
			ts.points, ts.total_points_scored, ts.total_matches, ts.final_position, tpr.race_pb, tpr.race_bf
		FROM tournament_standings ts
		LEFT JOIN tournament_player_races tpr ON ts.tournament_id = tpr.tournament_id AND ts.player_id = tpr.player_id
		WHERE ts.tournament_id = $1
		ORDER BY final_position ASC
	`

	rows, err := database.DB.Query(query, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tournament standings"})
		return
	}
	defer rows.Close()

	standings := []models.TournamentStanding{}
	for rows.Next() {
		var s models.TournamentStanding
		err := rows.Scan(
			&s.ID, &s.TournamentID, &s.PlayerID, &s.PlayerName, &s.MatchesPlayed,
			&s.Wins, &s.Ties, &s.Losses, &s.Points, &s.TotalPointsScored,
			&s.TotalMatches, &s.FinalPosition, &s.RacePB, &s.RaceBF,
		)
		if err != nil {
			continue
		}
		standings = append(standings, s)
	}

	c.JSON(http.StatusOK, standings)
}

// GetTournamentRounds returns rounds and matches for a specific tournament
func GetTournamentRounds(c *gin.Context) {
	tournamentID := c.Param("id")

	// Get tournament name
	var tournamentName string
	err := database.DB.QueryRow(`SELECT name FROM tournaments WHERE id = $1`, tournamentID).Scan(&tournamentName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	// Get rounds
	roundsQuery := `
		SELECT id, round_number, format
		FROM tournament_rounds
		WHERE tournament_id = $1
		ORDER BY round_number
	`

	rows, err := database.DB.Query(roundsQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
		return
	}
	defer rows.Close()

	roundsMap := make(map[int]*models.TournamentRoundDetail)

	for rows.Next() {
		var roundID, roundNumber int
		var format string
		err := rows.Scan(&roundID, &roundNumber, &format)
		if err != nil {
			continue
		}

		roundsMap[roundNumber] = &models.TournamentRoundDetail{
			Number:  roundNumber,
			Format:  format,
			Matches: []models.TournamentMatchInfo{},
		}

		// Get matches for this round
		matchesQuery := `
			SELECT id, player1_name, player2_name, score1, score2, completed
			FROM tournament_matches
			WHERE tournament_round_id = $1
			ORDER BY id
		`

		matchRows, err := database.DB.Query(matchesQuery, roundID)
		if err != nil {
			continue
		}

		for matchRows.Next() {
			var match models.TournamentMatchInfo
			err := matchRows.Scan(&match.ID, &match.Player1Name, &match.Player2Name, &match.Score1, &match.Score2, &match.Completed)
			if err != nil {
				continue
			}
			roundsMap[roundNumber].Matches = append(roundsMap[roundNumber].Matches, match)
		}
		matchRows.Close()
	}

	// Convert map to sorted slice
	rounds := []models.TournamentRoundDetail{}
	for i := 1; i <= len(roundsMap); i++ {
		if round, exists := roundsMap[i]; exists {
			rounds = append(rounds, *round)
		}
	}

	response := models.TournamentRoundsResponse{
		TournamentName: tournamentName,
		Rounds:         rounds,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteArchivedTournament deletes an archived tournament and all its associated data
func DeleteArchivedTournament(c *gin.Context) {
	tournamentID := c.Param("id")

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Check if tournament exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM tournaments WHERE id = $1)", tournamentID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check tournament existence"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	// Delete tournament (CASCADE will handle related records)
	_, err = tx.Exec("DELETE FROM tournaments WHERE id = $1", tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tournament"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit deletion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tournament deleted successfully",
	})
}

// GetTournamentPlayerRaces returns all players and their race selections for a specific tournament
func GetTournamentPlayerRaces(c *gin.Context) {
	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}

	query := `
		SELECT 
			tpr.id,
			tpr.tournament_id,
			tpr.player_id,
			p.name as player_name,
			tpr.race_pb,
			tpr.race_bf,
			tpr.notes,
			tpr.created_at,
			tpr.updated_at
		FROM tournament_player_races tpr
		LEFT JOIN players p ON tpr.player_id = p.id
		WHERE tpr.tournament_id = $1
		ORDER BY p.name
	`

	rows, err := database.DB.Query(query, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player races"})
		return
	}
	defer rows.Close()

	var playerRaces []models.TournamentPlayerRace

	for rows.Next() {
		var pr models.TournamentPlayerRace

		err := rows.Scan(
			&pr.ID,
			&pr.TournamentID,
			&pr.PlayerID,
			&pr.PlayerName,
			&pr.RacePB,
			&pr.RaceBF,
			&pr.Notes,
			&pr.CreatedAt,
			&pr.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning player races"})
			return
		}

		playerRaces = append(playerRaces, pr)
	}

	if playerRaces == nil {
		playerRaces = []models.TournamentPlayerRace{}
	}

	c.JSON(http.StatusOK, playerRaces)
}

// UpdatePlayerRace updates race selections for a player in a specific tournament
func UpdatePlayerRace(c *gin.Context) {
	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}
	playerIDStr := c.Param("player_id")
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	var req models.UpdatePlayerRaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if record exists, if not create it
	var exists bool
	err = database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM tournament_player_races WHERE tournament_id = $1 AND player_id = $2)",
		tournamentID, playerID,
	).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing record"})
		return
	}

	var query string
	var args []interface{}

	if exists {
		// Update existing record
		query = `
			UPDATE tournament_player_races 
			SET race_pb = $1, race_bf = $2, notes = $3, updated_at = CURRENT_TIMESTAMP
			WHERE tournament_id = $4 AND player_id = $5
		`
		args = []interface{}{req.RacePB, req.RaceBF, req.Notes, tournamentID, playerID}
	} else {
		// Insert new record
		query = `
			INSERT INTO tournament_player_races (tournament_id, player_id, race_pb, race_bf, notes)
			VALUES ($1, $2, $3, $4, $5)
		`
		args = []interface{}{tournamentID, playerID, req.RacePB, req.RaceBF, req.Notes}
	}

	_, err = database.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update player race"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Player race updated successfully",
	})
}

// GetArchivedTournamentPlayers returns all players who participated in a specific archived tournament
func GetArchivedTournamentPlayers(c *gin.Context) {
	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}

	query := `
		SELECT 
			player_id as id,
			player_name as name,
			total_matches,
			wins as total_wins,
			ties as total_ties,
			total_points_scored
		FROM tournament_standings
		WHERE tournament_id = $1
		ORDER BY player_name
	`

	rows, err := database.DB.Query(query, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tournament players"})
		return
	}
	defer rows.Close()

	type PlayerInfo struct {
		ID                int    `json:"id"`
		Name              string `json:"name"`
		TotalMatches      int    `json:"total_matches"`
		TotalWins         int    `json:"total_wins"`
		TotalTies         int    `json:"total_ties"`
		TotalPointsScored int    `json:"total_points_scored"`
	}

	var players []PlayerInfo

	for rows.Next() {
		var p PlayerInfo

		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.TotalMatches,
			&p.TotalWins,
			&p.TotalTies,
			&p.TotalPointsScored,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning players"})
			return
		}

		players = append(players, p)
	}

	if players == nil {
		players = []PlayerInfo{}
	}

	c.JSON(http.StatusOK, players)
}

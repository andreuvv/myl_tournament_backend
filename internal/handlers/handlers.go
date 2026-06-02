package handlers

import (
	"database/sql"
	"net/http"
	"sort"
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
			r.is_extra_round,
			m.id as match_id,
			m.subformat,
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
		var isExtraRound bool
		var matchID sql.NullInt64
		var subformat sql.NullString
		var player1Name sql.NullString
		var player2Name sql.NullString
		var score1 sql.NullInt64
		var score2 sql.NullInt64
		var completed sql.NullBool
		var updatedAt sql.NullTime

		err := rows.Scan(
			&roundNum,
			&format,
			&isExtraRound,
			&matchID,
			&subformat,
			&player1Name,
			&player2Name,
			&score1,
			&score2,
			&completed,
			&updatedAt,
		)
		if err != nil {
			continue
		}

		var roundSubformat *string
		if subformat.Valid {
			s := subformat.String
			roundSubformat = &s
		}

		if _, exists := roundsMap[roundNum]; !exists {
			roundsMap[roundNum] = &models.FixtureRound{
				Number:       roundNum,
				Format:       format,
				Subformat:    roundSubformat,
				IsExtraRound: isExtraRound,
				Matches:      []models.MatchDetail{},
			}
		}

		if !matchID.Valid {
			continue
		}

		match := models.MatchDetail{
			ID:           int(matchID.Int64),
			RoundNumber:  roundNum,
			Format:       format,
			Subformat:    roundSubformat,
			IsExtraRound: isExtraRound,
			Player1Name:  player1Name.String,
			Player2Name:  player2Name.String,
			Completed:    completed.Valid && completed.Bool,
		}
		if score1.Valid {
			s := int(score1.Int64)
			match.Score1 = &s
		}
		if score2.Valid {
			s := int(score2.Int64)
			match.Score2 = &s
		}
		if updatedAt.Valid {
			match.UpdatedAt = updatedAt.Time
		}

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

type archiveMatchData struct {
	Player1ID   sql.NullInt64
	Player2ID   sql.NullInt64
	Player1Name string
	Player2Name string
	Score1      sql.NullInt64
	Score2      sql.NullInt64
	Completed   bool
}

type archiveRoundData struct {
	ID           int
	Number       int
	Format       string
	IsExtraRound bool
	Matches      []archiveMatchData
}

func sortStandingsForArchive(standings []models.Standing) {
	sort.SliceStable(standings, func(i, j int) bool {
		if standings[i].Points != standings[j].Points {
			return standings[i].Points > standings[j].Points
		}
		if standings[i].TotalPointsScored != standings[j].TotalPointsScored {
			return standings[i].TotalPointsScored > standings[j].TotalPointsScored
		}
		if standings[i].Wins != standings[j].Wins {
			return standings[i].Wins > standings[j].Wins
		}
		return standings[i].ID < standings[j].ID
	})
}

func calculateArchivedFinalPositions(standings []models.Standing, rounds []archiveRoundData) map[int]int {
	extraMatches := make([]archiveMatchData, 0)
	for _, round := range rounds {
		if round.IsExtraRound {
			extraMatches = append(extraMatches, round.Matches...)
		}
	}

	finalPositions := make(map[int]int, len(standings))
	if len(standings) == 0 {
		return finalPositions
	}

	sortedCurrent := make([]models.Standing, len(standings))
	copy(sortedCurrent, standings)
	sortStandingsForArchive(sortedCurrent)

	if len(extraMatches) == 0 {
		for index, standing := range sortedCurrent {
			finalPositions[standing.ID] = index + 1
		}
		return finalPositions
	}

	preExtraStandings := make(map[int]*models.Standing, len(standings))
	for i := range standings {
		standingCopy := standings[i]
		preExtraStandings[standingCopy.ID] = &standingCopy
	}

	for _, match := range extraMatches {
		if !match.Completed || !match.Score1.Valid || !match.Score2.Valid || !match.Player1ID.Valid || !match.Player2ID.Valid {
			continue
		}

		player1 := preExtraStandings[int(match.Player1ID.Int64)]
		player2 := preExtraStandings[int(match.Player2ID.Int64)]
		if player1 == nil || player2 == nil {
			continue
		}

		score1 := int(match.Score1.Int64)
		score2 := int(match.Score2.Int64)
		if score1 > score2 {
			player1.Points -= 3
			player1.Wins--
		} else if score2 > score1 {
			player2.Points -= 3
			player2.Wins--
		} else {
			player1.Points--
			player2.Points--
			player1.Ties--
			player2.Ties--
		}
	}

	sortedPreExtra := make([]models.Standing, 0, len(preExtraStandings))
	for _, standing := range preExtraStandings {
		sortedPreExtra = append(sortedPreExtra, *standing)
	}
	sortStandingsForArchive(sortedPreExtra)

	seedPositions := make(map[int]int, len(sortedPreExtra))
	for index, standing := range sortedPreExtra {
		seedPositions[standing.ID] = index + 1
	}

	positionOverrides := make(map[int]int)
	playoffPlayers := make(map[int]bool)

	for _, match := range extraMatches {
		if !match.Completed || !match.Score1.Valid || !match.Score2.Valid || !match.Player1ID.Valid || !match.Player2ID.Valid {
			continue
		}

		position1, ok1 := seedPositions[int(match.Player1ID.Int64)]
		position2, ok2 := seedPositions[int(match.Player2ID.Int64)]
		if !ok1 || !ok2 {
			continue
		}

		higherPosition := position1
		lowerPosition := position2
		if lowerPosition < higherPosition {
			higherPosition, lowerPosition = lowerPosition, higherPosition
		}

		score1 := int(match.Score1.Int64)
		score2 := int(match.Score2.Int64)
		if score1 > score2 {
			positionOverrides[int(match.Player1ID.Int64)] = higherPosition
			positionOverrides[int(match.Player2ID.Int64)] = lowerPosition
			playoffPlayers[int(match.Player1ID.Int64)] = true
			playoffPlayers[int(match.Player2ID.Int64)] = true
		} else if score2 > score1 {
			positionOverrides[int(match.Player2ID.Int64)] = higherPosition
			positionOverrides[int(match.Player1ID.Int64)] = lowerPosition
			playoffPlayers[int(match.Player1ID.Int64)] = true
			playoffPlayers[int(match.Player2ID.Int64)] = true
		}
	}

	if len(positionOverrides) == 0 {
		for index, standing := range sortedPreExtra {
			finalPositions[standing.ID] = index + 1
		}
		return finalPositions
	}

	occupiedPositions := make(map[int]bool, len(positionOverrides))
	for playerID, position := range positionOverrides {
		finalPositions[playerID] = position
		occupiedPositions[position] = true
	}

	remainingPlayers := make([]models.Standing, 0, len(sortedPreExtra))
	for _, standing := range sortedPreExtra {
		if !playoffPlayers[standing.ID] {
			remainingPlayers = append(remainingPlayers, standing)
		}
	}

	remainingIndex := 0
	for position := 1; position <= len(sortedPreExtra); position++ {
		if occupiedPositions[position] {
			continue
		}
		if remainingIndex < len(remainingPlayers) {
			finalPositions[remainingPlayers[remainingIndex].ID] = position
			remainingIndex++
		}
	}

	return finalPositions
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

	// Check if BYE player is needed (look through matches for BYE)
	needsByePlayer := false
	for _, r := range req.Rounds {
		for _, m := range r.Matches {
			if m.Player1Name == "BYE" || m.Player2Name == "BYE" {
				needsByePlayer = true
				break
			}
		}
		if needsByePlayer {
			break
		}
	}

	// Create virtual BYE player if needed
	if needsByePlayer {
		var byeID int
		err := tx.QueryRow(
			"INSERT INTO players (name, confirmed) VALUES ($1, $2) RETURNING id",
			"BYE", false,
		).Scan(&byeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create BYE player"})
			return
		}
		playerMap["BYE"] = byeID
	}

	// Create rounds and matches
	for _, r := range req.Rounds {
		var roundID int
		err := tx.QueryRow(
			"INSERT INTO rounds (round_number, format, is_extra_round) VALUES ($1, $2, $3) RETURNING id",
			r.RoundNumber, r.Format, r.IsExtraRound,
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
				"INSERT INTO matches (round_id, player1_id, player2_id, subformat) VALUES ($1, $2, $3, $4)",
				roundID, player1ID, player2ID, m.Subformat,
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

	standingsRows, err := tx.Query(`
		SELECT id, name, matches_played, wins, ties, losses, points, total_points_scored, total_matches
		FROM standings
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch standings: " + err.Error()})
		return
	}
	defer standingsRows.Close()

	currentStandings := make([]models.Standing, 0)
	for standingsRows.Next() {
		var standing models.Standing
		if err := standingsRows.Scan(
			&standing.ID, &standing.Name, &standing.MatchesPlayed, &standing.Wins, &standing.Ties,
			&standing.Losses, &standing.Points, &standing.TotalPointsScored, &standing.TotalMatches,
		); err != nil {
			continue
		}
		currentStandings = append(currentStandings, standing)
	}

	roundRows, err := tx.Query(`
		SELECT
			r.id,
			r.round_number,
			r.format,
			r.is_extra_round,
			m.id,
			m.player1_id,
			m.player2_id,
			COALESCE(p1.name, 'Unknown'),
			COALESCE(p2.name, 'Unknown'),
			m.score1,
			m.score2,
			m.completed
		FROM rounds r
		LEFT JOIN matches m ON m.round_id = r.id
		LEFT JOIN players p1 ON m.player1_id = p1.id
		LEFT JOIN players p2 ON m.player2_id = p2.id
		ORDER BY r.round_number, m.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds: " + err.Error()})
		return
	}
	defer roundRows.Close()

	roundsMap := make(map[int]*archiveRoundData)
	for roundRows.Next() {
		var roundID, roundNumber int
		var format string
		var isExtraRound bool
		var matchID sql.NullInt64
		var player1ID sql.NullInt64
		var player2ID sql.NullInt64
		var player1Name string
		var player2Name string
		var score1 sql.NullInt64
		var score2 sql.NullInt64
		var completed sql.NullBool

		if err := roundRows.Scan(
			&roundID, &roundNumber, &format, &isExtraRound, &matchID, &player1ID, &player2ID,
			&player1Name, &player2Name, &score1, &score2, &completed,
		); err != nil {
			continue
		}

		if _, exists := roundsMap[roundNumber]; !exists {
			roundsMap[roundNumber] = &archiveRoundData{
				ID:           roundID,
				Number:       roundNumber,
				Format:       format,
				IsExtraRound: isExtraRound,
				Matches:      []archiveMatchData{},
			}
		}

		if !matchID.Valid {
			continue
		}

		roundsMap[roundNumber].Matches = append(roundsMap[roundNumber].Matches, archiveMatchData{
			Player1ID:   player1ID,
			Player2ID:   player2ID,
			Player1Name: player1Name,
			Player2Name: player2Name,
			Score1:      score1,
			Score2:      score2,
			Completed:   completed.Valid && completed.Bool,
		})
	}

	archivedRounds := make([]archiveRoundData, 0, len(roundsMap))
	for i := 1; i <= len(roundsMap); i++ {
		if round, exists := roundsMap[i]; exists {
			archivedRounds = append(archivedRounds, *round)
		}
	}

	finalPositions := calculateArchivedFinalPositions(currentStandings, archivedRounds)

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

	for _, standing := range currentStandings {
		finalPosition := finalPositions[standing.ID]
		if finalPosition == 0 {
			finalPosition = len(currentStandings)
		}

		_, err = tx.Exec(`
			INSERT INTO tournament_standings (
				tournament_id, player_id, player_name, matches_played, wins, ties, losses,
				points, total_points_scored, total_matches, final_position
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, tournamentID, standing.ID, standing.Name, standing.MatchesPlayed, standing.Wins, standing.Ties,
			standing.Losses, standing.Points, standing.TotalPointsScored, standing.TotalMatches, finalPosition)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive standings: " + err.Error()})
			return
		}
	}

	for _, round := range archivedRounds {
		// Create tournament round
		var tournamentRoundID int
		err = tx.QueryRow(`
			INSERT INTO tournament_rounds (tournament_id, round_number, format, is_extra_round)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, tournamentID, round.Number, round.Format, round.IsExtraRound).Scan(&tournamentRoundID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament round: " + err.Error()})
			return
		}

		// Archive matches for this round
		for _, match := range round.Matches {
			_, err = tx.Exec(`
				INSERT INTO tournament_matches (
					tournament_round_id, player1_id, player2_id, player1_name, player2_name,
					score1, score2, completed
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, tournamentRoundID, match.Player1ID.Int64, match.Player2ID.Int64, match.Player1Name, match.Player2Name,
				match.Score1, match.Score2, match.Completed)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive matches: " + err.Error()})
				return
			}
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
		SELECT id, name, month, year, type, format, subformat, start_date, end_date, created_at, archived_at
		FROM tournaments
		WHERE type = 'IN_PERSON'
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
		err := rows.Scan(&t.ID, &t.Name, &t.Month, &t.Year, &t.Type, &t.Format, &t.Subformat, &t.StartDate, &t.EndDate, &t.CreatedAt, &t.ArchivedAt)
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
			ts.points, ts.total_points_scored, ts.total_matches, ts.final_position, tpr.race_pb, tpr.race_bf, tpr.race_libre, tpr.race_edition_vcr
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
			&s.TotalMatches, &s.FinalPosition, &s.RacePB, &s.RaceBF, &s.RaceLibre, &s.RaceEditionVCR,
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
		SELECT id, round_number, format, is_extra_round, subformat
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
		var isExtraRound bool
		var subformat *string
		err := rows.Scan(&roundID, &roundNumber, &format, &isExtraRound, &subformat)
		if err != nil {
			continue
		}

		roundsMap[roundNumber] = &models.TournamentRoundDetail{
			Number:       roundNumber,
			Format:       format,
			IsExtraRound: isExtraRound,
			Subformat:    subformat,
			Matches:      []models.TournamentMatchInfo{},
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

// GetTournamentRaces returns race statistics for a specific tournament
func GetTournamentRaces(c *gin.Context) {
	tournamentID := c.Param("id")

	// Get PB race counts
	pbQuery := `
		SELECT race_pb, COUNT(*) as count
		FROM tournament_player_races
		WHERE tournament_id = $1 AND race_pb IS NOT NULL AND race_pb != ''
		GROUP BY race_pb
		ORDER BY count DESC
	`

	pbRows, err := database.DB.Query(pbQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PB races"})
		return
	}
	defer pbRows.Close()

	pbRaces := make(map[string]int)
	for pbRows.Next() {
		var race string
		var count int
		err := pbRows.Scan(&race, &count)
		if err != nil {
			continue
		}
		pbRaces[race] = count
	}

	// Get BF race counts
	bfQuery := `
		SELECT race_bf, COUNT(*) as count
		FROM tournament_player_races
		WHERE tournament_id = $1 AND race_bf IS NOT NULL AND race_bf != ''
		GROUP BY race_bf
		ORDER BY count DESC
	`

	bfRows, err := database.DB.Query(bfQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch BF races"})
		return
	}
	defer bfRows.Close()

	bfRaces := make(map[string]int)
	for bfRows.Next() {
		var race string
		var count int
		err := bfRows.Scan(&race, &count)
		if err != nil {
			continue
		}
		bfRaces[race] = count
	}

	// Get PB race winrates
	pbWinrateQuery := `
		SELECT tpr.race_pb, COUNT(*) as total_matches, 
		       SUM(CASE 
		             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
		             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
		             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
		             ELSE 0 
		           END) as win_points
		FROM tournament_player_races tpr
		JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
		JOIN tournament_matches m ON m.tournament_round_id = tr.id AND tr.format = 'PB' AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
		WHERE tpr.tournament_id = $1 AND tpr.race_pb IS NOT NULL AND tpr.race_pb != ''
		GROUP BY tpr.race_pb
	`

	pbWinrateRows, err := database.DB.Query(pbWinrateQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PB race winrates"})
		return
	}
	defer pbWinrateRows.Close()

	pbRaceWinrates := make(map[string]float64)
	for pbWinrateRows.Next() {
		var race string
		var totalMatches int
		var winPoints float64
		err := pbWinrateRows.Scan(&race, &totalMatches, &winPoints)
		if err != nil {
			continue
		}
		if totalMatches > 0 {
			pbRaceWinrates[race] = (winPoints * 100.0) / float64(totalMatches)
		} else {
			pbRaceWinrates[race] = 0.0
		}
	}

	// Get BF race winrates
	bfWinrateQuery := `
		SELECT tpr.race_bf, COUNT(*) as total_matches, 
		       SUM(CASE 
		             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
		             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
		             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
		             ELSE 0 
		           END) as win_points
		FROM tournament_player_races tpr
		JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
		JOIN tournament_matches m ON m.tournament_round_id = tr.id AND tr.format = 'BF' AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
		WHERE tpr.tournament_id = $1 AND tpr.race_bf IS NOT NULL AND tpr.race_bf != ''
		GROUP BY tpr.race_bf
	`

	bfWinrateRows, err := database.DB.Query(bfWinrateQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch BF race winrates"})
		return
	}
	defer bfWinrateRows.Close()

	bfRaceWinrates := make(map[string]float64)
	for bfWinrateRows.Next() {
		var race string
		var totalMatches int
		var winPoints float64
		err := bfWinrateRows.Scan(&race, &totalMatches, &winPoints)
		if err != nil {
			continue
		}
		if totalMatches > 0 {
			bfRaceWinrates[race] = (winPoints * 100.0) / float64(totalMatches)
		} else {
			bfRaceWinrates[race] = 0.0
		}
	}

	// Get Libre race counts
	libreQuery := `
		SELECT race_libre, COUNT(*) as count
		FROM tournament_player_races
		WHERE tournament_id = $1 AND race_libre IS NOT NULL AND race_libre != ''
		GROUP BY race_libre
		ORDER BY count DESC
	`

	libreRows, err := database.DB.Query(libreQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Libre races"})
		return
	}
	defer libreRows.Close()

	libreRaces := make(map[string]int)
	for libreRows.Next() {
		var race string
		var count int
		err := libreRows.Scan(&race, &count)
		if err != nil {
			continue
		}
		libreRaces[race] = count
	}

	// Get Libre race winrates
	libreWinrateQuery := `
		SELECT tpr.race_libre, COUNT(*) as total_matches, 
		       SUM(CASE 
		             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
		             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
		             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
		             ELSE 0 
		           END) as win_points
		FROM tournament_player_races tpr
		JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id AND COALESCE(tr.subformat, '') LIKE '%Libre'
		JOIN tournament_matches m ON m.tournament_round_id = tr.id AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
		WHERE tpr.tournament_id = $1 AND tpr.race_libre IS NOT NULL AND tpr.race_libre != ''
		GROUP BY tpr.race_libre
	`

	libreWinrateRows, err := database.DB.Query(libreWinrateQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Libre race winrates"})
		return
	}
	defer libreWinrateRows.Close()

	libreRaceWinrates := make(map[string]float64)
	for libreWinrateRows.Next() {
		var race string
		var totalMatches int
		var winPoints float64
		err := libreWinrateRows.Scan(&race, &totalMatches, &winPoints)
		if err != nil {
			continue
		}
		if totalMatches > 0 {
			libreRaceWinrates[race] = (winPoints * 100.0) / float64(totalMatches)
		} else {
			libreRaceWinrates[race] = 0.0
		}
	}

	// Get Edition VCR race counts
	vcrQuery := `
		SELECT race_edition_vcr, COUNT(*) as count
		FROM tournament_player_races
		WHERE tournament_id = $1 AND race_edition_vcr IS NOT NULL AND race_edition_vcr != ''
		GROUP BY race_edition_vcr
		ORDER BY count DESC
	`

	vcrRows, err := database.DB.Query(vcrQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Edition VCR races"})
		return
	}
	defer vcrRows.Close()

	vcrRaces := make(map[string]int)
	for vcrRows.Next() {
		var race string
		var count int
		err := vcrRows.Scan(&race, &count)
		if err != nil {
			continue
		}
		vcrRaces[race] = count
	}

	// Get VCR race winrates
	vcrWinrateQuery := `
		SELECT tpr.race_edition_vcr, COUNT(*) as total_matches, 
		       SUM(CASE 
		             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
		             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
		             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
		             ELSE 0 
		           END) as win_points
		FROM tournament_player_races tpr
			JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
				AND LOWER(COALESCE(tr.subformat, '')) IN ('pbre', 'bfvcr', 'vcr', 'edition', 'edición', 'edicion')
		JOIN tournament_matches m ON m.tournament_round_id = tr.id AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
		WHERE tpr.tournament_id = $1 AND tpr.race_edition_vcr IS NOT NULL AND tpr.race_edition_vcr != ''
		GROUP BY tpr.race_edition_vcr
	`

	vcrWinrateRows, err := database.DB.Query(vcrWinrateQuery, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch VCR race winrates"})
		return
	}
	defer vcrWinrateRows.Close()

	vcrRaceWinrates := make(map[string]float64)
	for vcrWinrateRows.Next() {
		var race string
		var totalMatches int
		var winPoints float64
		err := vcrWinrateRows.Scan(&race, &totalMatches, &winPoints)
		if err != nil {
			continue
		}
		if totalMatches > 0 {
			vcrRaceWinrates[race] = (winPoints * 100.0) / float64(totalMatches)
		} else {
			vcrRaceWinrates[race] = 0.0
		}
	}

	response := gin.H{
		"pb_races":            pbRaces,
		"bf_races":            bfRaces,
		"libre_races":         libreRaces,
		"vcr_races":           vcrRaces,
		"pb_race_winrates":    pbRaceWinrates,
		"bf_race_winrates":    bfRaceWinrates,
		"libre_race_winrates": libreRaceWinrates,
		"vcr_race_winrates":   vcrRaceWinrates,
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
			tpr.race_libre,
			tpr.race_edition_vcr,
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
			&pr.RaceLibre,
			&pr.RaceEditionVCR,
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
			SET player_name = $1, race_pb = $2, race_bf = $3, race_libre = $4, race_edition_vcr = $5, notes = $6, updated_at = CURRENT_TIMESTAMP
			WHERE tournament_id = $7 AND player_id = $8
		`
		args = []interface{}{req.PlayerName, req.RacePB, req.RaceBF, req.RaceLibre, req.RaceEditionVCR, req.Notes, tournamentID, playerID}
	} else {
		// Insert new record
		query = `
			INSERT INTO tournament_player_races (tournament_id, player_id, player_name, race_pb, race_bf, race_libre, race_edition_vcr, notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		args = []interface{}{tournamentID, playerID, req.PlayerName, req.RacePB, req.RaceBF, req.RaceLibre, req.RaceEditionVCR, req.Notes}
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

// GetPremierPlayers returns all players from the premier_players table
func GetPremierPlayers(c *gin.Context) {
	query := `SELECT id, name FROM premier_players ORDER BY name`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}
	defer rows.Close()

	type SimplifiedPlayer struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	players := []SimplifiedPlayer{}
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			continue
		}
		players = append(players, SimplifiedPlayer{ID: id, Name: name})
	}

	if players == nil {
		players = []SimplifiedPlayer{}
	}

	c.JSON(http.StatusOK, players)
}

// GetGlobalStandings returns aggregated standings from all archived tournaments
func GetGlobalStandings(c *gin.Context) {
	// Get all players with medal counts and most-played races from all tournaments
	query := `
		SELECT 
			pp.id,
			pp.name,
			COALESCE(SUM(CASE WHEN ts.final_position = 1 THEN 1 ELSE 0 END), 0) as first_place_count,
			COALESCE(SUM(CASE WHEN ts.final_position = 2 THEN 1 ELSE 0 END), 0) as second_place_count,
			COALESCE(SUM(CASE WHEN ts.final_position = 3 THEN 1 ELSE 0 END), 0) as third_place_count,
			-- Most played race PB (includes race_pb from mixed tournaments + race_libre/race_edition_vcr from PB-only tournaments)
			(
				SELECT race FROM (
					SELECT race_pb as race
					FROM tournament_player_races tpr2
					JOIN tournaments t2 ON t2.id = tpr2.tournament_id
					WHERE tpr2.player_name = pp.name AND tpr2.race_pb IS NOT NULL AND tpr2.race_pb != '' AND t2.format IS NULL
					UNION ALL
					SELECT race_libre as race
					FROM tournament_player_races tpr2
					JOIN tournaments t2 ON t2.id = tpr2.tournament_id
					WHERE tpr2.player_name = pp.name AND tpr2.race_libre IS NOT NULL AND tpr2.race_libre != '' AND t2.format = 'PB'
					UNION ALL
					SELECT race_edition_vcr as race
					FROM tournament_player_races tpr2
					JOIN tournaments t2 ON t2.id = tpr2.tournament_id
					WHERE tpr2.player_name = pp.name AND tpr2.race_edition_vcr IS NOT NULL AND tpr2.race_edition_vcr != '' AND t2.format = 'PB'
				) pb_all
				GROUP BY race
				ORDER BY COUNT(*) DESC
				LIMIT 1
			) as most_played_race_pb,
			-- Most played race BF (includes race_bf from mixed tournaments + race_libre/race_edition_vcr from BF-only tournaments)
			(
				SELECT race FROM (
					SELECT race_bf as race
					FROM tournament_player_races tpr3
					JOIN tournaments t3 ON t3.id = tpr3.tournament_id
					WHERE tpr3.player_name = pp.name AND tpr3.race_bf IS NOT NULL AND tpr3.race_bf != '' AND t3.format IS NULL
					UNION ALL
					SELECT race_libre as race
					FROM tournament_player_races tpr3
					JOIN tournaments t3 ON t3.id = tpr3.tournament_id
					WHERE tpr3.player_name = pp.name AND tpr3.race_libre IS NOT NULL AND tpr3.race_libre != '' AND t3.format = 'BF'
					UNION ALL
					SELECT race_edition_vcr as race
					FROM tournament_player_races tpr3
					JOIN tournaments t3 ON t3.id = tpr3.tournament_id
					WHERE tpr3.player_name = pp.name AND tpr3.race_edition_vcr IS NOT NULL AND tpr3.race_edition_vcr != '' AND t3.format = 'BF'
				) bf_all
				GROUP BY race
				ORDER BY COUNT(*) DESC
				LIMIT 1
			) as most_played_race_bf,
			-- PB stats aggregated from all tournaments for this player
			(
				SELECT COALESCE(SUM(CASE 
					WHEN tm.completed = true AND (
						(tm.player1_id = ts_inner.player_id AND tm.score1 > tm.score2) OR
						(tm.player2_id = ts_inner.player_id AND tm.score2 > tm.score1)
					) THEN 1 ELSE 0 
				END), 0)
				FROM tournament_standings ts_inner
				JOIN tournament_rounds tr2 ON tr2.tournament_id = ts_inner.tournament_id
				LEFT JOIN tournament_matches tm ON tr2.id = tm.tournament_round_id
				WHERE ts_inner.player_name = pp.name AND tr2.format = 'PB' AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id)
			) as pb_wins,
			(
				SELECT COALESCE(SUM(CASE 
					WHEN tm.completed = true AND tm.score1 = tm.score2 AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id) THEN 1 ELSE 0 
				END), 0)
				FROM tournament_standings ts_inner
				JOIN tournament_rounds tr2 ON tr2.tournament_id = ts_inner.tournament_id
				LEFT JOIN tournament_matches tm ON tr2.id = tm.tournament_round_id
				WHERE ts_inner.player_name = pp.name AND tr2.format = 'PB' AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id)
			) as pb_ties,
			(
				SELECT COALESCE(SUM(CASE 
					WHEN tm.completed = true AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id) THEN 1 ELSE 0 
				END), 0)
				FROM tournament_standings ts_inner
				JOIN tournament_rounds tr2 ON tr2.tournament_id = ts_inner.tournament_id
				LEFT JOIN tournament_matches tm ON tr2.id = tm.tournament_round_id
				WHERE ts_inner.player_name = pp.name AND tr2.format = 'PB' AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id)
			) as pb_matches,
			-- BF stats aggregated from all tournaments for this player
			(
				SELECT COALESCE(SUM(CASE 
					WHEN tm.completed = true AND (
						(tm.player1_id = ts_inner.player_id AND tm.score1 > tm.score2) OR
						(tm.player2_id = ts_inner.player_id AND tm.score2 > tm.score1)
					) THEN 1 ELSE 0 
				END), 0)
				FROM tournament_standings ts_inner
				JOIN tournament_rounds tr2 ON tr2.tournament_id = ts_inner.tournament_id
				LEFT JOIN tournament_matches tm ON tr2.id = tm.tournament_round_id
				WHERE ts_inner.player_name = pp.name AND tr2.format = 'BF' AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id)
			) as bf_wins,
			(
				SELECT COALESCE(SUM(CASE 
					WHEN tm.completed = true AND tm.score1 = tm.score2 AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id) THEN 1 ELSE 0 
				END), 0)
				FROM tournament_standings ts_inner
				JOIN tournament_rounds tr2 ON tr2.tournament_id = ts_inner.tournament_id
				LEFT JOIN tournament_matches tm ON tr2.id = tm.tournament_round_id
				WHERE ts_inner.player_name = pp.name AND tr2.format = 'BF' AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id)
			) as bf_ties,
			(
				SELECT COALESCE(SUM(CASE 
					WHEN tm.completed = true AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id) THEN 1 ELSE 0 
				END), 0)
				FROM tournament_standings ts_inner
				JOIN tournament_rounds tr2 ON tr2.tournament_id = ts_inner.tournament_id
				LEFT JOIN tournament_matches tm ON tr2.id = tm.tournament_round_id
				WHERE ts_inner.player_name = pp.name AND tr2.format = 'BF' AND (tm.player1_id = ts_inner.player_id OR tm.player2_id = ts_inner.player_id)
			) as bf_matches
		FROM premier_players pp
		LEFT JOIN tournament_standings ts ON ts.player_name = pp.name
		GROUP BY pp.id, pp.name
		ORDER BY first_place_count DESC, second_place_count DESC, third_place_count DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch global standings"})
		return
	}
	defer rows.Close()

	type GlobalStanding struct {
		PlayerID         int     `json:"player_id"`
		PlayerName       string  `json:"player_name"`
		FirstPlaceCount  int     `json:"first_place_count"`
		SecondPlaceCount int     `json:"second_place_count"`
		ThirdPlaceCount  int     `json:"third_place_count"`
		MostPlayedRacePB *string `json:"most_played_race_pb"`
		MostPlayedRaceBF *string `json:"most_played_race_bf"`
		WinratePB        float64 `json:"winrate_pb"`
		WinateBF         float64 `json:"winrate_bf"`
	}

	standings := []GlobalStanding{}
	for rows.Next() {
		var s GlobalStanding
		var pbWins int
		var pbTies int
		var pbMatches int
		var bfWins int
		var bfTies int
		var bfMatches int

		err := rows.Scan(
			&s.PlayerID,
			&s.PlayerName,
			&s.FirstPlaceCount,
			&s.SecondPlaceCount,
			&s.ThirdPlaceCount,
			&s.MostPlayedRacePB,
			&s.MostPlayedRaceBF,
			&pbWins,
			&pbTies,
			&pbMatches,
			&bfWins,
			&bfTies,
			&bfMatches,
		)
		if err != nil {
			continue
		}

		// Calculate PB winrate: (wins + 0.5*ties) / total_matches * 100
		if pbMatches > 0 {
			s.WinratePB = ((float64(pbWins) + 0.5*float64(pbTies)) / float64(pbMatches)) * 100.0
		}

		// Calculate BF winrate: (wins + 0.5*ties) / total_matches * 100
		if bfMatches > 0 {
			s.WinateBF = ((float64(bfWins) + 0.5*float64(bfTies)) / float64(bfMatches)) * 100.0
		}

		standings = append(standings, s)
	}

	if standings == nil {
		standings = []GlobalStanding{}
	}

	c.JSON(http.StatusOK, standings)
}

// GetGlobalRaces returns aggregated race statistics from all archived tournaments
func GetGlobalRaces(c *gin.Context) {
	// Get PB race counts from all tournaments (race_pb from mixed + race_libre/race_edition_vcr from PB-only)
	pbQuery := `
		SELECT race, SUM(cnt) as count FROM (
			SELECT race_pb as race, COUNT(*) as cnt
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			WHERE tpr.race_pb IS NOT NULL AND tpr.race_pb != '' AND t.format IS NULL
			GROUP BY race_pb
			UNION ALL
			SELECT race_libre as race, COUNT(*) as cnt
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			WHERE tpr.race_libre IS NOT NULL AND tpr.race_libre != '' AND t.format = 'PB'
			GROUP BY race_libre
			UNION ALL
			SELECT race_edition_vcr as race, COUNT(*) as cnt
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			WHERE tpr.race_edition_vcr IS NOT NULL AND tpr.race_edition_vcr != '' AND t.format = 'PB'
			GROUP BY race_edition_vcr
		) pb_all
		GROUP BY race
		ORDER BY count DESC
	`

	pbRows, err := database.DB.Query(pbQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch global PB races"})
		return
	}
	defer pbRows.Close()

	pbRaces := make(map[string]int)
	for pbRows.Next() {
		var race string
		var count int
		err := pbRows.Scan(&race, &count)
		if err != nil {
			continue
		}
		pbRaces[race] = count
	}

	// Get BF race counts from all tournaments (race_bf from mixed + race_libre/race_edition_vcr from BF-only)
	bfQuery := `
		SELECT race, SUM(cnt) as count FROM (
			SELECT race_bf as race, COUNT(*) as cnt
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			WHERE tpr.race_bf IS NOT NULL AND tpr.race_bf != '' AND t.format IS NULL
			GROUP BY race_bf
			UNION ALL
			SELECT race_libre as race, COUNT(*) as cnt
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			WHERE tpr.race_libre IS NOT NULL AND tpr.race_libre != '' AND t.format = 'BF'
			GROUP BY race_libre
			UNION ALL
			SELECT race_edition_vcr as race, COUNT(*) as cnt
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			WHERE tpr.race_edition_vcr IS NOT NULL AND tpr.race_edition_vcr != '' AND t.format = 'BF'
			GROUP BY race_edition_vcr
		) bf_all
		GROUP BY race
		ORDER BY count DESC
	`

	bfRows, err := database.DB.Query(bfQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch global BF races"})
		return
	}
	defer bfRows.Close()

	bfRaces := make(map[string]int)
	for bfRows.Next() {
		var race string
		var count int
		err := bfRows.Scan(&race, &count)
		if err != nil {
			continue
		}
		bfRaces[race] = count
	}

	// Get PB race winrates from all tournaments (race_pb from mixed + race_libre/race_edition_vcr from PB-only)
	pbWinrateQuery := `
		SELECT race, SUM(total_matches) as total_matches, SUM(win_points) as win_points FROM (
			SELECT tpr.race_pb as race, COUNT(*) as total_matches, 
			       SUM(CASE 
			             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
			             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
			             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
			             ELSE 0 
			           END) as win_points
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
			JOIN tournament_matches m ON m.tournament_round_id = tr.id AND tr.format = 'PB' AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
			WHERE tpr.race_pb IS NOT NULL AND tpr.race_pb != '' AND t.format IS NULL
			GROUP BY tpr.race_pb
			UNION ALL
			SELECT tpr.race_libre as race, COUNT(*) as total_matches, 
			       SUM(CASE 
			             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
			             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
			             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
			             ELSE 0 
			           END) as win_points
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
			JOIN tournament_matches m ON m.tournament_round_id = tr.id AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
			WHERE tpr.race_libre IS NOT NULL AND tpr.race_libre != '' AND t.format = 'PB'
				AND tr.subformat IN ('libre', 'pbrl', 'bfrl', 'Libre', 'PBRL', 'BFRL')
			GROUP BY tpr.race_libre
			UNION ALL
			SELECT tpr.race_edition_vcr as race, COUNT(*) as total_matches, 
			       SUM(CASE 
			             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
			             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
			             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
			             ELSE 0 
			           END) as win_points
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
			JOIN tournament_matches m ON m.tournament_round_id = tr.id AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
			WHERE tpr.race_edition_vcr IS NOT NULL AND tpr.race_edition_vcr != '' AND t.format = 'PB'
				AND tr.subformat IN ('pbre', 'bfvcr', 'vcr', 'edition', 'edición', 'edicion', 'PBRE', 'BFVCR', 'VCR', 'Edition', 'Edición', 'Edicion')
			GROUP BY tpr.race_edition_vcr
		) pb_all
		GROUP BY race
	`

	pbWinrateRows, err := database.DB.Query(pbWinrateQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch global PB race winrates"})
		return
	}
	defer pbWinrateRows.Close()

	pbRaceWinrates := make(map[string]float64)
	for pbWinrateRows.Next() {
		var race string
		var totalMatches int
		var winPoints float64
		err := pbWinrateRows.Scan(&race, &totalMatches, &winPoints)
		if err != nil {
			continue
		}
		if totalMatches > 0 {
			pbRaceWinrates[race] = (winPoints * 100.0) / float64(totalMatches)
		} else {
			pbRaceWinrates[race] = 0.0
		}
	}

	// Get BF race winrates from all tournaments (race_bf from mixed + race_libre/race_edition_vcr from BF-only)
	bfWinrateQuery := `
		SELECT race, SUM(total_matches) as total_matches, SUM(win_points) as win_points FROM (
			SELECT tpr.race_bf as race, COUNT(*) as total_matches, 
			       SUM(CASE 
			             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
			             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
			             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
			             ELSE 0 
			           END) as win_points
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
			JOIN tournament_matches m ON m.tournament_round_id = tr.id AND tr.format = 'BF' AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
			WHERE tpr.race_bf IS NOT NULL AND tpr.race_bf != '' AND t.format IS NULL
			GROUP BY tpr.race_bf
			UNION ALL
			SELECT tpr.race_libre as race, COUNT(*) as total_matches, 
			       SUM(CASE 
			             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
			             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
			             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
			             ELSE 0 
			           END) as win_points
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
			JOIN tournament_matches m ON m.tournament_round_id = tr.id AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
			WHERE tpr.race_libre IS NOT NULL AND tpr.race_libre != '' AND t.format = 'BF'
				AND tr.subformat IN ('libre', 'pbrl', 'bfrl', 'Libre', 'PBRL', 'BFRL')
			GROUP BY tpr.race_libre
			UNION ALL
			SELECT tpr.race_edition_vcr as race, COUNT(*) as total_matches, 
			       SUM(CASE 
			             WHEN m.player1_id = tpr.player_id AND m.score1 > m.score2 THEN 1
			             WHEN m.player2_id = tpr.player_id AND m.score2 > m.score1 THEN 1
			             WHEN m.score1 IS NOT NULL AND m.score2 IS NOT NULL AND m.score1 = m.score2 AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id) THEN 0.5
			             ELSE 0 
			           END) as win_points
			FROM tournament_player_races tpr
			JOIN tournaments t ON t.id = tpr.tournament_id
			JOIN tournament_rounds tr ON tr.tournament_id = tpr.tournament_id
			JOIN tournament_matches m ON m.tournament_round_id = tr.id AND (m.player1_id = tpr.player_id OR m.player2_id = tpr.player_id)
			WHERE tpr.race_edition_vcr IS NOT NULL AND tpr.race_edition_vcr != '' AND t.format = 'BF'
				AND tr.subformat IN ('pbre', 'bfvcr', 'vcr', 'edition', 'edición', 'edicion', 'PBRE', 'BFVCR', 'VCR', 'Edition', 'Edición', 'Edicion')
			GROUP BY tpr.race_edition_vcr
		) bf_all
		GROUP BY race
	`

	bfWinrateRows, err := database.DB.Query(bfWinrateQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch global BF race winrates"})
		return
	}
	defer bfWinrateRows.Close()

	bfRaceWinrates := make(map[string]float64)
	for bfWinrateRows.Next() {
		var race string
		var totalMatches int
		var winPoints float64
		err := bfWinrateRows.Scan(&race, &totalMatches, &winPoints)
		if err != nil {
			continue
		}
		if totalMatches > 0 {
			bfRaceWinrates[race] = (winPoints * 100.0) / float64(totalMatches)
		} else {
			bfRaceWinrates[race] = 0.0
		}
	}

	response := gin.H{
		"pb_races":         pbRaces,
		"bf_races":         bfRaces,
		"pb_race_winrates": pbRaceWinrates,
		"bf_race_winrates": bfRaceWinrates,
	}

	c.JSON(http.StatusOK, response)
}

// AddMatchesToRound adds matches to an existing round (used for the extra/final round)
func AddMatchesToRound(c *gin.Context) {
	roundNumberStr := c.Param("round_number")
	roundNumber, err := strconv.Atoi(roundNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid round number"})
		return
	}

	var req struct {
		Matches []struct {
			Player1Name string `json:"player1_name" binding:"required"`
			Player2Name string `json:"player2_name" binding:"required"`
		} `json:"matches" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Matches) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one match is required"})
		return
	}

	// Find the round
	var roundID int
	var isExtraRound bool
	err = database.DB.QueryRow(
		"SELECT id, is_extra_round FROM rounds WHERE round_number = $1", roundNumber,
	).Scan(&roundID, &isExtraRound)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Round not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find round"})
		return
	}

	if !isExtraRound {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only the extra round can be edited here"})
		return
	}

	// Resolve player names to IDs
	type playerPair struct {
		p1ID int
		p2ID int
	}
	pairs := make([]playerPair, 0, len(req.Matches))

	for _, m := range req.Matches {
		var p1ID, p2ID int
		err = database.DB.QueryRow(
			"SELECT id FROM players WHERE name = $1", m.Player1Name,
		).Scan(&p1ID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Player not found: " + m.Player1Name})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find player: " + m.Player1Name})
			return
		}

		err = database.DB.QueryRow(
			"SELECT id FROM players WHERE name = $1", m.Player2Name,
		).Scan(&p2ID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Player not found: " + m.Player2Name})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find player: " + m.Player2Name})
			return
		}

		if p1ID == p2ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A player cannot play against themselves: " + m.Player1Name})
			return
		}

		pairs = append(pairs, playerPair{p1ID: p1ID, p2ID: p2ID})
	}

	// Insert matches
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM matches WHERE round_id = $1", roundID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing matches"})
		return
	}

	for _, pair := range pairs {
		_, err := tx.Exec(
			"INSERT INTO matches (round_id, player1_id, player2_id) VALUES ($1, $2, $3)",
			roundID, pair.p1ID, pair.p2ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Matches added successfully",
		"round_number":    roundNumber,
		"matches_created": len(pairs),
	})
}

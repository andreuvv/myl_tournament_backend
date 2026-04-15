package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/andreuvv/premier_mitologico/backend/internal/database"
	"github.com/gin-gonic/gin"
)

// PlayerTournamentHistory represents a player's participation in a tournament
type PlayerTournamentHistory struct {
	TournamentID      int     `json:"tournament_id"`
	TournamentName    string  `json:"tournament_name"`
	Month             string  `json:"month"`
	Year              int     `json:"year"`
	Format            *string `json:"format"`
	Subformat         *string `json:"subformat"`
	FinalPosition     int     `json:"final_position"`
	MatchesPlayed     int     `json:"matches_played"`
	Wins              int     `json:"wins"`
	Ties              int     `json:"ties"`
	Losses            int     `json:"losses"`
	Points            int     `json:"points"`
	TotalPointsScored int     `json:"total_points_scored"`
	RacePB            *string `json:"race_pb"`
	RaceBF            *string `json:"race_bf"`
	RaceLibre         *string `json:"race_libre"`
	RaceEditionVCR    *string `json:"race_edition_vcr"`
	// Win data by format from actual match results
	PBWins    int `json:"pb_wins"`
	PBTies    int `json:"pb_ties"`
	PBMatches int `json:"pb_matches"`
	BFWins    int `json:"bf_wins"`
	BFTies    int `json:"bf_ties"`
	BFMatches int `json:"bf_matches"`
}

// GetPlayerTournamentHistory returns all tournament history for a specific player
func GetPlayerTournamentHistory(c *gin.Context) {
	// Accept either player_id or player_name parameter
	playerIDStr := c.Param("player_id")
	playerName := c.Query("name")

	fmt.Println("GetPlayerTournamentHistory called with player_id:", playerIDStr, "player_name:", playerName)

	// If player_name is provided as query param, use it directly
	if playerName != "" {
		fetchPlayerTournamentHistory(c, playerName)
		return
	}

	// Otherwise, try to get player name from player ID
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		fmt.Println("Error converting player_id to int:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	// Try to get the player name from the active players table first
	var pName string
	err = database.DB.QueryRow("SELECT name FROM players WHERE id = $1", playerID).Scan(&pName)
	if err != nil {
		// If not found in active players, try premier_players table
		err = database.DB.QueryRow("SELECT name FROM premier_players WHERE id = $1", playerID).Scan(&pName)
		if err != nil {
			fmt.Println("Error getting player name from either table:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
	}

	fetchPlayerTournamentHistory(c, pName)
}

// Helper function to fetch tournament history by player name
func fetchPlayerTournamentHistory(c *gin.Context, playerName string) {

	// Get all tournaments where the player participated, including match data by format
	// Use player_name from tournament_standings instead of player_id for matching
	// since player_id may be orphaned after removing foreign key constraints
	query := `
		SELECT 
			t.id,
			t.name,
			t.month,
			t.year,
			t.format,
			t.subformat,
			ts.final_position,
			ts.matches_played,
			ts.wins,
			ts.ties,
			ts.losses,
			ts.points,
			ts.total_points_scored,
			tpr.race_pb,
			tpr.race_bf,
			tpr.race_libre,
			tpr.race_edition_vcr,
			-- PB format statistics
			COALESCE(SUM(CASE 
				WHEN tr.format = 'PB' AND tm.completed = true AND (
					(tm.player1_id = ts.player_id AND tm.score1 > tm.score2) OR
					(tm.player2_id = ts.player_id AND tm.score2 > tm.score1)
				) THEN 1 ELSE 0 
			END), 0) as pb_wins,
			COALESCE(SUM(CASE 
				WHEN tr.format = 'PB' AND tm.completed = true AND tm.score1 = tm.score2 AND (tm.player1_id = ts.player_id OR tm.player2_id = ts.player_id) THEN 1 ELSE 0 
			END), 0) as pb_ties,
			COALESCE(SUM(CASE 
				WHEN tr.format = 'PB' AND tm.completed = true AND (tm.player1_id = ts.player_id OR tm.player2_id = ts.player_id) THEN 1 ELSE 0 
			END), 0) as pb_matches,
			-- BF format statistics
			COALESCE(SUM(CASE 
				WHEN tr.format = 'BF' AND tm.completed = true AND (
					(tm.player1_id = ts.player_id AND tm.score1 > tm.score2) OR
					(tm.player2_id = ts.player_id AND tm.score2 > tm.score1)
				) THEN 1 ELSE 0 
			END), 0) as bf_wins,
			COALESCE(SUM(CASE 
				WHEN tr.format = 'BF' AND tm.completed = true AND tm.score1 = tm.score2 AND (tm.player1_id = ts.player_id OR tm.player2_id = ts.player_id) THEN 1 ELSE 0 
			END), 0) as bf_ties,
			COALESCE(SUM(CASE 
				WHEN tr.format = 'BF' AND tm.completed = true AND (tm.player1_id = ts.player_id OR tm.player2_id = ts.player_id) THEN 1 ELSE 0 
			END), 0) as bf_matches
		FROM tournaments t
		INNER JOIN tournament_standings ts ON t.id = ts.tournament_id AND ts.player_name = $1
		LEFT JOIN tournament_player_races tpr ON t.id = tpr.tournament_id AND tpr.player_name = ts.player_name
		LEFT JOIN tournament_rounds tr ON t.id = tr.tournament_id
		LEFT JOIN tournament_matches tm ON tr.id = tm.tournament_round_id
		GROUP BY t.id, t.name, t.month, t.year, t.format, t.subformat, ts.id, tpr.race_pb, tpr.race_bf, tpr.race_libre, tpr.race_edition_vcr
		ORDER BY t.year DESC, t.month DESC
	`

	rows, err := database.DB.Query(query, playerName)
	if err != nil {
		fmt.Println("Database query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player tournament history"})
		return
	}
	defer rows.Close()

	var history []PlayerTournamentHistory

	for rows.Next() {
		var h PlayerTournamentHistory
		var racePB sql.NullString
		var raceBF sql.NullString
		var raceLibre sql.NullString
		var raceEditionVCR sql.NullString
		var format sql.NullString
		var subformat sql.NullString

		err := rows.Scan(
			&h.TournamentID,
			&h.TournamentName,
			&h.Month,
			&h.Year,
			&format,
			&subformat,
			&h.FinalPosition,
			&h.MatchesPlayed,
			&h.Wins,
			&h.Ties,
			&h.Losses,
			&h.Points,
			&h.TotalPointsScored,
			&racePB,
			&raceBF,
			&raceLibre,
			&raceEditionVCR,
			&h.PBWins,
			&h.PBTies,
			&h.PBMatches,
			&h.BFWins,
			&h.BFTies,
			&h.BFMatches,
		)
		if err != nil {
			continue
		}

		// Handle nullable race fields
		if racePB.Valid {
			h.RacePB = &racePB.String
		}
		if raceBF.Valid {
			h.RaceBF = &raceBF.String
		}
		if raceLibre.Valid {
			h.RaceLibre = &raceLibre.String
		}
		if raceEditionVCR.Valid {
			h.RaceEditionVCR = &raceEditionVCR.String
		}
		if format.Valid {
			h.Format = &format.String
		}
		if subformat.Valid {
			h.Subformat = &subformat.String
		}

		history = append(history, h)
	}

	if history == nil {
		history = []PlayerTournamentHistory{}
	}

	c.JSON(http.StatusOK, history)
}

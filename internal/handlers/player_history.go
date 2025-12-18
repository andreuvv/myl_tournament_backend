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
	FinalPosition     int     `json:"final_position"`
	MatchesPlayed     int     `json:"matches_played"`
	Wins              int     `json:"wins"`
	Ties              int     `json:"ties"`
	Losses            int     `json:"losses"`
	Points            int     `json:"points"`
	TotalPointsScored int     `json:"total_points_scored"`
	RacePB            *string `json:"race_pb"`
	RaceBF            *string `json:"race_bf"`
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
	playerIDStr := c.Param("player_id")
	fmt.Println("GetPlayerTournamentHistory called with player_id:", playerIDStr)

	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		fmt.Println("Error converting player_id to int:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	// Get all tournaments where the player participated, including match data by format
	query := `
		SELECT 
			t.id,
			t.name,
			t.month,
			t.year,
			ts.final_position,
			ts.matches_played,
			ts.wins,
			ts.ties,
			ts.losses,
			ts.points,
			ts.total_points_scored,
			tpr.race_pb,
			tpr.race_bf,
			-- PB format statistics
			COALESCE(SUM(CASE 
				WHEN tr.format = 'PB' AND tm.completed = true AND (
					(tm.player1_id = $1 AND tm.score1 > tm.score2) OR
					(tm.player2_id = $1 AND tm.score2 > tm.score1)
				) THEN 1 ELSE 0 
			END), 0) as pb_wins,
			COALESCE(SUM(CASE 
				WHEN tr.format = 'PB' AND tm.completed = true AND tm.score1 = tm.score2 AND (tm.player1_id = $1 OR tm.player2_id = $1) THEN 1 ELSE 0 
			END), 0) as pb_ties,
			COALESCE(SUM(CASE 
				WHEN tr.format = 'PB' AND tm.completed = true AND (tm.player1_id = $1 OR tm.player2_id = $1) THEN 1 ELSE 0 
			END), 0) as pb_matches,
			-- BF format statistics
			COALESCE(SUM(CASE 
				WHEN tr.format = 'BF' AND tm.completed = true AND (
					(tm.player1_id = $1 AND tm.score1 > tm.score2) OR
					(tm.player2_id = $1 AND tm.score2 > tm.score1)
				) THEN 1 ELSE 0 
			END), 0) as bf_wins,
			COALESCE(SUM(CASE 
				WHEN tr.format = 'BF' AND tm.completed = true AND tm.score1 = tm.score2 AND (tm.player1_id = $1 OR tm.player2_id = $1) THEN 1 ELSE 0 
			END), 0) as bf_ties,
			COALESCE(SUM(CASE 
				WHEN tr.format = 'BF' AND tm.completed = true AND (tm.player1_id = $1 OR tm.player2_id = $1) THEN 1 ELSE 0 
			END), 0) as bf_matches
		FROM tournaments t
		INNER JOIN tournament_standings ts ON t.id = ts.tournament_id AND ts.player_id = $1
		LEFT JOIN tournament_player_races tpr ON t.id = tpr.tournament_id AND ts.player_id = tpr.player_id
		LEFT JOIN tournament_rounds tr ON t.id = tr.tournament_id
		LEFT JOIN tournament_matches tm ON tr.id = tm.tournament_round_id
		GROUP BY t.id, t.name, t.month, t.year, ts.id, tpr.race_pb, tpr.race_bf
		ORDER BY t.year DESC, t.month DESC
	`

	rows, err := database.DB.Query(query, playerID)
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

		err := rows.Scan(
			&h.TournamentID,
			&h.TournamentName,
			&h.Month,
			&h.Year,
			&h.FinalPosition,
			&h.MatchesPlayed,
			&h.Wins,
			&h.Ties,
			&h.Losses,
			&h.Points,
			&h.TotalPointsScored,
			&racePB,
			&raceBF,
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

		history = append(history, h)
	}

	if history == nil {
		history = []PlayerTournamentHistory{}
	}

	c.JSON(http.StatusOK, history)
}

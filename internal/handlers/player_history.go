package handlers

import (
	"database/sql"
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
}

// GetPlayerTournamentHistory returns all tournament history for a specific player
func GetPlayerTournamentHistory(c *gin.Context) {
	playerIDStr := c.Param("player_id")
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	// Get all tournaments where the player participated along with standings and races
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
			tpr.race_bf
		FROM tournaments t
		INNER JOIN tournament_standings ts ON t.id = ts.tournament_id
		LEFT JOIN tournament_player_races tpr ON t.id = tpr.tournament_id AND ts.player_id = tpr.player_id
		WHERE ts.player_id = $1
		ORDER BY t.year DESC, t.month DESC
	`

	rows, err := database.DB.Query(query, playerID)
	if err != nil {
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

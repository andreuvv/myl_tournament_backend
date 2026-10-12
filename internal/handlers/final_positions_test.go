package handlers

import (
	"database/sql"
	"testing"

	"github.com/andreuvv/premier_mitologico/backend/internal/models"
)

func nullInt(v int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(v), Valid: true}
}

func TestJuneFinalKeepsThirdPlaceWinnerInThird(t *testing.T) {
	// Folo has the lower id, so an id tiebreak would seed him above Wesh.
	// Wesh beat Folo in the Swiss, and the final was Traukolin vs Wesh.
	// Folo beat Piter in the third-place match, so Folo stays 3rd.
	const (
		foloID      = 1
		weshID      = 2
		traukolinID = 10
		piterID     = 11
	)

	standings := []models.Standing{
		{ID: traukolinID, Name: "Traukolin", Points: 16, Wins: 5, TotalPointsScored: 9},
		{ID: weshID, Name: "Wesh", Points: 14, Wins: 4, TotalPointsScored: 10},
		{ID: foloID, Name: "Folo", Points: 14, Wins: 4, TotalPointsScored: 10},
		{ID: piterID, Name: "Piter", Points: 10, Wins: 3, TotalPointsScored: 8},
	}
	rounds := []archiveRoundData{
		{
			Number: 1,
			Matches: []archiveMatchData{{
				Player1ID: nullInt(foloID),
				Player2ID: nullInt(weshID),
				Score1:    nullInt(0),
				Score2:    nullInt(2),
				Completed: true,
			}},
		},
		{
			Number:       7,
			IsExtraRound: true,
			Matches: []archiveMatchData{
				{
					Player1ID: nullInt(traukolinID),
					Player2ID: nullInt(weshID),
					Score1:    nullInt(1),
					Score2:    nullInt(2),
					Completed: true,
				},
				{
					Player1ID: nullInt(foloID),
					Player2ID: nullInt(piterID),
					Score1:    nullInt(2),
					Score2:    nullInt(1),
					Completed: true,
				},
			},
		},
	}

	positions := calculateArchivedFinalPositions(standings, rounds)
	if positions[weshID] != 1 || positions[traukolinID] != 2 || positions[foloID] != 3 || positions[piterID] != 4 {
		t.Fatalf("final positions = Wesh %d, Traukolin %d, Folo %d, Piter %d; want 1, 2, 3, 4",
			positions[weshID], positions[traukolinID], positions[foloID], positions[piterID])
	}
}

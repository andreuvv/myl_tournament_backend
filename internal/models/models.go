package models

import "time"

type Player struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Confirmed bool      `json:"confirmed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Round struct {
	ID          int       `json:"id"`
	RoundNumber int       `json:"round_number"`
	Format      string    `json:"format"` // "PB" or "BF"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Match struct {
	ID        int       `json:"id"`
	RoundID   int       `json:"round_id"`
	Player1ID int       `json:"player1_id"`
	Player2ID int       `json:"player2_id"`
	Score1    *int      `json:"score1"`
	Score2    *int      `json:"score2"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MatchDetail struct {
	ID          int       `json:"id"`
	RoundNumber int       `json:"round_number"`
	Format      string    `json:"format"`
	Player1Name string    `json:"player1_name"`
	Player2Name string    `json:"player2_name"`
	Score1      *int      `json:"score1"`
	Score2      *int      `json:"score2"`
	Completed   bool      `json:"completed"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Standing struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	MatchesPlayed     int    `json:"matches_played"`
	Wins              int    `json:"wins"`
	Ties              int    `json:"ties"`
	Losses            int    `json:"losses"`
	TotalPointsScored int    `json:"total_points_scored"`
}

// Request/Response DTOs
type CreatePlayerRequest struct {
	Name      string `json:"name" binding:"required"`
	Confirmed bool   `json:"confirmed"`
}

type CreateRoundRequest struct {
	RoundNumber int    `json:"round_number" binding:"required"`
	Format      string `json:"format" binding:"required,oneof=PB BF"`
}

type CreateMatchRequest struct {
	RoundID   int `json:"round_id" binding:"required"`
	Player1ID int `json:"player1_id" binding:"required"`
	Player2ID int `json:"player2_id" binding:"required"`
}

type UpdateScoreRequest struct {
	Score1 int `json:"score1" binding:"gte=0"`
	Score2 int `json:"score2" binding:"gte=0"`
}

type FixtureRound struct {
	Number  int           `json:"number"`
	Format  string        `json:"format"`
	Matches []MatchDetail `json:"matches"`
}

type FixtureResponse struct {
	Rounds []FixtureRound `json:"rounds"`
}

type CreateFixtureRequest struct {
	Players []CreatePlayerRequest `json:"players" binding:"required"`
	Rounds  []struct {
		RoundNumber int    `json:"round_number" binding:"required"`
		Format      string `json:"format" binding:"required,oneof=PB BF"`
		Matches     []struct {
			Player1Name string `json:"player1_name" binding:"required"`
			Player2Name string `json:"player2_name" binding:"required"`
		} `json:"matches" binding:"required"`
	} `json:"rounds" binding:"required"`
}

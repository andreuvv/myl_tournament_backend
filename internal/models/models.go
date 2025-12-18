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
	Points            int    `json:"points"`
	TotalPointsScored int    `json:"total_points_scored"`
	TotalMatches      int    `json:"total_matches"`
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

// Tournament archive models
type Tournament struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Month      string    `json:"month"`
	Year       int       `json:"year"`
	StartDate  *string   `json:"start_date"`
	EndDate    *string   `json:"end_date"`
	CreatedAt  time.Time `json:"created_at"`
	ArchivedAt time.Time `json:"archived_at"`
}

type TournamentStanding struct {
	ID                int    `json:"id"`
	TournamentID      int    `json:"tournament_id"`
	PlayerID          int    `json:"player_id"`
	PlayerName        string `json:"player_name"`
	MatchesPlayed     int    `json:"matches_played"`
	Wins              int    `json:"wins"`
	Ties              int    `json:"ties"`
	Losses            int    `json:"losses"`
	Points            int    `json:"points"`
	TotalPointsScored int    `json:"total_points_scored"`
	TotalMatches      int    `json:"total_matches"`
	FinalPosition     int    `json:"final_position"`
}

type TournamentRound struct {
	ID           int       `json:"id"`
	TournamentID int       `json:"tournament_id"`
	RoundNumber  int       `json:"round_number"`
	Format       string    `json:"format"`
	CreatedAt    time.Time `json:"created_at"`
}

type TournamentMatch struct {
	ID                int       `json:"id"`
	TournamentRoundID int       `json:"tournament_round_id"`
	Player1ID         int       `json:"player1_id"`
	Player2ID         int       `json:"player2_id"`
	Player1Name       string    `json:"player1_name"`
	Player2Name       string    `json:"player2_name"`
	Score1            *int      `json:"score1"`
	Score2            *int      `json:"score2"`
	Completed         bool      `json:"completed"`
	CreatedAt         time.Time `json:"created_at"`
}

type ArchiveTournamentRequest struct {
	Name      string  `json:"name" binding:"required"`
	Month     string  `json:"month" binding:"required"`
	Year      int     `json:"year" binding:"required"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

type TournamentRoundDetail struct {
	Number  int                   `json:"number"`
	Format  string                `json:"format"`
	Matches []TournamentMatchInfo `json:"matches"`
}

type TournamentMatchInfo struct {
	ID          int    `json:"id"`
	Player1Name string `json:"player1_name"`
	Player2Name string `json:"player2_name"`
	Score1      *int   `json:"score1"`
	Score2      *int   `json:"score2"`
	Completed   bool   `json:"completed"`
}

type TournamentRoundsResponse struct {
	TournamentName string                  `json:"tournament_name"`
	Rounds         []TournamentRoundDetail `json:"rounds"`
}

type TournamentPlayerRace struct {
	ID           int       `json:"id"`
	TournamentID int       `json:"tournament_id"`
	PlayerID     int       `json:"player_id"`
	PlayerName   string    `json:"player_name"`
	RacePB       *string   `json:"race_pb"`
	RaceBF       *string   `json:"race_bf"`
	Notes        *string   `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdatePlayerRaceRequest struct {
	RacePB *string `json:"race_pb"`
	RaceBF *string `json:"race_bf"`
	Notes  *string `json:"notes"`
}

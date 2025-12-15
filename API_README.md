# Premier Mitológico API Documentation

Complete documentation for all API endpoints in the Premier Mitológico tournament management system.

**Base URL**: `https://your-api-domain.com/api`

---

## Table of Contents

- [Authentication](#authentication)
- [Public Endpoints](#public-endpoints)
  - [Health Check](#health-check)
  - [Get Fixture](#get-fixture)
  - [Get Standings](#get-standings)
  - [Get Players](#get-players)
  - [Get Tournaments (History)](#get-tournaments-history)
  - [Get Tournament Standings](#get-tournament-standings)
  - [Get Tournament Rounds](#get-tournament-rounds)
- [Protected Endpoints](#protected-endpoints)
  - [Create Player](#create-player)
  - [Toggle Player Confirmed](#toggle-player-confirmed)
  - [Get Confirmed Players](#get-confirmed-players)
  - [Create Fixture](#create-fixture)
  - [Update Match Score](#update-match-score)
  - [Archive Tournament](#archive-tournament)
  - [Clear Tournament](#clear-tournament)
- [Data Models](#data-models)
- [Error Handling](#error-handling)

---

## Authentication

Protected endpoints require an API key to be sent in the request headers.

**Header**: `X-API-Key: your-api-key-here`

Example:
```bash
curl -H "X-API-Key: your-api-key-here" \
  https://your-api-domain.com/api/players
```

---

## Public Endpoints

These endpoints are accessible without authentication.

### Health Check

Check if the API server is running.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "ok"
}
```

**Example**:
```bash
curl https://your-api-domain.com/health
```

---

### Get Fixture

Retrieve the complete tournament fixture with all rounds and matches.

**Endpoint**: `GET /api/fixture`

**Response**:
```json
{
  "rounds": [
    {
      "number": 1,
      "format": "PB",
      "matches": [
        {
          "id": 1,
          "round_number": 1,
          "format": "PB",
          "player1_name": "Player A",
          "player2_name": "Player B",
          "score1": 2,
          "score2": 1,
          "completed": true,
          "updated_at": "2025-12-13T15:30:00Z"
        }
      ]
    },
    {
      "number": 2,
      "format": "BF",
      "matches": [...]
    }
  ]
}
```

**Response Fields**:
- `rounds`: Array of tournament rounds
  - `number`: Round number (1, 2, 3, etc.)
  - `format`: Format code ("PB" = Primer Bloque, "BF" = Bloque Furia)
  - `matches`: Array of matches in this round
    - `id`: Unique match identifier
    - `player1_name`: First player's name
    - `player2_name`: Second player's name
    - `score1`: First player's score (null if not played)
    - `score2`: Second player's score (null if not played)
    - `completed`: Whether the match is finished
    - `updated_at`: Last update timestamp

**Example**:
```bash
curl https://your-api-domain.com/api/fixture
```

---

### Get Standings

Retrieve current tournament standings (leaderboard).

**Endpoint**: `GET /api/standings`

**Response**:
```json
[
  {
    "id": 1,
    "name": "Player A",
    "matches_played": 5,
    "wins": 4,
    "ties": 1,
    "losses": 0,
    "points": 13,
    "total_points_scored": 45,
    "total_matches": 15
  },
  {
    "id": 2,
    "name": "Player B",
    "matches_played": 5,
    "wins": 3,
    "ties": 1,
    "losses": 1,
    "points": 10,
    "total_points_scored": 38,
    "total_matches": 15
  }
]
```

**Response Fields**:
- `id`: Player ID
- `name`: Player name
- `matches_played`: Number of rounds played
- `wins`: Number of rounds won
- `ties`: Number of rounds tied
- `losses`: Number of rounds lost
- `points`: Total tournament points (3 per win, 1 per tie)
- `total_points_scored`: Total match points scored across all games
- `total_matches`: Total individual matches/games played

**Sorting**: Results are sorted by `points` (descending), then `total_points_scored` (descending).

**Example**:
```bash
curl https://your-api-domain.com/api/standings
```

---

### Get Players

Retrieve all registered players.

**Endpoint**: `GET /api/players`

**Response**:
```json
[
  {
    "id": 1,
    "name": "Player A",
    "confirmed": true,
    "created_at": "2025-12-01T10:00:00Z",
    "updated_at": "2025-12-10T14:30:00Z"
  },
  {
    "id": 2,
    "name": "Player B",
    "confirmed": false,
    "created_at": "2025-12-02T11:00:00Z",
    "updated_at": "2025-12-02T11:00:00Z"
  }
]
```

**Response Fields**:
- `id`: Unique player identifier
- `name`: Player name
- `confirmed`: Whether player is confirmed for the tournament
- `created_at`: Player registration timestamp
- `updated_at`: Last update timestamp

**Example**:
```bash
curl https://your-api-domain.com/api/players
```

---

### Get Tournaments (History)

Retrieve list of all archived tournaments.

**Endpoint**: `GET /api/tournaments`

**Response**:
```json
[
  {
    "id": 1,
    "name": "Copa K&T Diciembre 2025",
    "month": "Diciembre",
    "year": 2025,
    "start_date": "2025-12-13",
    "end_date": "2025-12-13",
    "created_at": "2025-12-14T20:00:00Z"
  },
  {
    "id": 2,
    "name": "Copa Navidad",
    "month": "Noviembre",
    "year": 2025,
    "start_date": "2025-11-20",
    "end_date": "2025-11-20",
    "created_at": "2025-11-21T18:00:00Z"
  }
]
```

**Response Fields**:
- `id`: Unique tournament identifier
- `name`: Tournament name
- `month`: Tournament month (Spanish)
- `year`: Tournament year
- `start_date`: Tournament start date (YYYY-MM-DD)
- `end_date`: Tournament end date (YYYY-MM-DD)
- `created_at`: Archive timestamp

**Sorting**: Results are sorted by year (descending), then month (descending using Spanish month order).

**Example**:
```bash
curl https://your-api-domain.com/api/tournaments
```

---

### Get Tournament Standings

Retrieve final standings for a specific archived tournament.

**Endpoint**: `GET /api/tournaments/:id/standings`

**URL Parameters**:
- `id`: Tournament ID (integer)

**Response**:
```json
[
  {
    "id": 1,
    "tournament_id": 1,
    "player_name": "Player A",
    "final_position": 1,
    "matches_played": 5,
    "wins": 4,
    "ties": 1,
    "losses": 0,
    "points": 13,
    "total_points_scored": 45,
    "total_matches": 15
  },
  {
    "id": 2,
    "tournament_id": 1,
    "player_name": "Player B",
    "final_position": 2,
    "matches_played": 5,
    "wins": 3,
    "ties": 2,
    "losses": 0,
    "points": 11,
    "total_points_scored": 42,
    "total_matches": 15
  }
]
```

**Response Fields**:
- `id`: Standing record ID
- `tournament_id`: Associated tournament ID
- `player_name`: Player name (denormalized for historical records)
- `final_position`: Final tournament position/ranking
- `matches_played`: Number of rounds played
- `wins`: Number of rounds won
- `ties`: Number of rounds tied
- `losses`: Number of rounds lost
- `points`: Total tournament points
- `total_points_scored`: Total match points scored
- `total_matches`: Total individual matches played

**Sorting**: Results are sorted by `final_position` (ascending).

**Example**:
```bash
curl https://your-api-domain.com/api/tournaments/1/standings
```

**Error Responses**:
- `404`: Tournament not found

---

### Get Tournament Rounds

Retrieve all rounds and matches for a specific archived tournament.

**Endpoint**: `GET /api/tournaments/:id/rounds`

**URL Parameters**:
- `id`: Tournament ID (integer)

**Response**:
```json
[
  {
    "id": 1,
    "tournament_id": 1,
    "round_number": 1,
    "format": "PB",
    "matches": [
      {
        "id": 1,
        "tournament_round_id": 1,
        "player1_name": "Player A",
        "player2_name": "Player B",
        "score1": 2,
        "score2": 1,
        "completed": true
      }
    ]
  },
  {
    "id": 2,
    "tournament_id": 1,
    "round_number": 2,
    "format": "BF",
    "matches": [...]
  }
]
```

**Response Fields**:
- `id`: Round record ID
- `tournament_id`: Associated tournament ID
- `round_number`: Round number in tournament
- `format`: Format code ("PB" or "BF")
- `matches`: Array of matches in this round
  - `id`: Match record ID
  - `tournament_round_id`: Associated round ID
  - `player1_name`: First player's name
  - `player2_name`: Second player's name
  - `score1`: First player's score
  - `score2`: Second player's score
  - `completed`: Whether match was completed

**Sorting**: Rounds sorted by `round_number`, matches sorted by `id`.

**Example**:
```bash
curl https://your-api-domain.com/api/tournaments/1/rounds
```

**Error Responses**:
- `404`: Tournament not found

---

## Protected Endpoints

These endpoints require authentication via `X-API-Key` header.

### Create Player

Add a new player to the roster.

**Endpoint**: `POST /api/players`

**Headers**:
```
X-API-Key: your-api-key-here
Content-Type: application/json
```

**Request Body**:
```json
{
  "name": "New Player Name"
}
```

**Request Fields**:
- `name`: Player name (required, string, max 255 characters)

**Response** (Success - 201):
```json
{
  "id": 3,
  "name": "New Player Name",
  "confirmed": false,
  "created_at": "2025-12-15T10:00:00Z",
  "updated_at": "2025-12-15T10:00:00Z"
}
```

**Example**:
```bash
curl -X POST https://your-api-domain.com/api/players \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{"name": "New Player"}'
```

**Error Responses**:
- `400`: Missing or invalid player name
- `401`: Missing or invalid API key
- `500`: Database error

---

### Toggle Player Confirmed

Toggle a player's confirmation status for the tournament.

**Endpoint**: `PATCH /api/players/:id/confirm`

**URL Parameters**:
- `id`: Player ID (integer)

**Headers**:
```
X-API-Key: your-api-key-here
```

**Response** (Success - 200):
```json
{
  "id": 1,
  "name": "Player A",
  "confirmed": true,
  "created_at": "2025-12-01T10:00:00Z",
  "updated_at": "2025-12-15T10:05:00Z"
}
```

**Example**:
```bash
curl -X PATCH https://your-api-domain.com/api/players/1/confirm \
  -H "X-API-Key: your-api-key-here"
```

**Error Responses**:
- `400`: Invalid player ID
- `401`: Missing or invalid API key
- `404`: Player not found
- `500`: Database error

---

### Get Confirmed Players

Retrieve only players who are confirmed for the tournament.

**Endpoint**: `GET /api/players/confirmed`

**Headers**:
```
X-API-Key: your-api-key-here
```

**Response**:
```json
[
  {
    "id": 1,
    "name": "Player A",
    "confirmed": true,
    "created_at": "2025-12-01T10:00:00Z",
    "updated_at": "2025-12-10T14:30:00Z"
  },
  {
    "id": 3,
    "name": "Player C",
    "confirmed": true,
    "created_at": "2025-12-05T09:00:00Z",
    "updated_at": "2025-12-12T16:00:00Z"
  }
]
```

**Example**:
```bash
curl https://your-api-domain.com/api/players/confirmed \
  -H "X-API-Key: your-api-key-here"
```

**Error Responses**:
- `401`: Missing or invalid API key
- `500`: Database error

---

### Create Fixture

Generate the complete tournament fixture with all rounds and matches. This creates the entire tournament structure based on confirmed players.

**Endpoint**: `POST /api/fixture`

**Headers**:
```
X-API-Key: your-api-key-here
Content-Type: application/json
```

**Request Body**:
```json
{
  "rounds": [
    {
      "format": "PB",
      "pairings": [
        { "player1_id": 1, "player2_id": 2 },
        { "player1_id": 3, "player2_id": 4 }
      ]
    },
    {
      "format": "BF",
      "pairings": [
        { "player1_id": 1, "player2_id": 3 },
        { "player1_id": 2, "player2_id": 4 }
      ]
    }
  ]
}
```

**Request Fields**:
- `rounds`: Array of tournament rounds (required)
  - `format`: Format code ("PB" or "BF", required)
  - `pairings`: Array of match pairings (required)
    - `player1_id`: First player ID (required, integer)
    - `player2_id`: Second player ID (required, integer)

**Response** (Success - 201):
```json
{
  "message": "Fixture created successfully",
  "rounds_created": 2,
  "matches_created": 4
}
```

**Example**:
```bash
curl -X POST https://your-api-domain.com/api/fixture \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "rounds": [
      {
        "format": "PB",
        "pairings": [
          {"player1_id": 1, "player2_id": 2}
        ]
      }
    ]
  }'
```

**Error Responses**:
- `400`: Invalid request body or missing required fields
- `401`: Missing or invalid API key
- `500`: Database error (e.g., duplicate round, invalid player IDs)

**Notes**:
- Creates rounds sequentially (1, 2, 3, etc.)
- Validates that all player IDs exist
- Uses database transaction to ensure atomicity

---

### Update Match Score

Update the score for a specific match.

**Endpoint**: `PATCH /api/matches/:id/score`

**URL Parameters**:
- `id`: Match ID (integer)

**Headers**:
```
X-API-Key: your-api-key-here
Content-Type: application/json
```

**Request Body**:
```json
{
  "score1": 2,
  "score2": 1
}
```

**Request Fields**:
- `score1`: First player's score (required, integer, 0-2)
- `score2`: Second player's score (required, integer, 0-2)

**Response** (Success - 200):
```json
{
  "id": 1,
  "round_id": 1,
  "player1_id": 1,
  "player2_id": 2,
  "score1": 2,
  "score2": 1,
  "completed": true,
  "created_at": "2025-12-13T15:00:00Z",
  "updated_at": "2025-12-15T10:30:00Z"
}
```

**Example**:
```bash
curl -X PATCH https://your-api-domain.com/api/matches/1/score \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{"score1": 2, "score2": 1}'
```

**Error Responses**:
- `400`: Invalid match ID, missing scores, or invalid score values
- `401`: Missing or invalid API key
- `404`: Match not found
- `500`: Database error

**Notes**:
- Valid scores are 0, 1, or 2 (Best of 3 format)
- Match is marked as `completed` when scores are updated
- Automatically updates the `standings` view via database triggers

---

### Archive Tournament

Save the current tournament (standings and matches) to the archive and optionally clear the active tournament.

**Endpoint**: `POST /api/tournaments/archive`

**Headers**:
```
X-API-Key: your-api-key-here
Content-Type: application/json
```

**Request Body**:
```json
{
  "name": "Copa K&T Diciembre 2025",
  "month": "Diciembre",
  "year": 2025,
  "start_date": "2025-12-13",
  "end_date": "2025-12-13",
  "clear_current": true
}
```

**Request Fields**:
- `name`: Tournament name (required, string)
- `month`: Tournament month in Spanish (required, string)
- `year`: Tournament year (required, integer)
- `start_date`: Tournament start date (required, string, format: YYYY-MM-DD)
- `end_date`: Tournament end date (required, string, format: YYYY-MM-DD)
- `clear_current`: Whether to clear the active tournament after archiving (optional, boolean, default: false)

**Response** (Success - 201):
```json
{
  "message": "Tournament archived successfully",
  "tournament_id": 1,
  "standings_archived": 8,
  "rounds_archived": 5,
  "matches_archived": 20
}
```

**Response Fields**:
- `message`: Success message
- `tournament_id`: ID of the newly created archive entry
- `standings_archived`: Number of player standings saved
- `rounds_archived`: Number of rounds saved
- `matches_archived`: Number of matches saved

**Example**:
```bash
curl -X POST https://your-api-domain.com/api/tournaments/archive \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Copa K&T Diciembre 2025",
    "month": "Diciembre",
    "year": 2025,
    "start_date": "2025-12-13",
    "end_date": "2025-12-13",
    "clear_current": true
  }'
```

**Error Responses**:
- `400`: Missing required fields or invalid date format
- `401`: Missing or invalid API key
- `500`: Database error (check for incomplete matches or data integrity issues)

**Notes**:
- Uses database transaction to ensure data consistency
- Archives current standings with player names (denormalized)
- Archives all rounds and matches
- If `clear_current` is true, clears matches, rounds, and unconfirms all players after successful archive
- Requires all matches to be completed before archiving

---

### Clear Tournament

Delete all tournament data (matches, rounds) and reset player confirmations. This prepares for a new tournament.

**Endpoint**: `DELETE /api/tournament`

**Headers**:
```
X-API-Key: your-api-key-here
Content-Type: application/json
```

**Request Body**:
```json
{
  "clear_players": false
}
```

**Request Fields**:
- `clear_players`: Whether to also delete all player records (optional, boolean, default: false)

**Response** (Success - 200):
```json
{
  "message": "Tournament data cleared successfully",
  "matches_deleted": 20,
  "rounds_deleted": 5,
  "players_unconfirmed": 8
}
```

**Response Fields**:
- `message`: Success message
- `matches_deleted`: Number of matches deleted
- `rounds_deleted`: Number of rounds deleted
- `players_unconfirmed`: Number of players set to unconfirmed (or deleted if `clear_players` was true)

**Example** (Clear tournament, keep players):
```bash
curl -X DELETE https://your-api-domain.com/api/tournament \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{"clear_players": false}'
```

**Example** (Clear tournament and delete all players):
```bash
curl -X DELETE https://your-api-domain.com/api/tournament \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{"clear_players": true}'
```

**Error Responses**:
- `401`: Missing or invalid API key
- `500`: Database error

**Notes**:
- Uses database transaction to ensure atomicity
- Deletes matches first, then rounds (due to foreign key constraints)
- If `clear_players` is false: sets all players' `confirmed` status to false
- If `clear_players` is true: deletes all player records
- The `standings` view is automatically updated via database triggers

---

## Data Models

### Player
```typescript
{
  id: number;
  name: string;
  confirmed: boolean;
  created_at: string; // ISO 8601 timestamp
  updated_at: string; // ISO 8601 timestamp
}
```

### Round
```typescript
{
  id: number;
  round_number: number;
  format: "PB" | "BF";
  created_at: string;
  updated_at: string;
}
```

### Match
```typescript
{
  id: number;
  round_id: number;
  player1_id: number;
  player2_id: number;
  score1: number | null; // 0-2 or null
  score2: number | null; // 0-2 or null
  completed: boolean;
  created_at: string;
  updated_at: string;
}
```

### Match Detail (in Fixture)
```typescript
{
  id: number;
  round_number: number;
  format: "PB" | "BF";
  player1_name: string;
  player2_name: string;
  score1: number | null;
  score2: number | null;
  completed: boolean;
  updated_at: string;
}
```

### Standing
```typescript
{
  id: number;
  name: string;
  matches_played: number;
  wins: number;
  ties: number;
  losses: number;
  points: number;
  total_points_scored: number;
  total_matches: number;
}
```

### Tournament (Archive)
```typescript
{
  id: number;
  name: string;
  month: string; // Spanish month name
  year: number;
  start_date: string; // YYYY-MM-DD
  end_date: string;   // YYYY-MM-DD
  created_at: string; // ISO 8601 timestamp
}
```

### Tournament Standing (Archive)
```typescript
{
  id: number;
  tournament_id: number;
  player_name: string; // Denormalized
  final_position: number;
  matches_played: number;
  wins: number;
  ties: number;
  losses: number;
  points: number;
  total_points_scored: number;
  total_matches: number;
}
```

### Tournament Round (Archive)
```typescript
{
  id: number;
  tournament_id: number;
  round_number: number;
  format: "PB" | "BF";
  matches: TournamentMatch[];
}
```

### Tournament Match (Archive)
```typescript
{
  id: number;
  tournament_round_id: number;
  player1_name: string; // Denormalized
  player2_name: string; // Denormalized
  score1: number;
  score2: number;
  completed: boolean;
}
```

---

## Error Handling

All endpoints follow consistent error response formats:

### Error Response Format
```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes

| Code | Meaning | When It Occurs |
|------|---------|----------------|
| 200 | OK | Successful GET, PATCH, or DELETE request |
| 201 | Created | Successful POST request creating a resource |
| 400 | Bad Request | Invalid request body, missing required fields, or invalid parameters |
| 401 | Unauthorized | Missing or invalid API key for protected endpoints |
| 404 | Not Found | Resource (player, match, tournament) not found |
| 500 | Internal Server Error | Database error or unexpected server error |

### Example Error Responses

**400 Bad Request**:
```json
{
  "error": "Player name is required"
}
```

**401 Unauthorized**:
```json
{
  "error": "Unauthorized"
}
```

**404 Not Found**:
```json
{
  "error": "Player not found"
}
```

**500 Internal Server Error**:
```json
{
  "error": "Failed to update match score"
}
```

---

## Rate Limiting

Currently, there is no rate limiting implemented. Consider implementing rate limiting for production use to prevent abuse.

---

## CORS Configuration

The API is configured with CORS middleware allowing:
- All origins (`*`)
- Methods: GET, POST, PATCH, DELETE, OPTIONS
- Headers: Content-Type, X-API-Key

Adjust CORS settings in `internal/middleware/cors.go` as needed for production.

---

## Database Schema

The API uses PostgreSQL with the following main tables:

- `players`: Player roster
- `rounds`: Tournament rounds
- `matches`: Individual matches
- `standings`: Materialized view of current standings (auto-updated via triggers)
- `tournaments`: Archived tournament metadata
- `tournament_standings`: Archived final standings
- `tournament_rounds`: Archived rounds
- `tournament_matches`: Archived matches

For complete schema details, see migration files in `/migrations`.

---

## Development & Testing

### Local Development

1. Set environment variables:
   ```bash
   DATABASE_URL=postgresql://user:pass@localhost/dbname
   API_KEY=your-dev-api-key
   PORT=8080
   ```

2. Run migrations:
   ```bash
   go run cmd/server/main.go
   ```

3. Server starts at `http://localhost:8080`

### Testing Endpoints

Use the provided `API_EXAMPLES.md` file for cURL examples of all endpoints.

### Health Check

Always available at `/health` without authentication:
```bash
curl http://localhost:8080/health
```

---

## Deployment

The API is deployed on Railway with automatic deployments from the main branch.

**Production URL**: See `DEPLOYMENT_SUCCESS.md` for the current production URL.

**Environment Variables** (set in Railway):
- `DATABASE_URL`: PostgreSQL connection string
- `API_KEY`: Secret API key for protected endpoints
- `PORT`: Server port (Railway provides this automatically)

---

## Support

For issues or questions:
- Check `API_EXAMPLES.md` for practical examples
- Review `DEPLOYMENT_README.md` for deployment details
- Check migration files in `/migrations` for database schema

---

## Version History

- **v1.0** (December 2025): Initial API with tournament management
- **v1.1** (December 2025): Added tournament archive functionality

---

**Last Updated**: December 15, 2025

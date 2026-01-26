# Online Tournament API Documentation

## Overview

The online tournament system allows you to create round-robin style tournaments where players schedule their own matches and report scores over an extended period.

## Key Differences from In-Person Tournaments

- **No rounds**: Matches are generated all at once as a complete round-robin
- **Single format**: Entire tournament uses one format (PB or BF)
- **Flexible scheduling**: Players play matches at their own pace
- **Same point system**: Win=3, Tie=1, Loss=0

## API Endpoints

### Create Online Tournament

**Endpoint**: `POST /api/tournaments/online`

**Headers**:
```
X-API-Key: your-api-key-here
Content-Type: application/json
```

**Request Body**:
```json
{
  "name": "Online Tournament January 2026",
  "month": "January",
  "year": 2026,
  "format": "PB",
  "player_ids": [1, 3, 5, 7, 9],
  "start_date": "2026-01-26",
  "end_date": "2026-02-15"
}
```

**Request Fields**:
- `name`: Tournament name (required, string)
- `month`: Month name (required, string)
- `year`: Year (required, integer)
- `format`: Tournament format (required, "PB" or "BF")
- `player_ids`: Array of player IDs to include (required, minimum 2 players)
- `start_date`: Optional start date (string, ISO 8601)
- `end_date`: Optional end date (string, ISO 8601)

**Response** (Success - 201):
```json
{
  "message": "Online tournament created successfully",
  "tournament_id": 42,
  "tournament_name": "Online Tournament January 2026",
  "format": "PB",
  "players_added": 5,
  "matches_generated": 10
}
```

**Auto-Generated Matches**:
- For N players: N*(N-1)/2 matches are created
- Each match has both player IDs and names stored
- All matches start as incomplete
- Example (5 players = 10 matches):
  - Player 1 vs 2, 1 vs 3, 1 vs 4, 1 vs 5
  - Player 2 vs 3, 2 vs 4, 2 vs 5
  - Player 3 vs 4, 3 vs 5
  - Player 4 vs 5

---

### Get All Online Tournament Matches

**Endpoint**: `GET /api/tournaments/online/:id/matches`

**Response**:
```json
[
  {
    "id": 1,
    "tournament_id": 42,
    "player1_id": 1,
    "player2_id": 3,
    "player1_name": "Troke",
    "player2_name": "Piter",
    "score1": null,
    "score2": null,
    "completed": false,
    "match_date": null,
    "created_at": "2026-01-26T10:00:00Z",
    "updated_at": "2026-01-26T10:00:00Z"
  },
  {
    "id": 2,
    "tournament_id": 42,
    "player1_id": 1,
    "player2_id": 5,
    "player1_name": "Troke",
    "player2_name": "Folo",
    "score1": 2,
    "score2": 0,
    "completed": true,
    "match_date": "2026-01-26T15:30:00Z",
    "created_at": "2026-01-26T10:00:00Z",
    "updated_at": "2026-01-26T15:35:00Z"
  }
]
```

---

### Get Pending Matches (Not Completed)

**Endpoint**: `GET /api/tournaments/online/:id/matches/pending`

**Response**: Same format as above, but only includes matches with `completed: false`

---

### Get Completed Matches

**Endpoint**: `GET /api/tournaments/online/:id/matches/completed`

**Response**: Same format as above, but only includes matches with `completed: true`

---

### Update Match Score

**Endpoint**: `PATCH /api/tournaments/online/matches/:matchId`

**Headers**:
```
X-API-Key: your-api-key-here
Content-Type: application/json
```

**Request Body**:
```json
{
  "score1": 2,
  "score2": 0
}
```

**Request Fields**:
- `score1`: Score for player 1 (required, integer ≥ 0)
- `score2`: Score for player 2 (required, integer ≥ 0)

**Response** (Success - 200):
```json
{
  "message": "Match score updated successfully",
  "match_id": 1,
  "score": "Troke 2-0 Piter"
}
```

**Notes**:
- Automatically marks match as `completed: true`
- Updates `updated_at` timestamp
- Standings are automatically recalculated

---

### Get Tournament Standings

**Endpoint**: `GET /api/tournaments/online/:id/standings`

**Response**:
```json
[
  {
    "tournament_id": 42,
    "player_id": 1,
    "player_name": "Troke",
    "matches_played": 4,
    "wins": 3,
    "ties": 1,
    "losses": 0,
    "points": 10
  },
  {
    "tournament_id": 42,
    "player_id": 3,
    "player_name": "Piter",
    "matches_played": 4,
    "wins": 2,
    "ties": 0,
    "losses": 2,
    "points": 6
  }
]
```

**Sorted by**: `points DESC, wins DESC`

---

### Get Tournament Info

**Endpoint**: `GET /api/tournaments/online/:id/info`

**Response**:
```json
{
  "id": 42,
  "name": "Online Tournament January 2026",
  "month": "January",
  "year": 2026,
  "format": "PB",
  "type": "ONLINE",
  "start_date": "2026-01-26",
  "end_date": "2026-02-15",
  "created_at": "2026-01-26T10:00:00Z"
}
```

---

### Get All Active Tournaments

**Endpoint**: `GET /api/tournaments/active`

**Response**:
```json
[
  {
    "id": 42,
    "name": "Online Tournament January 2026",
    "month": "January",
    "year": 2026,
    "type": "ONLINE",
    "format": "PB",
    "start_date": "2026-01-26",
    "end_date": "2026-02-15",
    "created_at": "2026-01-26T10:00:00Z"
  },
  {
    "id": 41,
    "name": "In-Person Tournament January 2026",
    "month": "January",
    "year": 2026,
    "type": "IN_PERSON",
    "format": null,
    "start_date": null,
    "end_date": null,
    "created_at": "2026-01-15T10:00:00Z"
  }
]
```

---

### Delete Online Tournament

**Endpoint**: `DELETE /api/tournaments/online/:id`

**Headers**:
```
X-API-Key: your-api-key-here
```

**Response** (Success - 200):
```json
{
  "message": "Online tournament deleted successfully"
}
```

**Notes**:
- Deletes the tournament record
- Deletes all associated matches
- Deletes all tournament player entries
- Cannot delete in-person tournaments with this endpoint (type check in place)

---

## Points System

| Result | Points |
|--------|--------|
| Win (score1 > score2)    | 3 |
| Tie (score1 = score2)    | 1 |
| Loss (score1 < score2)   | 0 |

---

## Flutter Integration Notes

For the Flutter admin app, the workflow should be:

1. **Tournament Setup Screen**:
   - Enter tournament name, month, year
   - Select format (PB/BF dropdown)
   - Select start and end dates

2. **Player Selection Screen**:
   - Display all available players (from premier_players)
   - Allow user to check/uncheck players
   - Show count of selected players
   - Validate minimum 2 players

3. **Create Tournament**:
   - Send all selections to `POST /api/tournaments/online`
   - Show response with confirmation of created matches

4. **Match Reporting Screen**:
   - Show list of pending matches (use `/matches/pending`)
   - Click on a match to enter scores
   - Update score via `PATCH /api/tournaments/online/matches/:matchId`

5. **Tournament Standings**:
   - Display standings from `GET /api/tournaments/online/:id/standings`
   - Auto-refresh or poll for live updates

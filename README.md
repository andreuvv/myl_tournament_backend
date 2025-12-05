# Tournament Management API - Backend

Go + PostgreSQL backend API for managing tournament fixtures, match results, and player standings.

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher ([Download](https://go.dev/dl/))
- PostgreSQL 14+ ([Download](https://www.postgresql.org/download/))
- Git

### Installation

1. **Install Go dependencies:**
   ```bash
   cd myl_app_go_backend
   go mod download
   ```

2. **Set up PostgreSQL database:**
   ```bash
   # Create database
   psql -U postgres
   CREATE DATABASE tournament_db;
   \q
   
   # Run migrations
   psql -U postgres -d tournament_db -f migrations/001_initial_schema.sql
   ```

3. **Configure environment variables:**
   ```bash
   # Copy example file
   cp .env.example .env
   
   # Edit .env with your settings
   # Update DB_PASSWORD and API_KEY
   ```

4. **Run the server:**
   ```bash
   go run cmd/server/main.go
   ```

   Server will start on `http://localhost:8080`

## 📡 API Endpoints

### Public Endpoints (No authentication)

#### Get Fixture
```http
GET /api/fixture
```
Returns all rounds with their matches.

**Response:**
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
          "player1_name": "Troke",
          "player2_name": "Timmy",
          "score1": null,
          "score2": null,
          "completed": false
        }
      ]
    }
  ]
}
```

#### Get Standings
```http
GET /api/standings
```
Returns current tournament standings.

**Response:**
```json
[
  {
    "id": 1,
    "name": "Troke",
    "matches_played": 5,
    "wins": 4,
    "losses": 1,
    "total_points_scored": 20,
    "total_points_against": 8
  }
]
```

#### Get Players
```http
GET /api/players
```
Returns all players.

### Protected Endpoints (Require API Key)

All protected endpoints require header:
```
X-API-Key: your_secret_api_key_here
```

#### Create Complete Fixture
```http
POST /api/fixture
```
Creates players, rounds, and matches in one transaction.

**Request:**
```json
{
  "players": [
    { "name": "Troke", "confirmed": true },
    { "name": "Timmy", "confirmed": true }
  ],
  "rounds": [
    {
      "round_number": 1,
      "format": "PB",
      "matches": [
        {
          "player1_name": "Troke",
          "player2_name": "Timmy"
        }
      ]
    }
  ]
}
```

#### Update Match Score
```http
PATCH /api/matches/:id/score
```

**Request:**
```json
{
  "score1": 2,
  "score2": 1
}
```

#### Create Player
```http
POST /api/players
```

**Request:**
```json
{
  "name": "Chester",
  "confirmed": true
}
```

## 🔐 Authentication

Protected endpoints use API Key authentication. Include the key in request headers:

```bash
curl -H "X-API-Key: your_secret_api_key_here" \
  http://localhost:8080/api/matches/1/score
```

## 📱 Mobile App Integration

### Posting Fixture from Mobile App

```javascript
// Example: React Native / Flutter HTTP request
const createFixture = async () => {
  const response = await fetch('http://your-server.com/api/fixture', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'your_secret_api_key_here'
    },
    body: JSON.stringify({
      players: [
        { name: "Troke", confirmed: true },
        { name: "Timmy", confirmed: true },
        // ... more players
      ],
      rounds: [
        {
          round_number: 1,
          format: "PB",
          matches: [
            { player1_name: "Troke", player2_name: "Timmy" }
          ]
        },
        // ... more rounds
      ]
    })
  });
  
  const result = await response.json();
  console.log(result);
};
```

### Updating Match Scores

```javascript
const updateScore = async (matchId, score1, score2) => {
  const response = await fetch(`http://your-server.com/api/matches/${matchId}/score`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'your_secret_api_key_here'
    },
    body: JSON.stringify({ score1, score2 })
  });
  
  const result = await response.json();
  console.log(result);
};
```

## 🌐 React Web App Integration

Update your React app to fetch from API:

```typescript
// src/services/api.ts
const API_BASE_URL = 'http://localhost:8080/api';

export const getFixture = async () => {
  const response = await fetch(`${API_BASE_URL}/fixture`);
  return response.json();
};

export const getStandings = async () => {
  const response = await fetch(`${API_BASE_URL}/standings`);
  return response.json();
};

// Use in component
import { useEffect, useState } from 'react';
import { getFixture } from '../services/api';

const FixturePage = () => {
  const [fixture, setFixture] = useState(null);
  
  useEffect(() => {
    getFixture().then(data => setFixture(data));
  }, []);
  
  // Render fixture...
};
```

## 🗄️ Database Schema

### Tables
- **players** - Tournament participants
- **rounds** - Tournament rounds with format (PB/BF)
- **matches** - Individual matches with scores
- **standings** (view) - Auto-calculated rankings

### Key Features
- Auto-updating timestamps
- Cascading deletes
- Performance indexes
- Data validation constraints

## 🚢 Deployment

### Option 1: Railway / Render
1. Push code to GitHub
2. Connect repository to Railway/Render
3. Add PostgreSQL addon
4. Set environment variables
5. Deploy!

### Option 2: VPS (DigitalOcean, etc.)
```bash
# Build binary
go build -o tournament-api cmd/server/main.go

# Run with systemd or PM2
./tournament-api
```

## 📝 Development

### Run with auto-reload (using Air)
```bash
go install github.com/cosmtrek/air@latest
air
```

### Run tests
```bash
go test ./...
```

## 🛠️ Troubleshooting

**"command not found: go"**
- Install Go from https://go.dev/dl/

**"connection refused" database error**
- Ensure PostgreSQL is running: `psql -U postgres`
- Check credentials in `.env`

**CORS errors in browser**
- Add your frontend URL to `ALLOWED_ORIGINS` in `.env`

## 📄 License

MIT

## 🤝 Support

For issues or questions, open an issue on GitHub.

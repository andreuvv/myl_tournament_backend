# Player Management API Endpoints

All protected endpoints require the `X-API-Key` header.

## Public Endpoints

### Get All Players
```
GET /api/players
```
Returns all players with their confirmed status.

**Response:**
```json
[
  {
    "id": 1,
    "name": "Troke",
    "confirmed": true,
    "created_at": "2025-12-11T...",
    "updated_at": "2025-12-11T..."
  }
]
```

## Protected Endpoints

### Get Confirmed Players Only
```
GET /api/players/confirmed
Headers: X-API-Key: your-api-key
```
Returns only players with `confirmed = true`.

**Response:**
```json
[
  {
    "id": 1,
    "name": "Troke",
    "confirmed": true,
    "created_at": "2025-12-11T...",
    "updated_at": "2025-12-11T..."
  }
]
```

### Create Player
```
POST /api/players
Headers: X-API-Key: your-api-key
Content-Type: application/json

{
  "name": "PlayerName",
  "confirmed": true
}
```

**Response:**
```json
{
  "id": 13,
  "name": "PlayerName",
  "confirmed": true,
  "created_at": "2025-12-11T...",
  "updated_at": "2025-12-11T..."
}
```

### Toggle Player Confirmation
```
PATCH /api/players/:id/confirm
Headers: X-API-Key: your-api-key
```
Toggles the confirmed status (true ↔ false).

**Response:**
```json
{
  "id": 1,
  "name": "Troke",
  "confirmed": false,
  "created_at": "2025-12-11T...",
  "updated_at": "2025-12-11T..."
}
```

## Usage Workflow

1. **Initial Setup**: Run migration 011 to insert all 12 players with confirmed=true
2. **Player Roster Management**: Use Flutter app's "Players Roster" page to:
   - View all players
   - Add new players
   - Toggle confirmation status (players are never deleted to maintain ID consistency for tracking records)
3. **Fixture Generation**: When creating a fixture, it automatically uses only confirmed players from `/api/players/confirmed`

**Note**: Players cannot be deleted to preserve unique IDs for historical record tracking. Use the confirmation toggle to mark players as inactive instead.

## Database Schema

```sql
CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    confirmed BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Initial Players (Migration 011)

All players start as confirmed:
- Troke
- Timmy
- Piter
- Folo
- Wesh
- Guari
- Vinny
- Chisco
- Clanso
- Traukolin
- Chester
- David

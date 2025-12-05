# API Testing Examples

## Using cURL

### Get Fixture
```bash
curl http://localhost:8080/api/fixture
```

### Get Standings
```bash
curl http://localhost:8080/api/standings
```

### Get Players
```bash
curl http://localhost:8080/api/players
```

### Create Complete Fixture (Protected)
```bash
curl -X POST http://localhost:8080/api/fixture \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_secret_api_key_here" \
  -d '{
    "players": [
      {"name": "Troke", "confirmed": true},
      {"name": "Timmy", "confirmed": true},
      {"name": "Wesh", "confirmed": true},
      {"name": "Folo", "confirmed": true},
      {"name": "Piter", "confirmed": true},
      {"name": "Clanso", "confirmed": true},
      {"name": "Chisco", "confirmed": true},
      {"name": "Traukolin", "confirmed": true},
      {"name": "Chester", "confirmed": true},
      {"name": "David", "confirmed": true}
    ],
    "rounds": [
      {
        "round_number": 1,
        "format": "PB",
        "matches": [
          {"player1_name": "Troke", "player2_name": "Timmy"},
          {"player1_name": "Wesh", "player2_name": "Folo"},
          {"player1_name": "Piter", "player2_name": "Clanso"},
          {"player1_name": "Chisco", "player2_name": "Traukolin"},
          {"player1_name": "Chester", "player2_name": "David"}
        ]
      },
      {
        "round_number": 2,
        "format": "BF",
        "matches": [
          {"player1_name": "Timmy", "player2_name": "Wesh"},
          {"player1_name": "Folo", "player2_name": "Piter"},
          {"player1_name": "Clanso", "player2_name": "Chisco"},
          {"player1_name": "Traukolin", "player2_name": "Chester"},
          {"player1_name": "David", "player2_name": "Troke"}
        ]
      }
    ]
  }'
```

### Update Match Score (Protected)
```bash
curl -X PATCH http://localhost:8080/api/matches/1/score \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_secret_api_key_here" \
  -d '{
    "score1": 2,
    "score2": 1
  }'
```

### Create Individual Player (Protected)
```bash
curl -X POST http://localhost:8080/api/players \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_secret_api_key_here" \
  -d '{
    "name": "NewPlayer",
    "confirmed": true
  }'
```

## Using Postman

1. Import collection from `postman_collection.json` (create if needed)
2. Set environment variable `API_KEY`
3. Test endpoints

## Using JavaScript (Browser Console / Node.js)

```javascript
// Get fixture
fetch('http://localhost:8080/api/fixture')
  .then(r => r.json())
  .then(console.log);

// Update score (with API key)
fetch('http://localhost:8080/api/matches/1/score', {
  method: 'PATCH',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your_secret_api_key_here'
  },
  body: JSON.stringify({ score1: 2, score2: 1 })
})
.then(r => r.json())
.then(console.log);
```

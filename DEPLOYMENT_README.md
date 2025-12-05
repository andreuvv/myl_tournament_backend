# Tournament Backend API

Go + PostgreSQL backend for tournament management system.

## 🚀 Deployed on Railway

**Live API:** https://your-app.up.railway.app (update after deployment)

## 📡 API Endpoints

### Public
- `GET /api/fixture` - Get all rounds and matches
- `GET /api/standings` - Get current standings
- `GET /api/players` - Get all players

### Protected (require X-API-Key header)
- `POST /api/fixture` - Create complete fixture
- `POST /api/players` - Add player
- `PATCH /api/matches/:id/score` - Update match score

## 🏗️ Local Development

See [README.md](./README.md) for local setup instructions.

## 🚂 Deployment

See [RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md) for deployment instructions.

## 📄 License

MIT

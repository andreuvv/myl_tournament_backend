# Railway Deployment Guide

## 🚂 Deploy to Railway

### Step 1: Push Code to GitHub

```bash
cd c:/Users/veanv/Documents/myl_tournament_web/myl_app_go_backend

# Initialize git (if not already done)
git init
git add .
git commit -m "Initial backend setup for Railway deployment"

# Create a new repo on GitHub named "tournament-backend"
# Then push:
git remote add origin https://github.com/andreuvv/tournament-backend.git
git branch -M main
git push -u origin main
```

### Step 2: Deploy on Railway

1. **Go to Railway:** https://railway.app/
2. **Sign up/Login** with GitHub
3. **Click "New Project"**
4. **Select "Deploy from GitHub repo"**
5. **Choose your `tournament-backend` repository**

### Step 3: Add PostgreSQL Database

1. In your Railway project, click **"+ New"**
2. Select **"Database"** → **"PostgreSQL"**
3. Railway will automatically create the database and set `DATABASE_URL` environment variable

### Step 4: Configure Environment Variables

In Railway project settings → **Variables**, add:

```
API_KEY=tournament_myl_secret_2025
PORT=8080
ALLOWED_ORIGINS=https://andreuvv.github.io,http://localhost:5173
```

**Note:** Railway automatically provides `DATABASE_URL`, so you don't need to set DB_HOST, DB_PORT, etc.

### Step 5: Run Database Migrations

In Railway project → **your service** → **Settings** → **Deploy**:

1. Wait for initial deployment to complete
2. Go to **Settings** → **Service** → **Variables**
3. Click on PostgreSQL database service
4. Click **"Query"** tab
5. Copy and paste contents of `migrations/001_initial_schema.sql`
6. Click **"Run"**

**OR** use Railway CLI:

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login
railway login

# Link to your project
railway link

# Run migrations
railway run psql $DATABASE_URL -f migrations/001_initial_schema.sql
```

### Step 6: Get Your API URL

1. Go to your Railway project
2. Click on your Go service
3. Go to **Settings** → **Networking**
4. Click **"Generate Domain"**
5. You'll get a URL like: `https://tournament-backend-production.up.railway.app`

### Step 7: Test Your API

```bash
# Health check
curl https://your-app.up.railway.app/health

# Get fixture
curl https://your-app.up.railway.app/api/fixture

# Create fixture (with API key)
curl -X POST https://your-app.up.railway.app/api/fixture \
  -H "Content-Type: application/json" \
  -H "X-API-Key: tournament_myl_secret_2025" \
  -d '{"players": [...], "rounds": [...]}'
```

## 🔄 Automatic Deployments

Every time you push to `main` branch, Railway will automatically:
1. Build your Go app
2. Deploy the new version
3. Zero downtime deployment

## 📱 Update Your Apps

### React Web App

Update API base URL in your React app:

```typescript
// src/config/api.ts
export const API_BASE_URL = process.env.NODE_ENV === 'production'
  ? 'https://your-app.up.railway.app/api'
  : 'http://localhost:8080/api';
```

### Mobile App

Update API endpoint:
```javascript
const API_URL = 'https://your-app.up.railway.app/api';
```

## 💰 Free Tier Limits

Railway free tier includes:
- $5 credit/month (~500 hours of uptime)
- 1GB RAM
- 1GB storage
- Shared CPU

This should be enough for your tournament app!

## 🔧 Troubleshooting

**Build fails?**
- Check Railway logs in the dashboard
- Ensure `go.mod` and `go.sum` are committed

**Database connection fails?**
- Verify migrations ran successfully
- Check `DATABASE_URL` is set automatically by Railway

**API returns 500 errors?**
- Check logs in Railway dashboard
- Verify environment variables are set

## 📊 Monitoring

View logs in real-time:
```bash
railway logs
```

Or in Railway dashboard → Your service → **Logs**

---

## Quick Start Summary

```bash
# 1. Push to GitHub
git init
git add .
git commit -m "Railway deployment ready"
git remote add origin https://github.com/andreuvv/tournament-backend.git
git push -u origin main

# 2. Deploy on Railway (via web interface)
# 3. Add PostgreSQL database
# 4. Run migrations
# 5. Generate domain
# 6. Update your apps with new URL
```

🎉 Your tournament API is now live!

@echo off
REM Setup script for Windows

echo Setting up Tournament API Backend...
echo.

REM Set paths
set "GOPATH=C:\Program Files\Go\bin"
set "PGPATH=C:\Program Files\PostgreSQL\18\bin"
set PATH=%GOPATH%;%PGPATH%;%PATH%

echo Checking installations...
go version
psql --version
echo.

echo Creating database...
echo You'll be prompted for your PostgreSQL password
psql -U postgres -c "CREATE DATABASE tournament_db;"
echo.

echo Running migrations...
psql -U postgres -d tournament_db -f migrations/001_initial_schema.sql
echo.

echo Setup complete!
echo.
echo Next steps:
echo 1. Edit .env file with your PostgreSQL password and API key
echo 2. Run: setup_and_run.bat
pause

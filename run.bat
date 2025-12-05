@echo off
REM Run server script for Windows

set "GOPATH=C:\Program Files\Go\bin"
set "PGPATH=C:\Program Files\PostgreSQL\18\bin"
set PATH=%GOPATH%;%PGPATH%;%PATH%

echo Starting Tournament API Server...
echo.
go run cmd/server/main.go

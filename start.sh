#!/bin/bash

# Skript pro spuštění backendu i frontendu současně

echo "🚀 Spouštím backend..."
cd backend || exit
go run ./cmd/server/main.go &   # Spustí backend na pozadí
BACKEND_PID=$!

echo "🚀 Spouštím frontend..."
cd ../frontend || exit
npm run preview -- --host 0.0.0.0 &                    # Spustí frontend na pozadí
FRONTEND_PID=$!

echo "✅ Backend PID: $BACKEND_PID"
echo "✅ Frontend PID: $FRONTEND_PID"

# Čekání na ukončení obou procesů
wait $BACKEND_PID $FRONTEND_PID


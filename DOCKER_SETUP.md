# 🐳 Docker Setup - Backend Events Fix

## Great News!
You're using Docker, which means:
- ✅ PostgreSQL is pre-configured
- ✅ Redis is pre-configured  
- ✅ Environment variables are set
- ✅ No local setup needed!

---

## Quick Start (3 Steps)

### Step 1: Start Docker Containers

```bash
cd d:\App\backend\docker

# Start all services (PostgreSQL, Redis, API)
docker-compose up --build

# First time: will build the Docker image and start containers
# This takes 2-3 minutes the first time

# Expected output:
# college-events-api   | Starting College Event Management API in development mode...
# college-events-api   | Listening on :8080
```

### Step 2: Verify Backend is Running

```bash
# In a new terminal, test the health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","service":"college-events-api"}
```

### Step 3: Reload Flutter

```bash
# In Flutter terminal
r   # Hot reload

# Try creating an event
# ✅ No more datetime parsing errors!
```

---

## What Docker-Compose Does

### 1. PostgreSQL Container
```
- Image: postgres:15-alpine
- Database: college_events
- User: postgres
- Password: postgres
- Port: 5432
- Volume: postgres_data (persists between restarts)
```

### 2. Redis Container
```
- Image: redis:7-alpine
- Port: 6379
- For caching (optional)
```

### 3. API Container
```
- Builds from Dockerfile
- Uses your fixed Go code ✅
- Runs migrations automatically
- Connects to PostgreSQL
- Listens on port 8080
```

---

## How It Works

```
Docker-Compose Flow:
────────────────────

docker-compose up --build
    │
    ├─→ Builds API image from Dockerfile
    │   ├─ Compiles Go backend with YOUR fixes ✅
    │   ├─ Includes models.go with JSONTime ✅
    │   ├─ Includes updated handlers.go ✅
    │   └─ Creates executable: ./api
    │
    ├─→ Starts PostgreSQL container
    │   ├─ Initializes college_events database
    │   └─ Waits for readiness (healthcheck)
    │
    ├─→ Starts Redis container
    │   └─ Ready for caching
    │
    └─→ Starts API container
        ├─ Runs ./api executable
        ├─ Connects to PostgreSQL at postgres:5432
        ├─ Runs migrations (creates tables)
        ├─ Listens on 0.0.0.0:8080
        └─ ✅ Ready for requests!
```

---

## Flutter Configuration

Your Flutter app is already configured for Docker:

```dart
// In lib/src/services/api_service.dart
static const String baseUrl = 'http://10.0.2.2:8080/api/v1';
```

**Why 10.0.2.2?**
- `localhost` = Your phone/emulator itself
- `10.0.2.2` = Host machine from Android emulator
- Docker containers run on host machine port 8080
- ✅ This configuration is correct!

---

## Docker Commands Cheat Sheet

```bash
# Start containers (first time)
cd d:\App\backend\docker
docker-compose up --build

# Start containers (if already built)
docker-compose up

# Stop containers
docker-compose down

# Restart containers
docker-compose restart

# View logs
docker-compose logs -f api

# Run migrations manually
docker-compose exec api ./migrate

# Access PostgreSQL from command line
docker-compose exec postgres psql -U postgres -d college_events

# Clean everything (removes data!)
docker-compose down -v
```

---

## Troubleshooting

### "Port 8080 already in use"
```bash
# Kill whatever is using port 8080
lsof -ti:8080 | xargs kill -9

# Or change docker-compose.yml ports section:
ports:
  - "8081:8080"  # Use 8081 instead
```

Then update Flutter API URL:
```dart
static const String baseUrl = 'http://10.0.2.2:8081/api/v1';
```

### "Cannot connect to postgres"
```bash
# Check if database is healthy
docker-compose ps

# Should show:
# postgres    "healthy"

# If not healthy, check logs:
docker-compose logs postgres
```

### "API container exits immediately"
```bash
# Check API logs
docker-compose logs api

# Common issues:
# - Port conflict
# - Database not ready
# - Wrong environment variables
```

### "Migrations fail"
```bash
# Run migrations manually
docker-compose exec api ./migrate

# Or check migration files
ls d:\App\backend\migrations\
```

---

## Complete Setup Workflow

```
1. Open terminal in d:\App\backend\docker
   └─ cd d:\App\backend\docker

2. Build and start containers
   └─ docker-compose up --build
   └─ Wait for "Listening on :8080"

3. In another terminal, test health
   └─ curl http://localhost:8080/health
   └─ Should return success JSON

4. In Flutter terminal
   └─ r (hot reload)
   └─ App reconnects to new backend ✅

5. Try creating an event
   └─ Fill form and submit
   └─ ✅ Success! Event created!
   └─ ❌ Still error? Check logs:
      └─ docker-compose logs api
```

---

## Why This Works

### Before (No Docker):
```
❌ Need to install PostgreSQL
❌ Need to configure credentials
❌ Need to run migrations manually
❌ Need to set environment variables
❌ Easy to misconfigure
```

### With Docker:
```
✅ Everything pre-configured
✅ Isolated environment
✅ Reproducible setup
✅ Works on any machine
✅ Same setup as production
```

---

## Database Persistence

By default, Docker-Compose creates a named volume:
```
postgres_data: contains your database files
```

This means:
- ✅ Data persists when you stop containers
- ✅ Data survives restarts
- ✅ Delete with: docker-compose down -v

---

## View PostgreSQL Data

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U postgres -d college_events

# Common commands
\dt          # List tables
SELECT * FROM events;   # View events
\q           # Exit
```

---

## Next Steps

1. ✅ Ensure your code changes are saved (they are)
2. Run: `docker-compose up --build` (from docker folder)
3. Wait for API to be ready
4. Hot reload Flutter
5. Test creating an event
6. **✅ Done!**

---

## Why Your Fix Will Work with Docker

```
Your Changes:
├─ models.go: Added JSONTime ✅
├─ handlers.go: Updated handlers ✅
├─ Tests: All pass ✅

Docker:
├─ Reads Dockerfile
├─ Copies source code (includes your changes)
├─ Runs: go build ./cmd/api (compiles with your changes)
├─ Creates image with your compiled code
├─ Runs container with your binary ✅

Result:
└─ New backend running with your fixes! 🎉
```

The fix is **already in your code**. Docker just needs to **build and run it**.

---

## TL;DR

```bash
cd d:\App\backend\docker
docker-compose up --build
# Wait for "Listening on :8080"
# In Flutter: r
# Try creating event
# ✅ Works!
```

**That's literally all you need to do!**

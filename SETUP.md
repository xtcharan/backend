# Backend Setup Guide

This guide will help you set up and run the college event management backend on your Windows machine.

## Prerequisites Installation

### 1. Install Go (Required)

1. Download Go from: https://go.dev/dl/
2. Download the Windows installer (e.g., `go1.21.6.windows-amd64.msi`)
3. Run the installer and follow the prompts
4. Verify installation by opening a new terminal and running:
   ```powershell
   go version
   ```

### 2. Install Docker Desktop (Recommended for easy setup)

1. Download Docker Desktop from: https://www.docker.com/products/docker-desktop/
2. Install and start Docker Desktop
3. Verify installation:
   ```powershell
   docker --version
   docker-compose --version
   ```

### 3. Alternative: Install PostgreSQL and Redis Manually

If you don't want to use Docker:

**PostgreSQL:**
1. Download from: https://www.postgresql.org/download/windows/
2. Install with default settings
3. Remember the password you set for the `postgres` user

**Redis:**
1. Download from: https://github.com/tporadowski/redis/releases
2. Extract and run `redis-server.exe`

## Setup Instructions

### Option 1: Using Docker (Recommended)

This is the easiest way to get started as it includes PostgreSQL and Redis.

1. **Navigate to the backend directory:**
   ```powershell
   cd d:\App\backend
   ```

2. **Create environment file:**
   ```powershell
   copy .env.example .env
   ```

3. **Edit `.env` file** (optional, Docker Compose has defaults):
   - Update `JWT_SECRET` to a secure random string
   - Update `INITIAL_ADMIN_PASSWORD` if desired

4. **Start all services with Docker:**
   ```powershell
   cd docker
   docker-compose up -d
   ```

5. **Check if everything is running:**
   ```powershell
   docker-compose ps
   docker-compose logs api
   ```

6. **Access the API:**
   - Health check: http://localhost:8080/health
   - API base: http://localhost:8080/api/v1

7. **Stop services:**
   ```powershell
   docker-compose down
   ```

### Option 2: Running Locally (For Development)

1. **Install Go dependencies:**
   ```powershell
   cd d:\App\backend
   go mod download
   go mod tidy
   ```

2. **Create and configure `.env` file:**
   ```powershell
   copy .env.example .env
   notepad .env
   ```
   
   Update these values in `.env`:
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=yourDBPassword
   DB_NAME=college_events
   JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
   REDIS_HOST=localhost
   REDIS_PORT=6379
   INITIAL_ADMIN_EMAIL=admin@college.edu
   INITIAL_ADMIN_PASSWORD=admin123
   ```

3. **Create the database:**
   ```sql
   -- Connect to PostgreSQL and run:
   CREATE DATABASE college_events;
   ```

4. **Run database migrations:**
   ```powershell
   go run cmd/migrate/main.go
   ```

5. **Start the API server:**
   ```powershell
   go run cmd/api/main.go
   ```

6. **The API should now be running at:**
   - Health check: http://localhost:8080/health
   - API base: http://localhost:8080/api/v1

## Testing the API

### 1. Health Check
```powershell
curl http://localhost:8080/health
```

### 2. Register a User
```powershell
curl -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{
    "email": "student@college.edu",
    "password": "password123",
    "full_name": "John Doe",
    "department": "Computer Science",
    "year": 2
  }'
```

### 3. Login as Admin
```powershell
curl -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{
    "email": "admin@college.edu",
    "password": "admin123"
  }'
```

Save the `access_token` from the response.

### 4. Create an Event (Admin Only)
```powershell
$token = "YOUR_ACCESS_TOKEN_HERE"

curl -X POST http://localhost:8080/api/v1/admin/events `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $token" `
  -d '{
    "title": "Tech Festival 2024",
    "description": "Annual technology festival",
    "start_date": "2024-03-15T10:00:00Z",
    "end_date": "2024-03-15T18:00:00Z",
    "location": "Main Auditorium",
    "category": "Technology",
    "max_capacity": 500
  }'
```

### 5. List All Events (Public)
```powershell
curl http://localhost:8080/api/v1/events
```

## Project Structure

```
backend/
├── cmd/
│   ├── api/              # Main API server
│   │   └── main.go
│   └── migrate/          # Database migration tool
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/     # HTTP request handlers
│   │   │   ├── auth.go   # Auth endpoints
│   │   │   └── events.go # Event endpoints
│   │   ├── middleware/   # Middleware
│   │   │   ├── auth.go   # JWT validation
│   │   │   └── cors.go   # CORS configuration
│   │   └── router.go     # Route definitions
│   ├── services/
│   │   └── auth/         # Authentication service
│   │       └── service.go
│   └── models/
│       └── models.go     # Data models
├── pkg/
│   ├── config/           # Configuration management
│   │   └── config.go
│   └── database/         # Database utilities
│       └── connection.go
├── migrations/           # SQL migration files
│   └── 001_initial_schema.sql
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── .env.example          # Environment variables template
├── .gitignore
├── go.mod                # Go module definition
├── Makefile              # Build automation
└── README.md
```

## Available Make Commands

If you have `make` installed (via Git Bash or WSL), you can use:

```bash
make help          # Show all available commands
make install       # Install dependencies
make dev           # Run API server locally
make migrate       # Run database migrations
make build         # Build binaries
make docker-up     # Start Docker services
make docker-down   # Stop Docker services
make test          # Run tests
```

## Development Workflow

1. **Make changes to the code**
2. **Run the server** (it will auto-compile):
   ```powershell
   go run cmd/api/main.go
   ```
3. **Test your changes** with curl or Postman
4. **Commit your changes** to Git

## Common Issues

### "go: command not found"
- Make sure Go is installed and in your PATH
- Restart your terminal after installing Go

### Database connection failed
- Check if PostgreSQL is running
- Verify credentials in `.env` file
- Make sure the database exists

### Port already in use
- Change the `PORT` in `.env` file
- Or stop the process using port 8080:
  ```powershell
  netstat -ano | findstr :8080
  taskkill /PID <PID> /F
  ```

## Next Steps

1. **Install Go** if not already installed
2. **Choose setup option** (Docker recommended)
3. **Test the API endpoints** with the examples above
4. **Integrate with Flutter app** (see Flutter integration guide)
5. **Set up admin dashboard** (Next.js - coming next)

## Cloud Deployment

Once tested locally, you can deploy to GCP:

1. **Build Docker image:**
   ```bash
   docker build -f docker/Dockerfile -t gcr.io/YOUR_PROJECT/college-events-api .
   ```

2. **Push to Google Container Registry:**
   ```bash
   docker push gcr.io/YOUR_PROJECT/college-events-api
   ```

3. **Deploy to Cloud Run:**
   ```bash
   gcloud run deploy college-events-api \
     --image gcr.io/YOUR_PROJECT/college-events-api \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated
   ```

## Support

For issues or questions, check:
- API documentation in `README.md`
- Implementation plan in the artifacts directory
- Database schema in `migrations/001_initial_schema.sql`

# College Event Management Backend

A scalable, cloud-agnostic backend API built with Go for managing college events, clubs, announcements, and more.

## Features

- 🔐 JWT-based authentication with role-based access control
- 📅 Event management (CRUD, registration, capacity tracking)
- 🏢 Club management with department organization
- 💬 Real-time chat with WebSocket support
- 📢 Announcements system
- 👥 User management and profiles
- 🖼️ Media upload to cloud storage
- 📊 Analytics and reporting
- ⚡ Redis caching for performance
- 🔄 Cloud-agnostic design (GCP/AWS/Azure compatible)

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Storage**: S3-compatible (GCS/S3)
- **Authentication**: JWT

## Project Structure

```
backend/
├── cmd/
│   ├── api/              # Main API server
│   └── migrate/          # Database migrations
├── internal/
│   ├── api/              # HTTP handlers & routes
│   ├── services/         # Business logic
│   ├── repository/       # Data access layer
│   └── models/           # Data models
├── pkg/
│   ├── storage/          # Cloud storage abstraction
│   ├── database/         # Database utilities
│   └── config/           # Configuration management
├── migrations/           # SQL migration files
├── docker/              # Docker configuration
└── configs/             # Cloud-specific configs
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- Redis 7+
- GCP account (or AWS for migration)

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and update values
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run database migrations:
   ```bash
   go run cmd/migrate/main.go
   ```
5. Start the server:
   ```bash
   go run cmd/api/main.go
   ```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh JWT token

### Events (Public)
- `GET /api/v1/events` - List all events
- `GET /api/v1/events/:id` - Get event details

### Events (Protected)
- `POST /api/v1/events/:id/register` - Register for event
- `DELETE /api/v1/events/:id/register` - Unregister from event

### Events (Admin)
- `POST /api/v1/admin/events` - Create event
- `PUT /api/v1/admin/events/:id` - Update event
- `DELETE /api/v1/admin/events/:id` - Delete event

[See full API documentation](./docs/api.md)

## Deployment

### GCP (Cloud Run)
```bash
# Build Docker image
docker build -f docker/Dockerfile -t gcr.io/PROJECT_ID/college-events-api .

# Push to GCR
docker push gcr.io/PROJECT_ID/college-events-api

# Deploy to Cloud Run
gcloud run deploy college-events-api \
  --image gcr.io/PROJECT_ID/college-events-api \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### AWS (ECS/Fargate)
See [AWS Deployment Guide](./docs/aws-deployment.md)

## Testing

```bash
# Run unit tests
go test ./... -v

# Run with coverage
go test ./... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

## Cloud Migration

This backend is designed to be cloud-agnostic. To migrate from GCP to AWS:

1. Update environment variables (DB_HOST, STORAGE_PROVIDER, etc.)
2. Change Terraform configs from GCP to AWS modules
3. Deploy same Docker images to AWS services
4. Update DNS to point to new load balancer

See [Cloud Migration Guide](./docs/cloud-migration.md)

## License

MIT

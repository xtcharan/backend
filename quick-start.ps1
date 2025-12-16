# Quick Start Script for Windows
# Run this script to set up and start the backend with Docker

Write-Host "============================================"
Write-Host "College Event Management Backend Setup"
Write-Host "============================================"
Write-Host ""

# Check if Docker is installed
try {
    docker --version | Out-Null
    Write-Host "✓ Docker is installed"
} catch {
    Write-Host "✗ Docker is not installed"
    Write-Host "Please install Docker Desktop from: https://www.docker.com/products/docker-desktop/"
    exit 1
}

# Check if Docker is running
try {
    docker ps | Out-Null
    Write-Host "✓ Docker is running"
} catch {
    Write-Host "✗ Docker is not running"
    Write-Host "Please start Docker Desktop and try again"
    exit 1
}

# Create .env file if it doesn't exist
if (-not (Test-Path ".env")) {
    Write-Host "Creating .env file from template..."
    Copy-Item ".env.example" ".env"
    Write-Host "✓ Created .env file"
    Write-Host "⚠ Please update .env with your configuration if needed"
} else {
    Write-Host "✓ .env file already exists"
}

Write-Host ""
Write-Host "Starting Docker services..."
Write-Host ""

# Navigate to docker directory and start services
Set-Location docker
docker-compose up -d

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "============================================"  
    Write-Host "✓ Backend is running!"
    Write-Host "============================================"
    Write-Host ""
    Write-Host "API Health Check: http://localhost:8080/health"
    Write-Host "API Base URL:     http://localhost:8080/api/v1"
    Write-Host ""
    Write-Host "Default Admin Credentials:"
    Write-Host "  Email:    admin@college.edu"
    Write-Host "  Password: admin123"
    Write-Host ""
    Write-Host "To view logs:     docker-compose logs -f"
    Write-Host "To stop services: docker-compose down"
    Write-Host "============================================"
} else {
    Write-Host ""
    Write-Host "✗ Failed to start services"
    Write-Host "Check the error messages above for details"
    exit 1
}

Set-Location ..

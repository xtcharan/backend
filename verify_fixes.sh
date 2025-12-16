#!/bin/bash
# Verification Script for Backend Events Bug Fixes

echo "=========================================="
echo "Backend Events Bug Fixes Verification"
echo "=========================================="
echo ""

echo "1. Checking if models.go has JSONTime type..."
if grep -q "type JSONTime time.Time" internal/models/models.go; then
    echo "   ✅ JSONTime type found"
else
    echo "   ❌ JSONTime type NOT found"
    exit 1
fi

echo ""
echo "2. Checking if UnmarshalJSON is implemented..."
if grep -q "func (jt \*JSONTime) UnmarshalJSON" internal/models/models.go; then
    echo "   ✅ UnmarshalJSON method found"
else
    echo "   ❌ UnmarshalJSON method NOT found"
    exit 1
fi

echo ""
echo "3. Checking if MarshalJSON is implemented..."
if grep -q "func (jt JSONTime) MarshalJSON" internal/models/models.go; then
    echo "   ✅ MarshalJSON method found"
else
    echo "   ❌ MarshalJSON method NOT found"
    exit 1
fi

echo ""
echo "4. Checking if CreateEventRequest uses JSONTime..."
if grep -q "StartDate   JSONTime" internal/models/models.go; then
    echo "   ✅ CreateEventRequest updated to use JSONTime"
else
    echo "   ❌ CreateEventRequest NOT using JSONTime"
    exit 1
fi

echo ""
echo "5. Running tests..."
if go test ./internal/models -v; then
    echo "   ✅ All tests passed"
else
    echo "   ❌ Tests failed"
    exit 1
fi

echo ""
echo "6. Checking event handler improvements..."
if grep -q "startTime := req.StartDate.Time()" internal/api/handlers/events.go; then
    echo "   ✅ CreateEvent handler updated with JSONTime conversion"
else
    echo "   ❌ CreateEvent handler NOT updated"
    exit 1
fi

echo ""
echo "7. Checking for end_date validation..."
if grep -q 'Error:   strPtr("end_date must be after start_date")' internal/api/handlers/events.go; then
    echo "   ✅ Time validation added to handlers"
else
    echo "   ❌ Time validation NOT found"
    exit 1
fi

echo ""
echo "=========================================="
echo "✅ All verifications passed!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Build the backend: go build ./cmd/api"
echo "2. Run tests: go test ./internal/models -v"
echo "3. Set up PostgreSQL with credentials in .env"
echo "4. Run migrations: go run ./cmd/migrate/main.go"
echo "5. Start the server: go run ./cmd/api/main.go"
echo "6. Test with Flutter app to create events"

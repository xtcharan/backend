Backend Events Bug Fixes - December 16, 2025
=============================================

## Issues Fixed

### 1. **Invalid DateTime Parsing Error**
**Error Message:** `Error: invalid request body parsing time "2025-12-16T21:26:00.000Z" as "2006-01-02T15:04:05Z07:00"`

**Root Cause:** 
- Flutter app sends datetime in format: `2025-12-16T21:26:00.000Z` (ISO8601 with milliseconds)
- Go backend expected RFC3339 format with timezone offset: `Z07:00`
- The timezone designator "Z" (UTC) doesn't match the Go parsing format `Z07:00`

### 2. **JSON Binding Issues**
- Default Gin JSON binding couldn't handle millisecond precision variations
- Multiple datetime formats from different clients weren't supported

---

## Solutions Implemented

### A. Custom JSONTime Type (models.go)

Created a custom `JSONTime` type with:

```go
type JSONTime time.Time
```

**Features:**
- ✅ Custom `UnmarshalJSON()` - Parses multiple datetime formats
- ✅ Custom `MarshalJSON()` - Outputs RFC3339Nano format
- ✅ Conversion helper - `Time()` method to get underlying `time.Time`

**Supported Input Formats:**
1. `2025-12-16T21:26:00.000Z` - Flutter/ISO8601 with milliseconds (PRIMARY)
2. `2025-12-16T21:26:00Z` - ISO8601 without milliseconds
3. `2025-12-16T21:26:00` - Date and time only

**Output Format:**
- `2025-12-16T21:26:00Z` - RFC3339Nano (compatible with all clients)

### B. Updated CreateEventRequest Struct (models.go)

```go
type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description *string   `json:"description"`
	ImageURL    *string   `json:"image_url"`
	StartDate   JSONTime  `json:"start_date" binding:"required"`  // Changed from time.Time
	EndDate     JSONTime  `json:"end_date" binding:"required"`    // Changed from time.Time
	Location    *string   `json:"location"`
	Category    *string   `json:"category"`
	MaxCapacity *int      `json:"max_capacity"`
}
```

### C. Improved Event Handlers (events.go)

**CreateEvent Handler:**
- ✅ Convert `JSONTime` to `time.Time` before database insertion
- ✅ Validate end_date > start_date with proper error messages
- ✅ Enhanced logging with `fmt.Printf` for debugging
- ✅ More informative error responses

**UpdateEvent Handler:**
- ✅ Same improvements as CreateEvent
- ✅ Added `updated_at` timestamp update
- ✅ Consistent datetime validation

---

## Testing

### Unit Tests Created (datetime_test.go)

Three comprehensive test suites:

1. **TestJSONTimeUnmarshal** ✅
   - Tests multiple input datetime formats
   - Validates error handling for invalid formats
   - All 5 test cases pass

2. **TestJSONTimeMarshal** ✅
   - Verifies output is valid RFC3339 format
   - Ensures round-trip compatibility

3. **TestCreateEventRequestUnmarshal** ✅
   - End-to-end event creation parsing
   - Validates all fields are properly parsed
   - Checks start_date < end_date logic

### Test Results:
```
PASS: TestJSONTimeUnmarshal (all sub-tests)
PASS: TestJSONTimeMarshal
PASS: TestCreateEventRequestUnmarshal
```

---

## Flutter App Compatibility

The Flutter `Event` model uses `toIso8601String()`:

```dart
Map<String, dynamic> toJson() {
  return {
    'start_date': startDate.toIso8601String(),  // Outputs: 2025-12-16T21:26:00.000Z
    'end_date': endDate.toIso8601String(),
    // ... other fields
  };
}
```

✅ **Perfect Alignment:** This format is now fully supported by the backend!

---

## Database Schema

No changes needed - PostgreSQL TIMESTAMP columns handle all UTC datetimes correctly:

```sql
CREATE TABLE events (
    ...
    start_date TIMESTAMP NOT NULL,  -- Stores as UTC
    end_date TIMESTAMP NOT NULL,
    ...
)
```

---

## Files Modified

1. **internal/models/models.go**
   - Added `encoding/json` and `fmt` imports
   - Added `JSONTime` type with marshal/unmarshal methods
   - Updated `CreateEventRequest` struct to use `JSONTime`

2. **internal/api/handlers/events.go**
   - Updated `CreateEvent()` handler with conversions and validation
   - Updated `UpdateEvent()` handler with conversions and validation
   - Enhanced error logging and messages

3. **internal/models/datetime_test.go** (NEW)
   - Comprehensive test coverage for datetime parsing
   - All tests passing

---

## How to Test

### 1. Build the Backend
```bash
cd backend
go build ./cmd/api
```

### 2. Run Tests
```bash
go test ./internal/models -v
```

### 3. Test with Flutter App
Create an event from the Flutter UI - it will now send:
```json
{
  "title": "Test Event",
  "description": "...",
  "start_date": "2025-12-16T21:26:00.000Z",
  "end_date": "2025-12-16T23:26:00.000Z",
  ...
}
```

✅ Backend will parse successfully!

---

## Error Handling Improvements

Before:
```
Error: invalid request body: parsing time "2025-12-16T21:26:00.000Z" as "2006-01-02T15:04:05Z07:00": cannot parse "Z" as "Z07:00"
```

After:
```
✅ Event created successfully (for valid datetimes)

❌ Error: invalid request body: unable to parse datetime '2025-12-16T21:26:00.000Z' in any supported format: ... (for invalid formats)

❌ Error: end_date must be after start_date (for time validation)
```

---

## Performance Impact

- ✅ Minimal - custom marshaling/unmarshaling is only during request/response
- ✅ No database changes needed
- ✅ No impact on query performance

---

## Backward Compatibility

- ✅ Fully compatible with existing Flutter app
- ✅ Supports multiple datetime formats (forwards compatible)
- ✅ No API contract changes - same endpoint structure

---

## Next Steps (Optional)

1. Set up PostgreSQL locally to test database operations
2. Add integration tests with full event CRUD flow
3. Add API documentation with example requests/responses
4. Consider adding timezone support if needed by business logic

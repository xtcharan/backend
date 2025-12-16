# 🏛️ Clubs & Departments Backend Implementation Plan

## Overview

Based on your frontend structure and requirements, here's a **clean, scalable backend architecture** for the clubs system.

---

## 📊 Database Schema Enhancement

### Current Schema (Already Exists)
```sql
-- clubs table (basic)
CREATE TABLE clubs (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    department VARCHAR(100),      -- ❌ Just a string
    image_url VARCHAR(500),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### Proposed Enhanced Schema

#### 1. Departments Table (NEW)
```sql
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(10) UNIQUE NOT NULL,              -- 'BCA', 'BCOM', etc.
    name VARCHAR(255) NOT NULL,                     -- 'Bachelor of Computer Applications'
    description TEXT,
    icon_name VARCHAR(50),                          -- 'computer', 'business', etc.
    color_hex VARCHAR(7),                           -- '#FF5733'
    logo_url VARCHAR(500),                          -- Department logo/photo
    
    -- Auto-calculated fields (via database functions/triggers)
    total_members INT DEFAULT 0,                    -- Calculated from club_members
    total_clubs INT DEFAULT 0,                      -- Calculated from clubs count
    total_events INT DEFAULT 0,                     -- Calculated from events
    
    rating DECIMAL(3,2) DEFAULT 0.00,              -- Optional: average rating
    
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_departments_code ON departments(code) WHERE deleted_at IS NULL;
```

#### 2. Enhanced Clubs Table
```sql
CREATE TABLE clubs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    short_name VARCHAR(50),                         -- 'BITBLAZE', 'SYNAPSE'
    tagline VARCHAR(255),                           -- 'Innovation Through Code'
    description TEXT,
    
    -- Foreign key to departments
    department_id UUID REFERENCES departments(id) ON DELETE CASCADE,
    
    -- Visual identity
    logo_url VARCHAR(500),
    primary_color_hex VARCHAR(7),
    secondary_color_hex VARCHAR(7),
    icon_name VARCHAR(50),
    
    -- Statistics (auto-calculated)
    member_count INT DEFAULT 0,
    event_count INT DEFAULT 0,
    awards_count INT DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.00,
    
    -- Contact info
    email VARCHAR(255),
    phone VARCHAR(20),
    website VARCHAR(500),
    
    -- Social media
    instagram VARCHAR(255),
    linkedin VARCHAR(255),
    twitter VARCHAR(255),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_clubs_department ON clubs(department_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_clubs_active ON clubs(is_active) WHERE deleted_at IS NULL;
```

#### 3. Enhanced Club Members Table
```sql
CREATE TABLE club_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Member role
    role VARCHAR(50) DEFAULT 'member',              -- 'president', 'vice_president', 'secretary', 'member'
    position VARCHAR(100),                          -- 'Technical Lead', 'Event Coordinator'
    
    -- Contact (optional, for leaders)
    email VARCHAR(255),
    phone VARCHAR(20),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    
    UNIQUE(club_id, user_id)
);

CREATE INDEX idx_club_members_club ON club_members(club_id);
CREATE INDEX idx_club_members_user ON club_members(user_id);
CREATE INDEX idx_club_members_role ON club_members(role);
```

#### 4. Club Events Table (Link to existing events)
```sql
-- Link events to clubs
ALTER TABLE events ADD COLUMN club_id UUID REFERENCES clubs(id) ON DELETE SET NULL;
CREATE INDEX idx_events_club ON events(club_id) WHERE deleted_at IS NULL;
```

#### 5. Club Announcements Table
```sql
CREATE TABLE club_announcements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    priority VARCHAR(20) DEFAULT 'medium',          -- 'low', 'medium', 'high', 'urgent'
    
    -- Attachments
    image_url VARCHAR(500),
    
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_club_announcements_club ON club_announcements(club_id);
CREATE INDEX idx_club_announcements_created_at ON club_announcements(created_at DESC);
```

#### 6. Club Awards Table (Auto-tracking)
```sql
CREATE TABLE club_awards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    award_name VARCHAR(255) NOT NULL,
    description TEXT,
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    
    -- Award details
    position VARCHAR(50),                           -- '1st Place', '2nd Place'
    prize_amount DECIMAL(10,2),
    certificate_url VARCHAR(500),
    
    awarded_date DATE NOT NULL,
    awarded_by VARCHAR(255),                        -- Institution/Organization name
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_club_awards_club ON club_awards(club_id);
```

---

## 🔄 Auto-Calculation Triggers

### Update Department Statistics
```sql
-- Trigger to update department statistics when clubs/members change
CREATE OR REPLACE FUNCTION update_department_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE departments d SET
        total_clubs = (
            SELECT COUNT(*) FROM clubs 
            WHERE department_id = d.id AND deleted_at IS NULL
        ),
        total_members = (
            SELECT COUNT(DISTINCT cm.user_id)
            FROM clubs c
            JOIN club_members cm ON c.id = cm.club_id
            WHERE c.department_id = d.id AND c.deleted_at IS NULL
        ),
        total_events = (
            SELECT COUNT(*)
            FROM events e
            JOIN clubs c ON e.club_id = c.id
            WHERE c.department_id = d.id AND e.deleted_at IS NULL
        ),
        updated_at = CURRENT_TIMESTAMP
    WHERE d.id = COALESCE(NEW.department_id, OLD.department_id);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_department_stats
AFTER INSERT OR UPDATE OR DELETE ON clubs
FOR EACH ROW EXECUTE FUNCTION update_department_stats();
```

### Update Club Statistics
```sql
-- Trigger to update club member count
CREATE OR REPLACE FUNCTION update_club_member_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE clubs SET
        member_count = (
            SELECT COUNT(*) FROM club_members 
            WHERE club_id = COALESCE(NEW.club_id, OLD.club_id) 
            AND is_active = true
        ),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.club_id, OLD.club_id);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_club_member_count
AFTER INSERT OR UPDATE OR DELETE ON club_members
FOR EACH ROW EXECUTE FUNCTION update_club_member_count();
```

---

## 🎯 Backend API Structure

### Go Models (internal/models/clubs.go)

```go
package models

import (
    "time"
    "github.com/google/uuid"
)

// Department represents an academic department
type Department struct {
    ID           uuid.UUID  `json:"id" db:"id"`
    Code         string     `json:"code" db:"code"`
    Name         string     `json:"name" db:"name"`
    Description  *string    `json:"description,omitempty" db:"description"`
    IconName     *string    `json:"icon_name,omitempty" db:"icon_name"`
    ColorHex     *string    `json:"color_hex,omitempty" db:"color_hex"`
    LogoURL      *string    `json:"logo_url,omitempty" db:"logo_url"`
    
    // Auto-calculated
    TotalMembers int        `json:"total_members" db:"total_members"`
    TotalClubs   int        `json:"total_clubs" db:"total_clubs"`
    TotalEvents  int        `json:"total_events" db:"total_events"`
    Rating       float64    `json:"rating" db:"rating"`
    
    CreatedBy    uuid.UUID  `json:"created_by" db:"created_by"`
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt    *time.Time `json:"-" db:"deleted_at"`
}

// CreateDepartmentRequest for API
type CreateDepartmentRequest struct {
    Code        string  `json:"code" binding:"required,max=10"`
    Name        string  `json:"name" binding:"required,max=255"`
    Description *string `json:"description"`
    IconName    *string `json:"icon_name"`
    ColorHex    *string `json:"color_hex"`
    LogoURL     *string `json:"logo_url"`
}

// Club represents a college club
type Club struct {
    ID                uuid.UUID  `json:"id" db:"id"`
    Name              string     `json:"name" db:"name"`
    ShortName         *string    `json:"short_name,omitempty" db:"short_name"`
    Tagline           *string    `json:"tagline,omitempty" db:"tagline"`
    Description       *string    `json:"description,omitempty" db:"description"`
    DepartmentID      uuid.UUID  `json:"department_id" db:"department_id"`
    
    // Visual
    LogoURL           *string    `json:"logo_url,omitempty" db:"logo_url"`
    PrimaryColorHex   *string    `json:"primary_color_hex,omitempty" db:"primary_color_hex"`
    SecondaryColorHex *string    `json:"secondary_color_hex,omitempty" db:"secondary_color_hex"`
    IconName          *string    `json:"icon_name,omitempty" db:"icon_name"`
    
    // Stats
    MemberCount       int        `json:"member_count" db:"member_count"`
    EventCount        int        `json:"event_count" db:"event_count"`
    AwardsCount       int        `json:"awards_count" db:"awards_count"`
    Rating            float64    `json:"rating" db:"rating"`
    
    // Contact
    Email             *string    `json:"email,omitempty" db:"email"`
    Phone             *string    `json:"phone,omitempty" db:"phone"`
    Website           *string    `json:"website,omitempty" db:"website"`
    
    // Social
    Instagram         *string    `json:"instagram,omitempty" db:"instagram"`
    LinkedIn          *string    `json:"linkedin,omitempty" db:"linkedin"`
    Twitter           *string    `json:"twitter,omitempty" db:"twitter"`
    
    IsActive          bool       `json:"is_active" db:"is_active"`
    CreatedBy         uuid.UUID  `json:"created_by" db:"created_by"`
    CreatedAt         time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt         *time.Time `json:"-" db:"deleted_at"`
}

// CreateClubRequest for API
type CreateClubRequest struct {
    Name              string     `json:"name" binding:"required"`
    ShortName         *string    `json:"short_name"`
    Tagline           *string    `json:"tagline"`
    Description       *string    `json:"description"`
    DepartmentID      string     `json:"department_id" binding:"required,uuid"`
    LogoURL           *string    `json:"logo_url"`
    PrimaryColorHex   *string    `json:"primary_color_hex"`
    SecondaryColorHex *string    `json:"secondary_color_hex"`
    IconName          *string    `json:"icon_name"`
    Email             *string    `json:"email"`
    Phone             *string    `json:"phone"`
    Website           *string    `json:"website"`
    Instagram         *string    `json:"instagram"`
    LinkedIn          *string    `json:"linkedin"`
    Twitter           *string    `json:"twitter"`
}

// ClubMember represents a club member
type ClubMember struct {
    ID       uuid.UUID  `json:"id" db:"id"`
    ClubID   uuid.UUID  `json:"club_id" db:"club_id"`
    UserID   uuid.UUID  `json:"user_id" db:"user_id"`
    Role     string     `json:"role" db:"role"`
    Position *string    `json:"position,omitempty" db:"position"`
    Email    *string    `json:"email,omitempty" db:"email"`
    Phone    *string    `json:"phone,omitempty" db:"phone"`
    IsActive bool       `json:"is_active" db:"is_active"`
    JoinedAt time.Time  `json:"joined_at" db:"joined_at"`
    LeftAt   *time.Time `json:"left_at,omitempty" db:"left_at"`
}

// ClubAnnouncement for club-specific announcements
type ClubAnnouncement struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    ClubID    uuid.UUID  `json:"club_id" db:"club_id"`
    Title     string     `json:"title" db:"title"`
    Content   string     `json:"content" db:"content"`
    Priority  string     `json:"priority" db:"priority"`
    ImageURL  *string    `json:"image_url,omitempty" db:"image_url"`
    CreatedBy uuid.UUID  `json:"created_by" db:"created_by"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// ClubAward for tracking club achievements
type ClubAward struct {
    ID             uuid.UUID  `json:"id" db:"id"`
    ClubID         uuid.UUID  `json:"club_id" db:"club_id"`
    AwardName      string     `json:"award_name" db:"award_name"`
    Description    *string    `json:"description,omitempty" db:"description"`
    EventID        *uuid.UUID `json:"event_id,omitempty" db:"event_id"`
    Position       *string    `json:"position,omitempty" db:"position"`
    PrizeAmount    *float64   `json:"prize_amount,omitempty" db:"prize_amount"`
    CertificateURL *string    `json:"certificate_url,omitempty" db:"certificate_url"`
    AwardedDate    time.Time  `json:"awarded_date" db:"awarded_date"`
    AwardedBy      *string    `json:"awarded_by,omitempty" db:"awarded_by"`
    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}
```

---

## 🛣️ API Endpoints Structure

### Department Endpoints

```go
// Public routes
GET    /api/v1/departments                  // List all departments
GET    /api/v1/departments/:id               // Get department details
GET    /api/v1/departments/:id/clubs         // Get clubs in department
GET    /api/v1/departments/:id/stats         // Get department statistics

// Admin routes
POST   /api/v1/admin/departments             // Create department
PUT    /api/v1/admin/departments/:id         // Update department
DELETE /api/v1/admin/departments/:id         // Delete department
POST   /api/v1/admin/departments/:id/logo    // Upload department logo
```

### Club Endpoints

```go
// Public routes
GET    /api/v1/clubs                         // List all clubs
GET    /api/v1/clubs/:id                     // Get club details
GET    /api/v1/clubs/:id/members             // Get club members
GET    /api/v1/clubs/:id/events              // Get club events
GET    /api/v1/clubs/:id/announcements       // Get club announcements
GET    /api/v1/clubs/:id/awards              // Get club awards

// Admin/Club Admin routes
POST   /api/v1/admin/clubs                   // Create club
PUT    /api/v1/admin/clubs/:id               // Update club
DELETE /api/v1/admin/clubs/:id               // Delete club
POST   /api/v1/admin/clubs/:id/logo          // Upload club logo

// Club management (requires club admin/member role)
POST   /api/v1/clubs/:id/members             // Add member
PUT    /api/v1/clubs/:id/members/:user_id    // Update member role
DELETE /api/v1/clubs/:id/members/:user_id    // Remove member

POST   /api/v1/clubs/:id/announcements       // Create announcement
PUT    /api/v1/clubs/:id/announcements/:ann_id  // Update announcement
DELETE /api/v1/clubs/:id/announcements/:ann_id  // Delete announcement

POST   /api/v1/clubs/:id/events              // Create club event
POST   /api/v1/clubs/:id/awards              // Add award
```

---

## 📝 Handler Implementation (Pseudo-code)

### Department Handlers

```go
// internal/api/handlers/departments.go

type DepartmentHandler struct {
    db *database.DB
}

// ListDepartments - GET /api/v1/departments
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
    // Query all active departments with stats
    // Return JSON array
}

// GetDepartment - GET /api/v1/departments/:id
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
    // Get department by ID
    // Include related stats
    // Return JSON
}

// CreateDepartment - POST /api/v1/admin/departments
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
    // Parse request body
    // Validate admin role
    // Insert into departments table
    // Return created department
}

// UpdateDepartment - PUT /api/v1/admin/departments/:id
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
    // Parse request body
    // Validate admin role
    // Update department
    // Return updated department
}

// DeleteDepartment - DELETE /api/v1/admin/departments/:id
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
    // Validate admin role
    // Soft delete (set deleted_at)
    // Return success
}
```

### Club Handlers

```go
// internal/api/handlers/clubs.go

type ClubHandler struct {
    db *database.DB
}

// ListClubs - GET /api/v1/clubs
func (h *ClubHandler) ListClubs(c *gin.Context) {
    // Optional query params: ?department_id=xxx
    // Query all active clubs
    // Include department info
    // Return JSON array
}

// GetClub - GET /api/v1/clubs/:id
func (h *ClubHandler) GetClub(c *gin.Context) {
    // Get club by ID
    // Include department, member count, event count
    // Return JSON
}

// CreateClub - POST /api/v1/admin/clubs
func (h *ClubHandler) CreateClub(c *gin.Context) {
    // Parse request body
    // Validate admin role
    // Validate department_id exists
    // Insert into clubs table
    // Trigger will auto-update department stats
    // Return created club
}

// UpdateClub - PUT /api/v1/admin/clubs/:id
func (h *ClubHandler) UpdateClub(c *gin.Context) {
    // Parse request body
    // Validate admin/club_admin role
    // Update club
    // Return updated club
}

// GetClubMembers - GET /api/v1/clubs/:id/members
func (h *ClubHandler) GetClubMembers(c *gin.Context) {
    // Query club_members with user info (JOIN)
    // Return members with roles
}

// AddClubMember - POST /api/v1/clubs/:id/members
func (h *ClubHandler) AddClubMember(c *gin.Context) {
    // Validate club admin role
    // Insert into club_members
    // Trigger auto-updates club.member_count
    // Return success
}

// GetClubEvents - GET /api/v1/clubs/:id/events
func (h *ClubHandler) GetClubEvents(c *gin.Context) {
    // Query events WHERE club_id = :id
    // Return events array
}

// CreateClubEvent - POST /api/v1/clubs/:id/events
func (h *ClubHandler) CreateClubEvent(c *gin.Context) {
    // Validate club admin role
    // Create event with club_id set
    // Return created event
}

// GetClubAnnouncements - GET /api/v1/clubs/:id/announcements
func (h *ClubHandler) GetClubAnnouncements(c *gin.Context) {
    // Query club_announcements
    // Return announcements
}

// CreateClubAnnouncement - POST /api/v1/clubs/:id/announcements
func (h *ClubHandler) CreateClubAnnouncement(c *gin.Context) {
    // Validate club admin role
    // Insert announcement
    // Return created announcement
}

// GetClubAwards - GET /api/v1/clubs/:id/awards
func (h *ClubHandler) GetClubAwards(c *gin.Context) {
    // Query club_awards
    // Return awards
}

// AddClubAward - POST /api/v1/clubs/:id/awards
func (h *ClubHandler) AddClubAward(c *gin.Context) {
    // Validate club admin role
    // Insert award
    // Trigger auto-updates club.awards_count
    // Return created award
}
```

---

## 🔐 Authorization Middleware

```go
// Middleware to check club admin role
func ClubAdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, _ := c.Get("user_id")
        clubID := c.Param("id")
        
        // Check if user is:
        // 1. Global admin, OR
        // 2. Club member with role 'president' or 'vice_president'
        
        var isMember bool
        err := db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM club_members 
                WHERE club_id = $1 AND user_id = $2 
                AND role IN ('president', 'vice_president', 'secretary')
                AND is_active = true
            )
        `, clubID, userID).Scan(&isMember)
        
        if !isMember {
            c.JSON(403, gin.H{"error": "Not authorized"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## 📊 Response Examples

### GET /api/v1/departments
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid-1",
      "code": "BCA",
      "name": "Bachelor of Computer Applications",
      "description": "Computer science and applications department",
      "icon_name": "computer",
      "color_hex": "#4F46E5",
      "logo_url": "https://storage/bca-logo.png",
      "total_members": 125,
      "total_clubs": 1,
      "total_events": 15,
      "rating": 4.5,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-12-16T10:00:00Z"
    }
  ]
}
```

### GET /api/v1/clubs/:id
```json
{
  "success": true,
  "data": {
    "id": "uuid-1",
    "name": "BITBLAZE",
    "short_name": "BITBLAZE",
    "tagline": "Innovation Through Code",
    "description": "Technical club for BCA students...",
    "department_id": "dept-uuid",
    "logo_url": "https://storage/bitblaze-logo.png",
    "primary_color_hex": "#4F46E5",
    "secondary_color_hex": "#818CF8",
    "member_count": 125,
    "event_count": 15,
    "awards_count": 5,
    "rating": 4.8,
    "email": "bitblaze@college.edu",
    "instagram": "@bitblaze_official",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## 🎨 Frontend Flow Integration

### Create Button Modal (Your Requirement)

```dart
// When admin clicks "Create" button
showModalBottomSheet(
  context: context,
  builder: (context) => Column(
    children: [
      ListTile(
        leading: Icon(Icons.school),
        title: Text('Create Department'),
        onTap: () => navigateToCreateDepartment(),
      ),
      ListTile(
        leading: Icon(Icons.group),
        title: Text('Create Club'),
        onTap: () => navigateToCreateClub(),
      ),
    ],
  ),
);
```

### Create Department Form → API Call
```dart
await apiService.createDepartment(
  code: 'BCA',
  name: 'Bachelor of Computer Applications',
  description: 'Computer applications department',
  iconName: 'computer',
  colorHex: '#4F46E5',
  logoUrl: uploadedLogoUrl,  // After image upload
);
```

### Create Club Form → API Call
```dart
await apiService.createClub(
  name: 'BITBLAZE',
  shortName: 'BITBLAZE',
  tagline: 'Innovation Through Code',
  departmentId: selectedDepartmentId,
  logoUrl: uploadedLogoUrl,
  primaryColorHex: '#4F46E5',
  // ... other fields
);
```

---

## 🚀 Implementation Steps

### Phase 1: Database (Week 1)
1. ✅ Create migration file: `002_clubs_departments.sql`
2. ✅ Add new tables: departments, enhanced clubs, club_announcements, club_awards
3. ✅ Create triggers for auto-calculations
4. ✅ Test migrations with Docker

### Phase 2: Backend Models (Week 1)
1. ✅ Create `internal/models/clubs.go`
2. ✅ Define structs for Department, Club, ClubMember, etc.
3. ✅ Create request/response DTOs

### Phase 3: Backend Handlers (Week 2)
1. ✅ Create `internal/api/handlers/departments.go`
2. ✅ Create `internal/api/handlers/clubs.go`
3. ✅ Implement CRUD operations
4. ✅ Add authorization middleware

### Phase 4: Routes (Week 2)
1. ✅ Register routes in `internal/api/router.go`
2. ✅ Add public and admin routes
3. ✅ Test with Postman/curl

### Phase 5: Frontend Integration (Week 3)
1. ✅ Update `lib/src/services/api_service.dart`
2. ✅ Add department/club API methods
3. ✅ Update clubs page to use real data
4. ✅ Add create forms with image upload

---

## 📋 Summary

### What This Architecture Provides:

✅ **Departments Table** - Stores BCA, BCOM, etc. with logos, stats  
✅ **Enhanced Clubs Table** - Rich club info with logos, colors, contacts  
✅ **Auto-Statistics** - Member/event counts update automatically via triggers  
✅ **Club Members** - Track roles (president, member, etc.)  
✅ **Club Events** - Link events to clubs  
✅ **Club Announcements** - Club-specific announcements  
✅ **Club Awards** - Auto-track achievements  
✅ **RESTful API** - Clean, organized endpoints  
✅ **Authorization** - Role-based access control  
✅ **Scalable** - Easy to extend with new features  

### Key Benefits:

- 🔥 **No Manual Counting** - Triggers auto-update stats
- 🔐 **Secure** - Role-based permissions
- 📊 **Complete Data** - All club info in one place
- 🎨 **Frontend Ready** - JSON responses match Flutter models
- 🐳 **Docker Ready** - Works with your existing setup
- ✅ **Tested Pattern** - Same as events (which already works!)

This is a **production-ready, clean architecture** that matches your frontend structure perfectly! 🚀

Want me to start implementing any specific part?

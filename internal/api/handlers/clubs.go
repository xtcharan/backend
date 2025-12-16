package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/college-event-backend/internal/models"
)

type ClubHandler struct {
	DB *sql.DB
}

// GetClubs retrieves all clubs
func (h *ClubHandler) GetClubs(c *gin.Context) {
	query := `
		SELECT id, department_id, name, tagline, description, logo_url,
		       primary_color, secondary_color, member_count, event_count,
		       awards_count, rating, email, phone, website, social_links,
		       created_at, updated_at
		FROM clubs
		ORDER BY name ASC
	`

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clubs"})
		return
	}
	defer rows.Close()

	clubs := []models.Club{}
	for rows.Next() {
		var club models.Club
		if err := rows.Scan(
			&club.ID, &club.DepartmentID, &club.Name, &club.Tagline, &club.Description,
			&club.LogoURL, &club.PrimaryColor, &club.SecondaryColor, &club.MemberCount,
			&club.EventCount, &club.AwardsCount, &club.Rating, &club.Email, &club.Phone,
			&club.Website, &club.SocialLinks, &club.CreatedAt, &club.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan club"})
			return
		}
		clubs = append(clubs, club)
	}

	c.JSON(http.StatusOK, gin.H{"data": clubs})
}

// GetClub retrieves a single club by ID
func (h *ClubHandler) GetClub(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	query := `
		SELECT id, department_id, name, tagline, description, logo_url,
		       primary_color, secondary_color, member_count, event_count,
		       awards_count, rating, email, phone, website, social_links,
		       created_at, updated_at
		FROM clubs
		WHERE id = $1
	`

	var club models.Club
	err = h.DB.QueryRow(query, clubID).Scan(
		&club.ID, &club.DepartmentID, &club.Name, &club.Tagline, &club.Description,
		&club.LogoURL, &club.PrimaryColor, &club.SecondaryColor, &club.MemberCount,
		&club.EventCount, &club.AwardsCount, &club.Rating, &club.Email, &club.Phone,
		&club.Website, &club.SocialLinks, &club.CreatedAt, &club.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Club not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch club"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": club})
}

// CreateClub creates a new club (admin only)
func (h *ClubHandler) CreateClub(c *gin.Context) {
	var req models.CreateClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	primaryColor := "#4F46E5"
	secondaryColor := "#818CF8"
	if req.PrimaryColor != nil {
		primaryColor = *req.PrimaryColor
	}
	if req.SecondaryColor != nil {
		secondaryColor = *req.SecondaryColor
	}

	query := `
		INSERT INTO clubs (
			department_id, name, tagline, description, logo_url,
			primary_color, secondary_color, email, phone, website, social_links
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, department_id, name, tagline, description, logo_url,
		          primary_color, secondary_color, member_count, event_count,
		          awards_count, rating, email, phone, website, social_links,
		          created_at, updated_at
	`

	var club models.Club
	err := h.DB.QueryRow(
		query, req.DepartmentID, req.Name, req.Tagline, req.Description,
		req.LogoURL, primaryColor, secondaryColor, req.Email, req.Phone,
		req.Website, req.SocialLinks,
	).Scan(
		&club.ID, &club.DepartmentID, &club.Name, &club.Tagline, &club.Description,
		&club.LogoURL, &club.PrimaryColor, &club.SecondaryColor, &club.MemberCount,
		&club.EventCount, &club.AwardsCount, &club.Rating, &club.Email, &club.Phone,
		&club.Website, &club.SocialLinks, &club.CreatedAt, &club.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create club"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": club})
}

// UpdateClub updates a club (admin only)
func (h *ClubHandler) UpdateClub(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	var req models.UpdateClubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE clubs
		SET department_id = COALESCE($1, department_id),
		    name = COALESCE($2, name),
		    tagline = COALESCE($3, tagline),
		    description = COALESCE($4, description),
		    logo_url = COALESCE($5, logo_url),
		    primary_color = COALESCE($6, primary_color),
		    secondary_color = COALESCE($7, secondary_color),
		    email = COALESCE($8, email),
		    phone = COALESCE($9, phone),
		    website = COALESCE($10, website),
		    social_links = COALESCE($11, social_links),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $12
		RETURNING id, department_id, name, tagline, description, logo_url,
		          primary_color, secondary_color, member_count, event_count,
		          awards_count, rating, email, phone, website, social_links,
		          created_at, updated_at
	`

	var club models.Club
	err = h.DB.QueryRow(
		query, req.DepartmentID, req.Name, req.Tagline, req.Description,
		req.LogoURL, req.PrimaryColor, req.SecondaryColor, req.Email,
		req.Phone, req.Website, req.SocialLinks, clubID,
	).Scan(
		&club.ID, &club.DepartmentID, &club.Name, &club.Tagline, &club.Description,
		&club.LogoURL, &club.PrimaryColor, &club.SecondaryColor, &club.MemberCount,
		&club.EventCount, &club.AwardsCount, &club.Rating, &club.Email, &club.Phone,
		&club.Website, &club.SocialLinks, &club.CreatedAt, &club.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Club not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update club"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": club})
}

// DeleteClub deletes a club (admin only)
func (h *ClubHandler) DeleteClub(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	result, err := h.DB.Exec("DELETE FROM clubs WHERE id = $1", clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete club"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Club not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Club deleted successfully"})
}

// ============================================================================
// CLUB MEMBERS
// ============================================================================

// GetClubMembers retrieves all members of a club
func (h *ClubHandler) GetClubMembers(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	query := `
		SELECT cm.id, cm.club_id, cm.user_id, cm.role, cm.position, cm.joined_at, cm.created_at,
		       u.id, u.email, u.full_name, u.role, u.avatar_url, u.department, u.year,
		       u.created_at, u.updated_at
		FROM club_members cm
		JOIN users u ON cm.user_id = u.id
		WHERE cm.club_id = $1
		ORDER BY cm.joined_at DESC
	`

	rows, err := h.DB.Query(query, clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}
	defer rows.Close()

	members := []models.ClubMemberWithUser{}
	for rows.Next() {
		var m models.ClubMemberWithUser
		if err := rows.Scan(
			&m.ID, &m.ClubID, &m.UserID, &m.Role, &m.Position, &m.JoinedAt, &m.CreatedAt,
			&m.User.ID, &m.User.Email, &m.User.FullName, &m.User.Role, &m.User.AvatarURL,
			&m.User.Department, &m.User.Year, &m.User.CreatedAt, &m.User.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan member"})
			return
		}
		members = append(members, m)
	}

	c.JSON(http.StatusOK, gin.H{"data": members})
}

// AddClubMember adds a member to a club
func (h *ClubHandler) AddClubMember(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	var req models.AddClubMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := "member"
	if req.Role != nil {
		role = *req.Role
	}

	query := `
		INSERT INTO club_members (club_id, user_id, role, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, club_id, user_id, role, position, joined_at, created_at
	`

	var member models.ClubMember
	err = h.DB.QueryRow(query, clubID, req.UserID, role, req.Position).Scan(
		&member.ID, &member.ClubID, &member.UserID, &member.Role,
		&member.Position, &member.JoinedAt, &member.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": member})
}

// UpdateClubMember updates a club member
func (h *ClubHandler) UpdateClubMember(c *gin.Context) {
	clubID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateClubMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE club_members
		SET role = COALESCE($1, role),
		    position = COALESCE($2, position)
		WHERE club_id = $3 AND user_id = $4
		RETURNING id, club_id, user_id, role, position, joined_at, created_at
	`

	var member models.ClubMember
	err = h.DB.QueryRow(query, req.Role, req.Position, clubID, userID).Scan(
		&member.ID, &member.ClubID, &member.UserID, &member.Role,
		&member.Position, &member.JoinedAt, &member.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": member})
}

// RemoveClubMember removes a member from a club
func (h *ClubHandler) RemoveClubMember(c *gin.Context) {
	clubID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result, err := h.DB.Exec("DELETE FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// ============================================================================
// CLUB ANNOUNCEMENTS
// ============================================================================

// GetClubAnnouncements retrieves all announcements for a club
func (h *ClubHandler) GetClubAnnouncements(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	query := `
		SELECT id, club_id, title, content, priority, is_pinned, created_by, created_at, updated_at
		FROM club_announcements
		WHERE club_id = $1
		ORDER BY is_pinned DESC, created_at DESC
	`

	rows, err := h.DB.Query(query, clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch announcements"})
		return
	}
	defer rows.Close()

	announcements := []models.ClubAnnouncement{}
	for rows.Next() {
		var a models.ClubAnnouncement
		if err := rows.Scan(
			&a.ID, &a.ClubID, &a.Title, &a.Content, &a.Priority,
			&a.IsPinned, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan announcement"})
			return
		}
		announcements = append(announcements, a)
	}

	c.JSON(http.StatusOK, gin.H{"data": announcements})
}

// CreateClubAnnouncement creates a new announcement
func (h *ClubHandler) CreateClubAnnouncement(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	var req models.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priority := "normal"
	isPinned := false
	if req.Priority != nil {
		priority = *req.Priority
	}
	if req.IsPinned != nil {
		isPinned = *req.IsPinned
	}

	// Get user ID from context (set by auth middleware)
	userID, _ := c.Get("userID")

	query := `
		INSERT INTO club_announcements (club_id, title, content, priority, is_pinned, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, club_id, title, content, priority, is_pinned, created_by, created_at, updated_at
	`

	var announcement models.ClubAnnouncement
	err = h.DB.QueryRow(query, clubID, req.Title, req.Content, priority, isPinned, userID).Scan(
		&announcement.ID, &announcement.ClubID, &announcement.Title, &announcement.Content,
		&announcement.Priority, &announcement.IsPinned, &announcement.CreatedBy,
		&announcement.CreatedAt, &announcement.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create announcement"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": announcement})
}

// UpdateClubAnnouncement updates an announcement
func (h *ClubHandler) UpdateClubAnnouncement(c *gin.Context) {
	clubID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	announcementID, err := uuid.Parse(c.Param("announcement_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement ID"})
		return
	}

	var req models.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE club_announcements
		SET title = COALESCE($1, title),
		    content = COALESCE($2, content),
		    priority = COALESCE($3, priority),
		    is_pinned = COALESCE($4, is_pinned),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND club_id = $6
		RETURNING id, club_id, title, content, priority, is_pinned, created_by, created_at, updated_at
	`

	var announcement models.ClubAnnouncement
	err = h.DB.QueryRow(query, req.Title, req.Content, req.Priority, req.IsPinned, announcementID, clubID).Scan(
		&announcement.ID, &announcement.ClubID, &announcement.Title, &announcement.Content,
		&announcement.Priority, &announcement.IsPinned, &announcement.CreatedBy,
		&announcement.CreatedAt, &announcement.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update announcement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": announcement})
}

// DeleteClubAnnouncement deletes an announcement
func (h *ClubHandler) DeleteClubAnnouncement(c *gin.Context) {
	clubID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	announcementID, err := uuid.Parse(c.Param("announcement_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid announcement ID"})
		return
	}

	result, err := h.DB.Exec("DELETE FROM club_announcements WHERE id = $1 AND club_id = $2", announcementID, clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete announcement"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Announcement deleted successfully"})
}

// ============================================================================
// CLUB AWARDS
// ============================================================================

// GetClubAwards retrieves all awards for a club
func (h *ClubHandler) GetClubAwards(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	query := `
		SELECT id, club_id, award_name, description, position, prize_amount,
		       event_name, awarded_date, certificate_url, created_at
		FROM club_awards
		WHERE club_id = $1
		ORDER BY awarded_date DESC NULLS LAST
	`

	rows, err := h.DB.Query(query, clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch awards"})
		return
	}
	defer rows.Close()

	awards := []models.ClubAward{}
	for rows.Next() {
		var a models.ClubAward
		if err := rows.Scan(
			&a.ID, &a.ClubID, &a.AwardName, &a.Description, &a.Position,
			&a.PrizeAmount, &a.EventName, &a.AwardedDate, &a.CertificateURL, &a.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan award"})
			return
		}
		awards = append(awards, a)
	}

	c.JSON(http.StatusOK, gin.H{"data": awards})
}

// CreateClubAward adds a new award
func (h *ClubHandler) CreateClubAward(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	var req models.CreateAwardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		INSERT INTO club_awards (
			club_id, award_name, description, position, prize_amount,
			event_name, awarded_date, certificate_url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, club_id, award_name, description, position, prize_amount,
		          event_name, awarded_date, certificate_url, created_at
	`

	var award models.ClubAward
	err = h.DB.QueryRow(
		query, clubID, req.AwardName, req.Description, req.Position,
		req.PrizeAmount, req.EventName, req.AwardedDate, req.CertificateURL,
	).Scan(
		&award.ID, &award.ClubID, &award.AwardName, &award.Description,
		&award.Position, &award.PrizeAmount, &award.EventName, &award.AwardedDate,
		&award.CertificateURL, &award.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create award"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": award})
}

// ============================================================================
// CLUB EVENTS
// ============================================================================

// GetClubEvents retrieves all events for a club
func (h *ClubHandler) GetClubEvents(c *gin.Context) {
	id := c.Param("id")
	clubID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid club ID"})
		return
	}

	query := `
		SELECT id, title, description, start_date, end_date, location,
		       banner_url, category, status, max_participants, current_participants,
		       registration_deadline, is_featured, club_id, created_at, updated_at
		FROM events
		WHERE club_id = $1
		ORDER BY start_date DESC
	`

	rows, err := h.DB.Query(query, clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}
	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.StartDate, &e.EndDate, &e.Location,
			&e.BannerURL, &e.Category, &e.Status, &e.MaxParticipants, &e.CurrentParticipants,
			&e.RegistrationDeadline, &e.IsFeatured, &e.ClubID, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan event"})
			return
		}
		events = append(events, e)
	}

	c.JSON(http.StatusOK, gin.H{"data": events})
}

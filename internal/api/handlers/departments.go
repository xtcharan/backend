package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/college-event-backend/internal/models"
)

type DepartmentHandler struct {
	DB *sql.DB
}

// GetDepartments retrieves all departments
func (h *DepartmentHandler) GetDepartments(c *gin.Context) {
	query := `
		SELECT id, code, name, description, logo_url, icon_name, color_hex,
		       total_members, total_clubs, total_events, created_at, updated_at
		FROM departments
		ORDER BY name ASC
	`

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch departments"})
		return
	}
	defer rows.Close()

	departments := []models.Department{}
	for rows.Next() {
		var dept models.Department
		if err := rows.Scan(
			&dept.ID, &dept.Code, &dept.Name, &dept.Description, &dept.LogoURL,
			&dept.IconName, &dept.ColorHex, &dept.TotalMembers, &dept.TotalClubs,
			&dept.TotalEvents, &dept.CreatedAt, &dept.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan department"})
			return
		}
		departments = append(departments, dept)
	}

	c.JSON(http.StatusOK, gin.H{"data": departments})
}

// GetDepartment retrieves a single department by ID
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	id := c.Param("id")
	departmentID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	query := `
		SELECT id, code, name, description, logo_url, icon_name, color_hex,
		       total_members, total_clubs, total_events, created_at, updated_at
		FROM departments
		WHERE id = $1
	`

	var dept models.Department
	err = h.DB.QueryRow(query, departmentID).Scan(
		&dept.ID, &dept.Code, &dept.Name, &dept.Description, &dept.LogoURL,
		&dept.IconName, &dept.ColorHex, &dept.TotalMembers, &dept.TotalClubs,
		&dept.TotalEvents, &dept.CreatedAt, &dept.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dept})
}

// GetDepartmentClubs retrieves all clubs in a department
func (h *DepartmentHandler) GetDepartmentClubs(c *gin.Context) {
	id := c.Param("id")
	departmentID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	query := `
		SELECT id, department_id, name, tagline, description, logo_url,
		       primary_color, secondary_color, member_count, event_count,
		       awards_count, rating, email, phone, website, social_links,
		       created_at, updated_at
		FROM clubs
		WHERE department_id = $1
		ORDER BY name ASC
	`

	rows, err := h.DB.Query(query, departmentID)
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

// CreateDepartment creates a new department (admin only)
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req models.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	colorHex := "#4F46E5"
	if req.ColorHex != nil {
		colorHex = *req.ColorHex
	}

	query := `
		INSERT INTO departments (code, name, description, logo_url, icon_name, color_hex)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, code, name, description, logo_url, icon_name, color_hex,
		          total_members, total_clubs, total_events, created_at, updated_at
	`

	var dept models.Department
	err := h.DB.QueryRow(
		query, req.Code, req.Name, req.Description, req.LogoURL, req.IconName, colorHex,
	).Scan(
		&dept.ID, &dept.Code, &dept.Name, &dept.Description, &dept.LogoURL,
		&dept.IconName, &dept.ColorHex, &dept.TotalMembers, &dept.TotalClubs,
		&dept.TotalEvents, &dept.CreatedAt, &dept.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create department"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dept})
}

// UpdateDepartment updates a department (admin only)
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id := c.Param("id")
	departmentID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	var req models.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE departments
		SET code = COALESCE($1, code),
		    name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    logo_url = COALESCE($4, logo_url),
		    icon_name = COALESCE($5, icon_name),
		    color_hex = COALESCE($6, color_hex),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING id, code, name, description, logo_url, icon_name, color_hex,
		          total_members, total_clubs, total_events, created_at, updated_at
	`

	var dept models.Department
	err = h.DB.QueryRow(
		query, req.Code, req.Name, req.Description, req.LogoURL,
		req.IconName, req.ColorHex, departmentID,
	).Scan(
		&dept.ID, &dept.Code, &dept.Name, &dept.Description, &dept.LogoURL,
		&dept.IconName, &dept.ColorHex, &dept.TotalMembers, &dept.TotalClubs,
		&dept.TotalEvents, &dept.CreatedAt, &dept.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update department"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dept})
}

// DeleteDepartment deletes a department (admin only)
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id := c.Param("id")
	departmentID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
		return
	}

	result, err := h.DB.Exec("DELETE FROM departments WHERE id = $1", departmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete department"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Department deleted successfully"})
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yven/backend/internal/middleware"
	"github.com/yven/backend/internal/models"
	"gorm.io/gorm"
)

type VolunteerHandler struct {
	DB *gorm.DB
}

type registerVolunteerRequest struct {
	FullName           string   `json:"fullName" binding:"required"`
	Email              string   `json:"email" binding:"required,email"`
	Phone              string   `json:"phone"`
	School             string   `json:"school"`
	GradeLevel         string   `json:"gradeLevel"`
	Skills             []string `json:"skills"`
	Interests          []string `json:"interests"`
	Availability       string   `json:"availability"` // raw JSON blob from the client form
	LocationPreference string   `json:"locationPreference"`
	MaxDistanceKm      *float64 `json:"maxDistanceKm"`
	EmergencyContact   string   `json:"emergencyContact" binding:"required"`
	Auth0Sub           string   `json:"auth0Sub" binding:"required"` // set post Auth0 signup redirect
}

// Register creates the User + Volunteer rows together. Called once,
// right after an Auth0 signup callback on the frontend.
func (h *VolunteerHandler) Register(c *gin.Context) {
	var req registerVolunteerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		user := models.User{
			Auth0Sub: req.Auth0Sub,
			Email:    req.Email,
			FullName: req.FullName,
			Phone:    req.Phone,
			Role:     models.RoleVolunteer,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		volunteer := models.Volunteer{
			UserID:             user.ID,
			School:             req.School,
			GradeLevel:         req.GradeLevel,
			Skills:             req.Skills,
			Interests:          req.Interests,
			AvailabilityJSON:   req.Availability,
			LocationPreference: req.LocationPreference,
			MaxDistanceKm:      req.MaxDistanceKm,
			EmergencyContact:   req.EmergencyContact,
		}
		return tx.Create(&volunteer).Error
	})

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "could not create volunteer account", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "volunteer account created"})
}

// Dashboard returns the aggregate view backing the volunteer dashboard:
// progress counters, plus the caller must separately hit /events and
// /service-logs for the list views (kept separate so each can be
// paginated/cached independently).
func (h *VolunteerHandler) Dashboard(c *gin.Context) {
	authUser := middleware.CurrentUser(c)

	var volunteer models.Volunteer
	if err := h.DB.Where("user_id = ?", authUser.UserID).First(&volunteer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "volunteer profile not found"})
		return
	}

	c.JSON(http.StatusOK, volunteer)
}

// UpdateProfile allows a volunteer to edit their own mutable profile
// fields — contact info, school, privacy controls. Identity fields
// (name, DOB if added later) intentionally excluded from this request
// shape; those would go through an admin-reviewed change path.
func (h *VolunteerHandler) UpdateProfile(c *gin.Context) {
	authUser := middleware.CurrentUser(c)

	var body struct {
		Phone              string   `json:"phone"`
		School             string   `json:"school"`
		Skills             []string `json:"skills"`
		Interests          []string `json:"interests"`
		LocationPreference string   `json:"locationPreference"`
		PrivacySettings    string   `json:"privacySettings"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var volunteer models.Volunteer
	if err := h.DB.Where("user_id = ?", authUser.UserID).First(&volunteer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "volunteer profile not found"})
		return
	}

	h.DB.Model(&volunteer).Updates(models.Volunteer{
		School:             body.School,
		Skills:             body.Skills,
		Interests:          body.Interests,
		LocationPreference: body.LocationPreference,
		PrivacySettings:    body.PrivacySettings,
	})

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

// volunteerIDForUser is a small shared helper other handlers use to
// resolve the current auth user into their Volunteer row ID.
func volunteerIDForUser(db *gorm.DB, userID uuid.UUID) (uuid.UUID, error) {
	var v models.Volunteer
	if err := db.Select("id").Where("user_id = ?", userID).First(&v).Error; err != nil {
		return uuid.Nil, err
	}
	return v.ID, nil
}

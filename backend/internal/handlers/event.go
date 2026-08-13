package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yven/backend/internal/middleware"
	"github.com/yven/backend/internal/models"
	"gorm.io/gorm"
)

type EventHandler struct {
	DB *gorm.DB
}

// List is the Opportunity Discovery endpoint. Supports the filter set
// from the spec: date range, event type (via requiredSkills overlap in
// a real search integration — see docs/ROADMAP.md for the Meilisearch
// wiring), location/remote, and availability.
func (h *EventHandler) List(c *gin.Context) {
	q := h.DB.Model(&models.Event{}).Where("published = ?", true)

	if remote := c.Query("remote"); remote != "" {
		q = q.Where("remote = ?", remote == "true")
	}
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			q = q.Where("starts_at >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			q = q.Where("starts_at <= ?", t)
		}
	}
	if loc := c.Query("location"); loc != "" {
		q = q.Where("location ILIKE ?", "%"+loc+"%")
	}

	var events []models.Event
	q.Order("starts_at asc").Limit(100).Find(&events)

	// NOTE: "Recommended for You" personalized matching against a
	// volunteer's skills/interests belongs in a separate ranking step
	// (ideally backed by Meilisearch) — tracked in docs/ROADMAP.md.

	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) Get(c *gin.Context) {
	var event models.Event
	if err := h.DB.First(&event, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

type createEventRequest struct {
	Title               string    `json:"title" binding:"required"`
	Description         string    `json:"description"`
	RoleExpectations    string    `json:"roleExpectations"`
	SupervisorUserID    uuid.UUID `json:"supervisorUserId" binding:"required"`
	VerificationMethod  string    `json:"verificationMethod"`
	ConductRequirements string    `json:"conductRequirements"`
	RequiredSkills      []string  `json:"requiredSkills"`
	StartsAt            time.Time `json:"startsAt" binding:"required"`
	EndsAt              time.Time `json:"endsAt" binding:"required"`
	Remote              bool      `json:"remote"`
	Location            string    `json:"location"`
	Capacity            int       `json:"capacity" binding:"required,gt=0"`
}

// Create is restricted (via RBAC + org-membership check) to org admins
// / event supervisors of an active organization.
func (h *EventHandler) Create(c *gin.Context) {
	authUser := middleware.CurrentUser(c)

	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !req.EndsAt.After(req.StartsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endsAt must be after startsAt"})
		return
	}

	orgID := c.Param("orgId")
	var org models.Organization
	if err := h.DB.First(&org, "id = ?", orgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	if org.Status != models.OrgStatusActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "only active organizations can publish events"})
		return
	}

	var membership models.OrgMembership
	err := h.DB.Where("user_id = ? AND organization_id = ?", authUser.UserID, org.ID).First(&membership).Error
	if err != nil || !membership.Active || (membership.OrgRole != models.OrgRoleAdmin && membership.OrgRole != models.OrgRoleCoordinator) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to create events for this organization"})
		return
	}

	event := models.Event{
		OrganizationID:      org.ID,
		Title:               req.Title,
		Description:         req.Description,
		RoleExpectations:    req.RoleExpectations,
		SupervisorUserID:    req.SupervisorUserID,
		VerificationMethod:  req.VerificationMethod,
		ConductRequirements: req.ConductRequirements,
		RequiredSkills:      req.RequiredSkills,
		StartsAt:            req.StartsAt,
		EndsAt:              req.EndsAt,
		Remote:              req.Remote,
		Location:            req.Location,
		Capacity:            req.Capacity,
		Published:           false, // drafts start unpublished; see Publish()
	}
	if err := h.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}
	c.JSON(http.StatusCreated, event)
}

// Publish flips a drafted event live.
func (h *EventHandler) Publish(c *gin.Context) {
	var event models.Event
	if err := h.DB.First(&event, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	event.Published = true
	h.DB.Save(&event)
	c.JSON(http.StatusOK, event)
}

// Register signs the calling volunteer up for an event, enforcing
// capacity and duplicate-registration rules.
func (h *EventHandler) Register(c *gin.Context) {
	authUser := middleware.CurrentUser(c)
	volunteerID, err := volunteerIDForUser(h.DB, authUser.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "volunteer profile not found"})
		return
	}

	eventID := c.Param("id")
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var event models.Event
		if err := tx.First(&event, "id = ?", eventID).Error; err != nil {
			return err
		}
		if !event.HasCapacity() {
			return gorm.ErrInvalidData
		}

		var existing int64
		tx.Model(&models.EventRegistration{}).
			Where("event_id = ? AND volunteer_id = ? AND withdrawn_at IS NULL", event.ID, volunteerID).
			Count(&existing)
		if existing > 0 {
			return gorm.ErrDuplicatedKey
		}

		reg := models.EventRegistration{EventID: event.ID, VolunteerID: volunteerID}
		if err := tx.Create(&reg).Error; err != nil {
			return err
		}
		return tx.Model(&event).Update("current_signups", gorm.Expr("current_signups + 1")).Error
	})

	switch err {
	case nil:
		c.JSON(http.StatusCreated, gin.H{"message": "registered"})
	case gorm.ErrInvalidData:
		c.JSON(http.StatusConflict, gin.H{"error": "event is at capacity"})
	case gorm.ErrDuplicatedKey:
		c.JSON(http.StatusConflict, gin.H{"error": "already registered for this event"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
	}
}

package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yven/backend/internal/middleware"
	"github.com/yven/backend/internal/models"
	"gorm.io/gorm"
)

type ServiceLogHandler struct {
	DB *gorm.DB
}

type createServiceLogRequest struct {
	EventID       uuid.UUID `json:"eventId" binding:"required"`
	RolePerformed string    `json:"rolePerformed"`
	ServiceDate   time.Time `json:"serviceDate" binding:"required"`
	HoursServed   float64   `json:"hoursServed" binding:"required,gt=0"`
	Location      string    `json:"location"`
	CheckInAt     *time.Time `json:"checkInAt,omitempty"`
	CheckOutAt    *time.Time `json:"checkOutAt,omitempty"`
	EvidenceURLs  []string  `json:"evidenceUrls,omitempty"`
}

// CreateDraft saves a new service log in Draft state. Validation here
// is deliberately permissive — the stricter integrity checks run at
// Submit time, so volunteers can save incomplete drafts freely.
func (h *ServiceLogHandler) CreateDraft(c *gin.Context) {
	authUser := middleware.CurrentUser(c)

	var req createServiceLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	volunteerID, err := volunteerIDForUser(h.DB, authUser.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "volunteer profile not found"})
		return
	}

	var event models.Event
	if err := h.DB.First(&event, "id = ?", req.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	// Prevent self-verification at the source: a volunteer can never be
	// their own assigned verifier.
	if event.SupervisorUserID == authUser.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you cannot log service for an event you supervise"})
		return
	}

	log := models.ServiceLog{
		EntryID:            generateEntryID(),
		VolunteerID:        volunteerID,
		EventID:            event.ID,
		OrganizationID:     event.OrganizationID,
		RolePerformed:      req.RolePerformed,
		ServiceDate:        req.ServiceDate,
		HoursServed:        req.HoursServed,
		Location:           req.Location,
		AssignedVerifierID: event.SupervisorUserID,
		CheckInAt:          req.CheckInAt,
		CheckOutAt:         req.CheckOutAt,
		EvidenceURLs:       req.EvidenceURLs,
		State:              models.StateDraft,
	}

	if err := h.DB.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save draft"})
		return
	}

	c.JSON(http.StatusCreated, log)
}

// Submit transitions a Draft log to Submitted, running the integrity
// checks required before it enters the verification queue.
func (h *ServiceLogHandler) Submit(c *gin.Context) {
	authUser := middleware.CurrentUser(c)
	logID := c.Param("id")

	var log models.ServiceLog
	if err := h.DB.First(&log, "id = ?", logID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service log not found"})
		return
	}

	volunteerID, err := volunteerIDForUser(h.DB, authUser.UserID)
	if err != nil || log.VolunteerID != volunteerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your service log"})
		return
	}

	if !models.CanTransition(log.State, models.StateSubmitted) {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("cannot submit a log in state %q", log.State)})
		return
	}

	if flags := h.integrityCheck(log); len(flags) > 0 {
		log.FlaggedReasons = flags
		log.State = models.StateFlagged
		h.DB.Save(&log)
		h.recordTransition(log.ID, models.StateDraft, models.StateFlagged, authUser.UserID, "automatic fraud-detection flag")
		c.JSON(http.StatusOK, gin.H{"message": "submission flagged for review", "flags": flags, "log": log})
		return
	}

	now := time.Now()
	log.State = models.StateSubmitted
	log.SubmittedAt = &now
	h.DB.Save(&log)
	h.recordTransition(log.ID, models.StateDraft, models.StateSubmitted, authUser.UserID, "")

	// TODO: enqueue notification to log.AssignedVerifierID (Resend/Twilio)
	// and emit a PostHog event for the "New service entry submitted"
	// supervisor notification requirement.

	c.JSON(http.StatusOK, log)
}

// Withdraw lets a volunteer pull back an entry before it's verified,
// per "Withdraw entries before verification".
func (h *ServiceLogHandler) Withdraw(c *gin.Context) {
	authUser := middleware.CurrentUser(c)
	logID := c.Param("id")

	var log models.ServiceLog
	if err := h.DB.First(&log, "id = ?", logID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service log not found"})
		return
	}
	volunteerID, err := volunteerIDForUser(h.DB, authUser.UserID)
	if err != nil || log.VolunteerID != volunteerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your service log"})
		return
	}
	if log.State == models.StateVerified {
		c.JSON(http.StatusConflict, gin.H{"error": "verified entries cannot be withdrawn"})
		return
	}

	h.DB.Delete(&log)
	c.JSON(http.StatusOK, gin.H{"message": "entry withdrawn"})
}

// integrityCheck implements the automatic fraud-detection rules:
// excessive hours in one day, duplicate submissions, supervisor
// mismatch. Returns a list of human-readable flag reasons; empty means
// clean.
func (h *ServiceLogHandler) integrityCheck(log models.ServiceLog) []string {
	var flags []string

	const maxReasonableHoursPerDay = 16.0
	if log.HoursServed > maxReasonableHoursPerDay {
		flags = append(flags, "excessive hours reported in a single day")
	}

	// Duplicate submission: same volunteer + event + date already
	// exists in a non-rejected state.
	var dupeCount int64
	h.DB.Model(&models.ServiceLog{}).
		Where("volunteer_id = ? AND event_id = ? AND service_date = ? AND state != ?",
			log.VolunteerID, log.EventID, log.ServiceDate, models.StateRejected).
		Where("id != ?", log.ID).
		Count(&dupeCount)
	if dupeCount > 0 {
		flags = append(flags, "duplicate submission for the same event and date")
	}

	// Overlapping time blocks for the same volunteer, regardless of event.
	if log.CheckInAt != nil && log.CheckOutAt != nil {
		var overlapCount int64
		h.DB.Model(&models.ServiceLog{}).
			Where("volunteer_id = ? AND id != ? AND state != ?", log.VolunteerID, log.ID, models.StateRejected).
			Where("check_in_at < ? AND check_out_at > ?", log.CheckOutAt, log.CheckInAt).
			Count(&overlapCount)
		if overlapCount > 0 {
			flags = append(flags, "overlapping time block with another logged entry")
		}
	}

	// Supervisor mismatch: assigned verifier must currently be an active
	// supervisor/admin on the log's organization.
	var membership models.OrgMembership
	err := h.DB.Where("user_id = ? AND organization_id = ?", log.AssignedVerifierID, log.OrganizationID).
		First(&membership).Error
	if err != nil || !membership.IsActiveSupervisorFor(log.OrganizationID, time.Now()) {
		flags = append(flags, "assigned verifier is not an active supervisor for this organization")
	}

	return flags
}

func (h *ServiceLogHandler) recordTransition(logID uuid.UUID, from, to models.ServiceLogState, actor uuid.UUID, reason string) {
	h.DB.Create(&models.StateTransition{
		ServiceLogID: logID,
		FromState:    from,
		ToState:      to,
		ActorUserID:  actor,
		Reason:       reason,
	})
}

func generateEntryID() string {
	return fmt.Sprintf("SL-%d-%06d", time.Now().Year(), rand.Intn(999999))
}

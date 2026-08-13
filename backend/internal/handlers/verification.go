package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yven/backend/internal/middleware"
	"github.com/yven/backend/internal/models"
	"gorm.io/gorm"
)

type VerificationHandler struct {
	DB *gorm.DB
}

type decideRequest struct {
	Decision      models.VerificationDecision `json:"decision" binding:"required"`
	AdjustedHours *float64                    `json:"adjustedHours,omitempty"`
	Attended      bool                        `json:"attended"`
	ConductOK     bool                        `json:"conductOk"`
	Reasoning     string                      `json:"reasoning" binding:"required"`
}

// Queue returns the pending service logs assigned to the calling
// supervisor — the core of the Supervisor Verification Portal.
func (h *VerificationHandler) Queue(c *gin.Context) {
	authUser := middleware.CurrentUser(c)

	var logs []models.ServiceLog
	h.DB.
		Where("assigned_verifier_id = ? AND state = ?", authUser.UserID, models.StateSubmitted).
		Order("submitted_at asc").
		Find(&logs)

	c.JSON(http.StatusOK, logs)
}

// Decide records a supervisor's verification decision and, on approval,
// triggers the VSR update. This is the single entry point for Approve /
// Reject / Return / Escalate — the decision value drives the outcome.
func (h *VerificationHandler) Decide(c *gin.Context) {
	authUser := middleware.CurrentUser(c)
	logID := c.Param("id")

	var req decideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var log models.ServiceLog
	if err := h.DB.First(&log, "id = ?", logID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service log not found"})
		return
	}

	// Conflict-of-interest + authority checks.
	if log.AssignedVerifierID != authUser.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not the assigned verifier for this entry"})
		return
	}
	if log.VolunteerID == authUser.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "self-verification is not permitted"})
		return
	}
	var membership models.OrgMembership
	err := h.DB.Where("user_id = ? AND organization_id = ?", authUser.UserID, log.OrganizationID).First(&membership).Error
	if err != nil || !membership.IsActiveSupervisorFor(log.OrganizationID, time.Now()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "verification authority expired, revoked, or not granted for this organization"})
		return
	}
	if log.State != models.StateSubmitted {
		c.JSON(http.StatusConflict, gin.H{"error": "entry is not awaiting verification"})
		return
	}

	var targetState models.ServiceLogState
	switch req.Decision {
	case models.DecisionApproved:
		targetState = models.StateVerified
	case models.DecisionRejected:
		targetState = models.StateRejected
	case models.DecisionReturned:
		targetState = models.StateDraft // "Returned for correction"
	case models.DecisionEscalated:
		targetState = models.StateFlagged
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unrecognized decision"})
		return
	}

	if !models.CanTransition(log.State, targetState) {
		c.JSON(http.StatusConflict, gin.H{"error": "illegal state transition"})
		return
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		verification := models.Verification{
			ServiceLogID:   log.ID,
			VerifierUserID: authUser.UserID,
			Decision:       req.Decision,
			AdjustedHours:  req.AdjustedHours,
			AttendedYN:     req.Attended,
			ConductOK:      req.ConductOK,
			Reasoning:      req.Reasoning,
		}
		if err := tx.Create(&verification).Error; err != nil {
			return err
		}

		fromState := log.State
		log.State = targetState
		if err := tx.Save(&log).Error; err != nil {
			return err
		}
		tx.Create(&models.StateTransition{
			ServiceLogID: log.ID,
			FromState:    fromState,
			ToState:      targetState,
			ActorUserID:  authUser.UserID,
			Reason:       req.Reasoning,
		})

		if targetState == models.StateVerified {
			if err := recomputeVSR(tx, log.VolunteerID); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record verification"})
		return
	}

	// TODO: notify volunteer of verification status (Resend/Twilio) per
	// "Notify volunteers of verification status".

	c.JSON(http.StatusOK, gin.H{"message": "decision recorded", "newState": targetState})
}

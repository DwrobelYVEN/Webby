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

type VSRHandler struct {
	DB *gorm.DB
}

// recomputeVSR rebuilds the single VSR summary row for a volunteer from
// verified ServiceLog rows — the only source of truth. Called inside
// the same transaction as a verification decision (see
// VerificationHandler.Decide) so the VSR can never drift from the
// underlying data, and manual edits to the VSR itself are impossible
// because nothing ever writes to it except this function.
func recomputeVSR(tx *gorm.DB, volunteerID uuid.UUID) error {
	var total float64
	if err := tx.Model(&models.ServiceLog{}).
		Where("volunteer_id = ? AND state = ?", volunteerID, models.StateVerified).
		Select("COALESCE(SUM(hours_served), 0)").
		Scan(&total).Error; err != nil {
		return err
	}

	var vsr models.VSR
	err := tx.Where("volunteer_id = ?", volunteerID).First(&vsr).Error
	if err == gorm.ErrRecordNotFound {
		vsr = models.VSR{VolunteerID: volunteerID, Locked: true}
	} else if err != nil {
		return err
	}
	vsr.TotalVerifiedHours = total
	vsr.LastUpdatedAt = time.Now()
	if err := tx.Save(&vsr).Error; err != nil {
		return err
	}

	// Keep the volunteer's denormalized dashboard counter in sync too.
	return tx.Model(&models.Volunteer{}).
		Where("id = ?", volunteerID).
		Update("total_hours_verified", total).Error
}

// GetMyVSR returns the calling volunteer's VSR plus the chronological
// list of verified entries backing it — Draft/Pending/Rejected entries
// are excluded per spec.
func (h *VSRHandler) GetMyVSR(c *gin.Context) {
	authUser := middleware.CurrentUser(c)
	volunteerID, err := volunteerIDForUser(h.DB, authUser.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "volunteer profile not found"})
		return
	}

	var vsr models.VSR
	h.DB.Where("volunteer_id = ?", volunteerID).First(&vsr)

	var entries []models.ServiceLog
	h.DB.Where("volunteer_id = ? AND state = ?", volunteerID, models.StateVerified).
		Order("service_date desc").
		Find(&entries)

	c.JSON(http.StatusOK, gin.H{"vsr": vsr, "entries": entries})
}

// Export generates a PDF or CSV of the volunteer's verified record and
// logs the export event. The actual file-rendering call (wkhtmltopdf,
// a PDF library, or a CSV writer) is a TODO — this wires up the
// authorization, data-selection, and audit-logging contract it must
// satisfy.
func (h *VSRHandler) Export(c *gin.Context) {
	authUser := middleware.CurrentUser(c)
	format := c.DefaultQuery("format", "pdf")
	if format != "pdf" && format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format must be 'pdf' or 'csv'"})
		return
	}

	volunteerID, err := volunteerIDForUser(h.DB, authUser.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "volunteer profile not found"})
		return
	}

	var entries []models.ServiceLog
	h.DB.Where("volunteer_id = ? AND state = ?", volunteerID, models.StateVerified).
		Order("service_date asc").
		Find(&entries)

	// TODO: render `entries` to PDF/CSV, upload to Azure Blob Storage,
	// and return a short-lived signed URL instead of FileURL below.
	export := models.VSRExport{
		VolunteerID:  volunteerID,
		ExportedByID: authUser.UserID,
		Format:       format,
		EntryCount:   len(entries),
		GeneratedAt:  time.Now(),
		FileURL:      "", // populated once blob upload is implemented
	}
	h.DB.Create(&export)

	c.JSON(http.StatusOK, gin.H{
		"message":    "export generated",
		"entryCount": len(entries),
		"format":     format,
		"exportId":   export.ID,
	})
}

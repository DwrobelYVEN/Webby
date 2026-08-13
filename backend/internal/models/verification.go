package models

import (
	"time"

	"github.com/google/uuid"
)

type VerificationDecision string

const (
	DecisionApproved  VerificationDecision = "verified"
	DecisionRejected  VerificationDecision = "rejected"
	DecisionReturned  VerificationDecision = "returned_for_correction"
	DecisionEscalated VerificationDecision = "escalated_to_yven"
)

// Verification is the immutable, non-repudiable record of a supervisor's
// decision on a ServiceLog. One ServiceLog may accumulate several rows
// over its lifetime (e.g. Returned, then re-Submitted, then Verified).
type Verification struct {
	ID             uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServiceLogID   uuid.UUID            `gorm:"type:uuid;not null;index" json:"serviceLogId"`
	VerifierUserID uuid.UUID            `gorm:"type:uuid;not null" json:"verifierUserId"`
	Decision       VerificationDecision `gorm:"not null" json:"decision"`
	AdjustedHours  *float64             `json:"adjustedHours,omitempty"` // "Approve partial hours"
	AttendedYN     bool                 `json:"attended"`
	ConductOK      bool                 `json:"conductOk"`
	Reasoning      string               `json:"reasoning"`
	CreatedAt      time.Time            `json:"createdAt"`
}

// VSR (Volunteer Service Record) is generated/updated automatically
// whenever a ServiceLog reaches Verified — never edited manually, and
// its included-entries list is derived, not stored redundantly.
type VSR struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VolunteerID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"volunteerId"`
	TotalVerifiedHours float64   `gorm:"default:0" json:"totalVerifiedHours"`
	LastUpdatedAt      time.Time `json:"lastUpdatedAt"`
	Locked             bool      `gorm:"default:true" json:"locked"` // always true; present for clarity/API shape
}

// VSRExport records every export event for audit purposes — "Log export
// activity", "Timestamp export generation".
type VSRExport struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VolunteerID    uuid.UUID `gorm:"type:uuid;not null;index" json:"volunteerId"`
	ExportedByID   uuid.UUID `gorm:"type:uuid;not null" json:"exportedById"`
	Format         string    `gorm:"not null" json:"format"` // "pdf" | "csv"
	EntryCount     int       `json:"entryCount"`
	GeneratedAt    time.Time `json:"generatedAt"`
	FileURL        string    `json:"fileUrl"` // Azure Blob Storage URL, short-lived signed link
}

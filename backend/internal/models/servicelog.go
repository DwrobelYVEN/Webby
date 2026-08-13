package models

import (
	"time"

	"github.com/google/uuid"
)

// ServiceLog is a single service-hour entry submitted by a volunteer.
// Once Submitted, content fields must never be mutated directly —
// only through the verification handlers, which preserve original
// values per "Preserve original submission values".
type ServiceLog struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EntryID          string          `gorm:"uniqueIndex;not null" json:"entryId"` // human-readable unique ID, e.g. SL-2026-000123
	VolunteerID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"volunteerId"`
	EventID          uuid.UUID       `gorm:"type:uuid;not null;index" json:"eventId"`
	OrganizationID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"organizationId"`
	RolePerformed    string          `json:"rolePerformed"`
	ServiceDate      time.Time       `gorm:"not null" json:"serviceDate"`
	HoursServed      float64         `gorm:"not null" json:"hoursServed"`
	Location         string          `json:"location"`
	AssignedVerifierID uuid.UUID     `gorm:"type:uuid;not null" json:"assignedVerifierId"`
	CheckInAt        *time.Time      `json:"checkInAt,omitempty"`
	CheckOutAt       *time.Time      `json:"checkOutAt,omitempty"`
	EvidenceURLs     []string        `gorm:"serializer:json" json:"evidenceUrls,omitempty"` // Azure Blob URLs
	QRConfirmed      bool            `gorm:"default:false" json:"qrConfirmed"`
	SupervisorNotes  string          `json:"supervisorNotes,omitempty"`

	State            ServiceLogState `gorm:"not null;default:'draft';index" json:"state"`
	SubmittedAt      *time.Time      `json:"submittedAt,omitempty"`

	// Fraud-detection flags, populated by the integrity check job before
	// an entry is allowed to move Draft -> Submitted.
	FlaggedReasons   []string        `gorm:"serializer:json" json:"flaggedReasons,omitempty"`

	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

// StateTransition is an immutable audit row: one per state change on a
// ServiceLog. Never updated or deleted — "Log every state transition",
// "Preserve full lifecycle history".
type StateTransition struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServiceLogID uuid.UUID       `gorm:"type:uuid;not null;index" json:"serviceLogId"`
	FromState    ServiceLogState `json:"fromState"`
	ToState      ServiceLogState `json:"toState"`
	ActorUserID  uuid.UUID       `gorm:"type:uuid;not null" json:"actorUserId"`
	Reason       string          `json:"reason,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
}

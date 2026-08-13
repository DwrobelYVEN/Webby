package models

import (
	"time"

	"github.com/google/uuid"
)

type EventVisibility string

const (
	VisibilityPublic       EventVisibility = "public"
	VisibilityRestricted   EventVisibility = "restricted"
	VisibilityInviteOnly   EventVisibility = "invitation_only"
)

// Event is a single volunteer opportunity hosted by an Organization.
type Event struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID     uuid.UUID       `gorm:"type:uuid;not null;index" json:"organizationId"`
	Title              string          `gorm:"not null" json:"title"`
	Description        string          `json:"description"`
	RoleExpectations   string          `json:"roleExpectations"`
	SupervisorUserID   uuid.UUID       `gorm:"type:uuid;not null" json:"supervisorUserId"`
	VerificationMethod string          `json:"verificationMethod"` // e.g. "QR check-in", "supervisor manual"
	ConductRequirements string         `json:"conductRequirements"`
	RequiredSkills     []string        `gorm:"serializer:json" json:"requiredSkills"`
	StartsAt           time.Time       `gorm:"not null" json:"startsAt"`
	EndsAt             time.Time       `gorm:"not null" json:"endsAt"`
	Remote             bool            `json:"remote"`
	Location           string          `json:"location"`
	MeetingPointNotes  string          `json:"meetingPointNotes"`
	Capacity           int             `json:"capacity"`
	CurrentSignups     int             `gorm:"default:0" json:"currentSignups"`
	Visibility         EventVisibility `gorm:"default:'public'" json:"visibility"`
	Published          bool            `gorm:"default:false" json:"published"`
	CutoffAt           *time.Time      `json:"cutoffAt,omitempty"` // edits blocked after this without admin override
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}

// HasCapacity reports whether another registration can be accepted.
func (e Event) HasCapacity() bool {
	return e.CurrentSignups < e.Capacity
}

// EventRegistration links a Volunteer to an Event they've signed up for.
type EventRegistration struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EventID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"eventId"`
	VolunteerID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"volunteerId"`
	ConfirmedAt     *time.Time `json:"confirmedAt,omitempty"`
	WithdrawnAt     *time.Time `json:"withdrawnAt,omitempty"`
	WithdrawReason  string     `json:"withdrawReason,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

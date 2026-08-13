package models

import (
	"time"

	"github.com/google/uuid"
)

// Case is a formal record on the Conflict Dashboard — a disputed log,
// rejected verification, misconduct report, etc. This model is wired
// into the DB and CRUD handlers now; the full admin case-management UI
// is tracked in docs/ROADMAP.md as a follow-up.
type Case struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CaseNumber         string     `gorm:"uniqueIndex;not null" json:"caseNumber"` // e.g. CASE-2026-0042
	Category           string     `gorm:"not null" json:"category"`              // disputed_log | misconduct | org_verification_failure | complaint
	Status             CaseStatus `gorm:"not null;default:'open'" json:"status"`
	AssignedAdminID    *uuid.UUID `gorm:"type:uuid" json:"assignedAdminId,omitempty"`
	ServiceLogID       *uuid.UUID `gorm:"type:uuid" json:"serviceLogId,omitempty"`
	VerificationID     *uuid.UUID `gorm:"type:uuid" json:"verificationId,omitempty"`
	OrganizationID     *uuid.UUID `gorm:"type:uuid" json:"organizationId,omitempty"`
	RaisedByUserID     uuid.UUID  `gorm:"type:uuid;not null" json:"raisedByUserId"`
	Description        string     `json:"description"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// CaseNote is an internal admin note or info request thread entry.
type CaseNote struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CaseID    uuid.UUID `gorm:"type:uuid;not null;index" json:"caseId"`
	AuthorID  uuid.UUID `gorm:"type:uuid;not null" json:"authorId"`
	Body      string    `json:"body"`
	IsRequestForInfo bool `gorm:"default:false" json:"isRequestForInfo"`
	CreatedAt time.Time `json:"createdAt"`
}

// CaseResolution is the immutable, logged outcome of a case.
type CaseResolution struct {
	ID              uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CaseID          uuid.UUID          `gorm:"type:uuid;not null;index" json:"caseId"`
	Action          string             `json:"action"` // uphold | reverse | require_reverification | freeze_vsr | restrict_access | escalate
	Category        ResolutionCategory `json:"category"`
	PerformedByID   uuid.UUID          `gorm:"type:uuid;not null" json:"performedById"`
	AffectedRecords []string           `gorm:"serializer:json" json:"affectedRecords"` // record IDs touched by this action
	ResolvedAt      time.Time          `json:"resolvedAt"`
}

// Policy is a versioned content record (Volunteer Conduct Standards,
// Verification Rules, etc.). Publishing creates a new PolicyVersion;
// prior versions are permanent and read-only.
type Policy struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug     string    `gorm:"uniqueIndex;not null" json:"slug"`
	Title    string    `gorm:"not null" json:"title"`
	Category string    `json:"category"` // volunteer | organization | verification | institutional
}

type PolicyVersion struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PolicyID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"policyId"`
	VersionNumber int        `gorm:"not null" json:"versionNumber"`
	BodyHTML      string     `json:"bodyHtml"`
	AuthorID      uuid.UUID  `gorm:"type:uuid;not null" json:"authorId"`
	EffectiveAt   time.Time  `json:"effectiveAt"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	Archived      bool       `gorm:"default:false" json:"archived"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// PolicyAcknowledgment logs a user accepting a specific policy version.
type PolicyAcknowledgment struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	PolicyVersionID uuid.UUID `gorm:"type:uuid;not null;index" json:"policyVersionId"`
	Method          string    `json:"method"` // click_to_accept | signature
	AcknowledgedAt  time.Time `json:"acknowledgedAt"`
}

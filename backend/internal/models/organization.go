package models

import (
	"time"

	"github.com/google/uuid"
)

// Organization is the event-hosting entity. Suspended/restricted orgs
// are blocked from publishing events at the handler layer — see
// "Suspensions" in the spec.
type Organization struct {
	ID                    uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LegalName             string             `gorm:"not null" json:"legalName"`
	DisplayName           string             `gorm:"not null" json:"displayName"`
	OrgType               string             `json:"orgType"`
	ContactEmail          string             `gorm:"not null" json:"contactEmail"`
	ContactPhone          string             `json:"contactPhone"`
	AuthorizedRepUserID   uuid.UUID          `gorm:"type:uuid;not null" json:"authorizedRepUserId"`
	LogoURL               string             `json:"logoUrl"` // Azure Blob Storage URL
	Bio                   string             `json:"bio"`
	Status                OrganizationStatus `gorm:"not null;default:'pending_approval'" json:"status"`
	SupportingDocsURL     string             `json:"supportingDocsUrl"` // Azure Blob Storage URL
	CreatedAt             time.Time          `json:"createdAt"`
	UpdatedAt             time.Time          `json:"updatedAt"`

	// Fields requiring admin approval + audit log to change (see
	// "Critical identity fields must require admin approval").
	LockedFields []string `gorm:"serializer:json" json:"-"`
}

// OrgMembership links a User to an Organization with an internal
// sub-role (Org Admin / Event Supervisor / Volunteer Coordinator).
// This is what "Restrict verification authority to designated
// supervisors" is enforced against.
type OrgMembership struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organizationId"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	OrgRole        OrgRole   `gorm:"not null" json:"orgRole"`
	Active         bool      `gorm:"default:true" json:"active"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"` // supports role expiration/time-bound authorization
	CreatedAt      time.Time `json:"createdAt"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
	RevokedByUserID *uuid.UUID `gorm:"type:uuid" json:"revokedByUserId,omitempty"`
}

// IsActiveSupervisorFor reports whether this membership currently grants
// verification authority — active, correct role, and not expired.
func (m OrgMembership) IsActiveSupervisorFor(orgID uuid.UUID, now time.Time) bool {
	if m.OrganizationID != orgID || !m.Active {
		return false
	}
	if m.OrgRole != OrgRoleSupervisor && m.OrgRole != OrgRoleAdmin {
		return false
	}
	if m.ExpiresAt != nil && now.After(*m.ExpiresAt) {
		return false
	}
	return true
}

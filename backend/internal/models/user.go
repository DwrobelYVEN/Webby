package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is the identity record shared by every account type. Auth0 is the
// identity provider of record — Sub stores Auth0's `sub` claim and is
// what JWT validation middleware matches against. We never store
// passwords here.
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Auth0Sub  string    `gorm:"uniqueIndex;not null" json:"-"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	FullName  string    `gorm:"not null" json:"fullName"`
	Phone     string    `json:"phone,omitempty"`
	Role      Role      `gorm:"not null;index" json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Volunteer holds the volunteer-specific profile fields. One-to-one with
// User (RoleVolunteer). Kept separate from User so org-side roles
// (supervisors, admins) don't carry unused volunteer columns.
type Volunteer struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"userId"`
	User               User      `json:"-"`
	School             string    `json:"school"`
	GradeLevel         string    `json:"gradeLevel"`
	Skills             []string  `gorm:"serializer:json" json:"skills"`
	Interests          []string  `gorm:"serializer:json" json:"interests"`
	AvailabilityJSON   string    `gorm:"type:jsonb" json:"availability"` // days/times/remote-in-person prefs
	LocationPreference string    `json:"locationPreference"`
	MaxDistanceKm      *float64  `json:"maxDistanceKm,omitempty"`
	EmergencyContact   string    `json:"emergencyContact"`
	FollowingPaused    bool      `gorm:"default:false" json:"followingPaused"`
	PrivacySettings    string    `gorm:"type:jsonb" json:"privacySettings"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`

	// Denormalized progress counters, recomputed by the VSR service
	// whenever a log is verified — avoids an aggregate query on every
	// dashboard load. Source of truth remains the ServiceLog table.
	TotalHoursVerified float64 `gorm:"default:0" json:"totalHoursVerified"`
	TotalHoursPending  float64 `gorm:"default:0" json:"totalHoursPending"`
	EventsAttended     int     `gorm:"default:0" json:"eventsAttended"`
}

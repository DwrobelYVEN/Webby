package models

// Role represents a system-wide RBAC role. A user can hold exactly one
// platform role, but may additionally hold org-scoped sub-roles (see
// OrgRole) — e.g. a Volunteer role user can also be an Event Supervisor
// for one specific organization.
type Role string

const (
	RoleVolunteer        Role = "volunteer"
	RoleOrgSupervisor    Role = "org_supervisor"
	RoleOrgAdmin         Role = "org_admin"
	RoleSchoolAdmin      Role = "school_admin"
	RoleYVENAdmin        Role = "yven_admin"
)

// OrgRole represents an internal sub-role within a single organization,
// per the "Role Controls" section of the spec (Org Admin / Event
// Supervisor / Volunteer Coordinator).
type OrgRole string

const (
	OrgRoleAdmin       OrgRole = "org_admin"
	OrgRoleSupervisor  OrgRole = "event_supervisor"
	OrgRoleCoordinator OrgRole = "volunteer_coordinator"
)

// ServiceLogState is the lifecycle state machine for a service entry.
// Transitions are enforced in the service log handlers/service layer —
// never mutated directly — per "State Management & Lifecycle Controls".
type ServiceLogState string

const (
	StateDraft     ServiceLogState = "draft"
	StateSubmitted ServiceLogState = "submitted" // pending verification
	StateVerified  ServiceLogState = "verified"
	StateRejected  ServiceLogState = "rejected"
	StateFlagged   ServiceLogState = "flagged"
	StateArchived  ServiceLogState = "archived"
)

// validTransitions encodes the only allowed state moves. Anything not
// listed here must be refused by TransitionServiceLog.
var validTransitions = map[ServiceLogState][]ServiceLogState{
	StateDraft:     {StateSubmitted},
	StateSubmitted: {StateVerified, StateRejected, StateFlagged},
	StateRejected:  {StateDraft}, // "Returned for correction"
	StateFlagged:   {StateSubmitted, StateArchived},
	StateVerified:  {StateArchived}, // archival only; content is otherwise locked
}

// CanTransition reports whether moving from `from` to `to` is a legal
// state-machine transition.
func CanTransition(from, to ServiceLogState) bool {
	for _, allowed := range validTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// OrganizationStatus tracks org approval lifecycle.
type OrganizationStatus string

const (
	OrgStatusPending   OrganizationStatus = "pending_approval"
	OrgStatusActive    OrganizationStatus = "active"
	OrgStatusSuspended OrganizationStatus = "suspended"
	OrgStatusDanger    OrganizationStatus = "restricted" // "deemed a danger" per spec

)

// CaseStatus tracks a Conflict Dashboard case.
type CaseStatus string

const (
	CaseOpen         CaseStatus = "open"
	CaseAwaitingInfo CaseStatus = "awaiting_info"
	CaseReviewing    CaseStatus = "reviewing"
	CaseResolved     CaseStatus = "resolved"
)

// ResolutionCategory classifies how a case was resolved.
type ResolutionCategory string

const (
	ResolutionError        ResolutionCategory = "error"
	ResolutionMisuse       ResolutionCategory = "misuse"
	ResolutionPolicyBreach ResolutionCategory = "policy_breach"
	ResolutionSystemIssue  ResolutionCategory = "system_issue"
)

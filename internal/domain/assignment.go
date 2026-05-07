package domain

import "time"

type ScopeType string

const (
	ScopeSubscription    ScopeType = "subscription"
	ScopeResourceGroup   ScopeType = "resource_group"
	ScopeManagementGroup ScopeType = "management_group"
)

type EligibleAssignment struct {
	ID            string
	PrincipalID   string
	Role          string
	RoleDefID     string
	ScopeType     ScopeType
	ScopeID       string
	ScopeName     string
	EligibleUntil time.Time
}

type ActivationRequest struct {
	ScopeID  string
	Role     string
	Reason   string
	Duration time.Duration
}

type ActivationResult struct {
	Role      string
	ScopeID   string
	ScopeName string
	Duration  time.Duration
	StartedAt time.Time
	ExpiresAt time.Time
	Reason    string
}

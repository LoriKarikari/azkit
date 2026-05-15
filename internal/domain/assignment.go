package domain

import (
	"strings"
	"time"
)

type ScopeType string

const (
	ScopeSubscription    ScopeType = "subscription"
	ScopeResourceGroup   ScopeType = "resource_group"
	ScopeManagementGroup ScopeType = "management_group"
)

type ActiveAssignmentStatus string

type DeactivationRequestStatus string

const (
	ActiveAssignmentActive ActiveAssignmentStatus = "active"

	DeactivationRequested DeactivationRequestStatus = "deactivation_requested"
)

type EligibleAssignment struct {
	ID               string
	PrincipalID      string
	Role             string
	RoleDefID        string
	ScopeType        ScopeType
	ScopeID          string
	ScopeName        string
	SubscriptionID   string
	SubscriptionName string
	EligibleUntil    time.Time
}

type ActiveAssignment struct {
	ID               string
	PrincipalID      string
	Role             string
	RoleDefID        string
	ScopeType        ScopeType
	ScopeID          string
	ScopeName        string
	SubscriptionID   string
	SubscriptionName string
	StartTime        time.Time
	EndTime          time.Time
	Status           ActiveAssignmentStatus
}

type ActivationRequest struct {
	ScopeID       string
	Subscription  string
	ResourceGroup string
	Role          string
	Reason        string
	Duration      time.Duration
}

type ActivationTarget struct {
	Assignment EligibleAssignment
	Reason     string
	Duration   time.Duration
}

type ActivationOutcome string

const (
	ActivationRequested     ActivationOutcome = "requested"
	ActivationActive        ActivationOutcome = "active"
	ActivationAlreadyActive ActivationOutcome = "already_active"
	ActivationPending       ActivationOutcome = "pending"
)

type ActivationResult struct {
	Role      string
	RoleDefID string
	ScopeID   string
	ScopeName string
	Duration  time.Duration
	StartedAt time.Time
	ExpiresAt time.Time
	Reason    string
	Outcome   ActivationOutcome
}

type DeactivationResult struct {
	Role         string
	ScopeID      string
	ScopeName    string
	ScopeType    ScopeType
	AssignmentID string
	Status       DeactivationRequestStatus
}

func ActiveAssignmentConfirmsResult(active ActiveAssignment, result *ActivationResult) bool {
	return ActiveMatchesRoleScope(active, result.Role, result.RoleDefID, result.ScopeID)
}

func ActiveAssignmentMatchesEligible(active ActiveAssignment, eligible EligibleAssignment) bool {
	return ActiveMatchesRoleScope(active, eligible.Role, eligible.RoleDefID, eligible.ScopeID)
}

func ActiveMatchesRoleScope(active ActiveAssignment, role, roleDefID, scopeID string) bool {
	isActive := active.Status == ActiveAssignmentActive
	hasExpiry := !active.EndTime.IsZero()
	matchesScope := strings.EqualFold(active.ScopeID, scopeID)
	if !isActive || !hasExpiry || !matchesScope {
		return false
	}
	if roleDefID != "" && strings.EqualFold(active.RoleDefID, roleDefID) {
		return true
	}
	return active.Role == role
}

func ActivationResultFromActive(active ActiveAssignment, outcome ActivationOutcome) ActivationResult {
	duration := max(active.EndTime.Sub(active.StartTime), 0)
	return ActivationResult{
		Role:      active.Role,
		RoleDefID: active.RoleDefID,
		ScopeID:   active.ScopeID,
		ScopeName: active.ScopeName,
		Duration:  duration,
		StartedAt: active.StartTime,
		ExpiresAt: active.EndTime,
		Outcome:   outcome,
	}
}

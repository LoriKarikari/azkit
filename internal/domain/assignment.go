package domain

import "time"

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

type ActivationResult struct {
	Role          string
	ScopeID       string
	ScopeName     string
	Duration      time.Duration
	StartedAt     time.Time
	ExpiresAt     time.Time
	Reason        string
	AlreadyActive bool
}

type DeactivationResult struct {
	Role         string
	ScopeID      string
	ScopeName    string
	ScopeType    ScopeType
	AssignmentID string
	Status       DeactivationRequestStatus
}

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
	Role          string
	ScopeType     ScopeType
	ScopeID       string
	ScopeName     string
	EligibleUntil time.Time
}

type ActiveAssignment struct {
	EligibleAssignment
	StartedAt string
	ExpiresAt string
}

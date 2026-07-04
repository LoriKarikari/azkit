package domain

import "time"

type Subscription struct {
	ID   string
	Name string
}

type SubscriptionCache struct {
	FetchedAt     time.Time
	Subscriptions []Subscription
}

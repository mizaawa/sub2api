package service

import (
	"time"
)

// AccountExpiryService is retained for compatibility with existing dependency
// injection and shutdown wiring. Account expiry no longer changes scheduling.
type AccountExpiryService struct{}

func NewAccountExpiryService(_ AccountRepository, _ time.Duration) *AccountExpiryService {
	return &AccountExpiryService{}
}

func (*AccountExpiryService) Start() {}

func (*AccountExpiryService) Stop() {}

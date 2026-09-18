package service

import (
	"testing"
	"time"
)

func TestAccountExpiryServiceIsCompatibilityNoop(t *testing.T) {
	svc := NewAccountExpiryService(nil, time.Nanosecond)
	svc.Start()
	svc.Stop()
}

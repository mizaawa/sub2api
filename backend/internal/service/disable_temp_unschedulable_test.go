package service

import (
	"testing"
	"time"
)

func TestDisableTempUnschedulableSwitchControlsAccountScheduling(t *testing.T) {
	t.Cleanup(func() { SetDisableTempUnschedulableRuntime(false) })
	future := time.Now().Add(time.Hour)

	account := &Account{
		Status:                 StatusActive,
		Schedulable:            true,
		TempUnschedulableUntil: &future,
		RateLimitResetAt:       &future,
		OverloadUntil:          &future,
		Credentials:            map[string]any{},
		Extra:                  map[string]any{},
	}

	SetDisableTempUnschedulableRuntime(false)
	if account.IsSchedulable() {
		t.Fatal("temporary scheduling blockers must apply when the switch is off")
	}

	SetDisableTempUnschedulableRuntime(true)
	if !account.IsSchedulable() {
		t.Fatal("temporary scheduling blockers must be ignored when the switch is on")
	}

	account.Schedulable = false
	if account.IsSchedulable() {
		t.Fatal("manual schedulable=false must remain enforced")
	}
	account.Schedulable = true
	account.Status = StatusError
	if account.IsSchedulable() {
		t.Fatal("non-active account status must remain enforced")
	}
}

func TestDisableTempUnschedulableSwitchKeepsExpiryAndHardQuotaEnforced(t *testing.T) {
	t.Cleanup(func() { SetDisableTempUnschedulableRuntime(false) })
	SetDisableTempUnschedulableRuntime(true)

	expiresAt := time.Now().Add(-time.Minute)
	expired := &Account{
		Status:             StatusActive,
		Schedulable:        true,
		AutoPauseOnExpired: true,
		ExpiresAt:          &expiresAt,
	}
	if expired.IsSchedulable() {
		t.Fatal("expired account must remain blocked")
	}

	quotaExceeded := &Account{
		Status:      StatusActive,
		Schedulable: true,
		Type:        AccountTypeAPIKey,
		Extra: map[string]any{
			"quota_limit": 100.0,
			"quota_used":  100.0,
		},
	}
	if quotaExceeded.IsSchedulable() {
		t.Fatal("hard account quota must remain enforced")
	}
}

func TestDisableTempUnschedulableSwitchSkipsModelCooldown(t *testing.T) {
	t.Cleanup(func() { SetDisableTempUnschedulableRuntime(false) })
	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	account := &Account{
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{},
		Extra: map[string]any{
			modelRateLimitsKey: map[string]any{
				"model-a": map[string]any{"rate_limit_reset_at": future},
			},
		},
	}

	SetDisableTempUnschedulableRuntime(false)
	if !account.isModelRateLimitedWithContext(nil, "model-a") {
		t.Fatal("model cooldown must apply when the switch is off")
	}
	SetDisableTempUnschedulableRuntime(true)
	if account.isModelRateLimitedWithContext(nil, "model-a") {
		t.Fatal("model cooldown must be ignored when the switch is on")
	}
	if got := account.GetModelRateLimitRemainingTimeWithContext(nil, "model-a"); got != 0 {
		t.Fatalf("remaining model cooldown = %s, want zero", got)
	}
}

func TestDisableTempUnschedulableSwitchSkipsOpenAIRuntimeBlocker(t *testing.T) {
	t.Cleanup(func() { SetDisableTempUnschedulableRuntime(false) })
	account := &Account{ID: 101, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	gateway := &OpenAIGatewayService{}

	SetDisableTempUnschedulableRuntime(false)
	gateway.BlockAccountScheduling(account, time.Now().Add(time.Hour), "test")
	if !gateway.isOpenAIAccountRuntimeBlocked(account) {
		t.Fatal("runtime blocker must apply when the switch is off")
	}

	SetDisableTempUnschedulableRuntime(true)
	if gateway.isOpenAIAccountRuntimeBlocked(account) {
		t.Fatal("runtime blocker must be ignored when the switch is on")
	}
}

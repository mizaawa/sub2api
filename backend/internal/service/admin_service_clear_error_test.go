//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type accountRepoStubForClearAccountError struct {
	mockAccountRepoForGemini
	account                  *Account
	clearErrorCalls          int
	clearRateLimitCalls      int
	clearAntigravityCalls    int
	clearModelRateLimitCalls int
	clearTempUnschedCalls    int
	setSchedulableCalls      int
}

func (r *accountRepoStubForClearAccountError) GetByID(ctx context.Context, id int64) (*Account, error) {
	return r.account, nil
}

func (r *accountRepoStubForClearAccountError) ClearError(ctx context.Context, id int64) error {
	r.clearErrorCalls++
	r.account.Status = StatusActive
	r.account.ErrorMessage = ""
	return nil
}

func (r *accountRepoStubForClearAccountError) SetSchedulable(_ context.Context, _ int64, schedulable bool) error {
	r.setSchedulableCalls++
	r.account.Schedulable = schedulable
	if r.account.Type == AccountTypeAPIKey {
		r.account.Status = StatusActive
		r.account.ErrorMessage = ""
		if schedulable {
			r.account.RateLimitedAt = nil
			r.account.RateLimitResetAt = nil
			r.account.OverloadUntil = nil
			r.account.TempUnschedulableUntil = nil
			r.account.TempUnschedulableReason = ""
			delete(r.account.Extra, modelRateLimitsKey)
		}
	}
	return nil
}

func (r *accountRepoStubForClearAccountError) ClearRateLimit(ctx context.Context, id int64) error {
	r.clearRateLimitCalls++
	r.account.RateLimitedAt = nil
	r.account.RateLimitResetAt = nil
	return nil
}

func (r *accountRepoStubForClearAccountError) ClearAntigravityQuotaScopes(ctx context.Context, id int64) error {
	r.clearAntigravityCalls++
	return nil
}

func (r *accountRepoStubForClearAccountError) ClearModelRateLimits(ctx context.Context, id int64) error {
	r.clearModelRateLimitCalls++
	return nil
}

func (r *accountRepoStubForClearAccountError) ClearTempUnschedulable(ctx context.Context, id int64) error {
	r.clearTempUnschedCalls++
	r.account.TempUnschedulableUntil = nil
	r.account.TempUnschedulableReason = ""
	return nil
}

func TestAdminService_ClearAccountError_AlsoClearsRecoverableRuntimeState(t *testing.T) {
	until := time.Now().Add(10 * time.Minute)
	resetAt := time.Now().Add(5 * time.Minute)
	repo := &accountRepoStubForClearAccountError{
		account: &Account{
			ID:                      31,
			Platform:                PlatformOpenAI,
			Type:                    AccountTypeOAuth,
			Status:                  StatusError,
			ErrorMessage:            "refresh failed",
			RateLimitResetAt:        &resetAt,
			TempUnschedulableUntil:  &until,
			TempUnschedulableReason: "missing refresh token",
		},
	}
	blocker := &runtimeBlockRecorder{}
	svc := &adminServiceImpl{accountRepo: repo, runtimeBlocker: blocker}

	updated, err := svc.ClearAccountError(context.Background(), 31)
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, 1, repo.clearErrorCalls)
	require.Equal(t, 1, repo.setSchedulableCalls)
	require.Equal(t, 1, repo.clearRateLimitCalls)
	require.Equal(t, 1, repo.clearAntigravityCalls)
	require.Equal(t, 1, repo.clearModelRateLimitCalls)
	require.Equal(t, 1, repo.clearTempUnschedCalls)
	require.Nil(t, updated.RateLimitResetAt)
	require.Nil(t, updated.TempUnschedulableUntil)
	require.Empty(t, updated.TempUnschedulableReason)
	require.True(t, updated.Schedulable)
	require.Equal(t, []int64{31}, blocker.clearedIDs)
}

func TestAdminService_SetAccountSchedulable_APIKeyClearsRuntimeBlock(t *testing.T) {
	blockedUntil := time.Now().Add(time.Hour)
	repo := &accountRepoStubForClearAccountError{
		account: &Account{
			ID: 44, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
			Status: StatusError, ErrorMessage: "legacy error", Schedulable: false,
			RateLimitResetAt: &blockedUntil, OverloadUntil: &blockedUntil,
			TempUnschedulableUntil: &blockedUntil, TempUnschedulableReason: "legacy block",
			Extra: map[string]any{modelRateLimitsKey: map[string]any{"gpt-image-2": map[string]any{}}},
		},
	}
	blocker := &runtimeBlockRecorder{}
	svc := &adminServiceImpl{accountRepo: repo, runtimeBlocker: blocker}

	updated, err := svc.SetAccountSchedulable(context.Background(), 44, true)
	require.NoError(t, err)
	require.Equal(t, StatusActive, updated.Status)
	require.True(t, updated.Schedulable)
	require.Empty(t, updated.ErrorMessage)
	require.Nil(t, updated.RateLimitResetAt)
	require.Nil(t, updated.OverloadUntil)
	require.Nil(t, updated.TempUnschedulableUntil)
	require.NotContains(t, updated.Extra, modelRateLimitsKey)
	require.Equal(t, []int64{44}, blocker.clearedIDs)

	updated, err = svc.SetAccountSchedulable(context.Background(), 44, false)
	require.NoError(t, err)
	require.Equal(t, StatusActive, updated.Status)
	require.False(t, updated.Schedulable)
	require.Equal(t, []int64{44}, blocker.clearedIDs)
}

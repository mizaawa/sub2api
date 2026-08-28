//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSettingService_IsDisableTempUnschedulableEnabled_FailsClosedAndResetsRuntime(t *testing.T) {
	t.Cleanup(func() { SetDisableTempUnschedulableRuntime(false) })
	SetDisableTempUnschedulableRuntime(true)

	svc := NewSettingService(&settingRepoStub{err: errors.New("settings unavailable")}, nil)

	require.False(t, svc.IsDisableTempUnschedulableEnabled(context.Background()))
	require.False(t, IsDisableTempUnschedulableEnabled())

	SetDisableTempUnschedulableRuntime(true)
	svc = NewSettingService(nil, nil)
	require.False(t, svc.IsDisableTempUnschedulableEnabled(context.Background()))
	require.False(t, IsDisableTempUnschedulableEnabled())
}

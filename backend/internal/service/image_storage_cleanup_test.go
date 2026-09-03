package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type cleanupStorage struct {
	deletedPrefix string
	deleted       int64
}

func (s *cleanupStorage) Save(context.Context, string, string, []byte) (string, error) {
	return "", nil
}
func (s *cleanupStorage) DeletePrefix(_ context.Context, prefix string) (int64, error) {
	s.deletedPrefix = prefix
	return s.deleted, nil
}

func TestNextImageCleanupBoundary(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	tests := []struct {
		name string
		now  string
		want string
	}{
		{"midnight", "2026-09-03T00:00:00+08:00", "2026-09-03T00:00:00+08:00"},
		{"after midnight", "2026-09-03T00:01:00+08:00", "2026-09-03T03:00:00+08:00"},
		{"before noon", "2026-09-03T11:59:59+08:00", "2026-09-03T12:00:00+08:00"},
		{"after 21", "2026-09-03T21:00:01+08:00", "2026-09-04T00:00:00+08:00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339, tc.now)
			require.NoError(t, err)
			got := NextImageCleanupBoundary(now.In(loc))
			want, err := time.Parse(time.RFC3339, tc.want)
			require.NoError(t, err)
			require.True(t, want.Equal(got), "got %s, want %s", got, want)
		})
	}
}

func TestNextImageCleanupBoundaryUsesWallClockAcrossDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("timezone data unavailable: %v", err)
	}
	// On the spring transition, adding three absolute hours to midnight would
	// produce 04:00 local. The configured boundary remains 03:00 local.
	spring := time.Date(2026, 3, 8, 1, 30, 0, 0, loc)
	wantSpring := time.Date(2026, 3, 8, 3, 0, 0, 0, loc)
	require.True(t, wantSpring.Equal(NextImageCleanupBoundary(spring)))

	// On the fall transition, adding three absolute hours would produce 02:00
	// local. The next wall-clock boundary is still 03:00 local.
	fall := time.Date(2026, 11, 1, 1, 30, 0, 0, loc)
	wantFall := time.Date(2026, 11, 1, 3, 0, 0, 0, loc)
	require.True(t, wantFall.Equal(NextImageCleanupBoundary(fall)))
}

func TestImageStorageCleanupRunOnceUsesManagedPrefix(t *testing.T) {
	storage := &cleanupStorage{deleted: 4}
	uploader := NewImageResultUploader(storage, "images/", 0, nil)
	settings := &ImageStorageSettingService{}
	settings.uploader = uploader
	settings.enabled = true
	settings.resolved = true
	svc := NewImageStorageCleanupService(settings)
	require.Equal(t, int64(4), svc.RunOnce(context.Background()))
	require.Equal(t, "images/users/", storage.deletedPrefix)
}

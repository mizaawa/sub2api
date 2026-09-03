package service

import (
	"context"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"go.uber.org/zap"
)

// ImageStorageCleanupService removes managed generated images at fixed local
// time boundaries: 00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00.
// It intentionally does not use object age or a rolling 24-hour window.
type ImageStorageCleanupService struct {
	settings *ImageStorageSettingService

	mu      sync.Mutex
	cancel  context.CancelFunc
	started bool
}

func NewImageStorageCleanupService(settings *ImageStorageSettingService) *ImageStorageCleanupService {
	return &ImageStorageCleanupService{settings: settings}
}

// ProvideImageStorageCleanupService creates and starts the periodic worker.
func ProvideImageStorageCleanupService(settings *ImageStorageSettingService) *ImageStorageCleanupService {
	svc := NewImageStorageCleanupService(settings)
	svc.Start()
	return svc
}

func (s *ImageStorageCleanupService) Start() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.started = true
	s.mu.Unlock()
	go s.run(ctx)
}

func (s *ImageStorageCleanupService) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.started = false
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *ImageStorageCleanupService) run(ctx context.Context) {
	for {
		wait := time.Until(NextImageCleanupBoundary(time.Now()))
		if wait < 0 {
			wait = 0
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
			s.RunOnce(ctx)
		}
	}
}

// RunOnce executes one best-effort full-prefix cleanup. A storage outage is
// logged and retried at the next boundary without affecting request handling.
func (s *ImageStorageCleanupService) RunOnce(ctx context.Context) int64 {
	if s == nil || s.settings == nil {
		return 0
	}
	uploader, enabled := s.settings.Resolver()()
	if !enabled || uploader == nil {
		return 0
	}
	deleted, err := uploader.DeleteAll(ctx)
	if err != nil {
		logger.L().Error("image_storage.cleanup_failed", zap.Error(err))
		return 0
	}
	if deleted > 0 {
		logger.L().Info("image_storage.cleanup_completed", zap.Int64("deleted_objects", deleted))
	}
	return deleted
}

// NextImageCleanupBoundary returns the next configured boundary in the
// timestamp's local location. At an exact boundary (including 00:00), it
// returns that same instant so a process started on the boundary cleans now.
func NextImageCleanupBoundary(now time.Time) time.Time {
	local := now.In(now.Location())
	hour := (local.Hour() / 3) * 3
	candidate := time.Date(local.Year(), local.Month(), local.Day(), hour, 0, 0, 0, local.Location())
	if local.Equal(candidate) {
		return candidate
	}
	if !candidate.After(local) {
		nextHour := hour + 3
		if nextHour >= 24 {
			return time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, local.Location())
		}
		// Build the next wall-clock boundary so DST transitions do not shift
		// a configured three-hour boundary by an hour.
		candidate = time.Date(local.Year(), local.Month(), local.Day(), nextHour, 0, 0, 0, local.Location())
	}
	return candidate
}

package lockmanager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

var (
	ErrBackendError = errors.New("lock manager backend error")
	ErrLockHeld     = errors.New("lock held by another owner")
)

type Backend interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (token string, acquired bool, err error)
	Release(ctx context.Context, key, token string) (bool, error)
}

type Manager struct {
	backend    Backend
	logger     *slog.Logger
	defaultTTL time.Duration
}

func New(backend Backend, logger *slog.Logger, defaultTTL time.Duration) *Manager {
	return &Manager{
		backend:    backend,
		logger:     logger,
		defaultTTL: defaultTTL,
	}
}

func (m *Manager) Acquire(ctx context.Context, key string) (string, error) {
	return m.AcquireWithTTL(ctx, key, m.defaultTTL)
}

func (m *Manager) AcquireWithTTL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token, ok, err := m.backend.Acquire(ctx, key, ttl)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrBackendError, err)
	}
	if !ok {
		m.logger.Info("lock held by another owner",
			slog.String("lock_key", key))
		return "", ErrLockHeld
	}
	return token, nil
}

func (m *Manager) Release(ctx context.Context, key, token string) error {
	ok, err := m.backend.Release(ctx, key, token)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBackendError, err)
	}
	if !ok {
		m.logger.Info("lock not released, token mismatch or already released",
			slog.String("lock_key", key))
	}
	return nil
}

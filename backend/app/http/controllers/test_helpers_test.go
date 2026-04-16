package controllers

import (
	"log/slog"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend/app/config"
	"github.com/nishiki/backend/app/container"
	"github.com/nishiki/backend/mocks"
)

// testMocks holds all mock dependencies used across controller tests.
type testMocks struct {
	ContainerRepo  *mocks.MockContainerRepository
	CollectionRepo *mocks.MockCollectionRepository
	AuthService    *mocks.MockAuthService
}

// newTestContainer creates a Container populated with mocks and a discard logger,
// ready for use with any NewXxxController constructor.
func newTestContainer(t *testing.T) (*container.Container, *testMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)

	m := &testMocks{
		ContainerRepo:  mocks.NewMockContainerRepository(ctrl),
		CollectionRepo: mocks.NewMockCollectionRepository(ctrl),
		AuthService:    mocks.NewMockAuthService(ctrl),
	}

	c := &container.Container{
		ContainerRepo:  m.ContainerRepo,
		CollectionRepo: m.CollectionRepo,
		AuthService:    m.AuthService,
	}
	c.SetConfig(&config.Config{})
	c.SetLogger(slog.New(slog.DiscardHandler))

	return c, m
}

package container

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/samber/slog-multi"
	"github.com/swczk/go-seqlogger"

	"github.com/nishiki/backend-go/app/config"
	"github.com/nishiki/backend-go/app/http/middleware"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/domain/services"
	"github.com/nishiki/backend-go/external/adapters"
	extRepos "github.com/nishiki/backend-go/external/repositories"
	extServices "github.com/nishiki/backend-go/external/services"
)

type Container struct {
	config *config.Config
	logger *slog.Logger

	database *adapters.MongoDatabase

	ContainerRepo  repositories.ContainerRepository
	CategoryRepo   repositories.CategoryRepository
	CollectionRepo repositories.CollectionRepository

	AuthService services.AuthService
}

func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{
		config: cfg,
	}

	if err := container.setupLogger(); err != nil {
		return nil, fmt.Errorf("failed to setup logger: %w", err)
	}

	if err := container.setupDatabase(); err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	if err := container.setupRepositories(); err != nil {
		return nil, fmt.Errorf("failed to setup repositories: %w", err)
	}

	if err := container.setupServices(); err != nil {
		return nil, fmt.Errorf("failed to setup services: %w", err)
	}

	return container, nil
}

func (c *Container) setupLogger() error {
	var level slog.Level
	switch c.config.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Always create console JSON handler
	consoleHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	// If Seq endpoint is configured, create multi-handler with both console and Seq
	if c.config.Logging.SeqEndpoint != "" {
		seqConfig := seqlogger.DefaultConfig(c.config.Logging.SeqEndpoint).
			WithLogLevel(level)

		if c.config.Logging.SeqAPIKey != "" {
			seqConfig = seqConfig.WithAPIKey(c.config.Logging.SeqAPIKey)
		}

		seqLogger := seqlogger.New(seqConfig)
		
		// Use slog-multi to combine console and Seq logging
		multiHandler := slogmulti.Fanout(
			consoleHandler,
			seqLogger.Handler(),
		)
		
		c.logger = slog.New(multiHandler)
	} else {
		// Only console logging
		c.logger = slog.New(consoleHandler)
	}

	return nil
}

func (c *Container) setupDatabase() error {
	c.database = adapters.NewMongoDatabase(c.config.Database)

	ctx := context.Background()
	if err := c.database.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	c.logger.Info("Database connected successfully",
		slog.String("database", c.config.Database.Database))

	return nil
}

func (c *Container) setupRepositories() error {
	c.ContainerRepo = extRepos.NewMongoContainerRepository(c.database)
	c.CategoryRepo = extRepos.NewMongoCategoryRepository(c.database)
	c.CollectionRepo = extRepos.NewMongoCollectionRepository(c.database)

	c.logger.Info("Repositories initialized successfully")
	return nil
}

func (c *Container) setupServices() error {
	var err error
	c.AuthService, err = extServices.NewAuthentikAuthService(c.config.Auth, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create auth service: %w", err)
	}

	c.logger.Info("Services initialized successfully")
	return nil
}

func (c *Container) Close() error {
	if c.database != nil {
		ctx := context.Background()
		if err := c.database.Disconnect(ctx); err != nil {
			c.logger.Error("Failed to disconnect from database", slog.Any("error", err))
			return err
		}
	}

	return nil
}

// Getters
func (c *Container) GetConfig() *config.Config {
	return c.config
}

func (c *Container) GetLogger() *slog.Logger {
	return c.logger
}

func (c *Container) GetAuthMiddleware() *middleware.AuthMiddleware {
	return middleware.NewAuthMiddleware(c.AuthService, c.logger)
}

// SetLogger sets the logger (primarily for testing purposes)
func (c *Container) SetLogger(logger *slog.Logger) {
	c.logger = logger
}

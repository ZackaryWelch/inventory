package adapters

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/nishiki/backend-go/app/config"
	"github.com/nishiki/backend-go/domain/adapters"
)

type MongoDatabase struct {
	client   *mongo.Client
	database *mongo.Database
	config   config.DatabaseConfig
}

type MongoTransaction struct {
	session *mongo.Session
	ctx     context.Context
}

func NewMongoDatabase(config config.DatabaseConfig) *MongoDatabase {
	return &MongoDatabase{
		config: config,
	}
}

func (m *MongoDatabase) Connect(ctx context.Context) error {
	timeout := time.Duration(m.config.Timeout) * time.Second
	uri := m.config.GetURI()
	clientOptions := options.Client().ApplyURI(uri).SetTimeout(timeout)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := client.Ping(timeoutCtx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.config.Database)

	return nil
}

func (m *MongoDatabase) Disconnect(ctx context.Context) error {
	if m.client == nil {
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return m.client.Disconnect(timeoutCtx)
}

func (m *MongoDatabase) Health(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("MongoDB client not connected")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return m.client.Ping(timeoutCtx, readpref.Primary())
}

func (m *MongoDatabase) StartTransaction(ctx context.Context) (adapters.Transaction, error) {
	if m.client == nil {
		return nil, fmt.Errorf("MongoDB client not connected")
	}

	session, err := m.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start MongoDB session: %w", err)
	}

	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return nil, fmt.Errorf("failed to start MongoDB transaction: %w", err)
	}

	return &MongoTransaction{
		session: session,
		ctx:     mongo.NewSessionContext(ctx, session),
	}, nil
}

func (m *MongoDatabase) Database() *mongo.Database {
	return m.database
}

func (m *MongoDatabase) Client() *mongo.Client {
	return m.client
}

func (t *MongoTransaction) Commit(ctx context.Context) error {
	defer t.session.EndSession(ctx)
	return t.session.CommitTransaction(ctx)
}

func (t *MongoTransaction) Rollback(ctx context.Context) error {
	defer t.session.EndSession(ctx)
	return t.session.AbortTransaction(ctx)
}

func (t *MongoTransaction) Context() context.Context {
	return t.ctx
}

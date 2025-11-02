package repositories

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/repositories"
	"github.com/nishiki/backend-go/external/adapters"
)

type objectDocument struct {
	ID         string                 `bson:"id"`
	Name       string                 `bson:"name"`
	ObjectType string                 `bson:"object_type"`
	Properties map[string]interface{} `bson:"properties"`
	Tags       []string               `bson:"tags"`
	CreatedAt  time.Time              `bson:"created_at"`
}

type containerDocument struct {
	ID                string           `bson:"_id"`
	CollectionID      string           `bson:"collection_id"`
	Name              string           `bson:"name"`
	Type              string           `bson:"type"`
	ParentContainerID *string          `bson:"parent_container_id,omitempty"`
	CategoryID        *string          `bson:"category_id,omitempty"`
	Objects           []objectDocument `bson:"objects"`
	Location          string           `bson:"location"`
	Width             *float64         `bson:"width,omitempty"`
	Depth             *float64         `bson:"depth,omitempty"`
	Rows              *int             `bson:"rows,omitempty"`
	Capacity          *float64         `bson:"capacity,omitempty"`
	CreatedAt         time.Time        `bson:"created_at"`
	UpdatedAt         time.Time        `bson:"updated_at"`
}

type MongoContainerRepository struct {
	db         *adapters.MongoDatabase
	collection *mongo.Collection
}

func NewMongoContainerRepository(db *adapters.MongoDatabase) repositories.ContainerRepository {
	return &MongoContainerRepository{
		db:         db,
		collection: db.Database().Collection("containers"),
	}
}

func (r *MongoContainerRepository) Create(ctx context.Context, container *entities.Container) error {
	doc := containerToDocument(container)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("container already exists: %w", err)
		}
		return fmt.Errorf("failed to create container: %w", err)
	}

	return nil
}

func (r *MongoContainerRepository) GetByID(ctx context.Context, id entities.ContainerID) (*entities.Container, error) {
	var doc containerDocument

	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("container not found")
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	return documentToContainer(&doc)
}

func (r *MongoContainerRepository) Update(ctx context.Context, container *entities.Container) error {
	doc := containerToDocument(container)

	filter := bson.M{"_id": container.ID().String()}
	update := bson.M{"$set": doc}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update container: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("container not found")
	}

	return nil
}

func (r *MongoContainerRepository) Delete(ctx context.Context, id entities.ContainerID) error {
	filter := bson.M{"_id": id.String()}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("container not found")
	}

	return nil
}

func (r *MongoContainerRepository) GetByGroupID(ctx context.Context, groupID entities.GroupID) ([]*entities.Container, error) {
	filter := bson.M{"group_id": groupID.String()}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers by group ID: %w", err)
	}
	defer cursor.Close(ctx)

	var containers []*entities.Container
	for cursor.Next(ctx) {
		var doc containerDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode container: %w", err)
		}

		container, err := documentToContainer(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert container: %w", err)
		}

		containers = append(containers, container)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return containers, nil
}

func (r *MongoContainerRepository) GetByCollectionID(ctx context.Context, collectionID entities.CollectionID) ([]*entities.Container, error) {
	filter := bson.M{"collection_id": collectionID.String()}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers by collection ID: %w", err)
	}
	defer cursor.Close(ctx)

	var containers []*entities.Container
	for cursor.Next(ctx) {
		var doc containerDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode container: %w", err)
		}

		container, err := documentToContainer(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert container: %w", err)
		}

		containers = append(containers, container)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return containers, nil
}

func (r *MongoContainerRepository) GetChildContainers(ctx context.Context, parentID entities.ContainerID) ([]*entities.Container, error) {
	filter := bson.M{"parent_container_id": parentID.String()}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get child containers: %w", err)
	}
	defer cursor.Close(ctx)

	var containers []*entities.Container
	for cursor.Next(ctx) {
		var doc containerDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode container: %w", err)
		}

		container, err := documentToContainer(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert container: %w", err)
		}

		containers = append(containers, container)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return containers, nil
}

func (r *MongoContainerRepository) List(ctx context.Context, limit, offset int) ([]*entities.Container, error) {
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	opts.SetSort(bson.M{"created_at": 1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	defer cursor.Close(ctx)

	var containers []*entities.Container
	for cursor.Next(ctx) {
		var doc containerDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode container: %w", err)
		}

		container, err := documentToContainer(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert container: %w", err)
		}

		containers = append(containers, container)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return containers, nil
}

func (r *MongoContainerRepository) Exists(ctx context.Context, id entities.ContainerID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return false, fmt.Errorf("failed to check container existence: %w", err)
	}

	return count > 0, nil
}

func (r *MongoContainerRepository) GetContainersWithExpiredFood(ctx context.Context, groupID entities.GroupID) ([]*entities.Container, error) {
	now := time.Now()
	filter := bson.M{
		"group_id":     groupID.String(),
		"foods.expiry": bson.M{"$lt": now},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers with expired food: %w", err)
	}
	defer cursor.Close(ctx)

	var containers []*entities.Container
	for cursor.Next(ctx) {
		var doc containerDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode container: %w", err)
		}

		container, err := documentToContainer(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert container: %w", err)
		}

		containers = append(containers, container)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return containers, nil
}

func containerToDocument(container *entities.Container) *containerDocument {
	objects := make([]objectDocument, len(container.Objects()))
	for i, object := range container.Objects() {
		objects[i] = objectToDocument(object)
	}

	var categoryID *string
	if container.CategoryID() != nil {
		id := container.CategoryID().String()
		categoryID = &id
	}

	var parentContainerID *string
	if container.ParentContainerID() != nil {
		id := container.ParentContainerID().String()
		parentContainerID = &id
	}

	return &containerDocument{
		ID:                container.ID().String(),
		CollectionID:      container.CollectionID().String(),
		Name:              container.Name().String(),
		Type:              string(container.ContainerType()),
		ParentContainerID: parentContainerID,
		CategoryID:        categoryID,
		Objects:           objects,
		Location:          container.Location(),
		Width:             container.Width(),
		Depth:             container.Depth(),
		Rows:              container.Rows(),
		Capacity:          container.Capacity(),
		CreatedAt:         container.CreatedAt(),
		UpdatedAt:         container.UpdatedAt(),
	}
}

func documentToContainer(doc *containerDocument) (*entities.Container, error) {
	id, err := entities.ContainerIDFromString(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid container ID: %w", err)
	}

	collectionID, err := entities.CollectionIDFromString(doc.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid collection ID: %w", err)
	}

	name, err := entities.NewContainerName(doc.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid container name: %w", err)
	}

	// Parse container type, default to general if not specified or invalid
	containerType := entities.ContainerTypeGeneral
	if doc.Type != "" {
		containerType = entities.ContainerType(doc.Type)
	}

	var categoryID *entities.CategoryID
	if doc.CategoryID != nil {
		cid, err := entities.CategoryIDFromHex(*doc.CategoryID)
		if err == nil {
			categoryID = &cid
		}
	}

	var parentContainerID *entities.ContainerID
	if doc.ParentContainerID != nil {
		pid, err := entities.ContainerIDFromString(*doc.ParentContainerID)
		if err == nil {
			parentContainerID = &pid
		}
	}

	objects := make([]entities.Object, len(doc.Objects))
	for i, objectDoc := range doc.Objects {
		object, err := documentToObject(objectDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert object: %w", err)
		}
		objects[i] = *object
	}

	return entities.ReconstructContainer(
		id,
		collectionID,
		name,
		containerType,
		parentContainerID,
		categoryID,
		objects,
		doc.Location,
		doc.Width,
		doc.Depth,
		doc.Rows,
		doc.Capacity,
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

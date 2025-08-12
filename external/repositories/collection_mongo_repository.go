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

type collectionDocument struct {
	ID         string    `bson:"_id"`
	UserID     string    `bson:"user_id"`
	GroupID    *string   `bson:"group_id,omitempty"`
	Name       string    `bson:"name"`
	CategoryID *string   `bson:"category_id,omitempty"`
	ObjectType string    `bson:"object_type"`
	Containers []string  `bson:"containers"` // Store container IDs, containers are stored separately
	Tags       []string  `bson:"tags"`
	Location   string    `bson:"location"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
}

type MongoCollectionRepository struct {
	db         *adapters.MongoDatabase
	collection *mongo.Collection
}

func NewMongoCollectionRepository(db *adapters.MongoDatabase) repositories.CollectionRepository {
	return &MongoCollectionRepository{
		db:         db,
		collection: db.Database().Collection("collections"),
	}
}

func (r *MongoCollectionRepository) Create(ctx context.Context, collection *entities.Collection) error {
	doc := collectionToDocument(collection)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("collection already exists: %w", err)
		}
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

func (r *MongoCollectionRepository) GetByID(ctx context.Context, id entities.CollectionID) (*entities.Collection, error) {
	var doc collectionDocument

	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("collection not found")
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	return documentToCollection(&doc)
}

func (r *MongoCollectionRepository) Update(ctx context.Context, collection *entities.Collection) error {
	doc := collectionToDocument(collection)

	filter := bson.M{"_id": collection.ID().String()}
	update := bson.M{"$set": doc}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("collection not found")
	}

	return nil
}

func (r *MongoCollectionRepository) Delete(ctx context.Context, id entities.CollectionID) error {
	filter := bson.M{"_id": id.String()}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("collection not found")
	}

	return nil
}

func (r *MongoCollectionRepository) GetByUserID(ctx context.Context, userID entities.UserID) ([]*entities.Collection, error) {
	filter := bson.M{"user_id": userID.String()}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections by user ID: %w", err)
	}
	defer cursor.Close(ctx)

	var collections []*entities.Collection
	for cursor.Next(ctx) {
		var doc collectionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode collection: %w", err)
		}

		collection, err := documentToCollection(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert collection: %w", err)
		}

		collections = append(collections, collection)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return collections, nil
}

func (r *MongoCollectionRepository) GetByGroupID(ctx context.Context, groupID entities.GroupID) ([]*entities.Collection, error) {
	filter := bson.M{"group_id": groupID.String()}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections by group ID: %w", err)
	}
	defer cursor.Close(ctx)

	var collections []*entities.Collection
	for cursor.Next(ctx) {
		var doc collectionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode collection: %w", err)
		}

		collection, err := documentToCollection(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert collection: %w", err)
		}

		collections = append(collections, collection)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return collections, nil
}

func (r *MongoCollectionRepository) List(ctx context.Context, limit, offset int) ([]*entities.Collection, error) {
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
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var collections []*entities.Collection
	for cursor.Next(ctx) {
		var doc collectionDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode collection: %w", err)
		}

		collection, err := documentToCollection(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert collection: %w", err)
		}

		collections = append(collections, collection)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return collections, nil
}

func (r *MongoCollectionRepository) Exists(ctx context.Context, id entities.CollectionID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return count > 0, nil
}

func collectionToDocument(collection *entities.Collection) *collectionDocument {
	containerIDs := make([]string, len(collection.Containers()))
	for i, container := range collection.Containers() {
		containerIDs[i] = container.ID().String()
	}

	doc := &collectionDocument{
		ID:         collection.ID().String(),
		UserID:     collection.UserID().String(),
		Name:       collection.Name().String(),
		ObjectType: collection.ObjectType().String(),
		Containers: containerIDs,
		Tags:       collection.Tags(),
		Location:   collection.Location(),
		CreatedAt:  collection.CreatedAt(),
		UpdatedAt:  collection.UpdatedAt(),
	}

	if collection.CategoryID() != nil {
		categoryID := collection.CategoryID().String()
		doc.CategoryID = &categoryID
	}

	if collection.GroupID() != nil {
		groupIDStr := collection.GroupID().String()
		doc.GroupID = &groupIDStr
	}

	return doc
}

func objectToDocument(object entities.Object) objectDocument {
	return objectDocument{
		ID:         object.ID().String(),
		Name:       object.Name().String(),
		ObjectType: object.ObjectType().String(),
		Properties: object.Properties(),
		Tags:       object.Tags(),
		CreatedAt:  object.CreatedAt(),
	}
}

func documentToCollection(doc *collectionDocument) (*entities.Collection, error) {
	id, err := entities.CollectionIDFromString(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid collection ID: %w", err)
	}

	userID, err := entities.UserIDFromString(doc.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var groupID *entities.GroupID
	if doc.GroupID != nil {
		gid, err := entities.GroupIDFromString(*doc.GroupID)
		if err != nil {
			return nil, fmt.Errorf("invalid group ID: %w", err)
		}
		groupID = &gid
	}

	name, err := entities.NewCollectionName(doc.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid collection name: %w", err)
	}

	var categoryID *entities.CategoryID
	if doc.CategoryID != nil {
		cid, err := entities.CategoryIDFromHex(*doc.CategoryID)
		if err == nil {
			categoryID = &cid
		}
	}

	// For now, we'll reconstruct collections with empty containers
	// In a full implementation, we'd load containers separately
	containers := make([]entities.Container, 0)

	return entities.ReconstructCollection(
		id,
		userID,
		groupID,
		name,
		categoryID,
		entities.ObjectType(doc.ObjectType),
		containers,
		doc.Tags,
		doc.Location,
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

func documentToObject(doc objectDocument) (*entities.Object, error) {
	id, err := entities.ObjectIDFromHex(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	name, err := entities.NewObjectName(doc.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid object name: %w", err)
	}

	return entities.ReconstructObject(
		id,
		name,
		entities.ObjectType(doc.ObjectType),
		doc.Properties,
		doc.Tags,
		doc.CreatedAt,
	), nil
}

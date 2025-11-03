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

type categoryDocument struct {
	ID          bson.ObjectID `bson:"_id"`
	Name        string        `bson:"name"`
	Description string        `bson:"description"`
	Icon        string        `bson:"icon"`
	Color       string        `bson:"color"`
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
}

type MongoCategoryRepository struct {
	db         *adapters.MongoDatabase
	collection *mongo.Collection
}

func NewMongoCategoryRepository(db *adapters.MongoDatabase) repositories.CategoryRepository {
	return &MongoCategoryRepository{
		db:         db,
		collection: db.Database().Collection("categories"),
	}
}

func (r *MongoCategoryRepository) Create(ctx context.Context, category *entities.Category) error {
	doc := categoryToDocument(category)

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("category already exists: %w", err)
		}
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *MongoCategoryRepository) GetByID(ctx context.Context, id entities.CategoryID) (*entities.Category, error) {
	var doc categoryDocument

	err := r.collection.FindOne(ctx, bson.M{"_id": id.ObjectID()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return documentToCategory(&doc)
}

func (r *MongoCategoryRepository) GetByName(ctx context.Context, name entities.CategoryName) (*entities.Category, error) {
	var doc categoryDocument

	err := r.collection.FindOne(ctx, bson.M{"name": name.String()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return documentToCategory(&doc)
}

func (r *MongoCategoryRepository) List(ctx context.Context, limit, offset int) ([]*entities.Category, error) {
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	opts.SetSort(bson.M{"name": 1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer cursor.Close(ctx)

	var categories []*entities.Category
	for cursor.Next(ctx) {
		var doc categoryDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode category: %w", err)
		}

		category, err := documentToCategory(&doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert category: %w", err)
		}

		categories = append(categories, category)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return categories, nil
}

func (r *MongoCategoryRepository) Update(ctx context.Context, category *entities.Category) error {
	doc := categoryToDocument(category)

	filter := bson.M{"_id": category.ID().ObjectID()}
	update := bson.M{"$set": doc}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *MongoCategoryRepository) Delete(ctx context.Context, id entities.CategoryID) error {
	filter := bson.M{"_id": id.ObjectID()}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *MongoCategoryRepository) Exists(ctx context.Context, id entities.CategoryID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id.ObjectID()})
	if err != nil {
		return false, fmt.Errorf("failed to check category existence: %w", err)
	}

	return count > 0, nil
}

func categoryToDocument(category *entities.Category) *categoryDocument {
	return &categoryDocument{
		ID:          category.ID().ObjectID(),
		Name:        category.Name().String(),
		Description: category.Description().String(),
		Icon:        category.Icon(),
		Color:       category.Color(),
		CreatedAt:   category.CreatedAt(),
		UpdatedAt:   category.UpdatedAt(),
	}
}

func documentToCategory(doc *categoryDocument) (*entities.Category, error) {
	id := entities.CategoryIDFromObjectID(doc.ID)

	name, err := entities.NewCategoryName(doc.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid category name: %w", err)
	}

	description := entities.NewCategoryDescription(doc.Description)

	return entities.ReconstructCategory(
		id,
		name,
		description,
		doc.Icon,
		doc.Color,
		doc.CreatedAt,
		doc.UpdatedAt,
	), nil
}

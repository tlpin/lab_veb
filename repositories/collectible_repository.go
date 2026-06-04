package repositories

import (
	"context"
	"time"

	"newyear-api/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CollectibleRepository struct {
	collection *mongo.Collection
}

func NewCollectibleRepository(db *mongo.Database) *CollectibleRepository {
	return &CollectibleRepository{
		collection: db.Collection("collectibles"),
	}
}

func (r *CollectibleRepository) Create(collectible *models.Collectible) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()

	if collectible.ID == "" {
		collectible.ID = uuid.New().String()
	}
	if collectible.CreatedAt.IsZero() {
		collectible.CreatedAt = now
	}
	collectible.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, collectible)
	return err
}

func (r *CollectibleRepository) GetAllByUser(userID string, limit, offset int) ([]models.Collectible, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var collectibles []models.Collectible
	if err := cursor.All(ctx, &collectibles); err != nil {
		return nil, 0, err
	}

	return collectibles, total, nil
}

func (r *CollectibleRepository) GetByID(id string) (*models.Collectible, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": id,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	var collectible models.Collectible
	err := r.collection.FindOne(ctx, filter).Decode(&collectible)
	if err != nil {
		return nil, err
	}

	return &collectible, nil
}

func (r *CollectibleRepository) Update(collectible *models.Collectible) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collectible.UpdatedAt = time.Now()

	filter := bson.M{
		"_id": collectible.ID,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	_, err := r.collection.ReplaceOne(ctx, filter, collectible)
	return err
}

func (r *CollectibleRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()

	filter := bson.M{
		"_id": id,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

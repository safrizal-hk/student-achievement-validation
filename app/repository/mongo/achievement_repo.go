package repository

import (
	"context"
	"time"

	model_mongo "github.com/safrizal-hk/uas-gofiber/app/model/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementMongoRepository interface {
	Create(ctx context.Context, achievement *model_mongo.AchievementMongo) (*model_mongo.AchievementMongo, error)
	SoftDelete(ctx context.Context, id primitive.ObjectID) error
	GetDetailByMongoIDs(ctx context.Context, ids []primitive.ObjectID) ([]model_mongo.AchievementMongo, error) // ⚠️ KONTRAK BARU
}

type achievementMongoRepositoryImpl struct {
	Collection *mongo.Collection
}

func NewAchievementMongoRepository(db *mongo.Database) AchievementMongoRepository {
	return &achievementMongoRepositoryImpl{
		Collection: db.Collection("achievements"),
	}
}

func (r *achievementMongoRepositoryImpl) Create(ctx context.Context, achievement *model_mongo.AchievementMongo) (*model_mongo.AchievementMongo, error) {
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	result, err := r.Collection.InsertOne(ctx, achievement)
	if err != nil {
		return nil, err
	}
	achievement.ID = result.InsertedID.(primitive.ObjectID)
	return achievement, nil
}

func (r *achievementMongoRepositoryImpl) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
    now := time.Now()
    update := bson.M{
        "$set": bson.M{
            "deletedAt": now,
            "updatedAt": now,
        },
    }
    _, err := r.Collection.UpdateByID(ctx, id, update)
    return err
}

// GetDetailByMongoIDs mengambil detail prestasi dari MongoDB berdasarkan daftar ObjectID
func (r *achievementMongoRepositoryImpl) GetDetailByMongoIDs(ctx context.Context, ids []primitive.ObjectID) ([]model_mongo.AchievementMongo, error) {
	// Query untuk mencari semua dokumen yang _id nya ada di dalam array 'ids'
	filter := bson.M{
		"_id": bson.M{"$in": ids},
		// Tambahkan filter soft delete agar hanya yang aktif yang terlihat (jika perlu)
		"deletedAt": nil, 
	}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []model_mongo.AchievementMongo
	
	// Dekode semua hasil sekaligus
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}


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
	GetDetailByID(ctx context.Context, id primitive.ObjectID) (*model_mongo.AchievementMongo, error)
	GetDetailsByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model_mongo.AchievementMongo, error)
	Update(ctx context.Context, id primitive.ObjectID, data *model_mongo.AchievementInput) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error
	DeleteByID(ctx context.Context, id primitive.ObjectID) error // Hard Delete (Rollback)
	AddAttachment(ctx context.Context, id primitive.ObjectID, attachment model_mongo.Attachment) error
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
	if err != nil { return nil, err }
	achievement.ID = result.InsertedID.(primitive.ObjectID)
	return achievement, nil
}

func (r *achievementMongoRepositoryImpl) GetDetailByID(ctx context.Context, id primitive.ObjectID) (*model_mongo.AchievementMongo, error) {
	var achievement model_mongo.AchievementMongo
	filter := bson.M{"_id": id, "deletedAt": nil}
	err := r.Collection.FindOne(ctx, filter).Decode(&achievement)
	if err != nil { return nil, err }
	return &achievement, nil
}

func (r *achievementMongoRepositoryImpl) GetDetailsByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model_mongo.AchievementMongo, error) {
	filter := bson.M{"_id": bson.M{"$in": ids}, "deletedAt": nil}
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil { return nil, err }
	defer cursor.Close(ctx)

	var achievements []model_mongo.AchievementMongo
	if err = cursor.All(ctx, &achievements); err != nil { return nil, err }
	return achievements, nil
}

func (r *achievementMongoRepositoryImpl) Update(ctx context.Context, id primitive.ObjectID, data *model_mongo.AchievementInput) error {
	update := bson.M{
		"$set": bson.M{
			"achievementType": data.AchievementType,
			"title":           data.Title,
			"description":     data.Description,
			"details":         data.Details,
			"tags":            data.Tags,
			"points":          data.Points,
			"updatedAt":       time.Now(),
		},
	}
	_, err := r.Collection.UpdateByID(ctx, id, update)
	return err
}

func (r *achievementMongoRepositoryImpl) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	update := bson.M{"$set": bson.M{"deletedAt": now, "updatedAt": now}}
	_, err := r.Collection.UpdateByID(ctx, id, update)
	return err
}

func (r *achievementMongoRepositoryImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *achievementMongoRepositoryImpl) AddAttachment(ctx context.Context, id primitive.ObjectID, attachment model_mongo.Attachment) error {
	update := bson.M{
		"$push": bson.M{"attachments": attachment},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	_, err := r.Collection.UpdateByID(ctx, id, update)
	return err
}
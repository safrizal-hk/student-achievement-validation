package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ReportMongoRepository Interface
type ReportMongoRepository interface {
	GetAchievementStatistics(ctx context.Context, studentIDs []string) ([]bson.M, error)
	GetStudentAchievementDetails(ctx context.Context, studentIDHex string) ([]bson.M, error)
}

type reportMongoRepositoryImpl struct {
	Collection *mongo.Collection
}

func NewReportMongoRepository(db *mongo.Database) ReportMongoRepository {
	return &reportMongoRepositoryImpl{
		Collection: db.Collection("achievements"),
	}
}

// GetAchievementStatistics: Total per Tipe, Distribusi Level (FR-011)
func (r *reportMongoRepositoryImpl) GetAchievementStatistics(ctx context.Context, studentIDs []string) ([]bson.M, error) {
	// Filter berdasarkan studentIDs yang diberikan (atau semua jika studentIDs kosong)
	var matchStage bson.D
	filter := bson.M{"deletedAt": nil}
	if len(studentIDs) > 0 {
		filter["studentId"] = bson.M{"$in": studentIDs}
	}
	matchStage = bson.D{{Key: "$match", Value: filter}}

	// Agregasi untuk Total per Tipe & Distribusi Level Kompetisi
	groupStage := bson.D{
		{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"type": "$achievementType",
				"level": "$details.competitionLevel", 
			},
			"count": bson.M{"$sum": 1},
		}},
	}

	pipeline := mongo.Pipeline{matchStage, groupStage}
	
	cursor, err := r.Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetStudentAchievementDetails: Mengambil detail prestasi untuk report spesifik mahasiswa
func (r *reportMongoRepositoryImpl) GetStudentAchievementDetails(ctx context.Context, studentIDHex string) ([]bson.M, error) {
	filter := bson.M{"studentId": studentIDHex, "deletedAt": nil}
	
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var achievements []bson.M
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}
	return achievements, nil
}
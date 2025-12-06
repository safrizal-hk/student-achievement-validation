package model

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AchievementInput: Payload dari Frontend
type AchievementInput struct {
	AchievementType string                 `json:"achievementType" validate:"required"`
	Title           string                 `json:"title" validate:"required"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Attachments     []Attachment           `json:"attachments"`
	Tags            []string               `json:"tags"`
	Points          int                    `json:"points"`
}

// AchievementMongo: Dokumen di Database MongoDB
type AchievementMongo struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	StudentID       string                 `bson:"studentId" json:"student_id"` // Referensi ke PG
	AchievementType string                 `bson:"achievementType" json:"achievement_type"`
	Title           string                 `bson:"title" json:"title"`
	Description     string                 `bson:"description" json:"description"`
	Details         map[string]interface{} `bson:"details" json:"details"`
	Attachments     []Attachment           `bson:"attachments" json:"attachments"`
	Tags            []string               `bson:"tags" json:"tags"`
	Points          int                    `bson:"points" json:"points"`
	
	CreatedAt       time.Time              `bson:"createdAt" json:"created_at"`
	UpdatedAt       time.Time              `bson:"updatedAt" json:"updated_at"`
	DeletedAt       *time.Time             `bson:"deletedAt,omitempty" json:"deleted_at,omitempty"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"file_name"`
	FileUrl    string    `bson:"fileUrl" json:"file_url"`
	FileType   string    `bson:"fileType" json:"file_type"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploaded_at"`
}
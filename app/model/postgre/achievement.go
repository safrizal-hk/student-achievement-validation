package model

import "time"

type AchievementStatus string 

const (
	StatusDraft     AchievementStatus = "draft"
	StatusSubmitted AchievementStatus = "submitted"
	StatusVerified  AchievementStatus = "verified"
	StatusRejected  AchievementStatus = "rejected"
	StatusDeleted   AchievementStatus = "deleted"
)

type AchievementReference struct {
	ID                 string            `json:"id"`
	StudentID          string            `json:"student_id"`
	MongoAchievementID string            `json:"mongo_achievement_id"` 
	Status             AchievementStatus `json:"status"`
	SubmittedAt        *time.Time        `json:"submitted_at"`
	VerifiedAt         *time.Time        `json:"verified_at"`
	VerifiedBy         *string           `json:"verified_by"` 
	RejectionNote      *string           `json:"rejection_note"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}
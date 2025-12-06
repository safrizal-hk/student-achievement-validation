package repository

import (
	"database/sql"
	"errors"
	"fmt"

	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
)

type ReportPGRepository interface {
	FindStudentIdByUserID(userID string) (string, error)
	FindStudentProfile(studentID string) (*model_postgre.Student, error)
	GetStudentAchievementReferences(studentID string) ([]model_postgre.AchievementReference, error)
}

type reportPGRepositoryImpl struct {
	DB *sql.DB
}

func NewReportPGRepository(db *sql.DB) ReportPGRepository {
	return &reportPGRepositoryImpl{DB: db}
}

func (r *reportPGRepositoryImpl) FindStudentIdByUserID(userID string) (string, error) {
	var studentID string
	query := `SELECT id FROM students WHERE user_id = $1`
	err := r.DB.QueryRow(query, userID).Scan(&studentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil 
		}
		return "", err
	}
	return studentID, nil
}

func (r *reportPGRepositoryImpl) FindStudentProfile(studentID string) (*model_postgre.Student, error) {
	profile := new(model_postgre.Student) 
	
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at 
		FROM students 
		WHERE id = $1
	`
	err := r.DB.QueryRow(query, studentID).Scan(
		&profile.ID, 
		&profile.UserID, 
		&profile.StudentID, // NIM
		&profile.ProgramStudy, 
		&profile.AcademicYear, 
		&profile.AdvisorID,
		&profile.CreatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("mahasiswa tidak ditemukan")
		}
		return nil, err
	}
	return profile, nil
}

func (r *reportPGRepositoryImpl) GetStudentAchievementReferences(studentID string) ([]model_postgre.AchievementReference, error) {
	query := `
		SELECT 
			id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, 
			verified_by, rejection_note, created_at
		FROM achievement_references
		WHERE student_id = $1 AND status != $2
	`
	rows, err := r.DB.Query(query, studentID, model_postgre.StatusDeleted)
	if err != nil {
		return nil, fmt.Errorf("SQL error in report repo: %w", err)
	}
	defer rows.Close()

	var references []model_postgre.AchievementReference
	for rows.Next() {
		var ref model_postgre.AchievementReference
		
		err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&ref.SubmittedAt, &ref.VerifiedAt, &ref.UpdatedAt,
			&ref.VerifiedBy, &ref.RejectionNote,
			&ref.CreatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning row in GetStudentAchievementReferences: %w", err)
		}
		references = append(references, ref)
	}
	return references, nil
}
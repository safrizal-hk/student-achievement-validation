package repository

import (
	"database/sql"
	"errors"
	"fmt"

	// ⚠️ KOREKSI: Gunakan alias yang konsisten
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
)

// ReportPGRepository Interface
type ReportPGRepository interface {
	// FindStudentIdByUserID (Lookup User -> Student)
	FindStudentIdByUserID(userID string) (string, error)
	
	// FindStudentProfile mengambil data profil mahasiswa (Sederhana)
	// NOTE: Pastikan struct Student ada di model/postgre
	FindStudentProfile(studentID string) (*model_postgre.Student, error)
	
	// GetStudentAchievementReferences mengambil daftar prestasi mahasiswa (kecuali deleted)
	GetStudentAchievementReferences(studentID string) ([]model_postgre.AchievementReference, error)
}

type reportPGRepositoryImpl struct {
	DB *sql.DB
}

func NewReportPGRepository(db *sql.DB) ReportPGRepository {
	return &reportPGRepositoryImpl{DB: db}
}

// Implementasi FindStudentIdByUserID
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

// Implementasi FindStudentProfile
func (r *reportPGRepositoryImpl) FindStudentProfile(studentID string) (*model_postgre.Student, error) {
	// Pastikan struct model_postgre.Student sudah didefinisikan di app/model/postgre/student.go
	profile := new(model_postgre.Student) 
	
	// Query mengambil data dasar mahasiswa
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

// Implementasi GetStudentAchievementReferences
func (r *reportPGRepositoryImpl) GetStudentAchievementReferences(studentID string) ([]model_postgre.AchievementReference, error) {
	// Ambil semua field yang diperlukan, termasuk created_at
	query := `
		SELECT 
			id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, 
			verified_by, rejection_note, created_at
		FROM achievement_references
		WHERE student_id = $1 AND status != $2
	`
	// Menggunakan StatusDeleted dari model_postgre
	rows, err := r.DB.Query(query, studentID, model_postgre.StatusDeleted)
	if err != nil {
		return nil, fmt.Errorf("SQL error in report repo: %w", err)
	}
	defer rows.Close()

	var references []model_postgre.AchievementReference
	for rows.Next() {
		var ref model_postgre.AchievementReference
		
		// ⚠️ KOREKSI: Scan semua variabel dengan urutan Pointer dan Non-Pointer yang benar
		// (Asumsi CreatedAt dan UpdatedAt di akhir)
		err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&ref.SubmittedAt, &ref.VerifiedAt, &ref.UpdatedAt,
			&ref.VerifiedBy, &ref.RejectionNote, // Pointers
			&ref.CreatedAt, // Non-Pointer
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning row in GetStudentAchievementReferences: %w", err)
		}
		references = append(references, ref)
	}
	return references, nil
}
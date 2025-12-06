package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq" 
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
)

type AchievementPGRepository interface {
	CreateReference(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error)
	GetReferenceByID(id string) (*model_postgre.AchievementReference, error)
	
	UpdateStatusToSubmitted(refID string) (*model_postgre.AchievementReference, error)
	VerifyAchievement(refID string, verifierID string) (*model_postgre.AchievementReference, error)
	RejectAchievement(refID string, verifierID string, rejectionNote string) (*model_postgre.AchievementReference, error)
	SoftDeleteReference(refID string, studentID string) (*model_postgre.AchievementReference, error)
	
	FindStudentIdByUserID(userID string) (string, error)
	FindLecturerIdByUserID(userID string) (string, error)
	GetAdviseeStudentIDs(lecturerID string) ([]string, error)
	
	GetMyAchievements(studentID string) ([]model_postgre.AchievementReference, error)
	GetAchievementsByStudentIDs(studentIDs []string) ([]model_postgre.AchievementReference, error)
	GetAllAchievementReferences() ([]model_postgre.AchievementReference, error)
}

type achievementPGRepositoryImpl struct {
	DB *sql.DB 
}

func NewAchievementPGRepository(db *sql.DB) AchievementPGRepository {
	return &achievementPGRepositoryImpl{DB: db}
}

func scanAchievementRow(scan func(dest ...interface{}) error) (*model_postgre.AchievementReference, error) {
	ref := new(model_postgre.AchievementReference)
	err := scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
		&ref.SubmittedAt, &ref.VerifiedAt, &ref.UpdatedAt,
		&ref.VerifiedBy, &ref.RejectionNote, &ref.CreatedAt,
	)
	return ref, err
}

func (r *achievementPGRepositoryImpl) CreateReference(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error) {
	query := `
		INSERT INTO achievement_references (student_id, mongo_achievement_id, status)
		VALUES ($1, $2, $3) RETURNING id, created_at, updated_at
	`
	ref.Status = model_postgre.StatusDraft 
	err := r.DB.QueryRow(query, ref.StudentID, ref.MongoAchievementID, ref.Status).Scan(&ref.ID, &ref.CreatedAt, &ref.UpdatedAt)
	if err != nil { return nil, err }
	return ref, nil
}

func (r *achievementPGRepositoryImpl) GetReferenceByID(id string) (*model_postgre.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
		FROM achievement_references WHERE id = $1 AND status != $2
	`
	return scanAchievementRow(r.DB.QueryRow(query, id, model_postgre.StatusDeleted).Scan)
}

func (r *achievementPGRepositoryImpl) UpdateStatusToSubmitted(refID string) (*model_postgre.AchievementReference, error) {
	query := `
		UPDATE achievement_references SET status = $1, submitted_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND status = $3
		RETURNING id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
	`
	ref, err := scanAchievementRow(r.DB.QueryRow(query, model_postgre.StatusSubmitted, refID, model_postgre.StatusDraft).Scan)
	if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("prestasi tidak ditemukan/status bukan draft") }
	return ref, err
}

func (r *achievementPGRepositoryImpl) VerifyAchievement(refID string, verifierID string) (*model_postgre.AchievementReference, error) {
	query := `
		UPDATE achievement_references SET status = $1, verified_by = $2, verified_at = NOW(), updated_at = NOW()
		WHERE id = $3 AND status = $4
		RETURNING id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
	`
	ref, err := scanAchievementRow(r.DB.QueryRow(query, model_postgre.StatusVerified, verifierID, refID, model_postgre.StatusSubmitted).Scan)
	if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("prestasi tidak ditemukan/status bukan submitted") }
	return ref, err
}

func (r *achievementPGRepositoryImpl) RejectAchievement(refID string, verifierID string, rejectionNote string) (*model_postgre.AchievementReference, error) {
	query := `
		UPDATE achievement_references SET status = $1, verified_by = $2, rejection_note = $3, verified_at = NOW(), updated_at = NOW()
		WHERE id = $4 AND status = $5
		RETURNING id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
	`
	ref, err := scanAchievementRow(r.DB.QueryRow(query, model_postgre.StatusRejected, verifierID, rejectionNote, refID, model_postgre.StatusSubmitted).Scan)
	if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("prestasi tidak ditemukan/status bukan submitted") }
	return ref, err
}

func (r *achievementPGRepositoryImpl) SoftDeleteReference(refID string, studentID string) (*model_postgre.AchievementReference, error) {
	query := `
		UPDATE achievement_references SET status = $1, updated_at = NOW()
		WHERE id = $2 AND student_id = $3 AND status = $4
		RETURNING id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
	`
	ref, err := scanAchievementRow(r.DB.QueryRow(query, model_postgre.StatusDeleted, refID, studentID, model_postgre.StatusDraft).Scan)
	if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("gagal hapus: ID salah, bukan pemilik, atau status bukan draft") }
	return ref, err
}

func (r *achievementPGRepositoryImpl) GetMyAchievements(studentID string) ([]model_postgre.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
		FROM achievement_references WHERE student_id = $1 AND status != $2
	`
	rows, err := r.DB.Query(query, studentID, model_postgre.StatusDeleted)
	if err != nil { return nil, err }
	defer rows.Close()
	
	var list []model_postgre.AchievementReference
	for rows.Next() {
		ref, err := scanAchievementRow(rows.Scan)
		if err != nil { return nil, err }
		list = append(list, *ref)
	}
	return list, nil
}

func (r *achievementPGRepositoryImpl) GetAchievementsByStudentIDs(studentIDs []string) ([]model_postgre.AchievementReference, error) {
	if len(studentIDs) == 0 { return []model_postgre.AchievementReference{}, nil }
	
	placeholders := make([]string, len(studentIDs))
	args := make([]interface{}, len(studentIDs))
	for i, id := range studentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args = append(args, model_postgre.StatusDraft, model_postgre.StatusDeleted)
	
	query := fmt.Sprintf(`
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
		FROM achievement_references WHERE student_id IN (%s) AND status != $%d AND status != $%d
	`, strings.Join(placeholders, ","), len(studentIDs)+1, len(studentIDs)+2)

	rows, err := r.DB.Query(query, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var list []model_postgre.AchievementReference
	for rows.Next() {
		ref, err := scanAchievementRow(rows.Scan)
		if err != nil { return nil, err }
		list = append(list, *ref)
	}
	return list, nil
}

func (r *achievementPGRepositoryImpl) GetAllAchievementReferences() ([]model_postgre.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
		FROM achievement_references WHERE status != $1
	`
	rows, err := r.DB.Query(query, model_postgre.StatusDeleted)
	if err != nil { return nil, err }
	defer rows.Close()

	var list []model_postgre.AchievementReference
	for rows.Next() {
		ref, err := scanAchievementRow(rows.Scan)
		if err != nil { return nil, err }
		list = append(list, *ref)
	}
	return list, nil
}

func (r *achievementPGRepositoryImpl) FindStudentIdByUserID(userID string) (string, error) {
	var sid string
	err := r.DB.QueryRow("SELECT id FROM students WHERE user_id = $1", userID).Scan(&sid)
	if errors.Is(err, sql.ErrNoRows) { return "", nil }
	return sid, err
}
func (r *achievementPGRepositoryImpl) FindLecturerIdByUserID(userID string) (string, error) {
	var lid string
	err := r.DB.QueryRow("SELECT id FROM lecturers WHERE user_id = $1", userID).Scan(&lid)
	if errors.Is(err, sql.ErrNoRows) { return "", nil }
	return lid, err
}
func (r *achievementPGRepositoryImpl) GetAdviseeStudentIDs(lecturerID string) ([]string, error) {
	rows, err := r.DB.Query("SELECT id FROM students WHERE advisor_id = $1", lecturerID)
	if err != nil { return nil, err }
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil { return nil, err }
		ids = append(ids, id)
	}
	return ids, nil
}
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
	CreateReference(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error) // FR-003
	FindStudentIdByUserID(userID string) (string, error) 
	UpdateStatusToSubmitted(refID string) (*model_postgre.AchievementReference, error) // FR-004
	SoftDeleteReference(refID string, studentID string) (*model_postgre.AchievementReference, error) // FR-005
	
	// FR-006 Lookup
	FindLecturerIdByUserID(userID string) (string, error)
	GetAdviseeStudentIDs(lecturerID string) ([]string, error)
	GetAchievementsByStudentIDs(studentIDs []string) ([]model_postgre.AchievementReference, error)
	GetMyAchievements(studentID string) ([]model_postgre.AchievementReference, error)
	GetAllAchievementReferences() ([]model_postgre.AchievementReference, error) // FR-010
	VerifyAchievement(refID string, verifierID string) (*model_postgre.AchievementReference, error)
	RejectAchievement(refID string, verifierID string, rejectionNote string) (*model_postgre.AchievementReference, error)
}

type achievementPGRepositoryImpl struct {
	DB *sql.DB 
}

func NewAchievementPGRepository(db *sql.DB) AchievementPGRepository {
	return &achievementPGRepositoryImpl{DB: db}
}

// Implementasi CreateReference (FR-003) - Tetap sama
func (r *achievementPGRepositoryImpl) CreateReference(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error) {
	query := `
		INSERT INTO achievement_references 
		(student_id, mongo_achievement_id, status)
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at
	`
	ref.Status = model_postgre.StatusDraft 
	// ... (QueryRow dan Scan) ...
	err := r.DB.QueryRow(query, ref.StudentID, ref.MongoAchievementID, ref.Status).Scan(
		&ref.ID, 
		&ref.CreatedAt, 
		&ref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

// Implementasi FindStudentIdByUserID - Tetap sama
func (r *achievementPGRepositoryImpl) FindStudentIdByUserID(userID string) (string, error) {
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

// Implementasi UpdateStatusToSubmitted (FR-004) - Tetap sama
func (r *achievementPGRepositoryImpl) UpdateStatusToSubmitted(refID string) (*model_postgre.AchievementReference, error) {
	ref := new(model_postgre.AchievementReference)
	query := `
		UPDATE achievement_references 
		SET status = $1, submitted_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND status = $3
		RETURNING id, student_id, mongo_achievement_id, status, submitted_at, updated_at
	`
	err := r.DB.QueryRow(query, model_postgre.StatusSubmitted, refID, model_postgre.StatusDraft).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status, &ref.SubmittedAt, &ref.UpdatedAt,
		// Perlu scan semua field yang di return, ini adalah contoh minimal:
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("prestasi tidak ditemukan atau status sudah berubah")
		}
		return nil, err
	}
	return ref, nil
}

// Implementasi SoftDeleteReference (FR-005) - Tetap sama
func (r *achievementPGRepositoryImpl) SoftDeleteReference(refID string, studentID string) (*model_postgre.AchievementReference, error) {
	ref := new(model_postgre.AchievementReference)
	query := `
		UPDATE achievement_references 
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND student_id = $3 AND status = $4
		RETURNING id, student_id, mongo_achievement_id, status, created_at, updated_at, submitted_at, verified_at, verified_by, rejection_note
	`
	err := r.DB.QueryRow(query, model_postgre.StatusDeleted, refID, studentID, model_postgre.StatusDraft).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status, &ref.CreatedAt, &ref.UpdatedAt,
		&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("prestasi tidak ditemukan, bukan milik Anda, atau status sudah disubmit/diverifikasi")
		}
		return nil, err
	}
	return ref, nil
}

// Implementasi FindLecturerIdByUserID (FR-006) - Tetap sama
func (r *achievementPGRepositoryImpl) FindLecturerIdByUserID(userID string) (string, error) {
	var lecturerID string
	query := `SELECT id FROM lecturers WHERE user_id = $1`
	err := r.DB.QueryRow(query, userID).Scan(&lecturerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return lecturerID, nil
}

// Implementasi GetAdviseeStudentIDs (FR-006) - Tetap sama
func (r *achievementPGRepositoryImpl) GetAdviseeStudentIDs(lecturerID string) ([]string, error) {
	query := `SELECT id FROM students WHERE advisor_id = $1`
	rows, err := r.DB.Query(query, lecturerID)
	// ... (logic query dan scan) ...
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, id)
	}
	return studentIDs, nil
}

func (r *achievementPGRepositoryImpl) GetMyAchievements(studentID string) ([]model_postgre.AchievementReference, error) {
    // Untuk Mahasiswa, hanya ada satu studentID, jadi kita bisa menggunakan $1 secara langsung.
    // Kita tetap menggunakan logic IN/ANY untuk konsistensi query, meskipun hanya 1 ID.
    
    // Jika tidak ada ID, kembalikan array kosong
    if studentID == "" {
        return []model_postgre.AchievementReference{}, nil
    }

    // Hanya perlu mengecualikan status DELETED
    deletedStatusIndex := 2
    args := []interface{}{studentID, model_postgre.StatusDeleted}

    query := fmt.Sprintf(`
        SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, 
            verified_by, rejection_note, created_at
        FROM achievement_references
        WHERE student_id = $1 AND status != $%d 
    `, deletedStatusIndex) // $1 adalah studentID
    
    rows, err := r.DB.Query(query, args...)
    
    if err != nil {
        return nil, fmt.Errorf("SQL execution error in GetMyAchievements: %w", err) 
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
            return nil, fmt.Errorf("error scanning row in GetMyAchievements: %w", err) 
        }
        references = append(references, ref)
    }
    return references, nil
}

// Implementasi GetAchievementsByStudentIDs (FR-006) - (Menampilkan prestasi yang telah disubmit)
func (r *achievementPGRepositoryImpl) GetAchievementsByStudentIDs(studentIDs []string) ([]model_postgre.AchievementReference, error) {
	
    // Jika tidak ada ID, kembalikan array kosong (mencegah error query)
    if len(studentIDs) == 0 {
        return []model_postgre.AchievementReference{}, nil
    }

    // 1. Buat placeholder dinamis ($1, $2, $3, ...)
    placeholders := make([]string, len(studentIDs))
    args := make([]interface{}, len(studentIDs))
    for i, id := range studentIDs {
        placeholders[i] = fmt.Sprintf("$%d", i+1)
        args[i] = id
    }
    
    // 2. Tambahkan parameter status ke akhir array args
    // Status Draft ($N+1) dan Deleted ($N+2)
    draftStatusIndex := len(args) + 1
    deletedStatusIndex := len(args) + 2
    args = append(args, model_postgre.StatusDraft, model_postgre.StatusDeleted)


	query := fmt.Sprintf(`
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, 
            verified_by, rejection_note, created_at
		FROM achievement_references
		WHERE student_id IN (%s) AND status != $%d AND status != $%d
	`, strings.Join(placeholders, ","), draftStatusIndex, deletedStatusIndex)
    
    // 3. Eksekusi Query
	rows, err := r.DB.Query(query, args...)
	
	if err != nil {
		// Log error SQL yang sebenarnya
		return nil, fmt.Errorf("SQL execution error (IN Clause): %w", err) 
	}
	defer rows.Close()

	// 4. Proses Scan (tetap sama)
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
			return nil, fmt.Errorf("error scanning row in GetAchievementsByStudentIDs: %w", err) 
		}
		references = append(references, ref)
	}
	return references, nil
}

// Implementasi GetAllAchievementReferences (FR-010) - Tetap sama
func (r *achievementPGRepositoryImpl) GetAllAchievementReferences() ([]model_postgre.AchievementReference, error) {
    query := `
        SELECT 
            id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, 
            verified_by, rejection_note, created_at /* ⬅️ KOREKSI: Tambahkan created_at */
        FROM achievement_references
        WHERE status != $1
    `
    rows, err := r.DB.Query(query, model_postgre.StatusDeleted)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var references []model_postgre.AchievementReference
    for rows.Next() {
        var ref model_postgre.AchievementReference
        // TOTAL HARUS ADA 10 VARIABEL SCAN
        err := rows.Scan(
            &ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
            &ref.SubmittedAt, &ref.VerifiedAt, &ref.UpdatedAt,
            &ref.VerifiedBy, &ref.RejectionNote, 
            &ref.CreatedAt, /* ⬅️ KOREKSI: Tambahkan &ref.CreatedAt */
        )
        if err != nil {
            // Ini akan mencetak error SQL yang detail di terminal Anda
            return nil, fmt.Errorf("error scanning row in GetAllAchievementReferences: %w", err)
        }
        references = append(references, ref)
    }
    return references, nil
}

// Implementasi VerifyAchievement (FR-007)
func (r *achievementPGRepositoryImpl) VerifyAchievement(refID string, verifierID string) (*model_postgre.AchievementReference, error) {
    ref := new(model_postgre.AchievementReference)
    
    query := `
        UPDATE achievement_references 
        SET status = $1, verified_by = $2, verified_at = NOW(), updated_at = NOW()
        WHERE id = $3 AND status = $4
        RETURNING id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
    `
    // Precondition: Status harus 'submitted'
    err := r.DB.QueryRow(query, 
        model_postgre.StatusVerified, 
        verifierID, 
        refID, 
        model_postgre.StatusSubmitted,
    ).Scan(
        &ref.ID, 
        &ref.StudentID,
        &ref.MongoAchievementID,
        &ref.Status,
        &ref.SubmittedAt,
        &ref.VerifiedAt,
        &ref.UpdatedAt,
        &ref.VerifiedBy,
        &ref.RejectionNote,
        &ref.CreatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // Gagal jika ID salah ATAU status BUKAN submitted
            return nil, errors.New("prestasi tidak ditemukan atau status bukan SUBMITTED")
        }
        return nil, err
    }
    return ref, nil
}

// Implementasi RejectAchievement (FR-008)
func (r *achievementPGRepositoryImpl) RejectAchievement(refID string, verifierID string, rejectionNote string) (*model_postgre.AchievementReference, error) {
    ref := new(model_postgre.AchievementReference)
    
    query := `
        UPDATE achievement_references 
        SET status = $1, verified_by = $2, rejection_note = $3, verified_at = NOW(), updated_at = NOW()
        WHERE id = $4 AND status = $5
        RETURNING id, student_id, mongo_achievement_id, status, submitted_at, verified_at, updated_at, verified_by, rejection_note, created_at
    `
    // Precondition: Status harus 'submitted'
    err := r.DB.QueryRow(query, 
        model_postgre.StatusRejected, // Status target: 'rejected'
        verifierID,                   // $2: Dosen Wali yang memverifikasi
        rejectionNote,                // $3: Catatan penolakan (string yang diinput)
        refID,                        // $4: ID Prestasi
        model_postgre.StatusSubmitted, // $5: Precondition status
    ).Scan(
        &ref.ID, 
        &ref.StudentID,
        &ref.MongoAchievementID,
        &ref.Status,
        &ref.SubmittedAt,
        &ref.VerifiedAt,
        &ref.UpdatedAt,
        &ref.VerifiedBy,
        &ref.RejectionNote,
        &ref.CreatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // Gagal jika ID tidak ada ATAU status BUKAN submitted
            return nil, errors.New("prestasi tidak ditemukan atau status bukan SUBMITTED")
        }
        return nil, err
    }
    return ref, nil
}
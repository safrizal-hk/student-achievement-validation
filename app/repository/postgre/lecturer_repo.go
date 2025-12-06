package repository

import (
	"database/sql"
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
)

type LecturerRepository interface {
	GetAllLecturers() ([]model_postgre.Lecturer, error)
	GetLecturerByID(id string) (*model_postgre.Lecturer, error)
	GetAdviseesByLecturerID(lecturerID string) ([]model_postgre.Student, error) 
}

type lecturerRepositoryImpl struct {
	DB *sql.DB
}

func NewLecturerRepository(db *sql.DB) LecturerRepository {
	return &lecturerRepositoryImpl{DB: db}
}

func (r *lecturerRepositoryImpl) GetAllLecturers() ([]model_postgre.Lecturer, error) {
	query := `
		SELECT l.id, l.user_id, l.lecturer_id, l.department, l.created_at, u.full_name, u.email
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE u.is_active = TRUE
	`
	rows, err := r.DB.Query(query)
	if err != nil { return nil, err }
	defer rows.Close()

	var lecturers []model_postgre.Lecturer
	for rows.Next() {
		var l model_postgre.Lecturer
		err := rows.Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt, &l.FullName, &l.Email)
		if err != nil { return nil, err }
		lecturers = append(lecturers, l)
	}
	return lecturers, nil
}

func (r *lecturerRepositoryImpl) GetLecturerByID(id string) (*model_postgre.Lecturer, error) {
	l := new(model_postgre.Lecturer)
	query := `
		SELECT l.id, l.user_id, l.lecturer_id, l.department, l.created_at, u.full_name, u.email
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE l.id = $1
	`
	err := r.DB.QueryRow(query, id).Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt, &l.FullName, &l.Email)
	if err != nil { return nil, err }
	return l, nil
}

// Mengambil daftar mahasiswa (Advisees) menggunakan Single Struct Student
func (r *lecturerRepositoryImpl) GetAdviseesByLecturerID(lecturerID string) ([]model_postgre.Student, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
			u.full_name, u.email
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.advisor_id = $1 AND u.is_active = TRUE
	`
	rows, err := r.DB.Query(query, lecturerID)
	if err != nil { return nil, err }
	defer rows.Close()

	var advisees []model_postgre.Student
	for rows.Next() {
		var s model_postgre.Student
		err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
			&s.FullName, &s.Email,
		)
		if err != nil { return nil, err }
		advisees = append(advisees, s)
	}
	return advisees, nil
}
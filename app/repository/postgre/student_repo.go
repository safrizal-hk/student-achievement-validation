package repository

import (
	"database/sql"
	"fmt"

	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
)

type StudentRepository interface {
	GetAllStudents() ([]model_postgre.Student, error)
	GetStudentDetail(studentID string) (*model_postgre.Student, error)
	SetStudentAdvisor(studentID string, advisorID string) error
}

type studentRepositoryImpl struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepositoryImpl{DB: db}
}

func (r *studentRepositoryImpl) GetAllStudents() ([]model_postgre.Student, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
			u.full_name, u.email
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE u.is_active = TRUE
	`
	rows, err := r.DB.Query(query)
	if err != nil { return nil, err }
	defer rows.Close()

	var students []model_postgre.Student
	for rows.Next() {
		var s model_postgre.Student
		err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
			&s.FullName, &s.Email,
		)
		if err != nil { return nil, err }
		students = append(students, s)
	}
	return students, nil
}

func (r *studentRepositoryImpl) GetStudentDetail(id string) (*model_postgre.Student, error) {
	s := new(model_postgre.Student)
	query := `
		SELECT 
			s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
			u.full_name, u.email
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = $1
	`
	err := r.DB.QueryRow(query, id).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
		&s.FullName, &s.Email,
	)
	if err != nil { return nil, err }
	return s, nil
}

func (r *studentRepositoryImpl) SetStudentAdvisor(studentID string, advisorID string) error {
	result, err := r.DB.Exec(`UPDATE students SET advisor_id = $1 WHERE id = $2`, advisorID, studentID)
	if err != nil { return err }
	rows, _ := result.RowsAffected()
	if rows == 0 { return fmt.Errorf("mahasiswa tidak ditemukan") }
	return nil
}
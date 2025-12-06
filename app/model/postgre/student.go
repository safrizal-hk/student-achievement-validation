package model

import "time"

// Student merepresentasikan tabel students
type Student struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	StudentID    string    `json:"student_id"` // NIM
	ProgramStudy string    `json:"program_study"`
	AcademicYear string    `json:"academic_year"`
	AdvisorID    *string   `json:"advisor_id"` // Pointer karena bisa NULL
	CreatedAt    time.Time `json:"created_at"`
}
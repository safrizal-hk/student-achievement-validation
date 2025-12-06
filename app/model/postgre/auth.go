package model

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refreshToken"`
	User         UserProfile `json:"user"`
}

type UserProfile struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"fullName"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// UserCreateRequest merepresentasikan payload untuk membuat user baru (Flow 1, 2, 3)
type UserCreateRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
	RoleName string `json:"role_name" validate:"required"` // Contoh: "Mahasiswa", "Admin", "Dosen Wali"

	// Profile data (Flow 3)
	StudentID string `json:"student_id,omitempty"` // NIM, jika role Mahasiswa
	LecturerID string `json:"lecturer_id,omitempty"` // NIP/NIDN, jika role Dosen Wali
	ProgramStudy string `json:"program_study,omitempty"`
	Department string `json:"department,omitempty"`
}

// UserUpdateRequest merepresentasikan payload untuk mengupdate data user
type UserUpdateRequest struct {
	FullName *string `json:"full_name"` // Gunakan pointer untuk update parsial
	Email    *string `json:"email" validate:"email"`
	IsActive *bool   `json:"is_active"`
}

// AssignRoleRequest (PUT /:id/role)
type AssignRoleRequest struct {
	RoleName string `json:"role_name" validate:"required"`
}

// SetAdvisorRequest (PUT /students/:id/advisor)
type SetAdvisorRequest struct {
	AdvisorID string `json:"advisor_id" validate:"required"` // UUID Lecturer
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}
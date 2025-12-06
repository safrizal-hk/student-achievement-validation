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

type UserCreateRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
	RoleName string `json:"role_name" validate:"required"`

	StudentID string `json:"student_id,omitempty"`
	LecturerID string `json:"lecturer_id,omitempty"`
	ProgramStudy string `json:"program_study,omitempty"`
	Department string `json:"department,omitempty"`
}

type UserUpdateRequest struct {
	FullName *string `json:"full_name"`
	Email    *string `json:"email" validate:"email"`
	IsActive *bool   `json:"is_active"`
}

type AssignRoleRequest struct {
	RoleName string `json:"role_name" validate:"required"`
}

type SetAdvisorRequest struct {
	AdvisorID string `json:"advisor_id" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}
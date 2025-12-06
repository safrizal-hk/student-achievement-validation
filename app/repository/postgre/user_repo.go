package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings" // Diperlukan untuk strings.Join
	// "time"    // Diperlukan untuk NOW() dan time.Time

	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
)

// Kontrak Repository (FR-009)
type AdminManageUsersRepository interface {
	// CRUD User
	CreateUser(req *model_postgre.UserCreateRequest, roleID string, passwordHash string) (*model_postgre.User, error)
	DeleteUser(userID string) error // Soft delete user (set is_active = false)
	UpdateUser(userID string, req *model_postgre.UserUpdateRequest) error
	GetUserByID(userID string) (*model_postgre.User, error)
	ListAllUsers() ([]model_postgre.User, error)
	// Role & Profile Flow
	GetRoleByName(roleName string) (*model_postgre.Role, error) // Helper untuk Auth/Create
	SetUserRole(userID string, roleID string) error
	SetStudentAdvisor(studentID string, advisorID string) error // FR-009 Flow 4
}

type adminManageUsersRepositoryImpl struct {
	DB *sql.DB
}

func NewAdminManageUsersRepository(db *sql.DB) AdminManageUsersRepository {
	return &adminManageUsersRepositoryImpl{DB: db}
}

// Implementasi CreateUser (FR-009 Flow 1, 2, 3)
func (r *adminManageUsersRepositoryImpl) CreateUser(req *model_postgre.UserCreateRequest, roleID string, passwordHash string) (*model_postgre.User, error) {
	// PENTING: GetRoleByName sudah dipanggil di Service
	
	tx, err := r.DB.Begin()
	if err != nil { return nil, err }
	defer func() {
		if r := recover(); r != nil || err != nil { tx.Rollback() }
	}()

	// 1. INSERT ke tabel users
	newUser := model_postgre.User{}
	userQuery := `
		INSERT INTO users (username, email, password_hash, full_name, role_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(userQuery, req.Username, req.Email, passwordHash, req.FullName, roleID).Scan(
		&newUser.ID, &newUser.CreatedAt, &newUser.UpdatedAt,
	)
	if err != nil { 
		tx.Rollback()
		return nil, fmt.Errorf("gagal insert user: %w", err) 
	}
	
	// 2. Set Profile (Mahasiswa/Dosen Wali)
	if req.RoleName == "Mahasiswa" && req.StudentID != "" {
		profileQuery := `
			INSERT INTO students (user_id, student_id, program_study) VALUES ($1, $2, $3)`
		_, err = tx.Exec(profileQuery, newUser.ID, req.StudentID, req.ProgramStudy)
	} else if req.RoleName == "Dosen Wali" && req.LecturerID != "" {
		profileQuery := `
			INSERT INTO lecturers (user_id, lecturer_id, department) VALUES ($1, $2, $3)`
		_, err = tx.Exec(profileQuery, newUser.ID, req.LecturerID, req.Department)
	}
	
	if err != nil { tx.Rollback(); return nil, fmt.Errorf("gagal set profile: %w", err) }

	if err := tx.Commit(); err != nil { return nil, fmt.Errorf("gagal commit: %w", err) }
	
	newUser.FullName = req.FullName
	newUser.RoleID = roleID
	return &newUser, nil
}

// Implementasi DeleteUser (Soft Delete: is_active = FALSE)
func (r *adminManageUsersRepositoryImpl) DeleteUser(userID string) error {
	result, err := r.DB.Exec(`UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

// Implementasi UpdateUser (Hanya update data dasar user)
func (r *adminManageUsersRepositoryImpl) UpdateUser(userID string, req *model_postgre.UserUpdateRequest) error {
	setClauses := []string{}
	args := []interface{}{}
	i := 1

	if req.FullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", i)); args = append(args, *req.FullName); i++
	}
	if req.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", i)); args = append(args, *req.Email); i++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", i)); args = append(args, *req.IsActive); i++
	}
	
	if len(setClauses) == 0 { return errors.New("tidak ada field yang diupdate") }

	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))
	
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), i)
	args = append(args, userID)

	result, err := r.DB.Exec(query, args...)
	if err != nil { return err }
	if rows, _ := result.RowsAffected(); rows == 0 { return errors.New("user tidak ditemukan") }
	return nil
}

// Implementasi GetUserByID (Mengambil user + nama role)
func (r *adminManageUsersRepositoryImpl) GetUserByID(userID string) (*model_postgre.User, error) {
	user := new(model_postgre.User)
	query := `
		SELECT u.id, u.username, u.email, u.full_name, r.name AS role_name, u.is_active, u.created_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`
	err := r.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.RoleName, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows { return nil, errors.New("user tidak ditemukan") }
		return nil, err
	}
	return user, nil
}

// Implementasi ListAllUsers (Mengambil semua user untuk list)
func (r *adminManageUsersRepositoryImpl) ListAllUsers() ([]model_postgre.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, r.name AS role_name, u.is_active, u.created_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.is_active = TRUE
		ORDER BY u.created_at DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil { return nil, err }
	defer rows.Close()

	var users []model_postgre.User
	for rows.Next() {
		var user model_postgre.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FullName, &user.RoleName, &user.IsActive, &user.CreatedAt,
		)
		if err != nil { return nil, err }
		users = append(users, user)
	}
	return users, nil
}

// Implementasi GetRoleByName (Helper) - Sudah benar
func (r *adminManageUsersRepositoryImpl) GetRoleByName(roleName string) (*model_postgre.Role, error) {
	// ... (Logika sudah benar) ...
	role := new(model_postgre.Role)
	query := `SELECT id, name, description FROM roles WHERE name = $1`
	err := r.DB.QueryRow(query, roleName).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		if err == sql.ErrNoRows { return nil, errors.New("role tidak ditemukan") }
		return nil, err
	}
	return role, nil
}

// Implementasi SetUserRole (FR-009 Flow 2)
func (r *adminManageUsersRepositoryImpl) SetUserRole(userID string, roleID string) error {
	result, err := r.DB.Exec(`UPDATE users SET role_id = $1, updated_at = NOW() WHERE id = $2`, roleID, userID)
	if err != nil { return err }
	if rows, _ := result.RowsAffected(); rows == 0 { return errors.New("user atau role tidak ditemukan") }
	return nil
}

// Implementasi SetStudentAdvisor (FR-009 Flow 4)
func (r *adminManageUsersRepositoryImpl) SetStudentAdvisor(studentID string, advisorID string) error {
	result, err := r.DB.Exec(`UPDATE students SET advisor_id = $1 WHERE id = $2`, advisorID, studentID)
	if err != nil { return err }
	if rows, _ := result.RowsAffected(); rows == 0 { return errors.New("mahasiswa atau advisor tidak valid") }
	return nil
}
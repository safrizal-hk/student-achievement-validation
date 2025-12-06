package repository

import (
	"database/sql"
	"errors"
	"fmt"
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
)

type AuthRepository interface {
	FindUserByEmailOrUsername(identifier string) (*model_postgre.User, string, error) 
	GetPermissionsByRoleID(roleID string) ([]string, error)
	FindUserByID(id string) (*model_postgre.User, string, error) // ⚠️ BARU
}

type authRepositoryImpl struct {
	DB *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepositoryImpl{DB: db}
}

func (r *authRepositoryImpl) FindUserByEmailOrUsername(identifier string) (*model_postgre.User, string, error) {
	user := new(model_postgre.User)
	var roleName string

	query := `
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1 OR u.username = $1
	`
	
	err := r.DB.QueryRow(query, identifier).Scan(
		&user.ID, 
		&user.Username, 
		&user.Email, 
		&user.PasswordHash, 
		&user.FullName, 
		&user.RoleID, 
		&user.IsActive, 
		&roleName,
	) 

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("repository error: %w", err)
	}
	return user, roleName, nil
}

func (r *authRepositoryImpl) GetPermissionsByRoleID(roleID string) ([]string, error) {
	query := `
		SELECT p.name 
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`
	rows, err := r.DB.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		permissions = append(permissions, name)
	}
	return permissions, nil
}

// FindUserByID (Profile & Refresh)
func (r *authRepositoryImpl) FindUserByID(id string) (*model_postgre.User, string, error) {
	user := new(model_postgre.User)
	var roleName string

	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`
	err := r.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, 
		&user.FullName, &user.RoleID, &user.IsActive, &roleName,
	) 

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, "", nil }
		return nil, "", fmt.Errorf("repository error: %w", err)
	}
	return user, roleName, nil
}
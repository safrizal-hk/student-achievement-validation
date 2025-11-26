package model

import "time"

type User struct {
	ID           string    
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	RoleID       string    
	IsActive     bool      
	CreatedAt    time.Time 
	UpdatedAt    time.Time 
	
	RoleName     string 
}
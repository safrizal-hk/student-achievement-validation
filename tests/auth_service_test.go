package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	"github.com/safrizal-hk/uas-gofiber/app/service"
	"github.com/safrizal-hk/uas-gofiber/utils"
)

// MOCK REPOSITORIES (Dynamic Function Fields)
type MockAuthRepo struct {
	FindUserByEmailOrUsernameFunc func(emailOrUsername string) (*model_postgre.User, string, error)
	GetPermissionsByRoleIDFunc    func(roleID string) ([]string, error)
	FindUserByIDFunc              func(userID string) (*model_postgre.User, string, error)
	
	CreateUserFunc func(req *model_postgre.UserCreateRequest, roleID, hash string) (*model_postgre.User, error)
}

func (m *MockAuthRepo) FindUserByEmailOrUsername(emailOrUsername string) (*model_postgre.User, string, error) {
	if m.FindUserByEmailOrUsernameFunc == nil { return nil, "", nil }
	return m.FindUserByEmailOrUsernameFunc(emailOrUsername)
}

func (m *MockAuthRepo) GetPermissionsByRoleID(roleID string) ([]string, error) {
	if m.GetPermissionsByRoleIDFunc == nil { return nil, nil }
	return m.GetPermissionsByRoleIDFunc(roleID)
}

func (m *MockAuthRepo) FindUserByID(userID string) (*model_postgre.User, string, error) {
	if m.FindUserByIDFunc == nil { return nil, "", nil }
	return m.FindUserByIDFunc(userID)
}

func (m *MockAuthRepo) CreateUser(req *model_postgre.UserCreateRequest, roleID, hash string) (*model_postgre.User, error) { return nil, nil }

// SETUP HELPER
func setupAuthServiceTestApp(mockRepo *MockAuthRepo) *fiber.App {
	app := fiber.New()
	svc := service.NewAuthService(mockRepo)

	app.Use(func(c *fiber.Ctx) error {
		role := c.Get("X-Test-Role")
		id := c.Get("X-Test-ID")
		
		if role != "" {
			c.Locals("userProfile", model_postgre.UserProfile{
				ID:   id,
				Role: role,
			})
		}
		return c.Next()
	})

	app.Post("/auth/login", svc.Login)
	app.Post("/auth/refresh", svc.RefreshToken)
	app.Post("/auth/logout", svc.Logout)
	app.Get("/auth/profile", svc.Profile)

	return app
}

func hashPassword(password string) string {
	hash, _ := utils.HashPassword(password)
	return hash
}

// UNIT TESTS
func TestLogin_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "rahasia-test-123")
	defer os.Unsetenv("JWT_SECRET")

	password := "rahasia123"
	hashed := hashPassword(password)

	mockRepo := &MockAuthRepo{
		FindUserByEmailOrUsernameFunc: func(email string) (*model_postgre.User, string, error) {
			if email == "budi@mail.com" {
				return &model_postgre.User{
					ID:           "user-1",
					Username:     "budi",
					FullName:     "Budi Santoso",
					PasswordHash: hashed,
					RoleID:       "role-1",
					IsActive:     true,
				}, "Mahasiswa", nil
			}
			return nil, "", nil
		},
		GetPermissionsByRoleIDFunc: func(roleID string) ([]string, error) {
			return []string{"read:achievement"}, nil
		},
	}
	app := setupAuthServiceTestApp(mockRepo)

	payload := model_postgre.LoginRequest{
		Username: "budi@mail.com",
		Password: password,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var respData map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&respData)
	assert.NoError(t, err)

	assert.Equal(t, "success", respData["status"])

	if data, ok := respData["data"].(map[string]interface{}); ok {
		assert.NotNil(t, data["token"])
	} else {
		t.Error("Response data is missing or invalid format")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	realHash := hashPassword("pass-asli")
	mockRepo := &MockAuthRepo{
		FindUserByEmailOrUsernameFunc: func(email string) (*model_postgre.User, string, error) {
			return &model_postgre.User{
				ID:           "user-1",
				PasswordHash: realHash,
				IsActive:     true,
			}, "Admin", nil
		},
	}
	app := setupAuthServiceTestApp(mockRepo)

	payload := model_postgre.LoginRequest{Username: "user", Password: "pass-salah"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := &MockAuthRepo{
		FindUserByEmailOrUsernameFunc: func(email string) (*model_postgre.User, string, error) {
			return nil, "", nil
		},
	}
	app := setupAuthServiceTestApp(mockRepo)

	payload := model_postgre.LoginRequest{Username: "unknown", Password: "any"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_UserInactive(t *testing.T) {
	mockRepo := &MockAuthRepo{
		FindUserByEmailOrUsernameFunc: func(email string) (*model_postgre.User, string, error) {
			return &model_postgre.User{IsActive: false}, "Admin", nil
		},
	}
	app := setupAuthServiceTestApp(mockRepo)

	payload := model_postgre.LoginRequest{Username: "inactive", Password: "any"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogout_Success(t *testing.T) {
	app := setupAuthServiceTestApp(&MockAuthRepo{}) // Repo tidak dipanggil
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProfile_Success(t *testing.T) {
	userID := "user-123"
	mockRepo := &MockAuthRepo{
		FindUserByIDFunc: func(id string) (*model_postgre.User, string, error) {
			if id == userID {
				return &model_postgre.User{
					ID:       userID,
					Username: "budi",
					RoleID:   "role-1",
				}, "Mahasiswa", nil
			}
			return nil, "", errors.New("not found")
		},
		GetPermissionsByRoleIDFunc: func(roleID string) ([]string, error) {
			return []string{"read"}, nil
		},
	}
	app := setupAuthServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/auth/profile", nil)
	req.Header.Set("X-Test-Role", "Mahasiswa")
	req.Header.Set("X-Test-ID", userID)
	
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProfile_NotFound(t *testing.T) {
	mockRepo := &MockAuthRepo{
		FindUserByIDFunc: func(id string) (*model_postgre.User, string, error) {
			return nil, "", nil
		},
	}
	app := setupAuthServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/auth/profile", nil)
	req.Header.Set("X-Test-ID", "unknown-id")
	req.Header.Set("X-Test-Role", "Guest")
	
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRefreshToken_BadRequest(t *testing.T) {
	app := setupAuthServiceTestApp(&MockAuthRepo{})
	
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
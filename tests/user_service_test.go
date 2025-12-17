package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	"github.com/safrizal-hk/uas-gofiber/app/service"
)

// MOCK REPOSITORY (Dynamic Function Fields)

type MockAdminUserRepo struct {
	GetRoleByNameFunc     func(name string) (*model_postgre.Role, error)
	CreateUserFunc        func(req *model_postgre.UserCreateRequest, roleID, hash string) (*model_postgre.User, error)
	ListAllUsersFunc      func() ([]model_postgre.User, error)
	GetUserByIDFunc       func(id string) (*model_postgre.User, error)
	UpdateUserFunc        func(id string, req *model_postgre.UserUpdateRequest) error
	DeleteUserFunc        func(id string) error
	SetUserRoleFunc       func(userID, roleID string) error
	SetStudentAdvisorFunc func(studentID, advisorID string) error

	FindUserByEmailOrUsernameFunc func(emailOrUsername string) (*model_postgre.User, string, error)
	FindUserByIDFunc              func(userID string) (*model_postgre.User, string, error)
	GetPermissionsByRoleIDFunc    func(roleID string) ([]string, error)
}

func (m *MockAdminUserRepo) GetRoleByName(name string) (*model_postgre.Role, error) {
	if m.GetRoleByNameFunc == nil { return nil, nil }
	return m.GetRoleByNameFunc(name)
}
func (m *MockAdminUserRepo) CreateUser(req *model_postgre.UserCreateRequest, roleID, hash string) (*model_postgre.User, error) {
	if m.CreateUserFunc == nil { return nil, nil }
	return m.CreateUserFunc(req, roleID, hash)
}
func (m *MockAdminUserRepo) ListAllUsers() ([]model_postgre.User, error) {
	if m.ListAllUsersFunc == nil { return nil, nil }
	return m.ListAllUsersFunc()
}
func (m *MockAdminUserRepo) GetUserByID(id string) (*model_postgre.User, error) {
	if m.GetUserByIDFunc == nil { return nil, nil }
	return m.GetUserByIDFunc(id)
}
func (m *MockAdminUserRepo) UpdateUser(id string, req *model_postgre.UserUpdateRequest) error {
	if m.UpdateUserFunc == nil { return nil }
	return m.UpdateUserFunc(id, req)
}
func (m *MockAdminUserRepo) DeleteUser(id string) error {
	if m.DeleteUserFunc == nil { return nil }
	return m.DeleteUserFunc(id)
}
func (m *MockAdminUserRepo) SetUserRole(userID, roleID string) error {
	if m.SetUserRoleFunc == nil { return nil }
	return m.SetUserRoleFunc(userID, roleID)
}
func (m *MockAdminUserRepo) SetStudentAdvisor(studentID, advisorID string) error {
	if m.SetStudentAdvisorFunc == nil { return nil }
	return m.SetStudentAdvisorFunc(studentID, advisorID)
}

func (m *MockAdminUserRepo) FindUserByEmailOrUsername(emailOrUsername string) (*model_postgre.User, string, error) { return nil, "", nil }
func (m *MockAdminUserRepo) FindUserByID(userID string) (*model_postgre.User, string, error) { return nil, "", nil }
func (m *MockAdminUserRepo) GetPermissionsByRoleID(roleID string) ([]string, error) { return nil, nil }

// SETUP HELPER
func setupAdminUserServiceTestApp(mockRepo *MockAdminUserRepo) *fiber.App {
	app := fiber.New()
	svc := service.NewAdminManageUsersService(mockRepo)

	app.Post("/users", svc.AdminCreateUser)
	app.Get("/users", svc.AdminListUsers)
	app.Get("/users/:id", svc.AdminGetUserByID)
	app.Put("/users/:id", svc.AdminUpdateUser)
	app.Delete("/users/:id", svc.AdminDeleteUser)
	app.Put("/users/:id/role", svc.AdminSetUserRole)
	app.Put("/students/:id/advisor", svc.AdminSetAdvisor)

	return app
}

// UNIT TESTS

func TestAdminCreateUser_Success(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		GetRoleByNameFunc: func(name string) (*model_postgre.Role, error) {
			return &model_postgre.Role{ID: "role-1", Name: name}, nil
		},
		CreateUserFunc: func(req *model_postgre.UserCreateRequest, roleID, hash string) (*model_postgre.User, error) {
			return &model_postgre.User{ID: "user-new", Username: req.Username}, nil
		},
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	payload := model_postgre.UserCreateRequest{
		Username: "newuser", Password: "password123", RoleName: "Mahasiswa",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestAdminCreateUser_Conflict(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		GetRoleByNameFunc: func(name string) (*model_postgre.Role, error) {
			return &model_postgre.Role{ID: "role-1", Name: name}, nil
		},
		CreateUserFunc: func(req *model_postgre.UserCreateRequest, roleID, hash string) (*model_postgre.User, error) {
			return nil, errors.New("duplicate key value violates unique constraint")
		},
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	payload := model_postgre.UserCreateRequest{Username: "exist", Password: "123", RoleName: "Admin"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestAdminListUsers_Success(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		ListAllUsersFunc: func() ([]model_postgre.User, error) {
			return []model_postgre.User{{ID: "1", Username: "User A"}}, nil
		},
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminGetUserByID_Success(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		GetUserByIDFunc: func(id string) (*model_postgre.User, error) {
			return &model_postgre.User{ID: id, Username: "Found"}, nil
		},
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/users/123", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminGetUserByID_NotFound(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		GetUserByIDFunc: func(id string) (*model_postgre.User, error) {
			return nil, errors.New("record not found")
		},
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/users/unknown", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAdminUpdateUser_Success(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		UpdateUserFunc: func(id string, req *model_postgre.UserUpdateRequest) error {
			return nil
		},
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	newFullName := "Updated Name"
	
	payload := model_postgre.UserUpdateRequest{
		FullName: &newFullName,
	}
	
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/users/123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminDeleteUser_Success(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		DeleteUserFunc: func(id string) error { return nil },
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	req := httptest.NewRequest("DELETE", "/users/123", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminSetUserRole_Success(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		GetRoleByNameFunc: func(name string) (*model_postgre.Role, error) {
			return &model_postgre.Role{ID: "role-new", Name: name}, nil
		},
		SetUserRoleFunc: func(userID, roleID string) error { return nil },
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	payload := model_postgre.AssignRoleRequest{RoleName: "Admin"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/users/123/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminSetUserRole_InvalidRole(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		GetRoleByNameFunc: func(name string) (*model_postgre.Role, error) {
			return nil, errors.New("role not found")
		},
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	payload := model_postgre.AssignRoleRequest{RoleName: "SuperGod"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/users/123/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAdminSetAdvisor_Success(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		SetStudentAdvisorFunc: func(studentID, advisorID string) error { return nil },
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	payload := model_postgre.SetAdvisorRequest{AdvisorID: "lec-1"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/students/stu-1/advisor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminSetAdvisor_BadRequest(t *testing.T) {
	mockRepo := &MockAdminUserRepo{
		SetStudentAdvisorFunc: func(studentID, advisorID string) error {
			return fmt.Errorf("student not found")
		},
	}
	app := setupAdminUserServiceTestApp(mockRepo)

	payload := model_postgre.SetAdvisorRequest{AdvisorID: "lec-1"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/students/stu-99/advisor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
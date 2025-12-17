package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	"github.com/safrizal-hk/uas-gofiber/app/service"
)

// MOCK REPOSITORIES

type MockStudentRepository struct {
	GetAllStudentsFunc    func() ([]model_postgre.Student, error)
	GetStudentDetailFunc  func(id string) (*model_postgre.Student, error)
	SetStudentAdvisorFunc func(studentID, advisorID string) error
}

func (m *MockStudentRepository) GetAllStudents() ([]model_postgre.Student, error) {
	return m.GetAllStudentsFunc()
}

func (m *MockStudentRepository) GetStudentDetail(id string) (*model_postgre.Student, error) {
	return m.GetStudentDetailFunc(id)
}

func (m *MockStudentRepository) SetStudentAdvisor(studentID, advisorID string) error {
	return m.SetStudentAdvisorFunc(studentID, advisorID)
}

type MockAchievementPGRepository struct {
	GetAchievementsByStudentIDsFunc func(ids []string) ([]model_postgre.AchievementReference, error)
	
	CreateReferenceFunc func(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error)
}

func (m *MockAchievementPGRepository) GetAchievementsByStudentIDs(ids []string) ([]model_postgre.AchievementReference, error) {
	return m.GetAchievementsByStudentIDsFunc(ids)
}

func (m *MockAchievementPGRepository) CreateReference(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error) { return nil, nil }
func (m *MockAchievementPGRepository) GetAllAchievementReferences() ([]model_postgre.AchievementReference, error) { return nil, nil }
func (m *MockAchievementPGRepository) GetReferenceByID(id string) (*model_postgre.AchievementReference, error) { return nil, nil }
func (m *MockAchievementPGRepository) SoftDeleteReference(id, studentID string) (*model_postgre.AchievementReference, error) { return nil, nil }
func (m *MockAchievementPGRepository) UpdateStatusToSubmitted(id string) (*model_postgre.AchievementReference, error) { return nil, nil }
func (m *MockAchievementPGRepository) VerifyAchievement(id, lecturerID string) (*model_postgre.AchievementReference, error) { return nil, nil }
func (m *MockAchievementPGRepository) RejectAchievement(id, lecturerID, note string) (*model_postgre.AchievementReference, error) { return nil, nil }
func (m *MockAchievementPGRepository) GetMyAchievements(studentID string) ([]model_postgre.AchievementReference, error) { return nil, nil }
func (m *MockAchievementPGRepository) FindStudentIdByUserID(userID string) (string, error) { return "", nil }
func (m *MockAchievementPGRepository) FindLecturerIdByUserID(userID string) (string, error) { return "", nil }
func (m *MockAchievementPGRepository) GetAdviseeStudentIDs(lecturerID string) ([]string, error) { return nil, nil }
func (m *MockAchievementPGRepository) GetAchievementsByStudentIDsAndStatus(studentIDs []string, status string) ([]model_postgre.AchievementReference, error) { return nil, nil }


// SETUP HELPER
func setupStudentServiceTestApp(mockStudentRepo *MockStudentRepository, mockAchRepo *MockAchievementPGRepository) *fiber.App {
	app := fiber.New()

	svc := service.NewStudentService(mockStudentRepo, mockAchRepo)

	app.Get("/students", svc.ListStudents)
	app.Get("/students/:id", svc.GetStudentDetail)
	app.Put("/students/:id/advisor", svc.SetAdvisor)
	app.Get("/students/:id/achievements", svc.GetStudentAchievements)

	return app
}

// UNIT TESTS
func TestListStudents_Success(t *testing.T) {
	mockStudentRepo := &MockStudentRepository{
		GetAllStudentsFunc: func() ([]model_postgre.Student, error) {
			return []model_postgre.Student{
				{ID: "1", FullName: "Budi"},
				{ID: "2", FullName: "Ani"},
			}, nil
		},
	}
	mockAchRepo := &MockAchievementPGRepository{}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListStudents_Error(t *testing.T) {
	mockStudentRepo := &MockStudentRepository{
		GetAllStudentsFunc: func() ([]model_postgre.Student, error) {
			return nil, errors.New("db connection error")
		},
	}
	mockAchRepo := &MockAchievementPGRepository{}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetStudentDetail_Success(t *testing.T) {
	targetID := "stu-123"
	mockStudentRepo := &MockStudentRepository{
		GetStudentDetailFunc: func(id string) (*model_postgre.Student, error) {
			if id == targetID {
				return &model_postgre.Student{ID: targetID, FullName: "Siti"}, nil
			}
			return nil, errors.New("not found")
		},
	}
	mockAchRepo := &MockAchievementPGRepository{}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	req := httptest.NewRequest(http.MethodGet, "/students/"+targetID, nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetStudentDetail_NotFound(t *testing.T) {
	mockStudentRepo := &MockStudentRepository{
		GetStudentDetailFunc: func(id string) (*model_postgre.Student, error) {
			return nil, errors.New("record not found")
		},
	}
	mockAchRepo := &MockAchievementPGRepository{}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	req := httptest.NewRequest(http.MethodGet, "/students/invalid-id", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSetAdvisor_Success(t *testing.T) {
	mockStudentRepo := &MockStudentRepository{
		SetStudentAdvisorFunc: func(studentID, advisorID string) error {
			return nil
		},
	}
	mockAchRepo := &MockAchievementPGRepository{}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	payload := model_postgre.SetAdvisorRequest{AdvisorID: "lec-99"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/students/stu-1/advisor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetAdvisor_BadRequest(t *testing.T) {
	mockStudentRepo := &MockStudentRepository{}
	mockAchRepo := &MockAchievementPGRepository{}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	req := httptest.NewRequest(http.MethodPut, "/students/stu-1/advisor", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSetAdvisor_DBError(t *testing.T) {
	mockStudentRepo := &MockStudentRepository{
		SetStudentAdvisorFunc: func(studentID, advisorID string) error {
			return errors.New("db write failed")
		},
	}
	mockAchRepo := &MockAchievementPGRepository{}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	payload := model_postgre.SetAdvisorRequest{AdvisorID: "lec-99"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/students/stu-1/advisor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetStudentAchievements_Success(t *testing.T) {
	mockStudentRepo := &MockStudentRepository{}
	mockAchRepo := &MockAchievementPGRepository{
		GetAchievementsByStudentIDsFunc: func(ids []string) ([]model_postgre.AchievementReference, error) {
			return []model_postgre.AchievementReference{
				{
					ID:                 "ach-1",
					StudentID:          "stu-1",
					MongoAchievementID: "mongo-123",
					Status:             "verified",
				},
			}, nil
		},
	}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	req := httptest.NewRequest(http.MethodGet, "/students/stu-1/achievements", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetStudentAchievements_Error(t *testing.T) {
	mockStudentRepo := &MockStudentRepository{}
	mockAchRepo := &MockAchievementPGRepository{
		GetAchievementsByStudentIDsFunc: func(ids []string) ([]model_postgre.AchievementReference, error) {
			return nil, errors.New("db error")
		},
	}
	app := setupStudentServiceTestApp(mockStudentRepo, mockAchRepo)

	req := httptest.NewRequest(http.MethodGet, "/students/stu-1/achievements", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
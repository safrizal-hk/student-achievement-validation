package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	"github.com/safrizal-hk/uas-gofiber/app/service"
)

// MOCK REPOSITORY (Dynamic Function Fields)
type MockLecturerRepo struct {
	GetAllLecturersFunc         func() ([]model_postgre.Lecturer, error)
	GetLecturerByIDFunc         func(id string) (*model_postgre.Lecturer, error)
	GetAdviseesByLecturerIDFunc func(id string) ([]model_postgre.Student, error)
}

func (m *MockLecturerRepo) GetAllLecturers() ([]model_postgre.Lecturer, error) {
	if m.GetAllLecturersFunc == nil {
		return nil, nil
	}
	return m.GetAllLecturersFunc()
}

func (m *MockLecturerRepo) GetLecturerByID(id string) (*model_postgre.Lecturer, error) {
	if m.GetLecturerByIDFunc == nil {
		return nil, nil
	}
	return m.GetLecturerByIDFunc(id)
}

func (m *MockLecturerRepo) GetAdviseesByLecturerID(id string) ([]model_postgre.Student, error) {
	if m.GetAdviseesByLecturerIDFunc == nil {
		return nil, nil
	}
	return m.GetAdviseesByLecturerIDFunc(id)
}

// SETUP HELPER
func setupLecturerServiceTestApp(mockRepo *MockLecturerRepo) *fiber.App {
	app := fiber.New()
	svc := service.NewLecturerService(mockRepo)

	app.Get("/lecturers", svc.ListLecturers)
	app.Get("/lecturers/:id/advisees", svc.GetLecturerAdvisees)

	return app
}

// UNIT TESTS
func TestListLecturers_Success(t *testing.T) {
	mockRepo := &MockLecturerRepo{
		GetAllLecturersFunc: func() ([]model_postgre.Lecturer, error) {
			return []model_postgre.Lecturer{
				{ID: "lec-1", FullName: "Dr. Budi"},
				{ID: "lec-2", FullName: "Prof. Siti"},
			}, nil
		},
	}
	app := setupLecturerServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/lecturers", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListLecturers_DBError(t *testing.T) {
	mockRepo := &MockLecturerRepo{
		GetAllLecturersFunc: func() ([]model_postgre.Lecturer, error) {
			return nil, errors.New("db connection failed")
		},
	}
	app := setupLecturerServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/lecturers", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetLecturerAdvisees_Success(t *testing.T) {
	targetLecturerID := "lec-123"
	mockRepo := &MockLecturerRepo{
		GetLecturerByIDFunc: func(id string) (*model_postgre.Lecturer, error) {
			if id == targetLecturerID {
				return &model_postgre.Lecturer{ID: id, FullName: "Dr. Budi"}, nil
			}
			return nil, errors.New("not found")
		},
		GetAdviseesByLecturerIDFunc: func(id string) ([]model_postgre.Student, error) {
			return []model_postgre.Student{
				{ID: "stu-1", FullName: "Mahasiswa A"},
				{ID: "stu-2", FullName: "Mahasiswa B"},
			}, nil
		},
	}
	app := setupLecturerServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/lecturers/"+targetLecturerID+"/advisees", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLecturerAdvisees_LecturerNotFound(t *testing.T) {
	mockRepo := &MockLecturerRepo{
		GetLecturerByIDFunc: func(id string) (*model_postgre.Lecturer, error) {
			return nil, errors.New("record not found")
		},
	}
	app := setupLecturerServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/lecturers/unknown-id/advisees", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetLecturerAdvisees_DBErrorOnAdvisees(t *testing.T) {
	targetLecturerID := "lec-123"
	mockRepo := &MockLecturerRepo{
		GetLecturerByIDFunc: func(id string) (*model_postgre.Lecturer, error) {
			return &model_postgre.Lecturer{ID: id}, nil
		},
		GetAdviseesByLecturerIDFunc: func(id string) ([]model_postgre.Student, error) {
			return nil, errors.New("failed to fetch students")
		},
	}
	app := setupLecturerServiceTestApp(mockRepo)

	req := httptest.NewRequest("GET", "/lecturers/"+targetLecturerID+"/advisees", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
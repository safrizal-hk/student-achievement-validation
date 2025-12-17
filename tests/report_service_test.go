package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	"github.com/safrizal-hk/uas-gofiber/app/service"
)

// MOCK REPOSITORIES (Dynamic Function Fields)

type MockReportPGRepo struct {
	FindStudentIdByUserIDFunc           func(userID string) (string, error)
	FindStudentProfileFunc              func(studentID string) (*model_postgre.Student, error)
	GetStudentAchievementReferencesFunc func(studentID string) ([]model_postgre.AchievementReference, error)
	FindLecturerIdByUserIDFunc          func(userID string) (string, error)
	GetAdviseeStudentIDsFunc            func(lecturerID string) ([]string, error)
	
	GetTotalStatisticsFunc    func() (map[string]int, error)
	GetStudentStatisticsFunc  func(studentID string) (map[string]int, error)
	GetLecturerStatisticsFunc func(userID string) (map[string]int, error)
}

func (m *MockReportPGRepo) FindStudentIdByUserID(userID string) (string, error) {
	if m.FindStudentIdByUserIDFunc == nil { return "", nil }
	return m.FindStudentIdByUserIDFunc(userID)
}
func (m *MockReportPGRepo) FindStudentProfile(studentID string) (*model_postgre.Student, error) {
	if m.FindStudentProfileFunc == nil { return nil, nil }
	return m.FindStudentProfileFunc(studentID)
}
func (m *MockReportPGRepo) GetStudentAchievementReferences(studentID string) ([]model_postgre.AchievementReference, error) {
	if m.GetStudentAchievementReferencesFunc == nil { return nil, nil }
	return m.GetStudentAchievementReferencesFunc(studentID)
}
func (m *MockReportPGRepo) FindLecturerIdByUserID(userID string) (string, error) {
	if m.FindLecturerIdByUserIDFunc == nil { return "", nil }
	return m.FindLecturerIdByUserIDFunc(userID)
}
func (m *MockReportPGRepo) GetAdviseeStudentIDs(lecturerID string) ([]string, error) {
	if m.GetAdviseeStudentIDsFunc == nil { return nil, nil }
	return m.GetAdviseeStudentIDsFunc(lecturerID)
}

func (m *MockReportPGRepo) GetTotalStatistics() (map[string]int, error) { return nil, nil }
func (m *MockReportPGRepo) GetStudentStatistics(studentID string) (map[string]int, error) { return nil, nil }
func (m *MockReportPGRepo) GetLecturerStatistics(userID string) (map[string]int, error) { return nil, nil }


type MockReportMongoRepo struct {
	GetAchievementStatisticsFunc     func(ctx context.Context, studentIDs []string) ([]bson.M, error)
	GetStudentAchievementDetailsFunc func(ctx context.Context, studentIDHex string) ([]bson.M, error)
}

func (m *MockReportMongoRepo) GetAchievementStatistics(ctx context.Context, studentIDs []string) ([]bson.M, error) {
	if m.GetAchievementStatisticsFunc == nil { return nil, nil }
	return m.GetAchievementStatisticsFunc(ctx, studentIDs)
}
func (m *MockReportMongoRepo) GetStudentAchievementDetails(ctx context.Context, studentIDHex string) ([]bson.M, error) {
	if m.GetStudentAchievementDetailsFunc == nil { return nil, nil }
	return m.GetStudentAchievementDetailsFunc(ctx, studentIDHex)
}

// SETUP HELPER
func setupReportServiceTestApp(mockMongo *MockReportMongoRepo, mockPg *MockReportPGRepo) *fiber.App {
	app := fiber.New()
	svc := service.NewReportService(mockMongo, mockPg)

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

	app.Get("/reports/statistics", svc.GetStatistics)
	app.Get("/reports/student/:id", svc.GetStudentReport)

	return app
}

// UNIT TESTS

func TestGetStatistics_Admin_Success(t *testing.T) {
	mockPg := &MockReportPGRepo{}
	mockMongo := &MockReportMongoRepo{
		GetAchievementStatisticsFunc: func(ctx context.Context, studentIDs []string) ([]bson.M, error) {
			return []bson.M{{"type": "Lomba", "count": 10}}, nil
		},
	}
	app := setupReportServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	req.Header.Set("X-Test-Role", "Admin")
	req.Header.Set("X-Test-ID", "admin-id")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetStatistics_Mahasiswa_Success(t *testing.T) {
	userID := "user-mhs"
	studentID := "stu-123"

	mockPg := &MockReportPGRepo{
		FindStudentIdByUserIDFunc: func(uid string) (string, error) {
			if uid == userID { return studentID, nil }
			return "", errors.New("not found")
		},
	}
	mockMongo := &MockReportMongoRepo{
		GetAchievementStatisticsFunc: func(ctx context.Context, studentIDs []string) ([]bson.M, error) {
			if len(studentIDs) > 0 && studentIDs[0] == studentID {
				return []bson.M{{"type": "Seminar", "count": 2}}, nil
			}
			return nil, nil
		},
	}
	app := setupReportServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	req.Header.Set("X-Test-Role", "Mahasiswa")
	req.Header.Set("X-Test-ID", userID)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetStatistics_DosenWali_Success(t *testing.T) {
	userID := "user-dosen"
	lecturerID := "lec-1"
	advisees := []string{"stu-A", "stu-B"}

	mockPg := &MockReportPGRepo{
		FindLecturerIdByUserIDFunc: func(uid string) (string, error) { return lecturerID, nil },
		GetAdviseeStudentIDsFunc:   func(lid string) ([]string, error) { return advisees, nil },
	}
	mockMongo := &MockReportMongoRepo{
		GetAchievementStatisticsFunc: func(ctx context.Context, studentIDs []string) ([]bson.M, error) {
			return []bson.M{{"type": "Juara", "count": 5}}, nil
		},
	}
	app := setupReportServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	req.Header.Set("X-Test-Role", "Dosen Wali")
	req.Header.Set("X-Test-ID", userID)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetStatistics_DosenWali_NoAdvisees(t *testing.T) {
	mockPg := &MockReportPGRepo{
		FindLecturerIdByUserIDFunc: func(uid string) (string, error) { return "lec-empty", nil },
		GetAdviseeStudentIDsFunc:   func(lid string) ([]string, error) { return []string{}, nil },
	}
	mockMongo := &MockReportMongoRepo{}
	app := setupReportServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	req.Header.Set("X-Test-Role", "Dosen Wali")
	req.Header.Set("X-Test-ID", "dosen-id")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetStatistics_ForbiddenRole(t *testing.T) {
	mockPg := &MockReportPGRepo{}
	mockMongo := &MockReportMongoRepo{}
	app := setupReportServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	req.Header.Set("X-Test-Role", "Satpam")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetStudentReport_Admin_Success(t *testing.T) {
	targetStudentID := "stu-123"
	mockPg := &MockReportPGRepo{
		GetStudentAchievementReferencesFunc: func(sid string) ([]model_postgre.AchievementReference, error) {
			if sid == targetStudentID {
				return []model_postgre.AchievementReference{{ID: "ref-1", Status: "verified"}}, nil
			}
			return nil, errors.New("not found")
		},
	}
	mockMongo := &MockReportMongoRepo{}
	app := setupReportServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/reports/student/"+targetStudentID, nil)
	req.Header.Set("X-Test-Role", "Admin")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetStudentReport_Mahasiswa_OwnData(t *testing.T) {
	userID := "user-mhs"
	myStudentID := "stu-123"

	mockPg := &MockReportPGRepo{
		FindStudentIdByUserIDFunc: func(uid string) (string, error) { return myStudentID, nil },
		GetStudentAchievementReferencesFunc: func(sid string) ([]model_postgre.AchievementReference, error) {
			return []model_postgre.AchievementReference{{ID: "ref-1"}}, nil
		},
	}
	mockMongo := &MockReportMongoRepo{}
	app := setupReportServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/reports/student/"+myStudentID, nil)
	req.Header.Set("X-Test-Role", "Mahasiswa")
	req.Header.Set("X-Test-ID", userID)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetStudentReport_Mahasiswa_OtherData_Forbidden(t *testing.T) {
	userID := "user-mhs"
	myStudentID := "stu-123"
	otherStudentID := "stu-999"

	mockPg := &MockReportPGRepo{
		FindStudentIdByUserIDFunc: func(uid string) (string, error) {
			return myStudentID, nil 
		},
	}
	mockMongo := &MockReportMongoRepo{}
	app := setupReportServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/reports/student/"+otherStudentID, nil)
	req.Header.Set("X-Test-Role", "Mahasiswa")
	req.Header.Set("X-Test-ID", userID)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
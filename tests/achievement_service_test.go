package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	model_mongo "github.com/safrizal-hk/uas-gofiber/app/model/mongo"
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	"github.com/safrizal-hk/uas-gofiber/app/service"
)

// MOCK REPOSITORIES (Dynamic Function Fields)
type MockAchievementMongoRepo struct {
	GetDetailsByIDsFunc func(ctx context.Context, ids []primitive.ObjectID) ([]model_mongo.AchievementMongo, error)
	GetDetailByIDFunc   func(ctx context.Context, id primitive.ObjectID) (*model_mongo.AchievementMongo, error)
	CreateFunc          func(ctx context.Context, achievement *model_mongo.AchievementMongo) (*model_mongo.AchievementMongo, error)
	UpdateFunc          func(ctx context.Context, id primitive.ObjectID, data *model_mongo.AchievementInput) error
	SoftDeleteFunc      func(ctx context.Context, id primitive.ObjectID) error
	DeleteByIDFunc      func(ctx context.Context, id primitive.ObjectID) error
	AddAttachmentFunc   func(ctx context.Context, id primitive.ObjectID, attachment model_mongo.Attachment) error
	
	GetAchievementStatisticsFunc func(ctx context.Context, studentIDs []string) ([]interface{}, error) 
	GetStudentAchievementDetailsFunc func(ctx context.Context, studentIDHex string) ([]interface{}, error)
}

func (m *MockAchievementMongoRepo) GetDetailsByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model_mongo.AchievementMongo, error) {
	if m.GetDetailsByIDsFunc == nil { return nil, nil }
	return m.GetDetailsByIDsFunc(ctx, ids)
}
func (m *MockAchievementMongoRepo) GetDetailByID(ctx context.Context, id primitive.ObjectID) (*model_mongo.AchievementMongo, error) {
	if m.GetDetailByIDFunc == nil { return nil, nil }
	return m.GetDetailByIDFunc(ctx, id)
}
func (m *MockAchievementMongoRepo) Create(ctx context.Context, achievement *model_mongo.AchievementMongo) (*model_mongo.AchievementMongo, error) {
	if m.CreateFunc == nil { return nil, nil }
	return m.CreateFunc(ctx, achievement)
}
func (m *MockAchievementMongoRepo) Update(ctx context.Context, id primitive.ObjectID, data *model_mongo.AchievementInput) error {
	if m.UpdateFunc == nil { return nil }
	return m.UpdateFunc(ctx, id, data)
}
func (m *MockAchievementMongoRepo) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	if m.SoftDeleteFunc == nil { return nil }
	return m.SoftDeleteFunc(ctx, id)
}
func (m *MockAchievementMongoRepo) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	if m.DeleteByIDFunc == nil { return nil }
	return m.DeleteByIDFunc(ctx, id)
}
func (m *MockAchievementMongoRepo) AddAttachment(ctx context.Context, id primitive.ObjectID, attachment model_mongo.Attachment) error {
	if m.AddAttachmentFunc == nil { return nil }
	return m.AddAttachmentFunc(ctx, id, attachment)
}
func (m *MockAchievementMongoRepo) GetAchievementStatistics(ctx context.Context, studentIDs []string) ([]interface{}, error) { return nil, nil } // Sesuaikan return type dg interface Anda (misal []bson.M)
func (m *MockAchievementMongoRepo) GetStudentAchievementDetails(ctx context.Context, studentIDHex string) ([]interface{}, error) { return nil, nil }

type MockAchievementPGRepo struct {
	GetAllAchievementReferencesFunc func() ([]model_postgre.AchievementReference, error)
	GetMyAchievementsFunc           func(studentID string) ([]model_postgre.AchievementReference, error)
	GetAchievementsByStudentIDsFunc func(ids []string) ([]model_postgre.AchievementReference, error)
	GetReferenceByIDFunc            func(id string) (*model_postgre.AchievementReference, error)
	CreateReferenceFunc             func(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error)
	SoftDeleteReferenceFunc         func(id, studentID string) (*model_postgre.AchievementReference, error)
	UpdateStatusToSubmittedFunc     func(id string) (*model_postgre.AchievementReference, error)
	VerifyAchievementFunc           func(id, lecturerID string) (*model_postgre.AchievementReference, error)
	RejectAchievementFunc           func(id, lecturerID, note string) (*model_postgre.AchievementReference, error)
	
	FindStudentIdByUserIDFunc       func(userID string) (string, error)
	FindLecturerIdByUserIDFunc      func(userID string) (string, error)
	GetAdviseeStudentIDsFunc        func(lecturerID string) ([]string, error)
}

func (m *MockAchievementPGRepo) GetAllAchievementReferences() ([]model_postgre.AchievementReference, error) {
	if m.GetAllAchievementReferencesFunc == nil { return nil, nil }
	return m.GetAllAchievementReferencesFunc()
}
func (m *MockAchievementPGRepo) GetMyAchievements(studentID string) ([]model_postgre.AchievementReference, error) {
	if m.GetMyAchievementsFunc == nil { return nil, nil }
	return m.GetMyAchievementsFunc(studentID)
}
func (m *MockAchievementPGRepo) GetAchievementsByStudentIDs(ids []string) ([]model_postgre.AchievementReference, error) {
	if m.GetAchievementsByStudentIDsFunc == nil { return nil, nil }
	return m.GetAchievementsByStudentIDsFunc(ids)
}
func (m *MockAchievementPGRepo) GetReferenceByID(id string) (*model_postgre.AchievementReference, error) {
	if m.GetReferenceByIDFunc == nil { return nil, nil }
	return m.GetReferenceByIDFunc(id)
}
func (m *MockAchievementPGRepo) CreateReference(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error) {
	if m.CreateReferenceFunc == nil { return nil, nil }
	return m.CreateReferenceFunc(ref)
}
func (m *MockAchievementPGRepo) SoftDeleteReference(id, studentID string) (*model_postgre.AchievementReference, error) {
	if m.SoftDeleteReferenceFunc == nil { return nil, nil }
	return m.SoftDeleteReferenceFunc(id, studentID)
}
func (m *MockAchievementPGRepo) UpdateStatusToSubmitted(id string) (*model_postgre.AchievementReference, error) {
	if m.UpdateStatusToSubmittedFunc == nil { return nil, nil }
	return m.UpdateStatusToSubmittedFunc(id)
}
func (m *MockAchievementPGRepo) VerifyAchievement(id, lecturerID string) (*model_postgre.AchievementReference, error) {
	if m.VerifyAchievementFunc == nil { return nil, nil }
	return m.VerifyAchievementFunc(id, lecturerID)
}
func (m *MockAchievementPGRepo) RejectAchievement(id, lecturerID, note string) (*model_postgre.AchievementReference, error) {
	if m.RejectAchievementFunc == nil { return nil, nil }
	return m.RejectAchievementFunc(id, lecturerID, note)
}
func (m *MockAchievementPGRepo) FindStudentIdByUserID(userID string) (string, error) {
	if m.FindStudentIdByUserIDFunc == nil { return "", nil }
	return m.FindStudentIdByUserIDFunc(userID)
}
func (m *MockAchievementPGRepo) FindLecturerIdByUserID(userID string) (string, error) {
	if m.FindLecturerIdByUserIDFunc == nil { return "", nil }
	return m.FindLecturerIdByUserIDFunc(userID)
}
func (m *MockAchievementPGRepo) GetAdviseeStudentIDs(lecturerID string) ([]string, error) {
	if m.GetAdviseeStudentIDsFunc == nil { return nil, nil }
	return m.GetAdviseeStudentIDsFunc(lecturerID)
}
func (m *MockAchievementPGRepo) GetAchievementsByStudentIDsAndStatus(ids []string, status string) ([]model_postgre.AchievementReference, error) { return nil, nil }


// 2. SETUP HELPER
func setupAchievementServiceTestApp(mockMongo *MockAchievementMongoRepo, mockPg *MockAchievementPGRepo) *fiber.App {
	app := fiber.New()
	svc := service.NewAchievementService(mockMongo, mockPg)

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

	app.Get("/achievements", svc.ListAllAchievements)
	app.Post("/achievements", svc.SubmitPrestasi)
	app.Get("/achievements/:id", svc.GetAchievementDetail)
	app.Put("/achievements/:id", svc.UpdatePrestasi)
	app.Delete("/achievements/:id", svc.DeletePrestasi)
	app.Post("/achievements/:id/submit", svc.SubmitForVerification)
	app.Post("/achievements/:id/verify", svc.VerifyPrestasi)
	app.Post("/achievements/:id/reject", svc.RejectPrestasi)
	app.Post("/achievements/:id/attachments", svc.AddAttachment)

	return app
}


// UNIT TESTS
func TestListAllAchievements_Mahasiswa_Success(t *testing.T) {
	mockPg := &MockAchievementPGRepo{
		FindStudentIdByUserIDFunc: func(userID string) (string, error) {
			return "stu-123", nil
		},
		GetMyAchievementsFunc: func(studentID string) ([]model_postgre.AchievementReference, error) {
			return []model_postgre.AchievementReference{
				{ID: "ref-1", MongoAchievementID: "64b0f1a2e4b0a1a2b3c4d5e6"},
			}, nil
		},
	}
	mockMongo := &MockAchievementMongoRepo{
		GetDetailsByIDsFunc: func(ctx context.Context, ids []primitive.ObjectID) ([]model_mongo.AchievementMongo, error) {
			oid, _ := primitive.ObjectIDFromHex("64b0f1a2e4b0a1a2b3c4d5e6")
			return []model_mongo.AchievementMongo{{ID: oid, Title: "Juara 1"}}, nil
		},
	}
	app := setupAchievementServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/achievements", nil)
	req.Header.Set("X-Test-Role", "Mahasiswa")
	req.Header.Set("X-Test-ID", "user-mhs")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListAllAchievements_Admin_Success(t *testing.T) {
	mockPg := &MockAchievementPGRepo{
		GetAllAchievementReferencesFunc: func() ([]model_postgre.AchievementReference, error) {
			return []model_postgre.AchievementReference{{ID: "ref-1", MongoAchievementID: "64b0f1a2e4b0a1a2b3c4d5e6"}}, nil
		},
	}
	mockMongo := &MockAchievementMongoRepo{
		GetDetailsByIDsFunc: func(ctx context.Context, ids []primitive.ObjectID) ([]model_mongo.AchievementMongo, error) {
			oid, _ := primitive.ObjectIDFromHex("64b0f1a2e4b0a1a2b3c4d5e6")
			return []model_mongo.AchievementMongo{{ID: oid, Title: "Juara 1"}}, nil
		},
	}
	app := setupAchievementServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("GET", "/achievements", nil)
	req.Header.Set("X-Test-Role", "Admin")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSubmitPrestasi_Success(t *testing.T) {
	mockPg := &MockAchievementPGRepo{
		FindStudentIdByUserIDFunc: func(userID string) (string, error) { return "stu-123", nil },
		CreateReferenceFunc: func(ref *model_postgre.AchievementReference) (*model_postgre.AchievementReference, error) {
			ref.ID = "ref-new"
			return ref, nil
		},
	}
	mockMongo := &MockAchievementMongoRepo{
		CreateFunc: func(ctx context.Context, achievement *model_mongo.AchievementMongo) (*model_mongo.AchievementMongo, error) {
			achievement.ID = primitive.NewObjectID()
			return achievement, nil
		},
	}
	app := setupAchievementServiceTestApp(mockMongo, mockPg)

	payload := model_mongo.AchievementInput{Title: "Juara 1 Lomba Coding"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", "Mahasiswa")
	req.Header.Set("X-Test-ID", "user-mhs")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestSubmitPrestasi_Forbidden(t *testing.T) {
	mockPg := &MockAchievementPGRepo{}
	mockMongo := &MockAchievementMongoRepo{}
	app := setupAchievementServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("POST", "/achievements", nil)
	req.Header.Set("X-Test-Role", "Dosen Wali") // Salah Role
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUpdatePrestasi_Success(t *testing.T) {
	mockPg := &MockAchievementPGRepo{
		GetReferenceByIDFunc: func(id string) (*model_postgre.AchievementReference, error) {
			return &model_postgre.AchievementReference{
				ID: "ref-1", StudentID: "stu-123", Status: "draft", MongoAchievementID: "64b0f1a2e4b0a1a2b3c4d5e6",
			}, nil
		},
		FindStudentIdByUserIDFunc: func(userID string) (string, error) { return "stu-123", nil },
	}
	mockMongo := &MockAchievementMongoRepo{
		UpdateFunc: func(ctx context.Context, id primitive.ObjectID, data *model_mongo.AchievementInput) error {
			return nil
		},
	}
	app := setupAchievementServiceTestApp(mockMongo, mockPg)

	payload := model_mongo.AchievementInput{Title: "Update Judul"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/achievements/ref-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", "Mahasiswa")
	req.Header.Set("X-Test-ID", "user-mhs")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSubmitForVerification_Success(t *testing.T) {
	mockPg := &MockAchievementPGRepo{
		GetReferenceByIDFunc: func(id string) (*model_postgre.AchievementReference, error) {
			return &model_postgre.AchievementReference{ID: "ref-1", StudentID: "stu-123", Status: "draft"}, nil
		},
		FindStudentIdByUserIDFunc: func(userID string) (string, error) { return "stu-123", nil },
		UpdateStatusToSubmittedFunc: func(id string) (*model_postgre.AchievementReference, error) {
			return &model_postgre.AchievementReference{Status: "submitted"}, nil
		},
	}
	mockMongo := &MockAchievementMongoRepo{}
	app := setupAchievementServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("POST", "/achievements/ref-1/submit", nil)
	req.Header.Set("X-Test-Role", "Mahasiswa")
	req.Header.Set("X-Test-ID", "user-mhs")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestVerifyPrestasi_Success(t *testing.T) {
	mockPg := &MockAchievementPGRepo{
		VerifyAchievementFunc: func(id, lecturerID string) (*model_postgre.AchievementReference, error) {
			return &model_postgre.AchievementReference{Status: "verified"}, nil
		},
	}
	mockMongo := &MockAchievementMongoRepo{}
	app := setupAchievementServiceTestApp(mockMongo, mockPg)

	req := httptest.NewRequest("POST", "/achievements/ref-1/verify", nil)
	req.Header.Set("X-Test-Role", "Dosen Wali")
	req.Header.Set("X-Test-ID", "user-dosen")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAddAttachment_Success(t *testing.T) {
	mockPg := &MockAchievementPGRepo{
		GetReferenceByIDFunc: func(id string) (*model_postgre.AchievementReference, error) {
			return &model_postgre.AchievementReference{MongoAchievementID: "64b0f1a2e4b0a1a2b3c4d5e6"}, nil
		},
	}
	mockMongo := &MockAchievementMongoRepo{
		AddAttachmentFunc: func(ctx context.Context, id primitive.ObjectID, attachment model_mongo.Attachment) error {
			return nil
		},
	}
	app := setupAchievementServiceTestApp(mockMongo, mockPg)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "sertifikat.pdf")
	part.Write([]byte("dummy content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/achievements/ref-1/attachments", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Test-Role", "Mahasiswa")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
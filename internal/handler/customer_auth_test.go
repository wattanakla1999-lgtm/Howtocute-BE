package handler

import (
	"bytes"
	"encoding/json"
	"nailly-back-end/internal/middleware"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"
	"nailly-back-end/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestLineLoginRequiresIDToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authHandler := NewAuthHandler(nil, nil)
	router.POST("/api/auth/line", authHandler.LineLogin)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/line", bytes.NewBufferString(`{}`)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestGetCustomerMeWithCustomerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lineUserID := "U123"
	phone := "0812345678"
	customer := model.User{
		Model:      gorm.Model{ID: 1},
		Name:       "Customer Name",
		Phone:      &phone,
		LineUserID: &lineUserID,
		PictureURL: "https://example.com/pic.jpg",
	}

	jwtManager := service.NewCustomerJWTManager("customer-secret", time.Hour)
	token, _, err := jwtManager.Generate(customer)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	router := gin.New()
	userHandler := NewUserHandler(fakeUserService{customer: customer})
	router.GET("/api/customers/me", middleware.RequireCustomer(jwtManager), userHandler.GetCustomerMe)

	request := httptest.NewRequest(http.MethodGet, "/api/customers/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", response.Code, response.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body["name"] != "Customer Name" || body["lineUserId"] != lineUserID || body["pictureUrl"] != "https://example.com/pic.jpg" {
		t.Fatalf("body = %#v", body)
	}
}

type fakeUserService struct {
	customer model.User
}

func (f fakeUserService) GetUsers(repository.UserFilter, utils.Pagination) ([]model.User, int64, error) {
	return nil, 0, nil
}

func (f fakeUserService) GetUserByID(string) (model.User, error) {
	return f.customer, nil
}

func (f fakeUserService) GetUserByEmail(string) (model.User, error) {
	return model.User{}, nil
}

func (f fakeUserService) GetUsersOlderThan(int) ([]model.User, error) {
	return nil, nil
}

func (f fakeUserService) CreateUser(model.User) (model.User, error) {
	return model.User{}, nil
}

func (f fakeUserService) UpdateUser(string, model.User) (model.User, error) {
	return model.User{}, nil
}

func (f fakeUserService) DeleteUser(string) error {
	return nil
}

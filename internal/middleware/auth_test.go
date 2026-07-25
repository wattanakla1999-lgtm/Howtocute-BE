package middleware

import (
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := service.NewJWTManager("middleware-test-secret", time.Hour)
	token, _, err := manager.Generate(model.Admin{
		Model: gorm.Model{ID: 7}, Username: "admin", Name: "Admin", Role: "admin",
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	router := gin.New()
	router.GET("/protected", RequireAdmin(manager), func(c *gin.Context) {
		claims, ok := AdminClaimsFromContext(c)
		if !ok || claims.AdminID != 7 {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	withoutToken := httptest.NewRecorder()
	router.ServeHTTP(withoutToken, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if withoutToken.Code != http.StatusUnauthorized {
		t.Fatalf("without token status = %d, want 401", withoutToken.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	withToken := httptest.NewRecorder()
	router.ServeHTTP(withToken, request)
	if withToken.Code != http.StatusOK {
		t.Fatalf("with token status = %d, want 200", withToken.Code)
	}
}

func TestRequireCustomer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := service.NewCustomerJWTManager("customer-middleware-secret", time.Hour)
	lineUserID := "U123"
	token, _, err := manager.Generate(model.User{
		Model: gorm.Model{ID: 9}, Name: "Customer", LineUserID: &lineUserID,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	router := gin.New()
	router.GET("/customer", RequireCustomer(manager), func(c *gin.Context) {
		claims, ok := CustomerClaimsFromContext(c)
		if !ok || claims.Subject != "9" || claims.LineUserID != lineUserID {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/customer", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
}

func TestRequireAdminRejectsCustomerJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	lineUserID := "U123"
	customerToken, _, err := service.NewCustomerJWTManager("customer-secret", time.Hour).Generate(model.User{
		Model: gorm.Model{ID: 9}, Name: "Customer", LineUserID: &lineUserID,
	})
	if err != nil {
		t.Fatalf("Generate customer token error = %v", err)
	}

	router := gin.New()
	router.GET("/admin", RequireAdmin(service.NewJWTManager("admin-secret", time.Hour)), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer "+customerToken)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestRequireCustomerRejectsAdminJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminToken, _, err := service.NewJWTManager("admin-secret", time.Hour).Generate(model.Admin{
		Model: gorm.Model{ID: 7}, Username: "admin", Name: "Admin", Role: "admin",
	})
	if err != nil {
		t.Fatalf("Generate admin token error = %v", err)
	}

	router := gin.New()
	router.GET("/customer", RequireCustomer(service.NewCustomerJWTManager("customer-secret", time.Hour)), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/customer", nil)
	request.Header.Set("Authorization", "Bearer "+adminToken)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

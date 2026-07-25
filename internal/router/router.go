package router

import (
	"nailly-back-end/internal/middleware"
	"nailly-back-end/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, allowOrigin string, jwtManager *service.JWTManager, customerJWTManager *service.CustomerJWTManager, lineLoginChannelID string) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware(allowOrigin))

	// Root
	r.GET("/", rootHandler)
	r.HEAD("/", rootHandler)

	api := r.Group("/api")
	requireAdmin := middleware.RequireAdmin(jwtManager)
	requireCustomer := middleware.RequireCustomer(customerJWTManager)

	// Keep-alive
	api.GET("/keep-alive", keepAliveHandler(db))
	api.HEAD("/keep-alive", keepAliveHandler(db))

	// Authentication
	RegisterAuthRoutes(api, db, jwtManager, customerJWTManager, lineLoginChannelID, requireAdmin)

	// Users
	RegisterUserRoutes(api, db, requireAdmin)

	// Customers
	RegisterCustomerRoutes(api, db, requireCustomer)

	// Services
	RegisterServiceRoutes(api, db, requireAdmin)

	// Nail Technicians
	RegisterNailTechnicianRoutes(api, db, requireAdmin)

	// Bookings
	RegisterBookingRoutes(api, db, requireAdmin, requireCustomer)

	// Reports
	RegisterReportRoutes(api, db, requireAdmin)

	// Dashboard
	RegisterDashboardRoutes(api, db, requireAdmin)

	// Shop Settings
	RegisterShopSettingRoutes(api, db, requireAdmin)

	return r
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "nailly-api",
		"status":  "running",
		"version": "v1.0.1",
	})
}

func keepAliveHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var result int
		if err := db.WithContext(c.Request.Context()).Raw("SELECT 1").Scan(&result).Error; err != nil {
			if c.Request.Method == http.MethodHead {
				c.Status(http.StatusServiceUnavailable)
				return
			}
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "unreachable",
			})
			return
		}

		if c.Request.Method == http.MethodHead {
			c.Status(http.StatusOK)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "active",
		})
	}
}

func corsMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

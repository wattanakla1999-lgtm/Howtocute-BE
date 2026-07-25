package middleware

import (
	"nailly-back-end/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const adminClaimsContextKey = "adminClaims"
const customerClaimsContextKey = "customerClaims"

func RequireAdmin(jwtManager *service.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := strings.TrimSpace(c.GetHeader("Authorization"))
		scheme, tokenValue, found := strings.Cut(authorization, " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(tokenValue) == "" {
			abortUnauthorized(c, "กรุณาเข้าสู่ระบบ")
			return
		}

		claims, err := jwtManager.Verify(strings.TrimSpace(tokenValue))
		if err != nil || claims.Role != "admin" {
			abortUnauthorized(c, "โทเคนไม่ถูกต้องหรือหมดอายุ")
			return
		}
		c.Set(adminClaimsContextKey, claims)
		c.Next()
	}
}

func AdminClaimsFromContext(c *gin.Context) (*service.AdminClaims, bool) {
	value, exists := c.Get(adminClaimsContextKey)
	claims, ok := value.(*service.AdminClaims)
	return claims, exists && ok
}

func RequireCustomer(jwtManager *service.CustomerJWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := strings.TrimSpace(c.GetHeader("Authorization"))
		scheme, tokenValue, found := strings.Cut(authorization, " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(tokenValue) == "" {
			abortUnauthorized(c, "กรุณาเข้าสู่ระบบ")
			return
		}

		claims, err := jwtManager.Verify(strings.TrimSpace(tokenValue))
		if err != nil {
			abortUnauthorized(c, "โทเคนลูกค้าไม่ถูกต้องหรือหมดอายุ")
			return
		}
		c.Set(customerClaimsContextKey, claims)
		c.Set("customerID", claims.Subject)
		c.Set("lineUserID", claims.LineUserID)
		c.Next()
	}
}

func CustomerClaimsFromContext(c *gin.Context) (*service.CustomerClaims, bool) {
	value, exists := c.Get(customerClaimsContextKey)
	claims, ok := value.(*service.CustomerClaims)
	return claims, exists && ok
}

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":  "UNAUTHORIZED",
		"error": message,
	})
}

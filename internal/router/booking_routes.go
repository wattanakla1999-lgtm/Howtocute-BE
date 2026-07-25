package router

import (
	"nailly-back-end/internal/handler"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterBookingRoutes(api *gin.RouterGroup, db *gorm.DB, requireAdmin gin.HandlerFunc, requireCustomer gin.HandlerFunc, bookingNotifier service.BookingNotifier) {
	bookingRepository := repository.NewBookingRepository(db)
	bookingService := service.NewBookingService(bookingRepository, bookingNotifier)
	bookingHandler := handler.NewBookingHandler(bookingService)

	bookings := api.Group("/bookings")
	bookings.GET("", requireAdmin, bookingHandler.GetBookings)
	bookings.GET("/busy-slots", bookingHandler.GetBusySlots)
	bookings.GET("/customer", bookingHandler.GetCustomerBookings)
	bookings.GET("/customer/me", requireCustomer, bookingHandler.GetMyBookings)
	bookings.GET("/:id", requireAdmin, bookingHandler.GetBookingByID)
	bookings.POST("", bookingHandler.CreateBooking)
	bookings.POST("/customer", requireCustomer, bookingHandler.CreateCustomerBooking)
	bookings.PATCH("/customer/:id/cancel", requireCustomer, bookingHandler.CancelCustomerBooking)
	bookings.PUT("/:id", requireAdmin, bookingHandler.UpdateBooking)
	bookings.DELETE("/:id", requireAdmin, bookingHandler.DeleteBooking)
	bookings.PATCH("/:id/status", requireAdmin, bookingHandler.UpdateBookingStatus)
}

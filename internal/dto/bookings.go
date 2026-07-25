package dto

import (
	"bytes"
	"encoding/json"
	"nailly-back-end/internal/model"
	"time"
)

type NullableUint struct {
	Value *uint
	Set   bool
}

func (n *NullableUint) UnmarshalJSON(data []byte) error {
	n.Set = true
	if bytes.Equal(data, []byte("null")) {
		n.Value = nil
		return nil
	}

	var value uint
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	n.Value = &value
	return nil
}

type CreateBookingRequest struct {
	UserID        *uint               `json:"userId"`
	ServiceID     uint                `json:"serviceId" binding:"required"`
	TechnicianID  *uint               `json:"technicianId"`
	StartAt       time.Time           `json:"startAt" binding:"required"`
	EndAt         *time.Time          `json:"endAt"`
	CustomerName  string              `json:"customerName" binding:"required"`
	CustomerPhone string              `json:"customerPhone" binding:"required"`
	PaymentMethod model.PaymentMethod `json:"paymentMethod"`
	Note          string              `json:"note"`
}

type UpdateBookingRequest struct {
	UserID        NullableUint         `json:"userId"`
	ServiceID     *uint                `json:"serviceId"`
	TechnicianID  NullableUint         `json:"technicianId"`
	StartAt       *time.Time           `json:"startAt"`
	EndAt         *time.Time           `json:"endAt"`
	CustomerName  *string              `json:"customerName"`
	CustomerPhone *string              `json:"customerPhone"`
	PaymentMethod *model.PaymentMethod `json:"paymentMethod"`
	Note          *string              `json:"note"`
}

type UpdateBookingStatusRequest struct {
	Status       model.BookingStatus `json:"status" binding:"required"`
	CancelReason string              `json:"cancelReason"`
}

type CancelBookingRequest struct {
	CancelReason string `json:"cancelReason"`
}

type BookingUserResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type BookingServiceResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Duration int    `json:"durationMinutes"`
}

type BookingTechnicianResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	ProfileImg string `json:"profileImg,omitempty"`
}

type BookingResponse struct {
	ID              uint                       `json:"id"`
	BookingNo       string                     `json:"bookingNo"`
	UserID          *uint                      `json:"userId"`
	ServiceID       uint                       `json:"serviceId"`
	TechnicianID    *uint                      `json:"technicianId"`
	StartAt         time.Time                  `json:"startAt"`
	EndAt           time.Time                  `json:"endAt"`
	CustomerName    string                     `json:"customerName"`
	CustomerPhone   string                     `json:"customerPhone"`
	ServiceName     string                     `json:"serviceName"`
	Price           int                        `json:"price"`
	DurationMinutes int                        `json:"durationMinutes"`
	Status          model.BookingStatus        `json:"status"`
	PaymentMethod   model.PaymentMethod        `json:"paymentMethod"`
	Note            string                     `json:"note,omitempty"`
	CancelReason    string                     `json:"cancelReason,omitempty"`
	User            *BookingUserResponse       `json:"user"`
	Service         BookingServiceResponse     `json:"service"`
	Technician      *BookingTechnicianResponse `json:"technician"`
	CreatedAt       time.Time                  `json:"createdAt"`
	UpdatedAt       time.Time                  `json:"updatedAt"`
}

func ToBookingResponse(booking model.Booking) BookingResponse {
	response := BookingResponse{
		ID:              booking.ID,
		BookingNo:       booking.BookingNo,
		UserID:          booking.UserID,
		ServiceID:       booking.ServiceID,
		TechnicianID:    booking.TechnicianID,
		StartAt:         booking.StartAt.In(thailandLocation),
		EndAt:           booking.EndAt.In(thailandLocation),
		CustomerName:    booking.CustomerName,
		CustomerPhone:   booking.CustomerPhone,
		ServiceName:     booking.ServiceName,
		Price:           booking.Price,
		DurationMinutes: booking.DurationMinutes,
		Status:          booking.Status,
		PaymentMethod:   booking.PaymentMethod,
		Note:            booking.Note,
		CancelReason:    booking.CancelReason,
		Service: BookingServiceResponse{
			ID:       booking.Service.ID,
			Name:     booking.Service.ServiceName,
			Price:    booking.Service.ServicePrice,
			Duration: booking.Service.Duration,
		},
		CreatedAt: booking.CreatedAt.In(thailandLocation),
		UpdatedAt: booking.UpdatedAt.In(thailandLocation),
	}
	if booking.User != nil {
		response.User = &BookingUserResponse{ID: booking.User.ID, Name: booking.User.Name}
	}

	if booking.Technician != nil {
		response.Technician = &BookingTechnicianResponse{
			ID:         booking.Technician.ID,
			Name:       booking.Technician.TechnicianName,
			ProfileImg: booking.Technician.ProfileImg,
		}
	}

	return response
}

func ToBookingResponses(bookings []model.Booking) []BookingResponse {
	responses := make([]BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		responses = append(responses, ToBookingResponse(booking))
	}
	return responses
}

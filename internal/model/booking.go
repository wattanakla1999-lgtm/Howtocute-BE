package model

import (
	"time"

	"gorm.io/gorm"
)

type BookingStatus string
type PaymentMethod string
type DepositStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusInService BookingStatus = "in_service"
	BookingStatusCompleted BookingStatus = "completed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusNoShow    BookingStatus = "no_show"
)

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodCard     PaymentMethod = "card"
)

const (
	DepositStatusNone     DepositStatus = "none"
	DepositStatusPending  DepositStatus = "pending"
	DepositStatusVerified DepositStatus = "verified"
	DepositStatusRejected DepositStatus = "rejected"
)

type Booking struct {
	gorm.Model

	BookingNo           string        `gorm:"type:varchar(50);not null;uniqueIndex" json:"bookingNo"`
	UserID              *uint         `gorm:"index" json:"userId"`
	ServiceID           uint          `gorm:"not null;index" json:"serviceId"`
	TechnicianID        *uint         `gorm:"index" json:"technicianId,omitempty"`
	StartAt             time.Time     `gorm:"not null;index" json:"startAt"`
	EndAt               time.Time     `gorm:"not null;check:end_at > start_at" json:"endAt"`
	CustomerName        string        `gorm:"type:varchar(255);not null" json:"customerName"`
	CustomerPhone       string        `gorm:"type:varchar(50);not null" json:"customerPhone"`
	ServiceName         string        `gorm:"type:varchar(255);not null" json:"serviceName"`
	Price               int           `gorm:"not null;check:price >= 0" json:"price"`
	DurationMinutes     int           `gorm:"not null;check:duration_minutes > 0" json:"durationMinutes"`
	Status              BookingStatus `gorm:"type:varchar(20);not null;default:pending;index;check:status IN ('pending','confirmed','in_service','completed','cancelled','no_show')" json:"status"`
	PaymentMethod       PaymentMethod `gorm:"type:varchar(20);not null;default:cash;index;check:payment_method IN ('cash','transfer','card')" json:"paymentMethod"`
	Note                string        `gorm:"type:text" json:"note,omitempty"`
	CancelReason        string        `gorm:"type:text" json:"cancelReason,omitempty"`
	DepositAmount       float64       `gorm:"type:numeric(10,2);not null;default:0;check:deposit_amount >= 0" json:"depositAmount"`
	DepositStatus       DepositStatus `gorm:"type:varchar(20);not null;default:none;index;check:deposit_status IN ('none','pending','verified','rejected')" json:"depositStatus"`
	SlipURL             string        `gorm:"type:text" json:"slipUrl,omitempty"`
	SlipUploadedAt      *time.Time    `json:"slipUploadedAt,omitempty"`
	DepositRejectReason string        `gorm:"type:text" json:"depositRejectReason,omitempty"`

	User       *User           `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
	Service    Service         `gorm:"foreignKey:ServiceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Technician *NailTechnician `gorm:"foreignKey:TechnicianID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
}

func IsValidDepositStatus(status DepositStatus) bool {
	switch status {
	case DepositStatusNone, DepositStatusPending, DepositStatusVerified, DepositStatusRejected:
		return true
	default:
		return false
	}
}

func IsValidPaymentMethod(method PaymentMethod) bool {
	switch method {
	case PaymentMethodCash, PaymentMethodTransfer, PaymentMethodCard:
		return true
	default:
		return false
	}
}

func (Booking) TableName() string {
	return "bookings"
}

func IsValidBookingStatus(status BookingStatus) bool {
	switch status {
	case BookingStatusPending, BookingStatusConfirmed, BookingStatusInService,
		BookingStatusCompleted, BookingStatusCancelled, BookingStatusNoShow:
		return true
	default:
		return false
	}
}

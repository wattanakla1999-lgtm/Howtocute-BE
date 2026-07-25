package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/pkg/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

type BookingStore interface {
	FindAll(filter repository.BookingFilter, pagination utils.Pagination) ([]model.Booking, int64, error)
	FindByID(id uint) (model.Booking, error)
	FindUserByID(id uint) (model.User, error)
	FindServiceByID(id uint) (model.Service, error)
	FindTechnicianByID(id uint) (model.NailTechnician, error)
	HasTechnicianOverlap(technicianID uint, startAt, endAt time.Time, excludeBookingID uint) (bool, error)
	FindBusyBookings(startAt, endAt time.Time, technicianID *uint) ([]model.Booking, error)
	FindActiveTechnicianIDs() ([]uint, error)
	Create(booking *model.Booking, walkInCustomer *model.User) error
	Update(booking *model.Booking) error
	Delete(booking *model.Booking) error
}

type CreateBookingInput struct {
	UserID        *uint
	ServiceID     uint
	TechnicianID  *uint
	StartAt       time.Time
	EndAt         *time.Time
	CustomerName  string
	CustomerPhone string
	PaymentMethod model.PaymentMethod
	Note          string
}

type UpdateBookingInput struct {
	UserID          *uint
	UserIDSet       bool
	ServiceID       *uint
	TechnicianID    *uint
	TechnicianIDSet bool
	StartAt         *time.Time
	EndAt           *time.Time
	CustomerName    *string
	CustomerPhone   *string
	PaymentMethod   *model.PaymentMethod
	Note            *string
}

type BookingService struct {
	repo             BookingStore
	bookingNoFactory func() (string, error)
	notifier         BookingNotifier
}

func NewBookingService(repo BookingStore, notifiers ...BookingNotifier) *BookingService {
	var notifier BookingNotifier
	if len(notifiers) > 0 {
		notifier = notifiers[0]
	}
	return &BookingService{repo: repo, bookingNoFactory: generateBookingNo, notifier: notifier}
}

func (s *BookingService) GetBookings(filter repository.BookingFilter, pagination utils.Pagination) ([]model.Booking, int64, error) {
	if filter.Status != "" && !model.IsValidBookingStatus(filter.Status) {
		return nil, 0, apperror.BadRequest("invalid booking status", apperror.ErrValidation)
	}
	if filter.DateFrom != nil && filter.DateTo != nil && filter.DateFrom.After(*filter.DateTo) {
		return nil, 0, apperror.BadRequest("dateFrom must be before dateTo", apperror.ErrValidation)
	}
	return s.repo.FindAll(filter, pagination)
}

func (s *BookingService) GetBookingByID(id uint) (model.Booking, error) {
	booking, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Booking{}, apperror.NotFound("booking not found", err)
	}
	return booking, err
}

func (s *BookingService) GetBusySlots(date time.Time, technicianID, serviceID *uint) ([]string, error) {
	if date.IsZero() {
		return nil, apperror.BadRequest("date is required", apperror.ErrValidation)
	}
	serviceDuration := time.Hour
	if serviceID != nil {
		serviceModel, err := s.findService(*serviceID)
		if err != nil {
			return nil, err
		}
		if serviceModel.Duration <= 0 {
			return nil, apperror.Internal("service has invalid duration", apperror.ErrValidation)
		}
		serviceDuration = time.Duration(serviceModel.Duration) * time.Minute
	}
	if technicianID != nil {
		technician, err := s.findTechnician(*technicianID)
		if err != nil {
			return nil, err
		}
		if !technician.Active {
			return nil, apperror.BadRequest("technician is inactive", apperror.ErrValidation)
		}
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	lastSlotEnd := time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location()).Add(serviceDuration)
	bookings, err := s.repo.FindBusyBookings(dayStart, lastSlotEnd, technicianID)
	if err != nil {
		return nil, err
	}

	if technicianID != nil {
		return busySlotsForTechnician(dayStart, serviceDuration, bookings), nil
	}
	activeTechnicianIDs, err := s.repo.FindActiveTechnicianIDs()
	if err != nil {
		return nil, err
	}
	return busySlotsForAnyTechnician(dayStart, serviceDuration, bookings, activeTechnicianIDs), nil
}

func (s *BookingService) CreateBooking(input CreateBookingInput) (model.Booking, error) {
	if input.ServiceID == 0 || input.StartAt.IsZero() {
		return model.Booking{}, apperror.BadRequest("serviceId and startAt are required", apperror.ErrValidation)
	}
	if strings.TrimSpace(input.CustomerName) == "" || strings.TrimSpace(input.CustomerPhone) == "" {
		return model.Booking{}, apperror.BadRequest("customerName and customerPhone are required", apperror.ErrValidation)
	}

	customerName := strings.TrimSpace(input.CustomerName)
	customerPhone := strings.TrimSpace(input.CustomerPhone)
	var user *model.User
	var walkInCustomer *model.User
	if input.UserID != nil {
		if *input.UserID == 0 {
			return model.Booking{}, apperror.BadRequest("userId must be greater than 0 or null", apperror.ErrValidation)
		}
		found, err := s.findUser(*input.UserID)
		if err != nil {
			return model.Booking{}, err
		}
		user = &found
	} else {
		walkInCustomer = &model.User{
			Name:  customerName,
			Email: walkInEmail(customerPhone),
			Phone: &customerPhone,
		}
	}
	serviceModel, err := s.findService(input.ServiceID)
	if err != nil {
		return model.Booking{}, err
	}
	if serviceModel.ServicePrice < 0 || serviceModel.Duration <= 0 {
		return model.Booking{}, apperror.Internal("service has invalid price or duration", apperror.ErrValidation)
	}

	var technician *model.NailTechnician
	if input.TechnicianID != nil {
		if *input.TechnicianID == 0 {
			return model.Booking{}, apperror.BadRequest("technicianId must be greater than 0", apperror.ErrValidation)
		}
		found, findErr := s.findTechnician(*input.TechnicianID)
		if findErr != nil {
			return model.Booking{}, findErr
		}
		technician = &found
	}

	endAt := input.StartAt.Add(time.Duration(serviceModel.Duration) * time.Minute)
	if input.EndAt != nil {
		endAt = *input.EndAt
	}
	if !input.StartAt.Before(endAt) {
		return model.Booking{}, apperror.BadRequest("startAt must be before endAt", apperror.ErrValidation)
	}
	if err := s.ensureNoOverlap(input.TechnicianID, input.StartAt, endAt, 0); err != nil {
		return model.Booking{}, err
	}

	bookingNo, err := s.bookingNoFactory()
	if err != nil {
		return model.Booking{}, apperror.Internal("could not generate booking number", err)
	}
	paymentMethod := input.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = model.PaymentMethodCash
	}
	if !model.IsValidPaymentMethod(paymentMethod) {
		return model.Booking{}, apperror.BadRequest("invalid payment method", apperror.ErrValidation)
	}
	booking := model.Booking{
		BookingNo:       bookingNo,
		UserID:          input.UserID,
		ServiceID:       input.ServiceID,
		TechnicianID:    input.TechnicianID,
		StartAt:         input.StartAt,
		EndAt:           endAt,
		CustomerName:    customerName,
		CustomerPhone:   customerPhone,
		ServiceName:     serviceModel.ServiceName,
		Price:           serviceModel.ServicePrice,
		DurationMinutes: serviceModel.Duration,
		Status:          model.BookingStatusPending,
		PaymentMethod:   paymentMethod,
		Note:            input.Note,
		User:            user,
		Service:         serviceModel,
		Technician:      technician,
	}
	if err := s.repo.Create(&booking, walkInCustomer); err != nil {
		if errors.Is(err, repository.ErrTechnicianOverlap) {
			return model.Booking{}, bookingTimeOverlapError(err)
		}
		return model.Booking{}, err
	}
	s.notifyOwnerBookingCreated(booking)
	return booking, nil
}

func walkInEmail(phone string) string {
	hash := sha256.Sum256([]byte(phone))
	return fmt.Sprintf("walkin-%x@nailly.local", hash[:8])
}

func (s *BookingService) UpdateBooking(id uint, input UpdateBookingInput) (model.Booking, error) {
	booking, err := s.GetBookingByID(id)
	if err != nil {
		return model.Booking{}, err
	}

	recalculateEnd := false
	if input.UserIDSet {
		booking.UserID = input.UserID
		booking.User = nil
		if input.UserID != nil {
			if *input.UserID == 0 {
				return model.Booking{}, apperror.BadRequest("userId must be greater than 0 or null", apperror.ErrValidation)
			}
			user, findErr := s.findUser(*input.UserID)
			if findErr != nil {
				return model.Booking{}, findErr
			}
			booking.User = &user
		}
	}
	if input.ServiceID != nil {
		serviceModel, findErr := s.findService(*input.ServiceID)
		if findErr != nil {
			return model.Booking{}, findErr
		}
		if serviceModel.ServicePrice < 0 || serviceModel.Duration <= 0 {
			return model.Booking{}, apperror.Internal("service has invalid price or duration", apperror.ErrValidation)
		}
		booking.ServiceID = *input.ServiceID
		booking.Service = serviceModel
		booking.ServiceName = serviceModel.ServiceName
		booking.Price = serviceModel.ServicePrice
		booking.DurationMinutes = serviceModel.Duration
		recalculateEnd = true
	}
	if input.TechnicianIDSet {
		booking.TechnicianID = input.TechnicianID
		booking.Technician = nil
		if input.TechnicianID != nil {
			technician, findErr := s.findTechnician(*input.TechnicianID)
			if findErr != nil {
				return model.Booking{}, findErr
			}
			booking.Technician = &technician
		}
	}
	if input.StartAt != nil {
		booking.StartAt = *input.StartAt
		recalculateEnd = true
	}
	if input.EndAt != nil {
		booking.EndAt = *input.EndAt
	} else if recalculateEnd {
		booking.EndAt = booking.StartAt.Add(time.Duration(booking.DurationMinutes) * time.Minute)
	}
	if input.CustomerName != nil {
		booking.CustomerName = strings.TrimSpace(*input.CustomerName)
	}
	if input.CustomerPhone != nil {
		booking.CustomerPhone = strings.TrimSpace(*input.CustomerPhone)
	}
	if input.PaymentMethod != nil {
		if !model.IsValidPaymentMethod(*input.PaymentMethod) {
			return model.Booking{}, apperror.BadRequest("invalid payment method", apperror.ErrValidation)
		}
		booking.PaymentMethod = *input.PaymentMethod
	}
	if input.Note != nil {
		booking.Note = *input.Note
	}
	if booking.CustomerName == "" || booking.CustomerPhone == "" || !booking.StartAt.Before(booking.EndAt) {
		return model.Booking{}, apperror.BadRequest("booking data or time range is invalid", apperror.ErrValidation)
	}
	if booking.Status != model.BookingStatusCancelled && booking.Status != model.BookingStatusNoShow {
		if err := s.ensureNoOverlap(booking.TechnicianID, booking.StartAt, booking.EndAt, booking.ID); err != nil {
			return model.Booking{}, err
		}
	}
	if err := s.repo.Update(&booking); err != nil {
		if errors.Is(err, repository.ErrTechnicianOverlap) {
			return model.Booking{}, bookingTimeOverlapError(err)
		}
		return model.Booking{}, err
	}
	return booking, nil
}

func (s *BookingService) CancelCustomerBooking(id, customerID uint, cancelReason string) (model.Booking, error) {
	booking, err := s.GetBookingByID(id)
	if err != nil {
		return model.Booking{}, err
	}
	if booking.UserID == nil || *booking.UserID != customerID {
		return model.Booking{}, apperror.NotFound("booking not found", gorm.ErrRecordNotFound)
	}
	if booking.Status == model.BookingStatusCancelled {
		return booking, nil
	}
	booking.Status = model.BookingStatusCancelled
	booking.CancelReason = strings.TrimSpace(cancelReason)
	if booking.CancelReason == "" {
		booking.CancelReason = "ยกเลิกโดยลูกค้า"
	}
	if err := s.repo.Update(&booking); err != nil {
		return model.Booking{}, err
	}
	s.notifyOwnerBookingCancelled(booking)
	return booking, nil
}

func (s *BookingService) UpdateBookingStatus(id uint, status model.BookingStatus, cancelReason string) (model.Booking, error) {
	if !model.IsValidBookingStatus(status) {
		return model.Booking{}, apperror.BadRequest("invalid booking status", apperror.ErrValidation)
	}
	booking, err := s.GetBookingByID(id)
	if err != nil {
		return model.Booking{}, err
	}
	if status != model.BookingStatusCancelled && status != model.BookingStatusNoShow {
		if err := s.ensureNoOverlap(booking.TechnicianID, booking.StartAt, booking.EndAt, booking.ID); err != nil {
			return model.Booking{}, err
		}
	}
	booking.Status = status
	booking.CancelReason = cancelReason
	if err := s.repo.Update(&booking); err != nil {
		if errors.Is(err, repository.ErrTechnicianOverlap) {
			return model.Booking{}, bookingTimeOverlapError(err)
		}
		return model.Booking{}, err
	}
	if status == model.BookingStatusConfirmed {
		s.notifyCustomerBookingConfirmed(booking)
	}
	return booking, nil
}

func (s *BookingService) DeleteBooking(id uint) error {
	booking, err := s.GetBookingByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(&booking)
}

func (s *BookingService) notifyOwnerBookingCreated(booking model.Booking) {
	if s.notifier == nil {
		return
	}
	err := s.notifier.NotifyOwnerBookingCreated(context.Background(), booking)
	logNotificationFailure("owner_booking_created", booking, err)
}

func (s *BookingService) notifyOwnerBookingCancelled(booking model.Booking) {
	if s.notifier == nil {
		return
	}
	err := s.notifier.NotifyOwnerBookingCancelled(context.Background(), booking)
	logNotificationFailure("owner_booking_cancelled", booking, err)
}

func (s *BookingService) notifyCustomerBookingConfirmed(booking model.Booking) {
	if s.notifier == nil {
		return
	}
	err := s.notifier.NotifyCustomerBookingConfirmed(context.Background(), booking)
	logNotificationFailure("customer_booking_confirmed", booking, err)
}

func (s *BookingService) findUser(id uint) (model.User, error) {
	user, err := s.repo.FindUserByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, apperror.NotFound("user not found", err)
	}
	return user, err
}

func (s *BookingService) findService(id uint) (model.Service, error) {
	serviceModel, err := s.repo.FindServiceByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Service{}, apperror.NotFound("service not found", err)
	}
	return serviceModel, err
}

func (s *BookingService) findTechnician(id uint) (model.NailTechnician, error) {
	technician, err := s.repo.FindTechnicianByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NailTechnician{}, apperror.NotFound("technician not found", err)
	}
	return technician, err
}

func (s *BookingService) ensureNoOverlap(technicianID *uint, startAt, endAt time.Time, excludeBookingID uint) error {
	if technicianID == nil {
		return nil
	}
	overlaps, err := s.repo.HasTechnicianOverlap(*technicianID, startAt, endAt, excludeBookingID)
	if err != nil {
		return err
	}
	if overlaps {
		return bookingTimeOverlapError(apperror.ErrValidation)
	}
	return nil
}

func bookingTimeOverlapError(err error) error {
	return apperror.ConflictWithCode(
		apperror.CodeBookingTimeOverlap,
		"ช่วงเวลานี้ทับซ้อนกับการจองอื่น",
		err,
	)
}

func busySlotsForTechnician(dayStart time.Time, serviceDuration time.Duration, bookings []model.Booking) []string {
	busySlots := make([]string, 0)
	for _, slotStart := range bookingSlotStarts(dayStart) {
		if hasBookingOverlap(bookings, slotStart, slotStart.Add(serviceDuration)) {
			busySlots = append(busySlots, slotStart.Format("15:04"))
		}
	}
	return busySlots
}

func busySlotsForAnyTechnician(dayStart time.Time, serviceDuration time.Duration, bookings []model.Booking, activeTechnicianIDs []uint) []string {
	busySlots := make([]string, 0)
	active := make(map[uint]struct{}, len(activeTechnicianIDs))
	for _, id := range activeTechnicianIDs {
		active[id] = struct{}{}
	}

	for _, slotStart := range bookingSlotStarts(dayStart) {
		if len(active) == 0 {
			busySlots = append(busySlots, slotStart.Format("15:04"))
			continue
		}
		slotEnd := slotStart.Add(serviceDuration)
		busyTechnicians := make(map[uint]struct{}, len(active))
		unassignedBookings := 0
		for _, booking := range bookings {
			if !booking.StartAt.Before(slotEnd) || !booking.EndAt.After(slotStart) {
				continue
			}
			if booking.TechnicianID == nil {
				unassignedBookings++
				continue
			}
			if _, isActive := active[*booking.TechnicianID]; isActive {
				busyTechnicians[*booking.TechnicianID] = struct{}{}
			}
		}
		if len(busyTechnicians)+unassignedBookings >= len(active) {
			busySlots = append(busySlots, slotStart.Format("15:04"))
		}
	}
	return busySlots
}

func bookingSlotStarts(dayStart time.Time) []time.Time {
	slots := make([]time.Time, 0, 9)
	for hour := 10; hour <= 18; hour++ {
		slots = append(slots, time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), hour, 0, 0, 0, dayStart.Location()))
	}
	return slots
}

func hasBookingOverlap(bookings []model.Booking, startAt, endAt time.Time) bool {
	for _, booking := range bookings {
		if booking.StartAt.Before(endAt) && booking.EndAt.After(startAt) {
			return true
		}
	}
	return false
}

func generateBookingNo() (string, error) {
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return fmt.Sprintf("BK-%s-%s", time.Now().Format("20060102"), strings.ToUpper(hex.EncodeToString(random))), nil
}

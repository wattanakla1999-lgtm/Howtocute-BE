package service

import (
	"context"
	"errors"
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/pkg/utils"
	"net/http"
	"sort"
	"testing"
	"time"

	"gorm.io/gorm"
)

type fakeBookingStore struct {
	users         map[uint]model.User
	services      map[uint]model.Service
	technicians   map[uint]model.NailTechnician
	bookings      map[uint]model.Booking
	nextUserID    uint
	nextBookingID uint
}

type fakeBookingNotifier struct {
	ownerCreated      []model.Booking
	ownerCancelled    []model.Booking
	customerConfirmed []model.Booking
	err               error
}

func (f *fakeBookingNotifier) NotifyOwnerBookingCreated(_ context.Context, booking model.Booking) error {
	f.ownerCreated = append(f.ownerCreated, booking)
	return f.err
}

func (f *fakeBookingNotifier) NotifyOwnerBookingCancelled(_ context.Context, booking model.Booking) error {
	f.ownerCancelled = append(f.ownerCancelled, booking)
	return f.err
}

func (f *fakeBookingNotifier) NotifyCustomerBookingConfirmed(_ context.Context, booking model.Booking) error {
	f.customerConfirmed = append(f.customerConfirmed, booking)
	return f.err
}

func newFakeBookingStore() *fakeBookingStore {
	return &fakeBookingStore{
		users: map[uint]model.User{
			1: {Model: gorm.Model{ID: 1}, Name: "Somying"},
			2: {Model: gorm.Model{ID: 2}, Name: "Mali"},
		},
		services: map[uint]model.Service{
			2: {Model: gorm.Model{ID: 2}, ServiceName: "Gel nails", ServicePrice: 300, Duration: 60},
			3: {Model: gorm.Model{ID: 3}, ServiceName: "Nail art", ServicePrice: 450, Duration: 90},
		},
		technicians: map[uint]model.NailTechnician{
			1: {Model: gorm.Model{ID: 1}, TechnicianName: "Nok"},
			2: {Model: gorm.Model{ID: 2}, TechnicianName: "Ploy"},
		},
		bookings:      make(map[uint]model.Booking),
		nextUserID:    3,
		nextBookingID: 1,
	}
}

func (f *fakeBookingStore) FindAll(filter repository.BookingFilter, pagination utils.Pagination) ([]model.Booking, int64, error) {
	bookings := make([]model.Booking, 0)
	for _, booking := range f.bookings {
		if booking.DeletedAt.Valid || filter.UserID != nil && (booking.UserID == nil || *booking.UserID != *filter.UserID) ||
			filter.ServiceID != nil && booking.ServiceID != *filter.ServiceID ||
			filter.TechnicianID != nil && (booking.TechnicianID == nil || *booking.TechnicianID != *filter.TechnicianID) ||
			filter.Status != "" && booking.Status != filter.Status ||
			filter.DateFrom != nil && booking.StartAt.Before(*filter.DateFrom) ||
			filter.DateTo != nil && booking.StartAt.After(*filter.DateTo) ||
			filter.CustomerPhone != "" && booking.CustomerPhone != filter.CustomerPhone {
			continue
		}
		bookings = append(bookings, f.hydrate(booking))
	}
	sort.Slice(bookings, func(i, j int) bool { return bookings[i].StartAt.Before(bookings[j].StartAt) })
	total := int64(len(bookings))
	start := pagination.Offset
	if start > len(bookings) {
		start = len(bookings)
	}
	end := start + pagination.Limit
	if end > len(bookings) {
		end = len(bookings)
	}
	return bookings[start:end], total, nil
}

func (f *fakeBookingStore) FindByID(id uint) (model.Booking, error) {
	booking, ok := f.bookings[id]
	if !ok || booking.DeletedAt.Valid {
		return model.Booking{}, gorm.ErrRecordNotFound
	}
	return f.hydrate(booking), nil
}

func (f *fakeBookingStore) FindUserByID(id uint) (model.User, error) {
	user, ok := f.users[id]
	if !ok {
		return model.User{}, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (f *fakeBookingStore) FindServiceByID(id uint) (model.Service, error) {
	serviceModel, ok := f.services[id]
	if !ok {
		return model.Service{}, gorm.ErrRecordNotFound
	}
	return serviceModel, nil
}

func (f *fakeBookingStore) FindTechnicianByID(id uint) (model.NailTechnician, error) {
	technician, ok := f.technicians[id]
	if !ok {
		return model.NailTechnician{}, gorm.ErrRecordNotFound
	}
	return technician, nil
}

func (f *fakeBookingStore) HasTechnicianOverlap(technicianID uint, startAt, endAt time.Time, excludeBookingID uint) (bool, error) {
	for _, booking := range f.bookings {
		if booking.ID == excludeBookingID || booking.DeletedAt.Valid || booking.TechnicianID == nil ||
			*booking.TechnicianID != technicianID || booking.Status == model.BookingStatusCancelled ||
			booking.Status == model.BookingStatusNoShow {
			continue
		}
		if booking.StartAt.Before(endAt) && booking.EndAt.After(startAt) {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeBookingStore) FindBusyBookings(startAt, endAt time.Time, technicianID *uint) ([]model.Booking, error) {
	bookings := make([]model.Booking, 0)
	for _, booking := range f.bookings {
		if booking.DeletedAt.Valid || booking.Status == model.BookingStatusCancelled || booking.Status == model.BookingStatusNoShow ||
			!booking.StartAt.Before(endAt) || !booking.EndAt.After(startAt) ||
			technicianID != nil && (booking.TechnicianID == nil || *booking.TechnicianID != *technicianID) {
			continue
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (f *fakeBookingStore) FindActiveTechnicianIDs() ([]uint, error) {
	ids := make([]uint, 0)
	for id, technician := range f.technicians {
		if technician.Active {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

func (f *fakeBookingStore) Create(booking *model.Booking, walkInCustomer *model.User) error {
	for _, current := range f.bookings {
		if current.BookingNo == booking.BookingNo {
			return errors.New("duplicate booking number")
		}
	}
	if walkInCustomer != nil {
		for _, user := range f.users {
			if user.Phone != nil && walkInCustomer.Phone != nil && *user.Phone == *walkInCustomer.Phone {
				walkInCustomer = &user
				break
			}
		}
		if walkInCustomer.ID == 0 {
			walkInCustomer.ID = f.nextUserID
			f.nextUserID++
			f.users[walkInCustomer.ID] = *walkInCustomer
		}
		booking.UserID = &walkInCustomer.ID
		booking.User = walkInCustomer
	}
	booking.ID = f.nextBookingID
	f.nextBookingID++
	f.bookings[booking.ID] = *booking
	return nil
}

func (f *fakeBookingStore) Update(booking *model.Booking) error {
	if _, ok := f.bookings[booking.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	f.bookings[booking.ID] = *booking
	return nil
}

func (f *fakeBookingStore) Delete(booking *model.Booking) error {
	booking.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	f.bookings[booking.ID] = *booking
	return nil
}

func (f *fakeBookingStore) hydrate(booking model.Booking) model.Booking {
	booking.User = nil
	if booking.UserID != nil {
		user := f.users[*booking.UserID]
		booking.User = &user
	}
	booking.Service = f.services[booking.ServiceID]
	booking.Technician = nil
	if booking.TechnicianID != nil {
		technician := f.technicians[*booking.TechnicianID]
		booking.Technician = &technician
	}
	return booking
}

func bookingServiceForTest(store *fakeBookingStore) *BookingService {
	bookingService := NewBookingService(store)
	sequence := 0
	bookingService.bookingNoFactory = func() (string, error) {
		sequence++
		return "BK-TEST-" + time.Unix(int64(sequence), 0).Format("150405"), nil
	}
	return bookingService
}

func validCreateBookingInput() CreateBookingInput {
	userID := uint(1)
	technicianID := uint(1)
	return CreateBookingInput{
		UserID: &userID, ServiceID: 2, TechnicianID: &technicianID,
		StartAt:      time.Date(2026, 7, 15, 10, 0, 0, 0, time.FixedZone("Asia/Bangkok", 7*60*60)),
		CustomerName: "Somying", CustomerPhone: "0812345678", Note: "flower pattern",
	}
}

func TestCreateBookingSuccess(t *testing.T) {
	store := newFakeBookingStore()
	booking, err := bookingServiceForTest(store).CreateBooking(validCreateBookingInput())
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}
	if booking.ID == 0 || booking.BookingNo == "" || booking.Status != model.BookingStatusPending {
		t.Fatalf("booking identifiers/status were not generated: %+v", booking)
	}
	if booking.ServiceName != "Gel nails" || booking.Price != 300 || booking.DurationMinutes != 60 {
		t.Fatalf("service snapshot is incorrect: %+v", booking)
	}
	if want := booking.StartAt.Add(60 * time.Minute); !booking.EndAt.Equal(want) {
		t.Fatalf("EndAt = %v, want %v", booking.EndAt, want)
	}
}

func TestCreateBookingNotifiesOwner(t *testing.T) {
	store := newFakeBookingStore()
	notifier := &fakeBookingNotifier{}
	bookingService := bookingServiceForTest(store)
	bookingService.notifier = notifier

	booking, err := bookingService.CreateBooking(validCreateBookingInput())
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}
	if len(notifier.ownerCreated) != 1 || notifier.ownerCreated[0].ID != booking.ID {
		t.Fatalf("ownerCreated = %+v, want booking %d", notifier.ownerCreated, booking.ID)
	}
}

func TestCreateBookingForeignKeyNotFound(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*CreateBookingInput)
		message string
	}{
		{name: "user", prepare: func(input *CreateBookingInput) { id := uint(99); input.UserID = &id }, message: "user not found"},
		{name: "service", prepare: func(input *CreateBookingInput) { input.ServiceID = 99 }, message: "service not found"},
		{name: "technician", prepare: func(input *CreateBookingInput) { id := uint(99); input.TechnicianID = &id }, message: "technician not found"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validCreateBookingInput()
			test.prepare(&input)
			_, err := bookingServiceForTest(newFakeBookingStore()).CreateBooking(input)
			assertAppError(t, err, http.StatusNotFound, test.message)
		})
	}
}

func TestCreateWalkInBookingWithoutUser(t *testing.T) {
	store := newFakeBookingStore()
	input := validCreateBookingInput()
	input.UserID = nil
	input.CustomerName = "คุณสมชาย เดินผ่านมา"
	input.CustomerPhone = "0812345678"

	booking, err := bookingServiceForTest(store).CreateBooking(input)
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}
	if booking.UserID == nil || booking.User == nil {
		t.Fatalf("walk-in user = (%v, %+v), want created customer", booking.UserID, booking.User)
	}
	if booking.User.Name != input.CustomerName || booking.User.Phone == nil || *booking.User.Phone != input.CustomerPhone {
		t.Fatalf("created customer = %+v, want booking name and phone", booking.User)
	}
	if _, ok := store.users[*booking.UserID]; !ok {
		t.Fatalf("created customer %d was not persisted", *booking.UserID)
	}
}

func TestCreateWalkInBookingReusesCustomerByPhone(t *testing.T) {
	store := newFakeBookingStore()
	phone := "0812345678"
	store.users[3] = model.User{Model: gorm.Model{ID: 3}, Name: "Existing Customer", Phone: &phone}
	store.nextUserID = 4

	input := validCreateBookingInput()
	input.UserID = nil
	input.CustomerPhone = phone
	booking, err := bookingServiceForTest(store).CreateBooking(input)
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}
	if booking.UserID == nil || *booking.UserID != 3 {
		t.Fatalf("UserID = %v, want existing customer 3", booking.UserID)
	}
	if len(store.users) != 3 {
		t.Fatalf("users count = %d, want 3", len(store.users))
	}
}

func TestUpdateBookingCanRemoveUser(t *testing.T) {
	store := newFakeBookingStore()
	input := validCreateBookingInput()
	store.bookings[1] = existingBooking(1, *input.TechnicianID, input.StartAt, model.BookingStatusPending)

	booking, err := bookingServiceForTest(store).UpdateBooking(1, UpdateBookingInput{UserIDSet: true, UserID: nil})
	if err != nil {
		t.Fatalf("UpdateBooking() error = %v", err)
	}
	if booking.UserID != nil || booking.User != nil {
		t.Fatalf("updated user = (%v, %+v), want nil", booking.UserID, booking.User)
	}
}

func TestCreateBookingRejectsInvalidTimeRange(t *testing.T) {
	input := validCreateBookingInput()
	endAt := input.StartAt.Add(-time.Minute)
	input.EndAt = &endAt
	_, err := bookingServiceForTest(newFakeBookingStore()).CreateBooking(input)
	assertAppError(t, err, http.StatusBadRequest, "startAt must be before endAt")
}

func TestCreateBookingRejectsTechnicianOverlap(t *testing.T) {
	store := newFakeBookingStore()
	input := validCreateBookingInput()
	store.bookings[1] = existingBooking(1, *input.TechnicianID, input.StartAt, model.BookingStatusConfirmed)
	store.nextBookingID = 2
	_, err := bookingServiceForTest(store).CreateBooking(input)
	assertAppError(t, err, http.StatusConflict, "ช่วงเวลานี้ทับซ้อนกับการจองอื่น")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeBookingTimeOverlap {
		t.Fatalf("error code = %q, want %q", appErr.Code, apperror.CodeBookingTimeOverlap)
	}
}

func TestCancelledAndNoShowDoNotBlockTimeSlot(t *testing.T) {
	for _, status := range []model.BookingStatus{model.BookingStatusCancelled, model.BookingStatusNoShow} {
		t.Run(string(status), func(t *testing.T) {
			store := newFakeBookingStore()
			input := validCreateBookingInput()
			store.bookings[1] = existingBooking(1, *input.TechnicianID, input.StartAt, status)
			store.nextBookingID = 2
			if _, err := bookingServiceForTest(store).CreateBooking(input); err != nil {
				t.Fatalf("CreateBooking() error = %v", err)
			}
		})
	}
}

func TestUpdateBookingStatusSuccess(t *testing.T) {
	store := newFakeBookingStore()
	input := validCreateBookingInput()
	store.bookings[1] = existingBooking(1, *input.TechnicianID, input.StartAt, model.BookingStatusPending)
	booking, err := bookingServiceForTest(store).UpdateBookingStatus(1, model.BookingStatusConfirmed, "")
	if err != nil {
		t.Fatalf("UpdateBookingStatus() error = %v", err)
	}
	if booking.Status != model.BookingStatusConfirmed {
		t.Fatalf("Status = %q, want confirmed", booking.Status)
	}
}

func TestUpdateBookingStatusConfirmedNotifiesCustomer(t *testing.T) {
	store := newFakeBookingStore()
	lineUserID := "U123"
	store.users[1] = model.User{Model: gorm.Model{ID: 1}, Name: "Somying", LineUserID: &lineUserID}
	input := validCreateBookingInput()
	store.bookings[1] = existingBooking(1, *input.TechnicianID, input.StartAt, model.BookingStatusPending)
	notifier := &fakeBookingNotifier{}
	bookingService := bookingServiceForTest(store)
	bookingService.notifier = notifier

	booking, err := bookingService.UpdateBookingStatus(1, model.BookingStatusConfirmed, "")
	if err != nil {
		t.Fatalf("UpdateBookingStatus() error = %v", err)
	}
	if len(notifier.customerConfirmed) != 1 || notifier.customerConfirmed[0].ID != booking.ID {
		t.Fatalf("customerConfirmed = %+v, want booking %d", notifier.customerConfirmed, booking.ID)
	}
}

func TestCancelCustomerBookingNotifiesOwner(t *testing.T) {
	store := newFakeBookingStore()
	input := validCreateBookingInput()
	store.bookings[1] = existingBooking(1, *input.TechnicianID, input.StartAt, model.BookingStatusPending)
	notifier := &fakeBookingNotifier{}
	bookingService := bookingServiceForTest(store)
	bookingService.notifier = notifier

	booking, err := bookingService.CancelCustomerBooking(1, 1, "ติดธุระ")
	if err != nil {
		t.Fatalf("CancelCustomerBooking() error = %v", err)
	}
	if booking.Status != model.BookingStatusCancelled || booking.CancelReason != "ติดธุระ" {
		t.Fatalf("booking = %+v, want cancelled with reason", booking)
	}
	if len(notifier.ownerCancelled) != 1 || notifier.ownerCancelled[0].ID != booking.ID {
		t.Fatalf("ownerCancelled = %+v, want booking %d", notifier.ownerCancelled, booking.ID)
	}
}

func TestCancelCustomerBookingRejectsOtherCustomer(t *testing.T) {
	store := newFakeBookingStore()
	input := validCreateBookingInput()
	store.bookings[1] = existingBooking(1, *input.TechnicianID, input.StartAt, model.BookingStatusPending)

	_, err := bookingServiceForTest(store).CancelCustomerBooking(1, 2, "")
	assertAppError(t, err, http.StatusNotFound, "booking not found")
}

func TestGetBookingsPaginationAndFilters(t *testing.T) {
	store := newFakeBookingStore()
	base := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
	for i := uint(1); i <= 8; i++ {
		booking := existingBooking(i, 1, base.Add(time.Duration(i)*time.Hour), model.BookingStatusConfirmed)
		if i > 6 {
			userID := uint(2)
			booking.UserID = &userID
		}
		if i == 6 {
			booking.Status = model.BookingStatusCancelled
		}
		store.bookings[i] = booking
	}
	userID := uint(1)
	from := base.Add(2 * time.Hour)
	to := base.Add(6 * time.Hour)
	filter := repository.BookingFilter{
		UserID: &userID, Status: model.BookingStatusConfirmed, DateFrom: &from, DateTo: &to,
	}
	bookings, total, err := bookingServiceForTest(store).GetBookings(filter, utils.Pagination{Page: 2, Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("GetBookings() error = %v", err)
	}
	if total != 4 || len(bookings) != 2 || bookings[0].ID != 4 || bookings[1].ID != 5 {
		t.Fatalf("pagination/filter result = ids %v, total %d; want [4 5], total 4", bookingIDs(bookings), total)
	}
}

func TestGetBusySlotsForTechnician(t *testing.T) {
	store := newFakeBookingStore()
	store.technicians[1] = model.NailTechnician{Model: gorm.Model{ID: 1}, TechnicianName: "Nok", Active: true}
	location := time.FixedZone("Asia/Bangkok", 7*60*60)
	date := time.Date(2026, 7, 16, 0, 0, 0, 0, location)
	store.bookings[1] = busySlotBooking(1, 1, time.Date(2026, 7, 16, 10, 30, 0, 0, location), 60, model.BookingStatusConfirmed)
	store.bookings[2] = busySlotBooking(2, 1, time.Date(2026, 7, 16, 14, 0, 0, 0, location), 60, model.BookingStatusCancelled)
	technicianID := uint(1)

	serviceID := uint(2)
	busySlots, err := bookingServiceForTest(store).GetBusySlots(date, &technicianID, &serviceID)
	if err != nil {
		t.Fatalf("GetBusySlots() error = %v", err)
	}
	if want := []string{"10:00", "11:00"}; !equalStrings(busySlots, want) {
		t.Fatalf("busySlots = %v, want %v", busySlots, want)
	}
}

func TestGetBusySlotsForAnyTechnician(t *testing.T) {
	store := newFakeBookingStore()
	store.technicians[1] = model.NailTechnician{Model: gorm.Model{ID: 1}, Active: true}
	store.technicians[2] = model.NailTechnician{Model: gorm.Model{ID: 2}, Active: true}
	location := time.FixedZone("Asia/Bangkok", 7*60*60)
	date := time.Date(2026, 7, 16, 0, 0, 0, 0, location)
	store.bookings[1] = busySlotBooking(1, 1, time.Date(2026, 7, 16, 11, 0, 0, 0, location), 60, model.BookingStatusConfirmed)
	store.bookings[2] = busySlotBooking(2, 2, time.Date(2026, 7, 16, 11, 0, 0, 0, location), 60, model.BookingStatusPending)
	store.bookings[3] = busySlotBooking(3, 1, time.Date(2026, 7, 16, 14, 0, 0, 0, location), 60, model.BookingStatusConfirmed)
	store.bookings[4] = busySlotBooking(4, 0, time.Date(2026, 7, 16, 14, 0, 0, 0, location), 60, model.BookingStatusPending)

	serviceID := uint(2)
	busySlots, err := bookingServiceForTest(store).GetBusySlots(date, nil, &serviceID)
	if err != nil {
		t.Fatalf("GetBusySlots() error = %v", err)
	}
	if want := []string{"11:00", "14:00"}; !equalStrings(busySlots, want) {
		t.Fatalf("busySlots = %v, want %v", busySlots, want)
	}
}

func TestGetBusySlotsUsesServiceDuration(t *testing.T) {
	store := newFakeBookingStore()
	store.technicians[1] = model.NailTechnician{Model: gorm.Model{ID: 1}, Active: true}
	store.services[5] = model.Service{Model: gorm.Model{ID: 5}, ServiceName: "Long service", ServicePrice: 300, Duration: 200}
	location := time.FixedZone("Asia/Bangkok", 7*60*60)
	date := time.Date(2026, 7, 16, 0, 0, 0, 0, location)
	store.bookings[1] = busySlotBooking(1, 1, time.Date(2026, 7, 16, 18, 0, 0, 0, location), 200, model.BookingStatusPending)
	technicianID, serviceID := uint(1), uint(5)

	busySlots, err := bookingServiceForTest(store).GetBusySlots(date, &technicianID, &serviceID)
	if err != nil {
		t.Fatalf("GetBusySlots() error = %v", err)
	}
	if want := []string{"15:00", "16:00", "17:00", "18:00"}; !equalStrings(busySlots, want) {
		t.Fatalf("busySlots = %v, want %v", busySlots, want)
	}
}

func busySlotBooking(id, technicianID uint, startAt time.Time, durationMinutes int, status model.BookingStatus) model.Booking {
	booking := model.Booking{
		Model: gorm.Model{ID: id}, StartAt: startAt,
		EndAt: startAt.Add(time.Duration(durationMinutes) * time.Minute), Status: status,
	}
	if technicianID != 0 {
		booking.TechnicianID = &technicianID
	}
	return booking
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func existingBooking(id, technicianID uint, startAt time.Time, status model.BookingStatus) model.Booking {
	userID := uint(1)
	return model.Booking{
		Model: gorm.Model{ID: id}, BookingNo: "BK-EXISTING", UserID: &userID, ServiceID: 2,
		TechnicianID: &technicianID, StartAt: startAt, EndAt: startAt.Add(time.Hour),
		CustomerName: "Customer", CustomerPhone: "0800000000", ServiceName: "Gel nails",
		Price: 300, DurationMinutes: 60, Status: status,
	}
}

func assertAppError(t *testing.T, err error, status int, message string) {
	t.Helper()
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %v, want AppError", err)
	}
	if appErr.Status != status || appErr.Message != message {
		t.Fatalf("AppError = (%d, %q), want (%d, %q)", appErr.Status, appErr.Message, status, message)
	}
}

func bookingIDs(bookings []model.Booking) []uint {
	ids := make([]uint, len(bookings))
	for i, booking := range bookings {
		ids[i] = booking.ID
	}
	return ids
}

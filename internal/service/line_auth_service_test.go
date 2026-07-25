package service

import (
	"context"
	"errors"
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"net/http"
	"strconv"
	"testing"
	"time"

	"gorm.io/gorm"
)

type fakeCustomerStore struct {
	usersByLineID map[string]model.User
	usersByID     map[string]model.User
	nextID        uint
}

func newFakeCustomerStore() *fakeCustomerStore {
	return &fakeCustomerStore{
		usersByLineID: make(map[string]model.User),
		usersByID:     make(map[string]model.User),
		nextID:        1,
	}
}

func (f *fakeCustomerStore) FindByID(id string) (model.User, error) {
	user, ok := f.usersByID[id]
	if !ok {
		return model.User{}, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (f *fakeCustomerStore) FindByLineUserID(lineUserID string) (model.User, error) {
	user, ok := f.usersByLineID[lineUserID]
	if !ok {
		return model.User{}, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (f *fakeCustomerStore) Create(user *model.User) error {
	user.ID = f.nextID
	f.nextID++
	f.save(*user)
	return nil
}

func (f *fakeCustomerStore) Save(user *model.User) error {
	f.save(*user)
	return nil
}

func (f *fakeCustomerStore) save(user model.User) {
	f.usersByID[itoa(user.ID)] = user
	if user.LineUserID != nil {
		f.usersByLineID[*user.LineUserID] = user
	}
}

func itoa(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

type fakeLineVerifier struct {
	claims LineTokenClaims
	err    error
}

func (f fakeLineVerifier) Verify(_ context.Context, _, _ string) (LineTokenClaims, error) {
	return f.claims, f.err
}

func TestLineAuthServiceRejectsEmptyIDToken(t *testing.T) {
	authService := NewLineAuthService(newFakeCustomerStore(), fakeLineVerifier{}, NewCustomerJWTManager("customer-secret", time.Hour), "channel-id")
	_, err := authService.Login(context.Background(), "")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Status != http.StatusBadRequest {
		t.Fatalf("Login() error = %v, want 400", err)
	}
}

func TestLineAuthServiceRejectsLineVerifyFailure(t *testing.T) {
	authService := NewLineAuthService(
		newFakeCustomerStore(),
		fakeLineVerifier{err: apperror.Unauthorized("invalid LINE token", apperror.ErrValidation)},
		NewCustomerJWTManager("customer-secret", time.Hour),
		"channel-id",
	)
	_, err := authService.Login(context.Background(), "id-token")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Status != http.StatusUnauthorized {
		t.Fatalf("Login() error = %v, want 401", err)
	}
}

func TestLineAuthServiceCreatesCustomer(t *testing.T) {
	store := newFakeCustomerStore()
	authService := NewLineAuthService(
		store,
		fakeLineVerifier{claims: LineTokenClaims{Subject: "U123", Audience: "channel-id", Name: "Mina", Picture: "https://example.com/pic.jpg"}},
		NewCustomerJWTManager("customer-secret", time.Hour),
		"channel-id",
	)

	result, err := authService.Login(context.Background(), "id-token")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Customer.ID == 0 || result.Customer.LineUserID == nil || *result.Customer.LineUserID != "U123" {
		t.Fatalf("customer = %+v", result.Customer)
	}
	if result.Customer.Name != "Mina" || result.Customer.PictureURL != "https://example.com/pic.jpg" {
		t.Fatalf("customer = %+v", result.Customer)
	}
	if result.Token == "" {
		t.Fatal("token is empty")
	}
}

func TestLineAuthServiceReusesExistingCustomer(t *testing.T) {
	store := newFakeCustomerStore()
	lineUserID := "U123"
	if err := store.Create(&model.User{Name: "Old Name", Email: "old@example.com", LineUserID: &lineUserID}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	authService := NewLineAuthService(
		store,
		fakeLineVerifier{claims: LineTokenClaims{Subject: lineUserID, Audience: "channel-id", Name: "New Name", Picture: "https://example.com/new.jpg"}},
		NewCustomerJWTManager("customer-secret", time.Hour),
		"channel-id",
	)

	result, err := authService.Login(context.Background(), "id-token")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Customer.ID != 1 || result.Customer.Name != "New Name" || result.Customer.PictureURL != "https://example.com/new.jpg" {
		t.Fatalf("customer = %+v", result.Customer)
	}
	if store.nextID != 2 {
		t.Fatalf("nextID = %d, want no new customer", store.nextID)
	}
}

func TestLineAuthServiceRejectsMismatchedAudience(t *testing.T) {
	authService := NewLineAuthService(
		newFakeCustomerStore(),
		fakeLineVerifier{claims: LineTokenClaims{Subject: "U123", Audience: "other-channel", Name: "Mina"}},
		NewCustomerJWTManager("customer-secret", time.Hour),
		"channel-id",
	)
	_, err := authService.Login(context.Background(), "id-token")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Status != http.StatusUnauthorized {
		t.Fatalf("Login() error = %v, want 401", err)
	}
}

func TestLineAuthServiceRejectsMissingSubject(t *testing.T) {
	authService := NewLineAuthService(
		newFakeCustomerStore(),
		fakeLineVerifier{claims: LineTokenClaims{Audience: "channel-id", Name: "Mina"}},
		NewCustomerJWTManager("customer-secret", time.Hour),
		"channel-id",
	)
	_, err := authService.Login(context.Background(), "id-token")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Status != http.StatusUnauthorized {
		t.Fatalf("Login() error = %v, want 401", err)
	}
}

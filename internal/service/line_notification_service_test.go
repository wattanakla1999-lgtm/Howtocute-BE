package service

import (
	"context"
	"encoding/json"
	"nailly-back-end/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestLineNotificationServicePushesOwnerCreatedFlexMessage(t *testing.T) {
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer channel-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPLinePushClient(server.Client())
	client.pushURL = server.URL
	notifier := NewLineNotificationService(client, "channel-token", "Uowner", "https://liff.line.me/2010841578-0yyF3gJw")

	if err := notifier.NotifyOwnerBookingCreated(context.Background(), notificationTestBooking()); err != nil {
		t.Fatalf("NotifyOwnerBookingCreated() error = %v", err)
	}
	if requestBody["to"] != "Uowner" {
		t.Fatalf("to = %v, want Uowner", requestBody["to"])
	}
	messages, ok := requestBody["messages"].([]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("messages = %#v", requestBody["messages"])
	}
	message := messages[0].(map[string]any)
	if message["type"] != "flex" || message["altText"] != "มีคำขอจองใหม่" {
		t.Fatalf("message = %#v", message)
	}
}

func TestLineNotificationServiceSkipsCustomerWithoutLineUserID(t *testing.T) {
	notifier := NewLineNotificationService(nil, "channel-token", "Uowner", "")
	booking := notificationTestBooking()
	booking.User = nil

	if err := notifier.NotifyCustomerBookingConfirmed(context.Background(), booking); err == nil {
		t.Fatal("NotifyCustomerBookingConfirmed() error = nil, want missing LINE user id")
	}
}

func notificationTestBooking() model.Booking {
	lineUserID := "Ucustomer"
	technicianID := uint(2)
	startAt := time.Date(2026, 7, 26, 10, 0, 0, 0, time.FixedZone("Asia/Bangkok", 7*60*60))
	return model.Booking{
		Model:           gorm.Model{ID: 10},
		BookingNo:       "BK-20260726-ABC123",
		UserID:          uintPtr(1),
		ServiceID:       1,
		TechnicianID:    &technicianID,
		StartAt:         startAt,
		EndAt:           startAt.Add(time.Hour),
		CustomerName:    "กล้า",
		CustomerPhone:   "0933587094",
		ServiceName:     "Gel nails",
		Price:           300,
		DurationMinutes: 60,
		Status:          model.BookingStatusPending,
		User:            &model.User{Model: gorm.Model{ID: 1}, Name: "กล้า", LineUserID: &lineUserID},
		Technician:      &model.NailTechnician{Model: gorm.Model{ID: technicianID}, TechnicianName: "Nok"},
	}
}

func uintPtr(value uint) *uint {
	return &value
}

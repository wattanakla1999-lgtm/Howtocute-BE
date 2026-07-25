package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"nailly-back-end/internal/model"
	"net/http"
	"strings"
	"time"
)

const linePushURL = "https://api.line.me/v2/bot/message/push"

type BookingNotifier interface {
	NotifyOwnerBookingCreated(context.Context, model.Booking) error
	NotifyOwnerBookingCancelled(context.Context, model.Booking) error
	NotifyCustomerBookingConfirmed(context.Context, model.Booking) error
}

type LinePushClient interface {
	Push(ctx context.Context, channelAccessToken, to string, messages []lineMessage) error
}

type HTTPLinePushClient struct {
	client  *http.Client
	pushURL string
}

func NewHTTPLinePushClient(client *http.Client) *HTTPLinePushClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPLinePushClient{client: client, pushURL: linePushURL}
}

func (c *HTTPLinePushClient) Push(ctx context.Context, channelAccessToken, to string, messages []lineMessage) error {
	body, err := json.Marshal(map[string]any{
		"to":       to,
		"messages": messages,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.pushURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+channelAccessToken)

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if readErr != nil {
		return readErr
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("LINE push failed status=%d response=%s", response.StatusCode, safeLogBody(responseBody))
	}
	return nil
}

type LineNotificationService struct {
	client             LinePushClient
	channelAccessToken string
	shopOwnerUserID    string
	bookingDetailsURL  string
}

func NewLineNotificationService(client LinePushClient, channelAccessToken, shopOwnerUserID, bookingDetailsURL string) *LineNotificationService {
	if client == nil {
		client = NewHTTPLinePushClient(nil)
	}
	return &LineNotificationService{
		client:             client,
		channelAccessToken: strings.TrimSpace(channelAccessToken),
		shopOwnerUserID:    strings.TrimSpace(shopOwnerUserID),
		bookingDetailsURL:  strings.TrimSpace(bookingDetailsURL),
	}
}

func (s *LineNotificationService) NotifyOwnerBookingCreated(ctx context.Context, booking model.Booking) error {
	return s.pushToOwner(ctx, "มีคำขอจองใหม่", booking, notificationStyle{
		Color:  "#06C755",
		Title:  "มีคำขอจองใหม่",
		Status: "รอแอดมินยืนยัน",
		Button: "ดูรายละเอียดการจอง",
	})
}

func (s *LineNotificationService) NotifyOwnerBookingCancelled(ctx context.Context, booking model.Booking) error {
	status := "ยกเลิกโดยลูกค้า"
	if strings.TrimSpace(booking.CancelReason) != "" {
		status = status + " (" + strings.TrimSpace(booking.CancelReason) + ")"
	}
	return s.pushToOwner(ctx, "ลูกค้ายกเลิกการจอง", booking, notificationStyle{
		Color:  "#F50016",
		Title:  "ลูกค้ายกเลิกการจอง",
		Status: status,
		Button: "ดูรายละเอียดการจอง",
	})
}

func (s *LineNotificationService) NotifyCustomerBookingConfirmed(ctx context.Context, booking model.Booking) error {
	if booking.User == nil || booking.User.LineUserID == nil || strings.TrimSpace(*booking.User.LineUserID) == "" {
		return errors.New("booking customer has no LINE user id")
	}
	return s.push(ctx, strings.TrimSpace(*booking.User.LineUserID), lineFlexMessage{
		Type:    "flex",
		AltText: "ยืนยันการจองแล้ว",
		Contents: bookingFlexBubble(booking, notificationStyle{
			Color:    "#0B74F0",
			Title:    "ยืนยันการจองแล้ว",
			Subtitle: "เจ้าของร้านได้ยืนยันรายการจองของคุณแล้ว",
			Status:   "ยืนยันแล้ว",
			Button:   "ดูประวัติการจอง",
		}, false, s.bookingDetailsURL),
	})
}

func (s *LineNotificationService) pushToOwner(ctx context.Context, altText string, booking model.Booking, style notificationStyle) error {
	if s.shopOwnerUserID == "" {
		return errors.New("LINE_SHOP_OWNER_USER_ID is not configured")
	}
	return s.push(ctx, s.shopOwnerUserID, lineFlexMessage{
		Type:     "flex",
		AltText:  altText,
		Contents: bookingFlexBubble(booking, style, true, s.bookingDetailsURL),
	})
}

func (s *LineNotificationService) push(ctx context.Context, to string, message lineFlexMessage) error {
	if s.channelAccessToken == "" {
		return errors.New("LINE_MESSAGING_CHANNEL_ACCESS_TOKEN is not configured")
	}
	if to == "" {
		return errors.New("LINE push destination is empty")
	}
	return s.client.Push(ctx, s.channelAccessToken, to, []lineMessage{message})
}

type lineMessage interface{}

type lineFlexMessage struct {
	Type     string `json:"type"`
	AltText  string `json:"altText"`
	Contents any    `json:"contents"`
}

type notificationStyle struct {
	Color    string
	Title    string
	Subtitle string
	Status   string
	Button   string
}

func bookingFlexBubble(booking model.Booking, style notificationStyle, includeCustomer bool, detailsURL string) map[string]any {
	headerContents := []any{
		flexText(style.Title, "#FFFFFF", "xxl", "bold", false),
	}
	if style.Subtitle != "" {
		headerContents = append(headerContents, flexText(style.Subtitle, "#FFFFFF", "md", "regular", true))
	} else {
		headerContents = append(headerContents, flexText("รหัสการจอง "+booking.BookingNo, "#FFFFFF", "md", "regular", true))
	}

	bodyContents := []any{
		flexText(booking.ServiceName, "#111827", "xl", "bold", true),
		map[string]any{"type": "separator", "margin": "md"},
	}

	if includeCustomer {
		bodyContents = append(bodyContents,
			bookingInfoRow("ลูกค้า", booking.CustomerName),
			bookingInfoRow("เบอร์โทร", booking.CustomerPhone),
		)
	}

	bodyContents = append(bodyContents,
		bookingInfoRow("ช่าง", technicianName(booking)),
		bookingInfoRow("วันที่", booking.StartAt.In(time.Local).Format("2006-01-02")),
		bookingInfoRow("เวลา", bookingTimeRange(booking)),
		bookingInfoRow("ยอดรวม", formatBaht(booking.Price)),
		bookingInfoRow("สถานะ", style.Status),
	)

	bubble := map[string]any{
		"type": "bubble",
		"header": map[string]any{
			"type":            "box",
			"layout":          "vertical",
			"paddingAll":      "20px",
			"contents":        headerContents,
			"backgroundColor": style.Color,
		},
		"body": map[string]any{
			"type":     "box",
			"layout":   "vertical",
			"spacing":  "md",
			"contents": bodyContents,
		},
	}

	if detailsURL != "" {
		bubble["footer"] = map[string]any{
			"type":    "box",
			"layout":  "vertical",
			"spacing": "sm",
			"contents": []any{
				map[string]any{
					"type":   "button",
					"style":  "primary",
					"color":  style.Color,
					"height": "sm",
					"action": map[string]any{
						"type":  "uri",
						"label": style.Button,
						"uri":   detailsURL,
					},
				},
			},
		}
	}

	return bubble
}

func flexText(text, color, size, weight string, wrap bool) map[string]any {
	return map[string]any{
		"type":   "text",
		"text":   text,
		"color":  color,
		"size":   size,
		"weight": weight,
		"wrap":   wrap,
	}
}

func bookingInfoRow(label, value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		value = "-"
	}
	return map[string]any{
		"type":   "box",
		"layout": "horizontal",
		"contents": []any{
			map[string]any{
				"type":  "text",
				"text":  label,
				"size":  "md",
				"color": "#6B7280",
				"flex":  2,
			},
			map[string]any{
				"type":  "text",
				"text":  value,
				"size":  "md",
				"color": "#111827",
				"align": "end",
				"wrap":  true,
				"flex":  3,
			},
		},
	}
}

func technicianName(booking model.Booking) string {
	if booking.Technician != nil && strings.TrimSpace(booking.Technician.TechnicianName) != "" {
		return booking.Technician.TechnicianName
	}
	return "เลือกอัตโนมัติ"
}

func bookingTimeRange(booking model.Booking) string {
	return booking.StartAt.In(time.Local).Format("15:04") + " - " + booking.EndAt.In(time.Local).Format("15:04")
}

func formatBaht(amount int) string {
	return fmt.Sprintf("฿%d", amount)
}

func logNotificationFailure(event string, booking model.Booking, err error) {
	if err != nil {
		log.Printf("LINE booking notification failed event=%s booking_id=%d booking_no=%s error=%v", event, booking.ID, booking.BookingNo, err)
	}
}

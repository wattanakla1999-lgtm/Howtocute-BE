package service

import (
	"fmt"
	"math"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"strconv"
	"time"
)

type DashboardStore interface {
	CountAppointments(startAt, endAt time.Time) (int64, error)
	CountCustomers() (int64, error)
	CountNewCustomers(startAt, endAt time.Time) (int64, error)
	CountServices() (int64, error)
	SumCompletedRevenue(startAt, endAt time.Time) (int64, error)
	FindAppointments(startAt, endAt time.Time, limit int) ([]model.Booking, error)
	FindPopularServices(limit int) ([]repository.PopularServiceRow, int64, error)
}

type DashboardAppointment struct {
	ID      string
	Name    string
	Service string
	Time    string
	Status  string
	ImgURL  string
}

type DashboardPopularService struct {
	Name    string
	Rate    float64
	Count   int64
	Percent int
	Color   string
}

type DashboardStats struct {
	TodayAppointments       int64
	TodayAppointmentsChange int64
	TotalCustomers          int64
	TotalCustomersChange    string
	ActiveServicesCount     int64
	TodayRevenue            int64
	Appointments            []DashboardAppointment
	PopularServices         []DashboardPopularService
}

type DashboardService struct {
	repo DashboardStore
	now  func() time.Time
}

func NewDashboardService(repo DashboardStore) *DashboardService {
	return &DashboardService{repo: repo, now: time.Now}
}

func (s *DashboardService) GetStats() (DashboardStats, error) {
	now := s.now().In(time.Local)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	todayAppointments, err := s.repo.CountAppointments(todayStart, tomorrowStart)
	if err != nil {
		return DashboardStats{}, err
	}
	yesterdayAppointments, err := s.repo.CountAppointments(yesterdayStart, todayStart)
	if err != nil {
		return DashboardStats{}, err
	}
	totalCustomers, err := s.repo.CountCustomers()
	if err != nil {
		return DashboardStats{}, err
	}
	newCustomers, err := s.repo.CountNewCustomers(todayStart, tomorrowStart)
	if err != nil {
		return DashboardStats{}, err
	}
	activeServices, err := s.repo.CountServices()
	if err != nil {
		return DashboardStats{}, err
	}
	todayRevenue, err := s.repo.SumCompletedRevenue(todayStart, tomorrowStart)
	if err != nil {
		return DashboardStats{}, err
	}
	bookings, err := s.repo.FindAppointments(todayStart, tomorrowStart, 5)
	if err != nil {
		return DashboardStats{}, err
	}
	popularRows, totalBookings, err := s.repo.FindPopularServices(4)
	if err != nil {
		return DashboardStats{}, err
	}

	return DashboardStats{
		TodayAppointments:       todayAppointments,
		TodayAppointmentsChange: todayAppointments - yesterdayAppointments,
		TotalCustomers:          totalCustomers,
		TotalCustomersChange:    newCustomerLabel(newCustomers),
		ActiveServicesCount:     activeServices,
		TodayRevenue:            todayRevenue,
		Appointments:            toDashboardAppointments(bookings),
		PopularServices:         toDashboardPopularServices(popularRows, totalBookings),
	}, nil
}

func newCustomerLabel(count int64) string {
	if count == 0 {
		return ""
	}
	return fmt.Sprintf("+%d ใหม่", count)
}

func toDashboardAppointments(bookings []model.Booking) []DashboardAppointment {
	items := make([]DashboardAppointment, 0, len(bookings))
	for _, booking := range bookings {
		var imgURL string
		if booking.User != nil {
			imgURL = booking.User.PictureURL
		}
		items = append(items, DashboardAppointment{
			ID: strconv.FormatUint(uint64(booking.ID), 10), Name: booking.CustomerName,
			Service: booking.ServiceName, Time: booking.StartAt.In(time.Local).Format("15:04"),
			Status: dashboardBookingStatus(booking.Status),
			ImgURL: imgURL,
		})
	}
	return items
}

func dashboardBookingStatus(status model.BookingStatus) string {
	switch status {
	case model.BookingStatusInService:
		return "active"
	case model.BookingStatusCompleted:
		return "done"
	case model.BookingStatusConfirmed:
		return "confirmed"
	case model.BookingStatusCancelled, model.BookingStatusNoShow:
		return "cancelled"
	default:
		return "pending"
	}
}

func toDashboardPopularServices(rows []repository.PopularServiceRow, total int64) []DashboardPopularService {
	colors := []string{
		"from-[#818CF8] to-[#FB923C]",
		"bg-[#FB923C]",
		"bg-[#a78bfa]",
		"bg-[#818CF8]",
	}
	items := make([]DashboardPopularService, 0, len(rows))
	for index, row := range rows {
		percent := 0
		if total > 0 {
			percent = int(math.Round(float64(row.Count) / float64(total) * 100))
		}
		items = append(items, DashboardPopularService{
			Name: row.Name, Rate: 0, Count: row.Count, Percent: percent, Color: colors[index%len(colors)],
		})
	}
	return items
}

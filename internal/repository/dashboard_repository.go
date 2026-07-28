package repository

import (
	"nailly-back-end/internal/model"
	"time"

	"gorm.io/gorm"
)

var dashboardExcludedStatuses = []model.BookingStatus{
	model.BookingStatusCancelled,
	model.BookingStatusNoShow,
}

type PopularServiceRow struct {
	Name  string
	Count int64
}

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) CountAppointments(startAt, endAt time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.Booking{}).
		Where("status NOT IN ?", dashboardExcludedStatuses).
		Where("start_at >= ? AND start_at < ?", startAt, endAt).
		Count(&count).Error
	return count, err
}

func (r *DashboardRepository) CountCustomers() (int64, error) {
	var count int64
	err := r.db.Model(&model.Booking{}).
		Where("customer_phone <> ''").
		Distinct("customer_phone").
		Count(&count).Error
	return count, err
}

func (r *DashboardRepository) CountNewCustomers(startAt, endAt time.Time) (int64, error) {
	firstBookings := r.db.Model(&model.Booking{}).
		Select("customer_phone").
		Where("customer_phone <> ''").
		Group("customer_phone").
		Having("MIN(created_at) >= ? AND MIN(created_at) < ?", startAt, endAt)
	var count int64
	err := r.db.Table("(?) AS first_bookings", firstBookings).Count(&count).Error
	return count, err
}

func (r *DashboardRepository) CountServices() (int64, error) {
	var count int64
	err := r.db.Model(&model.Service{}).Count(&count).Error
	return count, err
}

func (r *DashboardRepository) SumCompletedRevenue(startAt, endAt time.Time) (int64, error) {
	var revenue int64
	err := r.db.Model(&model.Booking{}).
		Select("COALESCE(SUM(price), 0)").
		Where("status = ?", model.BookingStatusCompleted).
		Where("start_at >= ? AND start_at < ?", startAt, endAt).
		Scan(&revenue).Error
	return revenue, err
}

func (r *DashboardRepository) FindAppointments(startAt, endAt time.Time, limit int) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.db.Preload("User").
		Where("status NOT IN ?", dashboardExcludedStatuses).
		Where("start_at >= ? AND start_at < ?", startAt, endAt).
		Order("start_at ASC").
		Limit(limit).
		Find(&bookings).Error
	return bookings, err
}

func (r *DashboardRepository) FindPopularServices(limit int) ([]PopularServiceRow, int64, error) {
	base := r.db.Model(&model.Booking{}).Where("status NOT IN ?", dashboardExcludedStatuses)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []PopularServiceRow
	err := base.
		Select("service_name AS name, COUNT(*) AS count").
		Where("service_name <> ''").
		Group("service_name").
		Order("count DESC, service_name ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, total, err
}

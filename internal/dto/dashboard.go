package dto

import "nailly-back-end/internal/service"

type DashboardAppointmentResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Service string `json:"service"`
	Time    string `json:"time"`
	Status  string `json:"status"`
	ImgURL  string `json:"imgUrl,omitempty"`
}

type DashboardPopularServiceResponse struct {
	Name    string  `json:"name"`
	Rate    float64 `json:"rate"`
	Count   int64   `json:"count"`
	Percent int     `json:"percent"`
	Color   string  `json:"color"`
}

type DashboardStatsResponse struct {
	TodayAppointments       int64                             `json:"todayAppointments"`
	TodayAppointmentsChange int64                             `json:"todayAppointmentsChange"`
	TotalCustomers          int64                             `json:"totalCustomers"`
	TotalCustomersChange    string                            `json:"totalCustomersChange"`
	ActiveServicesCount     int64                             `json:"activeServicesCount"`
	TodayRevenue            int64                             `json:"todayRevenue"`
	Appointments            []DashboardAppointmentResponse    `json:"appointments"`
	PopularServices         []DashboardPopularServiceResponse `json:"popularServices"`
}

func ToDashboardStatsResponse(stats service.DashboardStats) DashboardStatsResponse {
	response := DashboardStatsResponse{
		TodayAppointments:       stats.TodayAppointments,
		TodayAppointmentsChange: stats.TodayAppointmentsChange,
		TotalCustomers:          stats.TotalCustomers,
		TotalCustomersChange:    stats.TotalCustomersChange,
		ActiveServicesCount:     stats.ActiveServicesCount,
		TodayRevenue:            stats.TodayRevenue,
		Appointments:            make([]DashboardAppointmentResponse, 0, len(stats.Appointments)),
		PopularServices:         make([]DashboardPopularServiceResponse, 0, len(stats.PopularServices)),
	}
	for _, appointment := range stats.Appointments {
		response.Appointments = append(response.Appointments, DashboardAppointmentResponse(appointment))
	}
	for _, popular := range stats.PopularServices {
		response.PopularServices = append(response.PopularServices, DashboardPopularServiceResponse(popular))
	}
	return response
}

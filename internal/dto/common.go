package dto

import "time"

var thailandLocation = time.FixedZone("Asia/Bangkok", 7*60*60)

type PaginatedResponse struct {
	Data  any   `json:"data"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

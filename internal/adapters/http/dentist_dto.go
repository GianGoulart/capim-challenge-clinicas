package http

import (
	"time"

	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
)

type dentistRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

type dentistResponse struct {
	ID        string    `json:"id"`
	ClinicID  string    `json:"clinic_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toDentistResponse(d *dentistdomain.Dentist) dentistResponse {
	return dentistResponse{
		ID:        d.ID,
		ClinicID:  d.ClinicID,
		Name:      d.Name,
		Phone:     d.Phone,
		Email:     d.Email,
		IsAdmin:   d.IsAdmin,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func toDentistResponseList(dentists []*dentistdomain.Dentist) []dentistResponse {
	result := make([]dentistResponse, 0, len(dentists))
	for _, d := range dentists {
		result = append(result, toDentistResponse(d))
	}
	return result
}

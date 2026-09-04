package dto

import (
	"time"

	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
)

// DentistRequest is the request body for POST /clinics/{clinic_id}/dentists
// and PUT /dentists/{id}.
type DentistRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

// DentistResponse is the response body for all dentist endpoints.
type DentistResponse struct {
	ID        string    `json:"id"`
	ClinicID  string    `json:"clinic_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToDentistResponse converts a domain Dentist into its wire representation.
func ToDentistResponse(d *dentistdomain.Dentist) DentistResponse {
	return DentistResponse{
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

// ToDentistResponseList converts a slice of domain Dentists into their wire
// representation.
func ToDentistResponseList(dentists []*dentistdomain.Dentist) []DentistResponse {
	result := make([]DentistResponse, 0, len(dentists))
	for _, d := range dentists {
		result = append(result, ToDentistResponse(d))
	}
	return result
}

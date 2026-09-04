package dto

import (
	"time"

	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
)

// DentistRequest é o corpo da requisição para POST /clinics/{clinic_id}/dentists
// e PUT /dentists/{id}.
type DentistRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

// DentistResponse é o corpo da resposta de todos os endpoints de dentista.
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

// ToDentistResponse converte um Dentist de domínio para sua representação de wire format.
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

// ToDentistResponseList converte um slice de Dentists de domínio para sua
// representação de wire format.
func ToDentistResponseList(dentists []*dentistdomain.Dentist) []DentistResponse {
	result := make([]DentistResponse, 0, len(dentists))
	for _, d := range dentists {
		result = append(result, ToDentistResponse(d))
	}
	return result
}

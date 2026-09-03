package http

import (
	"time"

	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
)

type paymentRequest struct {
	ClinicID  string  `json:"clinic_id"`
	DentistID *string `json:"dentist_id,omitempty"`
	Cents     int64   `json:"amount_cents"`
}

type paymentResponse struct {
	ID          string    `json:"id"`
	ClinicID    string    `json:"clinic_id"`
	DentistID   *string   `json:"dentist_id,omitempty"`
	AmountCents int64     `json:"amount_cents"`
	Status      string    `json:"status"`
	PixCode     string    `json:"pix_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toPaymentResponse(p *paymentdomain.Payment) paymentResponse {
	return paymentResponse{
		ID:          p.ID,
		ClinicID:    p.ClinicID,
		DentistID:   p.DentistID,
		AmountCents: p.Amount.Cents(),
		Status:      string(p.Status),
		PixCode:     p.PixCode,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

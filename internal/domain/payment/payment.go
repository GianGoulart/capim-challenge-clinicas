package payment

import (
	"errors"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
)

var ErrInvalidTransition = errors.New("payment cannot be approved from its current status")

// Payment represents a simulated Pix charge. PixCode is populated after
// creation (see SetPixCode) once the PixProvider adapter generates it —
// this avoids a race between persisting the payment and the provider's
// asynchronous confirmation callback trying to find it.
type Payment struct {
	ID        string
	ClinicID  string
	DentistID *string
	Amount    Money
	Status    Status
	PixCode   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPayment(id, clinicID string, dentistID *string, amount Money, now time.Time) (*Payment, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if clinicID == "" {
		return nil, errors.New("clinic id is required")
	}
	return &Payment{
		ID:        id,
		ClinicID:  clinicID,
		DentistID: dentistID,
		Amount:    amount,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *Payment) SetPixCode(code string) {
	p.PixCode = code
}

func (p *Payment) Approve(now time.Time) error {
	if p.Status != StatusPending {
		return ErrInvalidTransition
	}
	p.Status = StatusApproved
	p.UpdatedAt = now
	return nil
}

package dentist

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var emailRE = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Dentist está sempre vinculado a uma clínica; IsAdmin marca o dentista
// como um dos administradores/representantes legais da clínica.
type Dentist struct {
	ID        string
	ClinicID  string
	Name      string
	Phone     string
	Email     string
	IsAdmin   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDentist(id, clinicID, name, phone, email string, isAdmin bool, now time.Time) (*Dentist, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is required")
	}
	if strings.TrimSpace(clinicID) == "" {
		return nil, errors.New("clinic id is required")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(phone) == "" {
		return nil, errors.New("phone is required")
	}
	if !emailRE.MatchString(email) {
		return nil, errors.New("invalid email")
	}
	return &Dentist{
		ID:        id,
		ClinicID:  clinicID,
		Name:      name,
		Phone:     phone,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (d *Dentist) Update(name, phone, email string, isAdmin bool, now time.Time) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(phone) == "" {
		return errors.New("phone is required")
	}
	if !emailRE.MatchString(email) {
		return errors.New("invalid email")
	}
	d.Name = name
	d.Phone = phone
	d.Email = email
	d.IsAdmin = isAdmin
	d.UpdatedAt = now
	return nil
}

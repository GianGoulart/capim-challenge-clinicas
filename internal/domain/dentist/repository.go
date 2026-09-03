package dentist

import "context"

type Repository interface {
	Save(ctx context.Context, dentist *Dentist) error
	FindByID(ctx context.Context, id string) (*Dentist, error)
	FindByClinicID(ctx context.Context, clinicID string) ([]*Dentist, error)
	Delete(ctx context.Context, id string) error
}

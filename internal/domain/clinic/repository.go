package clinic

import "context"

// Repository is the port through which the application layer persists and
// retrieves clinics. Implementations live in internal/adapters/*.
type Repository interface {
	Save(ctx context.Context, clinic *Clinic) error
	FindByID(ctx context.Context, id string) (*Clinic, error)
	FindAll(ctx context.Context) ([]*Clinic, error)
	Delete(ctx context.Context, id string) error
}

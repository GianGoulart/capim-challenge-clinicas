package clinic

import "context"

// Repository é o port através do qual a camada de aplicação persiste e
// recupera clínicas. As implementações vivem em internal/adapters/*.
type Repository interface {
	Save(ctx context.Context, clinic *Clinic) error
	FindByID(ctx context.Context, id string) (*Clinic, error)
	FindAll(ctx context.Context) ([]*Clinic, error)
	Delete(ctx context.Context, id string) error
}

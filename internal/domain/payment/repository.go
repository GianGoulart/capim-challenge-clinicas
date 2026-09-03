package payment

import "context"

type Repository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)
}

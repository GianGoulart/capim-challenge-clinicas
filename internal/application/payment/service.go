package payment

import (
	"context"
	"time"

	"github.com/google/uuid"

	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type Service struct {
	repo        paymentdomain.Repository
	clinicRepo  clinicdomain.Repository
	dentistRepo dentistdomain.Repository
	provider    paymentdomain.PixProvider
	now         func() time.Time
}

func NewService(repo paymentdomain.Repository, clinicRepo clinicdomain.Repository, dentistRepo dentistdomain.Repository, provider paymentdomain.PixProvider) *Service {
	return &Service{repo: repo, clinicRepo: clinicRepo, dentistRepo: dentistRepo, provider: provider, now: time.Now}
}

type CreateInput struct {
	ClinicID  string
	DentistID *string
	Cents     int64
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*paymentdomain.Payment, error) {
	if _, err := s.clinicRepo.FindByID(ctx, input.ClinicID); err != nil {
		return nil, err
	}
	if input.DentistID != nil {
		if _, err := s.dentistRepo.FindByID(ctx, *input.DentistID); err != nil {
			return nil, err
		}
	}
	amount, err := paymentdomain.NewMoney(input.Cents)
	if err != nil {
		return nil, apperrors.Validation("invalid amount", map[string]string{"amount": err.Error()})
	}

	p, err := paymentdomain.NewPayment(uuid.NewString(), input.ClinicID, input.DentistID, amount, s.now())
	if err != nil {
		return nil, apperrors.Validation("invalid payment data", map[string]string{"payment": err.Error()})
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}

	pixCode, err := s.provider.Simulate(p.ID, amount, s.onApproved)
	if err != nil {
		return nil, apperrors.Internal("failed to generate pix code")
	}
	p.SetPixCode(pixCode)
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// onApproved is the callback passed to the PixProvider. It is invoked
// asynchronously (from a goroutine, in the real adapter) once the
// simulated confirmation window elapses. The payment is guaranteed to
// already be persisted by the time this runs, since Create saves it
// before calling Simulate.
//
// Both failure branches below are intentionally silent: onApproved runs
// on a fire-and-forget goroutine with no request context to report back
// to, and the PixProvider port's callback signature (func(paymentID
// string), no error return) forbids surfacing an error even if we wanted
// to. Do not add panics or turn this into an error-returning callback
// without first revisiting the PixProvider contract.
func (s *Service) onApproved(paymentID string) {
	ctx := context.Background()
	p, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return
	}
	if err := p.Approve(s.now()); err != nil {
		return
	}
	_ = s.repo.Save(ctx, p)
}

func (s *Service) Get(ctx context.Context, id string) (*paymentdomain.Payment, error) {
	return s.repo.FindByID(ctx, id)
}

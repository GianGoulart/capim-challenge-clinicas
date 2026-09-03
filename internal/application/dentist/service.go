package dentist

import (
	"context"
	"time"

	"github.com/google/uuid"

	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type Service struct {
	repo       dentistdomain.Repository
	clinicRepo clinicdomain.Repository
	now        func() time.Time
}

func NewService(repo dentistdomain.Repository, clinicRepo clinicdomain.Repository) *Service {
	return &Service{repo: repo, clinicRepo: clinicRepo, now: time.Now}
}

type CreateInput struct {
	ClinicID string
	Name     string
	Phone    string
	Email    string
	IsAdmin  bool
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*dentistdomain.Dentist, error) {
	if _, err := s.clinicRepo.FindByID(ctx, input.ClinicID); err != nil {
		return nil, err
	}
	d, err := dentistdomain.NewDentist(uuid.NewString(), input.ClinicID, input.Name, input.Phone, input.Email, input.IsAdmin, s.now())
	if err != nil {
		return nil, apperrors.Validation("invalid dentist data", map[string]string{"dentist": err.Error()})
	}
	if err := s.repo.Save(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) Get(ctx context.Context, id string) (*dentistdomain.Dentist, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByClinic(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error) {
	if _, err := s.clinicRepo.FindByID(ctx, clinicID); err != nil {
		return nil, err
	}
	return s.repo.FindByClinicID(ctx, clinicID)
}

type UpdateInput struct {
	Name    string
	Phone   string
	Email   string
	IsAdmin bool
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (*dentistdomain.Dentist, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := d.Update(input.Name, input.Phone, input.Email, input.IsAdmin, s.now()); err != nil {
		return nil, apperrors.Validation("invalid dentist data", map[string]string{"dentist": err.Error()})
	}
	if err := s.repo.Save(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

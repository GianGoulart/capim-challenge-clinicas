package clinic

import (
	"context"
	"time"

	"github.com/google/uuid"

	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type Service struct {
	repo        clinicdomain.Repository
	dentistRepo dentistdomain.Repository
	now         func() time.Time
}

func NewService(repo clinicdomain.Repository, dentistRepo dentistdomain.Repository) *Service {
	return &Service{repo: repo, dentistRepo: dentistRepo, now: time.Now}
}

type CreateInput struct {
	Document      string
	CorporateName string
	TradeName     string
	BankCode      string
	Agency        string
	Account       string
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*clinicdomain.Clinic, error) {
	doc, err := clinicdomain.NewDocument(input.Document)
	if err != nil {
		return nil, apperrors.Validation("invalid document", map[string]string{"document": err.Error()})
	}
	bankAccount, err := clinicdomain.NewBankAccount(input.BankCode, input.Agency, input.Account)
	if err != nil {
		return nil, apperrors.Validation("invalid bank account", map[string]string{"bank_account": err.Error()})
	}
	c, err := clinicdomain.NewClinic(uuid.NewString(), doc, input.CorporateName, input.TradeName, bankAccount, s.now())
	if err != nil {
		return nil, apperrors.Validation("invalid clinic data", map[string]string{"clinic": err.Error()})
	}
	if err := s.repo.Save(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Get(ctx context.Context, id string) (*clinicdomain.Clinic, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*clinicdomain.Clinic, error) {
	return s.repo.FindAll(ctx)
}

type UpdateInput struct {
	CorporateName string
	TradeName     string
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (*clinicdomain.Clinic, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := c.UpdateInfo(input.CorporateName, input.TradeName, s.now()); err != nil {
		return nil, apperrors.Validation("invalid clinic data", map[string]string{"clinic": err.Error()})
	}
	if err := s.repo.Save(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

type UpdateBankAccountInput struct {
	BankCode string
	Agency   string
	Account  string
}

func (s *Service) UpdateBankAccount(ctx context.Context, id string, input UpdateBankAccountInput) (*clinicdomain.Clinic, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bankAccount, err := clinicdomain.NewBankAccount(input.BankCode, input.Agency, input.Account)
	if err != nil {
		return nil, apperrors.Validation("invalid bank account", map[string]string{"bank_account": err.Error()})
	}
	c.UpdateBankAccount(bankAccount, s.now())
	if err := s.repo.Save(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Delete remove uma clínica, a menos que ela ainda tenha dentistas vinculados a
// ela. Uma clínica com dentistas ativos não pode ser excluída: quem chamar deve
// remover (ou realocar) seus dentistas primeiro. Isso evita deixar registros de
// dentistas órfãos, que fariam referência a uma clínica inexistente.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	dentists, err := s.dentistRepo.FindByClinicID(ctx, id)
	if err != nil {
		return err
	}
	if len(dentists) > 0 {
		return apperrors.Conflict("cannot delete clinic with dentists still linked to it")
	}
	return s.repo.Delete(ctx, id)
}

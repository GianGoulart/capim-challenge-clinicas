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
		dentist, err := s.dentistRepo.FindByID(ctx, *input.DentistID)
		if err != nil {
			return nil, err
		}
		if dentist.ClinicID != input.ClinicID {
			return nil, apperrors.Validation("dentist does not belong to clinic", map[string]string{
				"dentist_id": "dentist is not associated with the given clinic_id",
			})
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

// onApproved é o callback passado para o PixProvider. Ele é invocado de forma
// assíncrona (a partir de uma goroutine, no adapter real) assim que a janela
// de confirmação simulada se esgota. O pagamento tem a garantia de já estar
// persistido no momento em que isso roda, já que Create o salva antes de
// chamar Simulate.
//
// Os dois ramos de falha abaixo são silenciosos de forma intencional:
// onApproved roda numa goroutine fire-and-forget, sem request context para
// reportar de volta, e a assinatura do callback do port PixProvider
// (func(paymentID string), sem retorno de erro) proíbe expor um erro mesmo
// que quiséssemos. Não adicione panics nem transforme isso num callback que
// retorna erro sem antes revisitar o contrato do PixProvider.
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

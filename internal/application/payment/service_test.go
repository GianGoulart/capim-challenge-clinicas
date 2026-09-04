package payment_test

import (
	"context"
	"testing"

	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClinicRepository struct{ existing map[string]bool }

func (f *fakeClinicRepository) Save(_ context.Context, _ *clinicdomain.Clinic) error { return nil }
func (f *fakeClinicRepository) FindByID(_ context.Context, id string) (*clinicdomain.Clinic, error) {
	if !f.existing[id] {
		return nil, apperrors.NotFound("clinic not found")
	}
	return &clinicdomain.Clinic{ID: id}, nil
}
func (f *fakeClinicRepository) FindAll(_ context.Context) ([]*clinicdomain.Clinic, error) {
	return nil, nil
}
func (f *fakeClinicRepository) Delete(_ context.Context, _ string) error { return nil }

// fakeDentistRepository mapeia um ID de dentista para o ID da clínica à qual
// ele pertence. Um ID de dentista ausente do mapa é tratado como não
// encontrado.
type fakeDentistRepository struct{ existing map[string]string }

func (f *fakeDentistRepository) Save(_ context.Context, _ *dentistdomain.Dentist) error { return nil }
func (f *fakeDentistRepository) FindByID(_ context.Context, id string) (*dentistdomain.Dentist, error) {
	clinicID, ok := f.existing[id]
	if !ok {
		return nil, apperrors.NotFound("dentist not found")
	}
	return &dentistdomain.Dentist{ID: id, ClinicID: clinicID}, nil
}
func (f *fakeDentistRepository) FindByClinicID(_ context.Context, _ string) ([]*dentistdomain.Dentist, error) {
	return nil, nil
}
func (f *fakeDentistRepository) Delete(_ context.Context, _ string) error { return nil }

type fakePaymentRepository struct {
	data map[string]*paymentdomain.Payment
}

func newFakePaymentRepository() *fakePaymentRepository {
	return &fakePaymentRepository{data: make(map[string]*paymentdomain.Payment)}
}
func (f *fakePaymentRepository) Save(_ context.Context, p *paymentdomain.Payment) error {
	f.data[p.ID] = p
	return nil
}
func (f *fakePaymentRepository) FindByID(_ context.Context, id string) (*paymentdomain.Payment, error) {
	p, ok := f.data[id]
	if !ok {
		return nil, apperrors.NotFound("payment not found")
	}
	return p, nil
}

// fakePixProvider invoca onApproved de forma síncrona (se autoApprove for
// true), o que é determinístico e rápido para os testes — o código real usa
// o adapter assíncrono pix.Simulator.
type fakePixProvider struct {
	autoApprove bool
	simulateErr error
	code        string
}

func (f *fakePixProvider) Simulate(paymentID string, _ paymentdomain.Money, onApproved func(paymentID string)) (string, error) {
	if f.simulateErr != nil {
		return "", f.simulateErr
	}
	if f.autoApprove {
		onApproved(paymentID)
	}
	code := f.code
	if code == "" {
		code = "FAKE-PIX-CODE"
	}
	return code, nil
}

func TestService_Create_Success(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	p, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", Cents: 1000})

	require.NoError(t, err)
	assert.Equal(t, paymentdomain.StatusPending, p.Status)
	assert.Equal(t, "FAKE-PIX-CODE", p.PixCode)
}

func TestService_Create_ClinicNotFound(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "missing", Cents: 1000})

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Create_DentistNotFound(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)
	dentistID := "missing-dentist"

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", DentistID: &dentistID, Cents: 1000})

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Create_DentistBelongsToDifferentClinic(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true, "clinic-2": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{"dentist-1": "clinic-2"}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)
	dentistID := "dentist-1"

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", DentistID: &dentistID, Cents: 1000})

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Create_DentistBelongsToSameClinic_Success(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{"dentist-1": "clinic-1"}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)
	dentistID := "dentist-1"

	p, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", DentistID: &dentistID, Cents: 1000})

	require.NoError(t, err)
	assert.Equal(t, "dentist-1", *p.DentistID)
}

func TestService_Create_InvalidAmount(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", Cents: 0})

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Create_ProviderError(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{simulateErr: assert.AnError}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", Cents: 1000})

	assert.True(t, apperrors.Is(err, apperrors.KindInternal))
}

func TestService_Create_AutoApprovedByProviderCallback(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{autoApprove: true}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	p, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", Cents: 1000})
	require.NoError(t, err)

	found, err := svc.Get(context.Background(), p.ID)
	require.NoError(t, err)
	assert.Equal(t, paymentdomain.StatusApproved, found.Status)
}

func TestService_Get_NotFound(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{}}
	dentistRepo := &fakeDentistRepository{existing: map[string]string{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	_, err := svc.Get(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

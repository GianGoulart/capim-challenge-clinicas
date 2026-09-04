package clinic_test

import (
	"context"
	"testing"

	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClinicRepository struct {
	data    map[string]*clinicdomain.Clinic
	saveErr error
	findErr error
}

func newFakeClinicRepository() *fakeClinicRepository {
	return &fakeClinicRepository{data: make(map[string]*clinicdomain.Clinic)}
}

func (f *fakeClinicRepository) Save(_ context.Context, c *clinicdomain.Clinic) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.data[c.ID] = c
	return nil
}

func (f *fakeClinicRepository) FindByID(_ context.Context, id string) (*clinicdomain.Clinic, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	c, ok := f.data[id]
	if !ok {
		return nil, apperrors.NotFound("clinic not found")
	}
	return c, nil
}

func (f *fakeClinicRepository) FindAll(_ context.Context) ([]*clinicdomain.Clinic, error) {
	result := make([]*clinicdomain.Clinic, 0, len(f.data))
	for _, c := range f.data {
		result = append(result, c)
	}
	return result, nil
}

func (f *fakeClinicRepository) Delete(_ context.Context, id string) error {
	if _, ok := f.data[id]; !ok {
		return apperrors.NotFound("clinic not found")
	}
	delete(f.data, id)
	return nil
}

// fakeDentistRepository mapeia um ID de clínica para o número de dentistas
// vinculados a ela, para exercitar a validação cross-aggregate de
// Clinic.Delete de forma isolada (nunca precisa inspecionar os campos reais
// do dentista).
type fakeDentistRepository struct {
	byClinicID map[string][]*dentistdomain.Dentist
}

func newFakeDentistRepository() *fakeDentistRepository {
	return &fakeDentistRepository{byClinicID: make(map[string][]*dentistdomain.Dentist)}
}

func (f *fakeDentistRepository) Save(_ context.Context, _ *dentistdomain.Dentist) error { return nil }
func (f *fakeDentistRepository) FindByID(_ context.Context, _ string) (*dentistdomain.Dentist, error) {
	return nil, apperrors.NotFound("dentist not found")
}
func (f *fakeDentistRepository) FindByClinicID(_ context.Context, clinicID string) ([]*dentistdomain.Dentist, error) {
	return f.byClinicID[clinicID], nil
}
func (f *fakeDentistRepository) Delete(_ context.Context, _ string) error { return nil }

func validCreateInput() clinicapp.CreateInput {
	return clinicapp.CreateInput{
		Document:      "52998224725",
		CorporateName: "Clinica Sorriso LTDA",
		TradeName:     "Clinica Sorriso",
		BankCode:      "341",
		Agency:        "1234",
		Account:       "56789-0",
	}
}

func TestService_Create_Success(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())

	c, err := svc.Create(context.Background(), validCreateInput())

	require.NoError(t, err)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "Clinica Sorriso LTDA", c.CorporateName)
	assert.Len(t, repo.data, 1)
}

func TestService_Create_InvalidDocument(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())
	input := validCreateInput()
	input.Document = "123"

	_, err := svc.Create(context.Background(), input)

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Create_InvalidBankAccount(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())
	input := validCreateInput()
	input.BankCode = ""

	_, err := svc.Create(context.Background(), input)

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Get_NotFound(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())

	_, err := svc.Get(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_List(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())
	_, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)

	all, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestService_Update_Success(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())
	created, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)

	updated, err := svc.Update(context.Background(), created.ID, clinicapp.UpdateInput{
		CorporateName: "New Corp",
		TradeName:     "New Trade",
	})

	require.NoError(t, err)
	assert.Equal(t, "New Corp", updated.CorporateName)
}

func TestService_Update_NotFound(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())

	_, err := svc.Update(context.Background(), "missing", clinicapp.UpdateInput{CorporateName: "A", TradeName: "B"})

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_UpdateBankAccount_Success(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())
	created, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)

	updated, err := svc.UpdateBankAccount(context.Background(), created.ID, clinicapp.UpdateBankAccountInput{
		BankCode: "001", Agency: "0001", Account: "111-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "001", updated.BankAccount.BankCode)
}

func TestService_UpdateBankAccount_InvalidData(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())
	created, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)

	_, err = svc.UpdateBankAccount(context.Background(), created.ID, clinicapp.UpdateBankAccountInput{})

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Delete_Success(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())
	created, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)

	err = svc.Delete(context.Background(), created.ID)

	require.NoError(t, err)
	_, err = svc.Get(context.Background(), created.ID)
	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo, newFakeDentistRepository())

	err := svc.Delete(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Delete_ConflictWhenDentistsLinked(t *testing.T) {
	repo := newFakeClinicRepository()
	dentistRepo := newFakeDentistRepository()
	svc := clinicapp.NewService(repo, dentistRepo)
	created, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)
	dentistRepo.byClinicID[created.ID] = []*dentistdomain.Dentist{{ID: "dentist-1", ClinicID: created.ID}}

	err = svc.Delete(context.Background(), created.ID)

	assert.True(t, apperrors.Is(err, apperrors.KindConflict))
	_, getErr := svc.Get(context.Background(), created.ID)
	require.NoError(t, getErr, "clinic must not have been deleted")
}

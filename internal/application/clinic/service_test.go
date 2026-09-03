package clinic_test

import (
	"context"
	"testing"

	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
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
	svc := clinicapp.NewService(repo)

	c, err := svc.Create(context.Background(), validCreateInput())

	require.NoError(t, err)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "Clinica Sorriso LTDA", c.CorporateName)
	assert.Len(t, repo.data, 1)
}

func TestService_Create_InvalidDocument(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo)
	input := validCreateInput()
	input.Document = "123"

	_, err := svc.Create(context.Background(), input)

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Create_InvalidBankAccount(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo)
	input := validCreateInput()
	input.BankCode = ""

	_, err := svc.Create(context.Background(), input)

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Get_NotFound(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo)

	_, err := svc.Get(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_List(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo)
	_, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)

	all, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestService_Update_Success(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo)
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
	svc := clinicapp.NewService(repo)

	_, err := svc.Update(context.Background(), "missing", clinicapp.UpdateInput{CorporateName: "A", TradeName: "B"})

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_UpdateBankAccount_Success(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo)
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
	svc := clinicapp.NewService(repo)
	created, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)

	_, err = svc.UpdateBankAccount(context.Background(), created.ID, clinicapp.UpdateBankAccountInput{})

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Delete_Success(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo)
	created, err := svc.Create(context.Background(), validCreateInput())
	require.NoError(t, err)

	err = svc.Delete(context.Background(), created.ID)

	require.NoError(t, err)
	_, err = svc.Get(context.Background(), created.ID)
	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := newFakeClinicRepository()
	svc := clinicapp.NewService(repo)

	err := svc.Delete(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

package dentist_test

import (
	"context"
	"testing"

	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClinicRepository struct {
	existing map[string]bool
}

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

type fakeDentistRepository struct {
	data map[string]*dentistdomain.Dentist
}

func newFakeDentistRepository() *fakeDentistRepository {
	return &fakeDentistRepository{data: make(map[string]*dentistdomain.Dentist)}
}

func (f *fakeDentistRepository) Save(_ context.Context, d *dentistdomain.Dentist) error {
	f.data[d.ID] = d
	return nil
}
func (f *fakeDentistRepository) FindByID(_ context.Context, id string) (*dentistdomain.Dentist, error) {
	d, ok := f.data[id]
	if !ok {
		return nil, apperrors.NotFound("dentist not found")
	}
	return d, nil
}
func (f *fakeDentistRepository) FindByClinicID(_ context.Context, clinicID string) ([]*dentistdomain.Dentist, error) {
	result := make([]*dentistdomain.Dentist, 0)
	for _, d := range f.data {
		if d.ClinicID == clinicID {
			result = append(result, d)
		}
	}
	return result, nil
}
func (f *fakeDentistRepository) Delete(_ context.Context, id string) error {
	if _, ok := f.data[id]; !ok {
		return apperrors.NotFound("dentist not found")
	}
	delete(f.data, id)
	return nil
}

func TestService_Create_Success(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	repo := newFakeDentistRepository()
	svc := dentistapp.NewService(repo, clinicRepo)

	d, err := svc.Create(context.Background(), dentistapp.CreateInput{
		ClinicID: "clinic-1", Name: "Dra. Ana", Phone: "+55 11 90000-0000", Email: "ana@example.com", IsAdmin: true,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, d.ID)
	assert.True(t, d.IsAdmin)
}

func TestService_Create_ClinicNotFound(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{}}
	repo := newFakeDentistRepository()
	svc := dentistapp.NewService(repo, clinicRepo)

	_, err := svc.Create(context.Background(), dentistapp.CreateInput{
		ClinicID: "missing", Name: "Dra. Ana", Phone: "+55 11 90000-0000", Email: "ana@example.com",
	})

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Create_InvalidData(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	repo := newFakeDentistRepository()
	svc := dentistapp.NewService(repo, clinicRepo)

	_, err := svc.Create(context.Background(), dentistapp.CreateInput{
		ClinicID: "clinic-1", Name: "", Phone: "+55 11 90000-0000", Email: "ana@example.com",
	})

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_ListByClinic(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	repo := newFakeDentistRepository()
	svc := dentistapp.NewService(repo, clinicRepo)
	_, err := svc.Create(context.Background(), dentistapp.CreateInput{
		ClinicID: "clinic-1", Name: "Dra. Ana", Phone: "+55 11 90000-0000", Email: "ana@example.com",
	})
	require.NoError(t, err)

	all, err := svc.ListByClinic(context.Background(), "clinic-1")

	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestService_ListByClinic_ClinicNotFound(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{}}
	repo := newFakeDentistRepository()
	svc := dentistapp.NewService(repo, clinicRepo)

	_, err := svc.ListByClinic(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Update_Success(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	repo := newFakeDentistRepository()
	svc := dentistapp.NewService(repo, clinicRepo)
	created, err := svc.Create(context.Background(), dentistapp.CreateInput{
		ClinicID: "clinic-1", Name: "Dra. Ana", Phone: "+55 11 90000-0000", Email: "ana@example.com",
	})
	require.NoError(t, err)

	updated, err := svc.Update(context.Background(), created.ID, dentistapp.UpdateInput{
		Name: "Dra. Ana Silva", Phone: "+55 11 91111-1111", Email: "ana.silva@example.com", IsAdmin: true,
	})

	require.NoError(t, err)
	assert.Equal(t, "Dra. Ana Silva", updated.Name)
}

func TestService_Delete_Success(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	repo := newFakeDentistRepository()
	svc := dentistapp.NewService(repo, clinicRepo)
	created, err := svc.Create(context.Background(), dentistapp.CreateInput{
		ClinicID: "clinic-1", Name: "Dra. Ana", Phone: "+55 11 90000-0000", Email: "ana@example.com",
	})
	require.NoError(t, err)

	err = svc.Delete(context.Background(), created.ID)

	require.NoError(t, err)
}

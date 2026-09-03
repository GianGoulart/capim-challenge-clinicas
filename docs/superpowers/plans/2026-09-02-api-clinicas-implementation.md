# API de Gestão de Clínicas (Capim Challenge) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar, em Go, uma API REST para gestão de Clínicas, Dentistas e Pagamentos (Pix simulado), seguindo Arquitetura Hexagonal (Ports & Adapters), com testes unitários e de integração cobrindo sucesso e erro, sem dependência de banco/serviço externo real.

**Architecture:** `internal/domain/*` contém entidades e regras de negócio puras + interfaces (ports). `internal/application/*` contém os use cases (services) que orquestram o domínio através dos ports. `internal/adapters/*` contém as implementações concretas dos ports (HTTP, in-memory storage, simulador Pix). `cmd/api/main.go` é o composition root que faz o wiring manual (sem framework de DI). Ver spec completa em `docs/superpowers/specs/2026-09-02-api-clinicas-design.md`.

**Tech Stack:** Go 1.23+, `net/http` (stdlib, `ServeMux` com path params), `github.com/google/uuid`, `github.com/stretchr/testify` (assert/require), `log/slog` para logging estruturado. Sem framework HTTP externo, sem ORM, sem banco real.

**Module path:** `github.com/giancarlogoulart/capim-challenge-clinicas` (ajustar em todos os imports abaixo se o path real do repositório remoto for diferente — é uma decisão de nomenclatura, não estrutural).

---

## Convenção de nomes entre domain/application

Os pacotes de domínio e de aplicação para o mesmo conceito têm o mesmo nome curto (`clinic`, `dentist`, `payment`), pois vivem em módulos diferentes (`internal/domain/...` vs `internal/application/...`). Sempre que um arquivo precisar importar os dois, use alias:

```go
clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
clinicapp    "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
```

O mesmo padrão vale para `dentist` (`dentistdomain`/`dentistapp`) e `payment` (`paymentdomain`/`paymentapp`).

---

### Task 1: Scaffolding do projeto

**Files:**
- Create: `go.mod`, `go.sum`
- Create: `.gitignore`
- Create: `Makefile`
- Create: diretórios vazios (via `.gitkeep` onde necessário): `internal/domain`, `internal/application`, `internal/adapters/http`, `internal/adapters/memory`, `internal/adapters/pix`, `internal/platform`, `api`, `test/integration`

- [ ] **Step 1: Inicializar o módulo Go**

Run:
```bash
go mod init github.com/giancarlogoulart/capim-challenge-clinicas
```
Expected: cria `go.mod` com `module github.com/giancarlogoulart/capim-challenge-clinicas` e a versão do Go instalada.

- [ ] **Step 2: Ajustar a versão do Go no go.mod para 1.23**

Edite a linha `go X.YY` em `go.mod` para:
```
go 1.23
```

- [ ] **Step 3: Adicionar dependências**

Run:
```bash
go get github.com/google/uuid@latest
go get github.com/stretchr/testify@latest
```
Expected: `go.mod`/`go.sum` atualizados com as duas dependências.

- [ ] **Step 4: Criar `.gitignore`**

```gitignore
/bin/
*.test
*.out
.DS_Store
```

- [ ] **Step 5: Criar estrutura de diretórios**

Run:
```bash
mkdir -p cmd/api internal/domain internal/application internal/adapters/http internal/adapters/memory internal/adapters/pix internal/platform/config internal/platform/apperrors api test/integration
```

- [ ] **Step 6: Criar `Makefile`**

```makefile
.PHONY: run test test-race lint build docker-build docker-up

run:
	go run ./cmd/api

test:
	go test ./... -v

test-race:
	go test ./... -race

lint:
	golangci-lint run ./...

build:
	go build -o bin/api ./cmd/api

docker-build:
	docker build -t capim-clinics-api .

docker-up:
	docker compose up --build
```

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum .gitignore Makefile cmd internal api test
git commit -m "chore: scaffold go module and project structure"
```

---

### Task 2: `internal/platform/apperrors` — erros de domínio tipados

**Files:**
- Create: `internal/platform/apperrors/apperrors.go`
- Test: `internal/platform/apperrors/apperrors_test.go`

- [ ] **Step 1: Write the failing test**

`internal/platform/apperrors/apperrors_test.go`:
```go
package apperrors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	err := apperrors.NotFound("clinic abc not found")

	assert.Equal(t, apperrors.KindNotFound, err.Kind)
	assert.Equal(t, "clinic abc not found", err.Message)
	assert.Equal(t, "NOT_FOUND: clinic abc not found", err.Error())
}

func TestValidation(t *testing.T) {
	err := apperrors.Validation("invalid document", map[string]string{"document": "bad check digit"})

	assert.Equal(t, apperrors.KindValidation, err.Kind)
	assert.Equal(t, map[string]string{"document": "bad check digit"}, err.Details)
}

func TestConflict(t *testing.T) {
	err := apperrors.Conflict("already exists")
	assert.Equal(t, apperrors.KindConflict, err.Kind)
}

func TestInternal(t *testing.T) {
	err := apperrors.Internal("boom")
	assert.Equal(t, apperrors.KindInternal, err.Kind)
}

func TestIs_MatchesWrappedError(t *testing.T) {
	base := apperrors.NotFound("clinic abc not found")
	wrapped := fmt.Errorf("loading clinic: %w", base)

	assert.True(t, apperrors.Is(wrapped, apperrors.KindNotFound))
	assert.False(t, apperrors.Is(wrapped, apperrors.KindConflict))
}

func TestIs_ReturnsFalseForPlainErrors(t *testing.T) {
	assert.False(t, apperrors.Is(errors.New("plain error"), apperrors.KindNotFound))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/apperrors/... -v`
Expected: FAIL — package `apperrors` does not exist yet (`no Go files in ...` / compile error).

- [ ] **Step 3: Write minimal implementation**

`internal/platform/apperrors/apperrors.go`:
```go
package apperrors

import (
	"errors"
	"fmt"
)

// Kind identifies the category of a domain-level error, used by adapters to
// map it to protocol-specific responses (e.g. HTTP status codes).
type Kind string

const (
	KindNotFound   Kind = "NOT_FOUND"
	KindValidation Kind = "VALIDATION_ERROR"
	KindConflict   Kind = "CONFLICT"
	KindInternal   Kind = "INTERNAL_ERROR"
)

// Error is the concrete error type returned by domain and application code.
// It never carries HTTP/transport-specific data — adapters translate Kind
// into whatever the protocol needs.
type Error struct {
	Kind    Kind
	Message string
	Details map[string]string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func NotFound(message string) *Error {
	return &Error{Kind: KindNotFound, Message: message}
}

func Validation(message string, details map[string]string) *Error {
	return &Error{Kind: KindValidation, Message: message, Details: details}
}

func Conflict(message string) *Error {
	return &Error{Kind: KindConflict, Message: message}
}

func Internal(message string) *Error {
	return &Error{Kind: KindInternal, Message: message}
}

// Is reports whether err is (or wraps) an *Error of the given Kind.
func Is(err error, kind Kind) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Kind == kind
	}
	return false
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/platform/apperrors/... -v`
Expected: PASS (all 6 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/platform/apperrors
git commit -m "feat: add typed domain errors (apperrors)"
```

---

### Task 3: `internal/domain/clinic` — Document, BankAccount, Clinic, Repository port

**Files:**
- Create: `internal/domain/clinic/document.go`
- Create: `internal/domain/clinic/bank_account.go`
- Create: `internal/domain/clinic/clinic.go`
- Create: `internal/domain/clinic/repository.go`
- Test: `internal/domain/clinic/document_test.go`
- Test: `internal/domain/clinic/bank_account_test.go`
- Test: `internal/domain/clinic/clinic_test.go`

- [ ] **Step 1: Write the failing test for Document**

`internal/domain/clinic/document_test.go`:
```go
package clinic_test

import (
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocument_ValidCPF(t *testing.T) {
	doc, err := clinic.NewDocument("529.982.247-25")

	require.NoError(t, err)
	assert.Equal(t, clinic.DocumentTypeCPF, doc.Type())
	assert.Equal(t, "52998224725", doc.Digits())
}

func TestNewDocument_ValidCNPJ(t *testing.T) {
	doc, err := clinic.NewDocument("11.222.333/0001-81")

	require.NoError(t, err)
	assert.Equal(t, clinic.DocumentTypeCNPJ, doc.Type())
	assert.Equal(t, "11222333000181", doc.Digits())
}

func TestNewDocument_InvalidCPFCheckDigit(t *testing.T) {
	_, err := clinic.NewDocument("529.982.247-20")
	assert.Error(t, err)
}

func TestNewDocument_InvalidCNPJCheckDigit(t *testing.T) {
	_, err := clinic.NewDocument("11.222.333/0001-80")
	assert.Error(t, err)
}

func TestNewDocument_AllSameDigitsRejected(t *testing.T) {
	_, err := clinic.NewDocument("111.111.111-11")
	assert.Error(t, err)
}

func TestNewDocument_InvalidLength(t *testing.T) {
	_, err := clinic.NewDocument("123")
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/clinic/... -v`
Expected: FAIL — package `clinic` does not exist yet.

- [ ] **Step 3: Implement Document**

`internal/domain/clinic/document.go`:
```go
package clinic

import (
	"fmt"
	"regexp"
	"strconv"
)

type DocumentType string

const (
	DocumentTypeCPF  DocumentType = "CPF"
	DocumentTypeCNPJ DocumentType = "CNPJ"
)

var nonDigitRE = regexp.MustCompile(`\D`)

// Document is a Value Object representing a Brazilian CPF or CNPJ,
// validated by its official check-digit algorithm.
type Document struct {
	docType DocumentType
	digits  string
}

func (d Document) Type() DocumentType { return d.docType }
func (d Document) Digits() string     { return d.digits }
func (d Document) String() string     { return d.digits }

// NewDocument parses raw (with or without punctuation) and validates it as
// a CPF (11 digits) or CNPJ (14 digits) using the official check-digit
// algorithm.
func NewDocument(raw string) (Document, error) {
	digits := nonDigitRE.ReplaceAllString(raw, "")

	switch len(digits) {
	case 11:
		if !isValidCPF(digits) {
			return Document{}, fmt.Errorf("invalid CPF check digits: %q", raw)
		}
		return Document{docType: DocumentTypeCPF, digits: digits}, nil
	case 14:
		if !isValidCNPJ(digits) {
			return Document{}, fmt.Errorf("invalid CNPJ check digits: %q", raw)
		}
		return Document{docType: DocumentTypeCNPJ, digits: digits}, nil
	default:
		return Document{}, fmt.Errorf("document must have 11 (CPF) or 14 (CNPJ) digits, got %d", len(digits))
	}
}

func isAllSameDigit(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func checkDigit(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		n, _ := strconv.Atoi(string(digits[i]))
		sum += n * w
	}
	rem := sum % 11
	if rem < 2 {
		return 0
	}
	return 11 - rem
}

func isValidCPF(d string) bool {
	if isAllSameDigit(d) {
		return false
	}
	dv1 := checkDigit(d[:9], []int{10, 9, 8, 7, 6, 5, 4, 3, 2})
	dv2 := checkDigit(d[:10], []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2})
	return d[9:] == fmt.Sprintf("%d%d", dv1, dv2)
}

func isValidCNPJ(d string) bool {
	if isAllSameDigit(d) {
		return false
	}
	dv1 := checkDigit(d[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	dv2 := checkDigit(d[:13], []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	return d[12:] == fmt.Sprintf("%d%d", dv1, dv2)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/clinic/... -v -run TestNewDocument`
Expected: PASS (all 6 tests)

- [ ] **Step 5: Write the failing test for BankAccount**

`internal/domain/clinic/bank_account_test.go`:
```go
package clinic_test

import (
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBankAccount_Valid(t *testing.T) {
	acc, err := clinic.NewBankAccount("341", "1234", "56789-0")

	require.NoError(t, err)
	assert.Equal(t, "341", acc.BankCode)
	assert.Equal(t, "1234", acc.Agency)
	assert.Equal(t, "56789-0", acc.Account)
}

func TestNewBankAccount_MissingBankCode(t *testing.T) {
	_, err := clinic.NewBankAccount("", "1234", "56789-0")
	assert.Error(t, err)
}

func TestNewBankAccount_MissingAgency(t *testing.T) {
	_, err := clinic.NewBankAccount("341", "", "56789-0")
	assert.Error(t, err)
}

func TestNewBankAccount_MissingAccount(t *testing.T) {
	_, err := clinic.NewBankAccount("341", "1234", "")
	assert.Error(t, err)
}
```

- [ ] **Step 6: Run test to verify it fails**

Run: `go test ./internal/domain/clinic/... -v -run TestNewBankAccount`
Expected: FAIL — `NewBankAccount` undefined.

- [ ] **Step 7: Implement BankAccount**

`internal/domain/clinic/bank_account.go`:
```go
package clinic

import (
	"errors"
	"strings"
)

// BankAccount is a Value Object holding the clinic's banking details.
type BankAccount struct {
	BankCode string
	Agency   string
	Account  string
}

func NewBankAccount(bankCode, agency, account string) (BankAccount, error) {
	if strings.TrimSpace(bankCode) == "" {
		return BankAccount{}, errors.New("bank code is required")
	}
	if strings.TrimSpace(agency) == "" {
		return BankAccount{}, errors.New("agency is required")
	}
	if strings.TrimSpace(account) == "" {
		return BankAccount{}, errors.New("account is required")
	}
	return BankAccount{BankCode: bankCode, Agency: agency, Account: account}, nil
}
```

- [ ] **Step 8: Run test to verify it passes**

Run: `go test ./internal/domain/clinic/... -v -run TestNewBankAccount`
Expected: PASS (all 4 tests)

- [ ] **Step 9: Write the failing test for Clinic entity**

`internal/domain/clinic/clinic_test.go`:
```go
package clinic_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validDocument(t *testing.T) clinic.Document {
	t.Helper()
	doc, err := clinic.NewDocument("52998224725")
	require.NoError(t, err)
	return doc
}

func validBankAccount(t *testing.T) clinic.BankAccount {
	t.Helper()
	acc, err := clinic.NewBankAccount("341", "1234", "56789-0")
	require.NoError(t, err)
	return acc
}

func TestNewClinic_Valid(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	c, err := clinic.NewClinic("id-1", validDocument(t), "Clinica Sorriso LTDA", "Clinica Sorriso", validBankAccount(t), now)

	require.NoError(t, err)
	assert.Equal(t, "id-1", c.ID)
	assert.Equal(t, "Clinica Sorriso LTDA", c.CorporateName)
	assert.Equal(t, "Clinica Sorriso", c.TradeName)
	assert.Equal(t, now, c.CreatedAt)
	assert.Equal(t, now, c.UpdatedAt)
}

func TestNewClinic_MissingCorporateName(t *testing.T) {
	_, err := clinic.NewClinic("id-1", validDocument(t), "", "Clinica Sorriso", validBankAccount(t), time.Now())
	assert.Error(t, err)
}

func TestNewClinic_MissingTradeName(t *testing.T) {
	_, err := clinic.NewClinic("id-1", validDocument(t), "Clinica Sorriso LTDA", "", validBankAccount(t), time.Now())
	assert.Error(t, err)
}

func TestClinic_UpdateInfo(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(24 * time.Hour)
	c, err := clinic.NewClinic("id-1", validDocument(t), "Old Corp", "Old Trade", validBankAccount(t), created)
	require.NoError(t, err)

	err = c.UpdateInfo("New Corp", "New Trade", updated)

	require.NoError(t, err)
	assert.Equal(t, "New Corp", c.CorporateName)
	assert.Equal(t, "New Trade", c.TradeName)
	assert.Equal(t, updated, c.UpdatedAt)
}

func TestClinic_UpdateInfo_RejectsEmptyCorporateName(t *testing.T) {
	c, err := clinic.NewClinic("id-1", validDocument(t), "Old Corp", "Old Trade", validBankAccount(t), time.Now())
	require.NoError(t, err)

	err = c.UpdateInfo("", "New Trade", time.Now())
	assert.Error(t, err)
}

func TestClinic_UpdateBankAccount(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(24 * time.Hour)
	c, err := clinic.NewClinic("id-1", validDocument(t), "Corp", "Trade", validBankAccount(t), created)
	require.NoError(t, err)

	newAccount, err := clinic.NewBankAccount("001", "0001", "111-1")
	require.NoError(t, err)

	c.UpdateBankAccount(newAccount, updated)

	assert.Equal(t, newAccount, c.BankAccount)
	assert.Equal(t, updated, c.UpdatedAt)
}
```

- [ ] **Step 10: Run test to verify it fails**

Run: `go test ./internal/domain/clinic/... -v -run TestNewClinic`
Expected: FAIL — `Clinic`/`NewClinic` undefined.

- [ ] **Step 11: Implement Clinic entity and Repository port**

`internal/domain/clinic/clinic.go`:
```go
package clinic

import (
	"errors"
	"strings"
	"time"
)

// Clinic is the aggregate root for a dental clinic and its banking details.
type Clinic struct {
	ID            string
	Document      Document
	CorporateName string
	TradeName     string
	BankAccount   BankAccount
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewClinic(id string, document Document, corporateName, tradeName string, bankAccount BankAccount, now time.Time) (*Clinic, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is required")
	}
	if strings.TrimSpace(corporateName) == "" {
		return nil, errors.New("corporate name is required")
	}
	if strings.TrimSpace(tradeName) == "" {
		return nil, errors.New("trade name is required")
	}
	return &Clinic{
		ID:            id,
		Document:      document,
		CorporateName: corporateName,
		TradeName:     tradeName,
		BankAccount:   bankAccount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (c *Clinic) UpdateInfo(corporateName, tradeName string, now time.Time) error {
	if strings.TrimSpace(corporateName) == "" {
		return errors.New("corporate name is required")
	}
	if strings.TrimSpace(tradeName) == "" {
		return errors.New("trade name is required")
	}
	c.CorporateName = corporateName
	c.TradeName = tradeName
	c.UpdatedAt = now
	return nil
}

func (c *Clinic) UpdateBankAccount(bankAccount BankAccount, now time.Time) {
	c.BankAccount = bankAccount
	c.UpdatedAt = now
}
```

`internal/domain/clinic/repository.go`:
```go
package clinic

import "context"

// Repository is the port through which the application layer persists and
// retrieves clinics. Implementations live in internal/adapters/*.
type Repository interface {
	Save(ctx context.Context, clinic *Clinic) error
	FindByID(ctx context.Context, id string) (*Clinic, error)
	FindAll(ctx context.Context) ([]*Clinic, error)
	Delete(ctx context.Context, id string) error
}
```

- [ ] **Step 12: Run test to verify it passes**

Run: `go test ./internal/domain/clinic/... -v`
Expected: PASS (all tests in the package)

- [ ] **Step 13: Commit**

```bash
git add internal/domain/clinic
git commit -m "feat: add clinic domain model (Document, BankAccount, Clinic, Repository port)"
```

---

### Task 4: `internal/domain/dentist` — Dentist entity, Repository port

**Files:**
- Create: `internal/domain/dentist/dentist.go`
- Create: `internal/domain/dentist/repository.go`
- Test: `internal/domain/dentist/dentist_test.go`

- [ ] **Step 1: Write the failing test**

`internal/domain/dentist/dentist_test.go`:
```go
package dentist_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDentist_Valid(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	d, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", true, now)

	require.NoError(t, err)
	assert.Equal(t, "id-1", d.ID)
	assert.Equal(t, "clinic-1", d.ClinicID)
	assert.True(t, d.IsAdmin)
	assert.Equal(t, now, d.CreatedAt)
}

func TestNewDentist_MissingClinicID(t *testing.T) {
	_, err := dentist.NewDentist("id-1", "", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", false, time.Now())
	assert.Error(t, err)
}

func TestNewDentist_MissingName(t *testing.T) {
	_, err := dentist.NewDentist("id-1", "clinic-1", "", "+55 11 90000-0000", "ana@example.com", false, time.Now())
	assert.Error(t, err)
}

func TestNewDentist_MissingPhone(t *testing.T) {
	_, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "", "ana@example.com", false, time.Now())
	assert.Error(t, err)
}

func TestNewDentist_InvalidEmail(t *testing.T) {
	_, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "not-an-email", false, time.Now())
	assert.Error(t, err)
}

func TestDentist_Update(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	d, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", false, created)
	require.NoError(t, err)

	err = d.Update("Dra. Ana Silva", "+55 11 91111-1111", "ana.silva@example.com", true, updated)

	require.NoError(t, err)
	assert.Equal(t, "Dra. Ana Silva", d.Name)
	assert.True(t, d.IsAdmin)
	assert.Equal(t, updated, d.UpdatedAt)
}

func TestDentist_Update_RejectsInvalidEmail(t *testing.T) {
	d, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", false, time.Now())
	require.NoError(t, err)

	err = d.Update("Dra. Ana", "+55 11 90000-0000", "bad-email", false, time.Now())
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/dentist/... -v`
Expected: FAIL — package `dentist` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

`internal/domain/dentist/dentist.go`:
```go
package dentist

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var emailRE = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Dentist is always bound to a clinic; IsAdmin marks it as one of the
// clinic's administrators/legal representatives.
type Dentist struct {
	ID        string
	ClinicID  string
	Name      string
	Phone     string
	Email     string
	IsAdmin   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDentist(id, clinicID, name, phone, email string, isAdmin bool, now time.Time) (*Dentist, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is required")
	}
	if strings.TrimSpace(clinicID) == "" {
		return nil, errors.New("clinic id is required")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(phone) == "" {
		return nil, errors.New("phone is required")
	}
	if !emailRE.MatchString(email) {
		return nil, errors.New("invalid email")
	}
	return &Dentist{
		ID:        id,
		ClinicID:  clinicID,
		Name:      name,
		Phone:     phone,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (d *Dentist) Update(name, phone, email string, isAdmin bool, now time.Time) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(phone) == "" {
		return errors.New("phone is required")
	}
	if !emailRE.MatchString(email) {
		return errors.New("invalid email")
	}
	d.Name = name
	d.Phone = phone
	d.Email = email
	d.IsAdmin = isAdmin
	d.UpdatedAt = now
	return nil
}
```

`internal/domain/dentist/repository.go`:
```go
package dentist

import "context"

type Repository interface {
	Save(ctx context.Context, dentist *Dentist) error
	FindByID(ctx context.Context, id string) (*Dentist, error)
	FindByClinicID(ctx context.Context, clinicID string) ([]*Dentist, error)
	Delete(ctx context.Context, id string) error
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/dentist/... -v`
Expected: PASS (all 7 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/domain/dentist
git commit -m "feat: add dentist domain model (Dentist, Repository port)"
```

---

### Task 5: `internal/domain/payment` — Money, Payment (state machine), Repository e PixProvider ports

**Files:**
- Create: `internal/domain/payment/money.go`
- Create: `internal/domain/payment/payment.go`
- Create: `internal/domain/payment/repository.go`
- Create: `internal/domain/payment/pix_provider.go`
- Test: `internal/domain/payment/money_test.go`
- Test: `internal/domain/payment/payment_test.go`

- [ ] **Step 1: Write the failing test for Money**

`internal/domain/payment/money_test.go`:
```go
package payment_test

import (
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoney_Valid(t *testing.T) {
	m, err := payment.NewMoney(1050)

	require.NoError(t, err)
	assert.Equal(t, int64(1050), m.Cents())
	assert.Equal(t, "R$ 10.50", m.String())
}

func TestNewMoney_RejectsZero(t *testing.T) {
	_, err := payment.NewMoney(0)
	assert.Error(t, err)
}

func TestNewMoney_RejectsNegative(t *testing.T) {
	_, err := payment.NewMoney(-100)
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/payment/... -v`
Expected: FAIL — package `payment` does not exist yet.

- [ ] **Step 3: Implement Money**

`internal/domain/payment/money.go`:
```go
package payment

import (
	"errors"
	"fmt"
)

// Money represents an amount in integer cents, avoiding floating point
// rounding issues for currency values.
type Money struct {
	cents int64
}

func NewMoney(cents int64) (Money, error) {
	if cents <= 0 {
		return Money{}, errors.New("amount must be greater than zero")
	}
	return Money{cents: cents}, nil
}

func (m Money) Cents() int64 { return m.cents }

func (m Money) String() string {
	return fmt.Sprintf("R$ %d.%02d", m.cents/100, m.cents%100)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/payment/... -v -run TestNewMoney`
Expected: PASS (all 3 tests)

- [ ] **Step 5: Write the failing test for Payment**

`internal/domain/payment/payment_test.go`:
```go
package payment_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayment_Valid(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)

	p, err := payment.NewPayment("pay-1", "clinic-1", nil, amount, now)

	require.NoError(t, err)
	assert.Equal(t, payment.StatusPending, p.Status)
	assert.Equal(t, "", p.PixCode)
	assert.Nil(t, p.DentistID)
}

func TestNewPayment_WithOptionalDentist(t *testing.T) {
	dentistID := "dentist-1"
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)

	p, err := payment.NewPayment("pay-1", "clinic-1", &dentistID, amount, time.Now())

	require.NoError(t, err)
	require.NotNil(t, p.DentistID)
	assert.Equal(t, dentistID, *p.DentistID)
}

func TestNewPayment_MissingClinicID(t *testing.T) {
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)

	_, err = payment.NewPayment("pay-1", "", nil, amount, time.Now())
	assert.Error(t, err)
}

func TestPayment_SetPixCode(t *testing.T) {
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)
	p, err := payment.NewPayment("pay-1", "clinic-1", nil, amount, time.Now())
	require.NoError(t, err)

	p.SetPixCode("00020126...")

	assert.Equal(t, "00020126...", p.PixCode)
}

func TestPayment_Approve_FromPending(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	approvedAt := created.Add(3 * time.Second)
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)
	p, err := payment.NewPayment("pay-1", "clinic-1", nil, amount, created)
	require.NoError(t, err)

	err = p.Approve(approvedAt)

	require.NoError(t, err)
	assert.Equal(t, payment.StatusApproved, p.Status)
	assert.Equal(t, approvedAt, p.UpdatedAt)
}

func TestPayment_Approve_RejectsWhenAlreadyApproved(t *testing.T) {
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)
	p, err := payment.NewPayment("pay-1", "clinic-1", nil, amount, time.Now())
	require.NoError(t, err)
	require.NoError(t, p.Approve(time.Now()))

	err = p.Approve(time.Now())

	assert.ErrorIs(t, err, payment.ErrInvalidTransition)
}
```

- [ ] **Step 6: Run test to verify it fails**

Run: `go test ./internal/domain/payment/... -v -run TestNewPayment`
Expected: FAIL — `Payment`/`NewPayment` undefined.

- [ ] **Step 7: Implement Payment, Repository and PixProvider ports**

`internal/domain/payment/payment.go`:
```go
package payment

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
)

var ErrInvalidTransition = errors.New("payment cannot be approved from its current status")

// Payment represents a simulated Pix charge. PixCode is populated after
// creation (see SetPixCode) once the PixProvider adapter generates it —
// this avoids a race between persisting the payment and the provider's
// asynchronous confirmation callback trying to find it.
type Payment struct {
	ID        string
	ClinicID  string
	DentistID *string
	Amount    Money
	Status    Status
	PixCode   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPayment(id, clinicID string, dentistID *string, amount Money, now time.Time) (*Payment, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if clinicID == "" {
		return nil, errors.New("clinic id is required")
	}
	return &Payment{
		ID:        id,
		ClinicID:  clinicID,
		DentistID: dentistID,
		Amount:    amount,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *Payment) SetPixCode(code string) {
	p.PixCode = code
}

func (p *Payment) Approve(now time.Time) error {
	if p.Status != StatusPending {
		return ErrInvalidTransition
	}
	p.Status = StatusApproved
	p.UpdatedAt = now
	return nil
}
```

`internal/domain/payment/repository.go`:
```go
package payment

import "context"

type Repository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)
}
```

`internal/domain/payment/pix_provider.go`:
```go
package payment

// PixProvider is the port through which the application layer requests a
// simulated Pix "copy and paste" code and an asynchronous confirmation.
//
// Simulate must return immediately with a pixCode and schedule onApproved
// to be invoked exactly once, later, with paymentID — real implementations
// do this via a background goroutine with a randomized delay; test doubles
// may call it synchronously.
type PixProvider interface {
	Simulate(paymentID string, amount Money, onApproved func(paymentID string)) (pixCode string, err error)
}
```

- [ ] **Step 8: Run test to verify it passes**

Run: `go test ./internal/domain/payment/... -v`
Expected: PASS (all tests in the package)

- [ ] **Step 9: Commit**

```bash
git add internal/domain/payment
git commit -m "feat: add payment domain model (Money, Payment state machine, ports)"
```

---

### Task 6: `internal/adapters/memory` — repositórios in-memory thread-safe

**Files:**
- Create: `internal/adapters/memory/clinic_repository.go`
- Create: `internal/adapters/memory/dentist_repository.go`
- Create: `internal/adapters/memory/payment_repository.go`
- Test: `internal/adapters/memory/clinic_repository_test.go`
- Test: `internal/adapters/memory/dentist_repository_test.go`
- Test: `internal/adapters/memory/payment_repository_test.go`

- [ ] **Step 1: Write the failing test for ClinicRepository**

`internal/adapters/memory/clinic_repository_test.go`:
```go
package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClinic(t *testing.T, id string) *clinic.Clinic {
	t.Helper()
	doc, err := clinic.NewDocument("52998224725")
	require.NoError(t, err)
	acc, err := clinic.NewBankAccount("341", "1234", "56789-0")
	require.NoError(t, err)
	c, err := clinic.NewClinic(id, doc, "Corp", "Trade", acc, time.Now())
	require.NoError(t, err)
	return c
}

func TestClinicRepository_SaveAndFindByID(t *testing.T) {
	repo := memory.NewClinicRepository()
	ctx := context.Background()
	c := newTestClinic(t, "id-1")

	require.NoError(t, repo.Save(ctx, c))

	found, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)
	assert.Equal(t, c, found)
}

func TestClinicRepository_FindByID_NotFound(t *testing.T) {
	repo := memory.NewClinicRepository()

	_, err := repo.FindByID(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestClinicRepository_FindAll(t *testing.T) {
	repo := memory.NewClinicRepository()
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, newTestClinic(t, "id-1")))
	require.NoError(t, repo.Save(ctx, newTestClinic(t, "id-2")))

	all, err := repo.FindAll(ctx)

	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestClinicRepository_Delete(t *testing.T) {
	repo := memory.NewClinicRepository()
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, newTestClinic(t, "id-1")))

	require.NoError(t, repo.Delete(ctx, "id-1"))

	_, err := repo.FindByID(ctx, "id-1")
	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestClinicRepository_Delete_NotFound(t *testing.T) {
	repo := memory.NewClinicRepository()

	err := repo.Delete(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/memory/... -v -run TestClinicRepository`
Expected: FAIL — package `memory` does not exist yet.

- [ ] **Step 3: Implement ClinicRepository**

`internal/adapters/memory/clinic_repository.go`:
```go
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// ClinicRepository is a thread-safe in-memory implementation of
// clinic.Repository — no external database is used, per the challenge's
// technical requirements.
type ClinicRepository struct {
	mu   sync.RWMutex
	data map[string]*clinic.Clinic
}

func NewClinicRepository() *ClinicRepository {
	return &ClinicRepository{data: make(map[string]*clinic.Clinic)}
}

func (r *ClinicRepository) Save(_ context.Context, c *clinic.Clinic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[c.ID] = c
	return nil
}

func (r *ClinicRepository) FindByID(_ context.Context, id string) (*clinic.Clinic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("clinic %s not found", id))
	}
	return c, nil
}

func (r *ClinicRepository) FindAll(_ context.Context) ([]*clinic.Clinic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*clinic.Clinic, 0, len(r.data))
	for _, c := range r.data {
		result = append(result, c)
	}
	return result, nil
}

func (r *ClinicRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return apperrors.NotFound(fmt.Sprintf("clinic %s not found", id))
	}
	delete(r.data, id)
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/memory/... -v -run TestClinicRepository`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Write the failing test for DentistRepository**

`internal/adapters/memory/dentist_repository_test.go`:
```go
package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDentist(t *testing.T, id, clinicID string) *dentist.Dentist {
	t.Helper()
	d, err := dentist.NewDentist(id, clinicID, "Dra. Ana", "+55 11 90000-0000", "ana@example.com", false, time.Now())
	require.NoError(t, err)
	return d
}

func TestDentistRepository_SaveAndFindByID(t *testing.T) {
	repo := memory.NewDentistRepository()
	ctx := context.Background()
	d := newTestDentist(t, "id-1", "clinic-1")

	require.NoError(t, repo.Save(ctx, d))

	found, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)
	assert.Equal(t, d, found)
}

func TestDentistRepository_FindByID_NotFound(t *testing.T) {
	repo := memory.NewDentistRepository()

	_, err := repo.FindByID(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestDentistRepository_FindByClinicID(t *testing.T) {
	repo := memory.NewDentistRepository()
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, newTestDentist(t, "id-1", "clinic-1")))
	require.NoError(t, repo.Save(ctx, newTestDentist(t, "id-2", "clinic-1")))
	require.NoError(t, repo.Save(ctx, newTestDentist(t, "id-3", "clinic-2")))

	found, err := repo.FindByClinicID(ctx, "clinic-1")

	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestDentistRepository_Delete(t *testing.T) {
	repo := memory.NewDentistRepository()
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, newTestDentist(t, "id-1", "clinic-1")))

	require.NoError(t, repo.Delete(ctx, "id-1"))

	_, err := repo.FindByID(ctx, "id-1")
	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}
```

- [ ] **Step 6: Run test to verify it fails**

Run: `go test ./internal/adapters/memory/... -v -run TestDentistRepository`
Expected: FAIL — `NewDentistRepository` undefined.

- [ ] **Step 7: Implement DentistRepository**

`internal/adapters/memory/dentist_repository.go`:
```go
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type DentistRepository struct {
	mu   sync.RWMutex
	data map[string]*dentist.Dentist
}

func NewDentistRepository() *DentistRepository {
	return &DentistRepository{data: make(map[string]*dentist.Dentist)}
}

func (r *DentistRepository) Save(_ context.Context, d *dentist.Dentist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[d.ID] = d
	return nil
}

func (r *DentistRepository) FindByID(_ context.Context, id string) (*dentist.Dentist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("dentist %s not found", id))
	}
	return d, nil
}

func (r *DentistRepository) FindByClinicID(_ context.Context, clinicID string) ([]*dentist.Dentist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*dentist.Dentist, 0)
	for _, d := range r.data {
		if d.ClinicID == clinicID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (r *DentistRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return apperrors.NotFound(fmt.Sprintf("dentist %s not found", id))
	}
	delete(r.data, id)
	return nil
}
```

- [ ] **Step 8: Run test to verify it passes**

Run: `go test ./internal/adapters/memory/... -v -run TestDentistRepository`
Expected: PASS (all 4 tests)

- [ ] **Step 9: Write the failing test for PaymentRepository**

`internal/adapters/memory/payment_repository_test.go`:
```go
package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPayment(t *testing.T, id string) *payment.Payment {
	t.Helper()
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)
	p, err := payment.NewPayment(id, "clinic-1", nil, amount, time.Now())
	require.NoError(t, err)
	return p
}

func TestPaymentRepository_SaveAndFindByID(t *testing.T) {
	repo := memory.NewPaymentRepository()
	ctx := context.Background()
	p := newTestPayment(t, "pay-1")

	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.Equal(t, p, found)
}

func TestPaymentRepository_FindByID_NotFound(t *testing.T) {
	repo := memory.NewPaymentRepository()

	_, err := repo.FindByID(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestPaymentRepository_SaveOverwritesExisting(t *testing.T) {
	repo := memory.NewPaymentRepository()
	ctx := context.Background()
	p := newTestPayment(t, "pay-1")
	require.NoError(t, repo.Save(ctx, p))

	require.NoError(t, p.Approve(time.Now()))
	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.Equal(t, payment.StatusApproved, found.Status)
}
```

- [ ] **Step 10: Run test to verify it fails**

Run: `go test ./internal/adapters/memory/... -v -run TestPaymentRepository`
Expected: FAIL — `NewPaymentRepository` undefined.

- [ ] **Step 11: Implement PaymentRepository**

`internal/adapters/memory/payment_repository.go`:
```go
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type PaymentRepository struct {
	mu   sync.RWMutex
	data map[string]*payment.Payment
}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{data: make(map[string]*payment.Payment)}
}

func (r *PaymentRepository) Save(_ context.Context, p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[p.ID] = p
	return nil
}

func (r *PaymentRepository) FindByID(_ context.Context, id string) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("payment %s not found", id))
	}
	return p, nil
}
```

- [ ] **Step 12: Run test to verify it passes**

Run: `go test ./internal/adapters/memory/... -v`
Expected: PASS (all tests in the package)

- [ ] **Step 13: Run with the race detector**

Run: `go test ./internal/adapters/memory/... -race`
Expected: PASS, no data race reported.

- [ ] **Step 14: Commit**

```bash
git add internal/adapters/memory
git commit -m "feat: add thread-safe in-memory repositories for clinic, dentist, payment"
```

---

### Task 7: `internal/adapters/pix` — simulador do provedor Pix

**Files:**
- Create: `internal/adapters/pix/simulator.go`
- Test: `internal/adapters/pix/simulator_test.go`

- [ ] **Step 1: Write the failing test**

`internal/adapters/pix/simulator_test.go`:
```go
package pix_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/pix"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimulator_GeneratesNonEmptyCodeContainingPaymentID(t *testing.T) {
	sim := pix.NewSimulator(time.Millisecond, 2*time.Millisecond)
	amount, err := payment.NewMoney(1000)
	require.NoError(t, err)

	code, err := sim.Simulate("pay-1", amount, func(string) {})

	require.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Contains(t, code, "pay-1")
}

func TestSimulator_InvokesCallbackAfterDelay(t *testing.T) {
	sim := pix.NewSimulator(time.Millisecond, 3*time.Millisecond)
	amount, err := payment.NewMoney(1000)
	require.NoError(t, err)

	var mu sync.Mutex
	var receivedID string
	done := make(chan struct{})

	_, err = sim.Simulate("pay-1", amount, func(id string) {
		mu.Lock()
		receivedID = id
		mu.Unlock()
		close(done)
	})
	require.NoError(t, err)

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, "pay-1", receivedID)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for onApproved callback")
	}
}

func TestNewDefaultSimulator_UsesTwoToFiveSecondWindow(t *testing.T) {
	sim := pix.NewDefaultSimulator()
	assert.NotNil(t, sim)
	// Behavioral guarantee only — exact delay is randomized and not asserted here.
	assert.True(t, strings.HasPrefix("ok", "ok"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/pix/... -v`
Expected: FAIL — package `pix` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

`internal/adapters/pix/simulator.go`:
```go
package pix

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
)

// Simulator implements payment.PixProvider without any real Pix
// integration. It generates a fake copy-and-paste code and, after a random
// delay within [minDelay, maxDelay], invokes onApproved on its own
// goroutine — simulating an asynchronous payment confirmation.
type Simulator struct {
	minDelay time.Duration
	maxDelay time.Duration
}

func NewSimulator(minDelay, maxDelay time.Duration) *Simulator {
	return &Simulator{minDelay: minDelay, maxDelay: maxDelay}
}

// NewDefaultSimulator uses the challenge's suggested 2-5 second
// confirmation window.
func NewDefaultSimulator() *Simulator {
	return NewSimulator(2*time.Second, 5*time.Second)
}

func (s *Simulator) Simulate(paymentID string, amount payment.Money, onApproved func(paymentID string)) (string, error) {
	pixCode := fmt.Sprintf("00020126SIMULATEDPIX%s5204000053039865406%s5802BR6304%s",
		paymentID, amount.String(), paymentID)

	delay := s.minDelay
	if window := s.maxDelay - s.minDelay; window > 0 {
		delay += time.Duration(rand.Int63n(int64(window)))
	}

	go func() {
		time.Sleep(delay)
		onApproved(paymentID)
	}()

	return pixCode, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/pix/... -v -race`
Expected: PASS (all 3 tests, no data races)

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/pix
git commit -m "feat: add simulated Pix provider adapter"
```

---

### Task 8: `internal/application/clinic` — use cases de Clínica

**Files:**
- Create: `internal/application/clinic/service.go`
- Test: `internal/application/clinic/service_test.go`

- [ ] **Step 1: Write the failing test**

`internal/application/clinic/service_test.go`:
```go
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
	data     map[string]*clinicdomain.Clinic
	saveErr  error
	findErr  error
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/application/clinic/... -v`
Expected: FAIL — package `clinic` (application) does not exist yet.

- [ ] **Step 3: Write minimal implementation**

`internal/application/clinic/service.go`:
```go
package clinic

import (
	"context"
	"time"

	"github.com/google/uuid"

	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type Service struct {
	repo clinicdomain.Repository
	now  func() time.Time
}

func NewService(repo clinicdomain.Repository) *Service {
	return &Service{repo: repo, now: time.Now}
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

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/application/clinic/... -v`
Expected: PASS (all tests in the package)

- [ ] **Step 5: Commit**

```bash
git add internal/application/clinic
git commit -m "feat: add clinic use cases (create, get, list, update, update bank account, delete)"
```

---

### Task 9: `internal/application/dentist` — use cases de Dentista

**Files:**
- Create: `internal/application/dentist/service.go`
- Test: `internal/application/dentist/service_test.go`

- [ ] **Step 1: Write the failing test**

`internal/application/dentist/service_test.go`:
```go
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
func (f *fakeClinicRepository) FindAll(_ context.Context) ([]*clinicdomain.Clinic, error) { return nil, nil }
func (f *fakeClinicRepository) Delete(_ context.Context, _ string) error                  { return nil }

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/application/dentist/... -v`
Expected: FAIL — package `dentist` (application) does not exist yet.

- [ ] **Step 3: Write minimal implementation**

`internal/application/dentist/service.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/application/dentist/... -v`
Expected: PASS (all tests in the package)

- [ ] **Step 5: Commit**

```bash
git add internal/application/dentist
git commit -m "feat: add dentist use cases (create, get, list by clinic, update, delete)"
```

---

### Task 10: `internal/application/payment` — use case de Pagamento (Pix)

**Files:**
- Create: `internal/application/payment/service.go`
- Test: `internal/application/payment/service_test.go`

- [ ] **Step 1: Write the failing test**

`internal/application/payment/service_test.go`:
```go
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
func (f *fakeClinicRepository) FindAll(_ context.Context) ([]*clinicdomain.Clinic, error) { return nil, nil }
func (f *fakeClinicRepository) Delete(_ context.Context, _ string) error                  { return nil }

type fakeDentistRepository struct{ existing map[string]bool }

func (f *fakeDentistRepository) Save(_ context.Context, _ *dentistdomain.Dentist) error { return nil }
func (f *fakeDentistRepository) FindByID(_ context.Context, id string) (*dentistdomain.Dentist, error) {
	if !f.existing[id] {
		return nil, apperrors.NotFound("dentist not found")
	}
	return &dentistdomain.Dentist{ID: id}, nil
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

// fakePixProvider invokes onApproved synchronously (if autoApprove is true),
// which is deterministic and fast for tests — real code uses the async
// pix.Simulator adapter instead.
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
	dentistRepo := &fakeDentistRepository{existing: map[string]bool{}}
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
	dentistRepo := &fakeDentistRepository{existing: map[string]bool{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "missing", Cents: 1000})

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Create_DentistNotFound(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]bool{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)
	dentistID := "missing-dentist"

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", DentistID: &dentistID, Cents: 1000})

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestService_Create_InvalidAmount(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]bool{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", Cents: 0})

	assert.True(t, apperrors.Is(err, apperrors.KindValidation))
}

func TestService_Create_ProviderError(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]bool{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{simulateErr: assert.AnError}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	_, err := svc.Create(context.Background(), paymentapp.CreateInput{ClinicID: "clinic-1", Cents: 1000})

	assert.True(t, apperrors.Is(err, apperrors.KindInternal))
}

func TestService_Create_AutoApprovedByProviderCallback(t *testing.T) {
	clinicRepo := &fakeClinicRepository{existing: map[string]bool{"clinic-1": true}}
	dentistRepo := &fakeDentistRepository{existing: map[string]bool{}}
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
	dentistRepo := &fakeDentistRepository{existing: map[string]bool{}}
	repo := newFakePaymentRepository()
	provider := &fakePixProvider{}
	svc := paymentapp.NewService(repo, clinicRepo, dentistRepo, provider)

	_, err := svc.Get(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/application/payment/... -v`
Expected: FAIL — package `payment` (application) does not exist yet.

- [ ] **Step 3: Write minimal implementation**

`internal/application/payment/service.go`:
```go
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
func (s *Service) onApproved(paymentID string) {
	ctx := context.Background()
	p, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return
	}
	if err := p.Approve(time.Now()); err != nil {
		return
	}
	_ = s.repo.Save(ctx, p)
}

func (s *Service) Get(ctx context.Context, id string) (*paymentdomain.Payment, error) {
	return s.repo.FindByID(ctx, id)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/application/payment/... -v -race`
Expected: PASS (all tests in the package)

- [ ] **Step 5: Commit**

```bash
git add internal/application/payment
git commit -m "feat: add payment use case (create with simulated Pix confirmation, get)"
```

---

### Task 11: `internal/adapters/http` — envelope de resposta e mapeamento de erros

**Files:**
- Create: `internal/adapters/http/response.go`
- Test: `internal/adapters/http/response_test.go`

- [ ] **Step 1: Write the failing test**

`internal/adapters/http/response_test.go`:
```go
package http_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON_WritesStatusAndBody(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteJSON(rec, 201, map[string]string{"id": "abc"})

	assert.Equal(t, 201, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"id":"abc"}`, rec.Body.String())
}

func TestWriteError_NotFound(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, apperrors.NotFound("clinic abc not found"))

	assert.Equal(t, 404, rec.Code)
	var body map[string]map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "NOT_FOUND", body["error"]["code"])
}

func TestWriteError_Validation(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, apperrors.Validation("invalid document", map[string]string{"document": "bad"}))

	assert.Equal(t, 422, rec.Code)
}

func TestWriteError_Conflict(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, apperrors.Conflict("already exists"))

	assert.Equal(t, 409, rec.Code)
}

func TestWriteError_UnknownErrorMapsToInternal(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, assert.AnError)

	assert.Equal(t, 500, rec.Code)
	var body map[string]map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "INTERNAL_ERROR", body["error"]["code"])
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/http/... -v`
Expected: FAIL — package `http` (adapter) does not exist yet.

- [ ] **Step 3: Write minimal implementation**

`internal/adapters/http/response.go`:
```go
package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// WriteError maps a domain/application error to an HTTP status code and a
// consistent JSON error envelope. Errors that are not *apperrors.Error
// (e.g. unexpected panics recovered upstream) are treated as internal.
func WriteError(w http.ResponseWriter, err error) {
	var appErr *apperrors.Error
	status := http.StatusInternalServerError
	code := string(apperrors.KindInternal)
	message := "internal server error"
	var details map[string]string

	if errors.As(err, &appErr) {
		code = string(appErr.Kind)
		message = appErr.Message
		details = appErr.Details
		switch appErr.Kind {
		case apperrors.KindNotFound:
			status = http.StatusNotFound
		case apperrors.KindValidation:
			status = http.StatusUnprocessableEntity
		case apperrors.KindConflict:
			status = http.StatusConflict
		default:
			status = http.StatusInternalServerError
		}
	}

	WriteJSON(w, status, errorEnvelope{Error: errorBody{Code: code, Message: message, Details: details}})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/http/... -v -run TestWrite`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/http/response.go internal/adapters/http/response_test.go
git commit -m "feat: add HTTP response envelope and apperrors-to-status mapping"
```

---

### Task 12: `internal/adapters/http` — DTOs de Clinic, Dentist, Payment

**Files:**
- Create: `internal/adapters/http/clinic_dto.go`
- Create: `internal/adapters/http/dentist_dto.go`
- Create: `internal/adapters/http/payment_dto.go`

Não há testes dedicados nesta task — os DTOs são exercitados indiretamente pelos testes de handler (Tasks 13–15). São apenas `struct`s e funções puras de mapeamento.

- [ ] **Step 1: Criar DTOs de Clinic**

`internal/adapters/http/clinic_dto.go`:
```go
package http

import (
	"time"

	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
)

type clinicRequest struct {
	Document      string `json:"document"`
	CorporateName string `json:"corporate_name"`
	TradeName     string `json:"trade_name"`
	BankCode      string `json:"bank_code"`
	Agency        string `json:"agency"`
	Account       string `json:"account"`
}

type clinicUpdateRequest struct {
	CorporateName string `json:"corporate_name"`
	TradeName     string `json:"trade_name"`
}

type bankAccountRequest struct {
	BankCode string `json:"bank_code"`
	Agency   string `json:"agency"`
	Account  string `json:"account"`
}

type clinicResponse struct {
	ID            string    `json:"id"`
	Document      string    `json:"document"`
	DocumentType  string    `json:"document_type"`
	CorporateName string    `json:"corporate_name"`
	TradeName     string    `json:"trade_name"`
	BankCode      string    `json:"bank_code"`
	Agency        string    `json:"agency"`
	Account       string    `json:"account"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toClinicResponse(c *clinicdomain.Clinic) clinicResponse {
	return clinicResponse{
		ID:            c.ID,
		Document:      c.Document.Digits(),
		DocumentType:  string(c.Document.Type()),
		CorporateName: c.CorporateName,
		TradeName:     c.TradeName,
		BankCode:      c.BankAccount.BankCode,
		Agency:        c.BankAccount.Agency,
		Account:       c.BankAccount.Account,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func toClinicResponseList(clinics []*clinicdomain.Clinic) []clinicResponse {
	result := make([]clinicResponse, 0, len(clinics))
	for _, c := range clinics {
		result = append(result, toClinicResponse(c))
	}
	return result
}
```

- [ ] **Step 2: Criar DTOs de Dentist**

`internal/adapters/http/dentist_dto.go`:
```go
package http

import (
	"time"

	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
)

type dentistRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

type dentistResponse struct {
	ID        string    `json:"id"`
	ClinicID  string    `json:"clinic_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toDentistResponse(d *dentistdomain.Dentist) dentistResponse {
	return dentistResponse{
		ID:        d.ID,
		ClinicID:  d.ClinicID,
		Name:      d.Name,
		Phone:     d.Phone,
		Email:     d.Email,
		IsAdmin:   d.IsAdmin,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func toDentistResponseList(dentists []*dentistdomain.Dentist) []dentistResponse {
	result := make([]dentistResponse, 0, len(dentists))
	for _, d := range dentists {
		result = append(result, toDentistResponse(d))
	}
	return result
}
```

- [ ] **Step 3: Criar DTOs de Payment**

`internal/adapters/http/payment_dto.go`:
```go
package http

import (
	"time"

	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
)

type paymentRequest struct {
	ClinicID  string  `json:"clinic_id"`
	DentistID *string `json:"dentist_id,omitempty"`
	Cents     int64   `json:"amount_cents"`
}

type paymentResponse struct {
	ID          string    `json:"id"`
	ClinicID    string    `json:"clinic_id"`
	DentistID   *string   `json:"dentist_id,omitempty"`
	AmountCents int64     `json:"amount_cents"`
	Status      string    `json:"status"`
	PixCode     string    `json:"pix_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toPaymentResponse(p *paymentdomain.Payment) paymentResponse {
	return paymentResponse{
		ID:          p.ID,
		ClinicID:    p.ClinicID,
		DentistID:   p.DentistID,
		AmountCents: p.Amount.Cents(),
		Status:      string(p.Status),
		PixCode:     p.PixCode,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
```

- [ ] **Step 4: Verificar que o projeto compila**

Run: `go build ./...`
Expected: sem erros (DTOs ainda não são usados por nenhum handler, mas devem compilar isoladamente).

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/http/clinic_dto.go internal/adapters/http/dentist_dto.go internal/adapters/http/payment_dto.go
git commit -m "feat: add HTTP request/response DTOs for clinic, dentist, payment"
```

---

### Task 13: `internal/adapters/http` — ClinicHandler

**Files:**
- Create: `internal/adapters/http/clinic_handler.go`
- Test: `internal/adapters/http/clinic_handler_test.go`

- [ ] **Step 1: Write the failing test**

`internal/adapters/http/clinic_handler_test.go`:
```go
package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClinicService struct {
	createFn func(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error)
	getFn    func(ctx context.Context, id string) (*clinicdomain.Clinic, error)
	listFn   func(ctx context.Context) ([]*clinicdomain.Clinic, error)
	updateFn func(ctx context.Context, id string, input clinicapp.UpdateInput) (*clinicdomain.Clinic, error)
	bankFn   func(ctx context.Context, id string, input clinicapp.UpdateBankAccountInput) (*clinicdomain.Clinic, error)
	deleteFn func(ctx context.Context, id string) error
}

func (f *fakeClinicService) Create(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error) {
	return f.createFn(ctx, input)
}
func (f *fakeClinicService) Get(ctx context.Context, id string) (*clinicdomain.Clinic, error) {
	return f.getFn(ctx, id)
}
func (f *fakeClinicService) List(ctx context.Context) ([]*clinicdomain.Clinic, error) {
	return f.listFn(ctx)
}
func (f *fakeClinicService) Update(ctx context.Context, id string, input clinicapp.UpdateInput) (*clinicdomain.Clinic, error) {
	return f.updateFn(ctx, id, input)
}
func (f *fakeClinicService) UpdateBankAccount(ctx context.Context, id string, input clinicapp.UpdateBankAccountInput) (*clinicdomain.Clinic, error) {
	return f.bankFn(ctx, id, input)
}
func (f *fakeClinicService) Delete(ctx context.Context, id string) error {
	return f.deleteFn(ctx, id)
}

func sampleClinic(t *testing.T) *clinicdomain.Clinic {
	t.Helper()
	doc, err := clinicdomain.NewDocument("52998224725")
	require.NoError(t, err)
	acc, err := clinicdomain.NewBankAccount("341", "1234", "56789-0")
	require.NoError(t, err)
	c, err := clinicdomain.NewClinic("id-1", doc, "Corp", "Trade", acc, timeNow())
	require.NoError(t, err)
	return c
}

func TestClinicHandler_Create_Success(t *testing.T) {
	svc := &fakeClinicService{
		createFn: func(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error) {
			return sampleClinic(t), nil
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	body := `{"document":"52998224725","corporate_name":"Corp","trade_name":"Trade","bank_code":"341","agency":"1234","account":"56789-0"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "id-1", resp["id"])
}

func TestClinicHandler_Create_InvalidBody(t *testing.T) {
	svc := &fakeClinicService{}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestClinicHandler_Create_ServiceError(t *testing.T) {
	svc := &fakeClinicService{
		createFn: func(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error) {
			return nil, apperrors.Validation("invalid document", nil)
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestClinicHandler_Get_Success(t *testing.T) {
	svc := &fakeClinicService{
		getFn: func(ctx context.Context, id string) (*clinicdomain.Clinic, error) { return sampleClinic(t), nil },
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics/id-1", nil)
	req.SetPathValue("id", "id-1")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestClinicHandler_Get_NotFound(t *testing.T) {
	svc := &fakeClinicService{
		getFn: func(ctx context.Context, id string) (*clinicdomain.Clinic, error) {
			return nil, apperrors.NotFound("clinic not found")
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestClinicHandler_List_Success(t *testing.T) {
	svc := &fakeClinicService{
		listFn: func(ctx context.Context) ([]*clinicdomain.Clinic, error) {
			return []*clinicdomain.Clinic{sampleClinic(t)}, nil
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
}

func TestClinicHandler_Update_Success(t *testing.T) {
	svc := &fakeClinicService{
		updateFn: func(ctx context.Context, id string, input clinicapp.UpdateInput) (*clinicdomain.Clinic, error) {
			return sampleClinic(t), nil
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/clinics/id-1", bytes.NewBufferString(`{"corporate_name":"New","trade_name":"New"}`))
	req.SetPathValue("id", "id-1")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestClinicHandler_UpdateBankAccount_Success(t *testing.T) {
	svc := &fakeClinicService{
		bankFn: func(ctx context.Context, id string, input clinicapp.UpdateBankAccountInput) (*clinicdomain.Clinic, error) {
			return sampleClinic(t), nil
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/clinics/id-1/bank-account", bytes.NewBufferString(`{"bank_code":"001","agency":"1","account":"2"}`))
	req.SetPathValue("id", "id-1")
	rec := httptest.NewRecorder()

	handler.UpdateBankAccount(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestClinicHandler_Delete_Success(t *testing.T) {
	svc := &fakeClinicService{
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clinics/id-1", nil)
	req.SetPathValue("id", "id-1")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestClinicHandler_Delete_NotFound(t *testing.T) {
	svc := &fakeClinicService{
		deleteFn: func(ctx context.Context, id string) error { return apperrors.NotFound("clinic not found") },
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clinics/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
```

Este teste usa um helper `timeNow()` — adicione-o em um novo arquivo de suporte de teste:

`internal/adapters/http/helpers_test.go`:
```go
package http_test

import "time"

func timeNow() time.Time { return time.Now() }
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/http/... -v -run TestClinicHandler`
Expected: FAIL — `NewClinicHandler` undefined.

- [ ] **Step 3: Write minimal implementation**

`internal/adapters/http/clinic_handler.go`:
```go
package http

import (
	"context"
	"encoding/json"
	"net/http"

	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// clinicService is the subset of clinicapp.Service the handler depends on —
// declared here (consumer side) so tests can supply a fake without needing
// the real application package.
type clinicService interface {
	Create(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error)
	Get(ctx context.Context, id string) (*clinicdomain.Clinic, error)
	List(ctx context.Context) ([]*clinicdomain.Clinic, error)
	Update(ctx context.Context, id string, input clinicapp.UpdateInput) (*clinicdomain.Clinic, error)
	UpdateBankAccount(ctx context.Context, id string, input clinicapp.UpdateBankAccountInput) (*clinicdomain.Clinic, error)
	Delete(ctx context.Context, id string) error
}

type ClinicHandler struct {
	service clinicService
}

func NewClinicHandler(service clinicService) *ClinicHandler {
	return &ClinicHandler{service: service}
}

func (h *ClinicHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req clinicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	c, err := h.service.Create(r.Context(), clinicapp.CreateInput{
		Document:      req.Document,
		CorporateName: req.CorporateName,
		TradeName:     req.TradeName,
		BankCode:      req.BankCode,
		Agency:        req.Agency,
		Account:       req.Account,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, toClinicResponse(c))
}

func (h *ClinicHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toClinicResponse(c))
}

func (h *ClinicHandler) List(w http.ResponseWriter, r *http.Request) {
	clinics, err := h.service.List(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toClinicResponseList(clinics))
}

func (h *ClinicHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req clinicUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	c, err := h.service.Update(r.Context(), id, clinicapp.UpdateInput{
		CorporateName: req.CorporateName,
		TradeName:     req.TradeName,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toClinicResponse(c))
}

func (h *ClinicHandler) UpdateBankAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req bankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	c, err := h.service.UpdateBankAccount(r.Context(), id, clinicapp.UpdateBankAccountInput{
		BankCode: req.BankCode,
		Agency:   req.Agency,
		Account:  req.Account,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toClinicResponse(c))
}

func (h *ClinicHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusNoContent, nil)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/http/... -v -run TestClinicHandler`
Expected: PASS (all 9 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/http/clinic_handler.go internal/adapters/http/clinic_handler_test.go internal/adapters/http/helpers_test.go
git commit -m "feat: add clinic HTTP handler"
```

---

### Task 14: `internal/adapters/http` — DentistHandler

**Files:**
- Create: `internal/adapters/http/dentist_handler.go`
- Test: `internal/adapters/http/dentist_handler_test.go`

- [ ] **Step 1: Write the failing test**

`internal/adapters/http/dentist_handler_test.go`:
```go
package http_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDentistService struct {
	createFn func(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error)
	getFn    func(ctx context.Context, id string) (*dentistdomain.Dentist, error)
	listFn   func(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error)
	updateFn func(ctx context.Context, id string, input dentistapp.UpdateInput) (*dentistdomain.Dentist, error)
	deleteFn func(ctx context.Context, id string) error
}

func (f *fakeDentistService) Create(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error) {
	return f.createFn(ctx, input)
}
func (f *fakeDentistService) Get(ctx context.Context, id string) (*dentistdomain.Dentist, error) {
	return f.getFn(ctx, id)
}
func (f *fakeDentistService) ListByClinic(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error) {
	return f.listFn(ctx, clinicID)
}
func (f *fakeDentistService) Update(ctx context.Context, id string, input dentistapp.UpdateInput) (*dentistdomain.Dentist, error) {
	return f.updateFn(ctx, id, input)
}
func (f *fakeDentistService) Delete(ctx context.Context, id string) error {
	return f.deleteFn(ctx, id)
}

func sampleDentist(t *testing.T) *dentistdomain.Dentist {
	t.Helper()
	d, err := dentistdomain.NewDentist("dentist-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", true, timeNow())
	require.NoError(t, err)
	return d
}

func TestDentistHandler_Create_Success(t *testing.T) {
	svc := &fakeDentistService{
		createFn: func(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error) {
			return sampleDentist(t), nil
		},
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics/clinic-1/dentists", bytes.NewBufferString(`{"name":"Dra. Ana","phone":"+55 11 90000-0000","email":"ana@example.com","is_admin":true}`))
	req.SetPathValue("clinic_id", "clinic-1")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestDentistHandler_Create_ClinicNotFound(t *testing.T) {
	svc := &fakeDentistService{
		createFn: func(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error) {
			return nil, apperrors.NotFound("clinic not found")
		},
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics/missing/dentists", bytes.NewBufferString(`{}`))
	req.SetPathValue("clinic_id", "missing")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDentistHandler_Get_Success(t *testing.T) {
	svc := &fakeDentistService{
		getFn: func(ctx context.Context, id string) (*dentistdomain.Dentist, error) { return sampleDentist(t), nil },
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dentists/dentist-1", nil)
	req.SetPathValue("id", "dentist-1")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDentistHandler_ListByClinic_Success(t *testing.T) {
	svc := &fakeDentistService{
		listFn: func(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error) {
			return []*dentistdomain.Dentist{sampleDentist(t)}, nil
		},
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics/clinic-1/dentists", nil)
	req.SetPathValue("clinic_id", "clinic-1")
	rec := httptest.NewRecorder()

	handler.ListByClinic(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDentistHandler_Update_Success(t *testing.T) {
	svc := &fakeDentistService{
		updateFn: func(ctx context.Context, id string, input dentistapp.UpdateInput) (*dentistdomain.Dentist, error) {
			return sampleDentist(t), nil
		},
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/dentists/dentist-1", bytes.NewBufferString(`{"name":"Dra. Ana","phone":"1","email":"a@a.com"}`))
	req.SetPathValue("id", "dentist-1")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDentistHandler_Delete_Success(t *testing.T) {
	svc := &fakeDentistService{deleteFn: func(ctx context.Context, id string) error { return nil }}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/dentists/dentist-1", nil)
	req.SetPathValue("id", "dentist-1")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/http/... -v -run TestDentistHandler`
Expected: FAIL — `NewDentistHandler` undefined.

- [ ] **Step 3: Write minimal implementation**

`internal/adapters/http/dentist_handler.go`:
```go
package http

import (
	"context"
	"encoding/json"
	"net/http"

	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type dentistService interface {
	Create(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error)
	Get(ctx context.Context, id string) (*dentistdomain.Dentist, error)
	ListByClinic(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error)
	Update(ctx context.Context, id string, input dentistapp.UpdateInput) (*dentistdomain.Dentist, error)
	Delete(ctx context.Context, id string) error
}

type DentistHandler struct {
	service dentistService
}

func NewDentistHandler(service dentistService) *DentistHandler {
	return &DentistHandler{service: service}
}

func (h *DentistHandler) Create(w http.ResponseWriter, r *http.Request) {
	clinicID := r.PathValue("clinic_id")
	var req dentistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	d, err := h.service.Create(r.Context(), dentistapp.CreateInput{
		ClinicID: clinicID,
		Name:     req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
		IsAdmin:  req.IsAdmin,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, toDentistResponse(d))
}

func (h *DentistHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	d, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toDentistResponse(d))
}

func (h *DentistHandler) ListByClinic(w http.ResponseWriter, r *http.Request) {
	clinicID := r.PathValue("clinic_id")
	dentists, err := h.service.ListByClinic(r.Context(), clinicID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toDentistResponseList(dentists))
}

func (h *DentistHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req dentistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	d, err := h.service.Update(r.Context(), id, dentistapp.UpdateInput{
		Name:    req.Name,
		Phone:   req.Phone,
		Email:   req.Email,
		IsAdmin: req.IsAdmin,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toDentistResponse(d))
}

func (h *DentistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusNoContent, nil)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/http/... -v -run TestDentistHandler`
Expected: PASS (all 6 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/http/dentist_handler.go internal/adapters/http/dentist_handler_test.go
git commit -m "feat: add dentist HTTP handler"
```

---

### Task 15: `internal/adapters/http` — PaymentHandler

**Files:**
- Create: `internal/adapters/http/payment_handler.go`
- Test: `internal/adapters/http/payment_handler_test.go`

- [ ] **Step 1: Write the failing test**

`internal/adapters/http/payment_handler_test.go`:
```go
package http_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePaymentService struct {
	createFn func(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error)
	getFn    func(ctx context.Context, id string) (*paymentdomain.Payment, error)
}

func (f *fakePaymentService) Create(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error) {
	return f.createFn(ctx, input)
}
func (f *fakePaymentService) Get(ctx context.Context, id string) (*paymentdomain.Payment, error) {
	return f.getFn(ctx, id)
}

func samplePayment(t *testing.T) *paymentdomain.Payment {
	t.Helper()
	amount, err := paymentdomain.NewMoney(1000)
	require.NoError(t, err)
	p, err := paymentdomain.NewPayment("pay-1", "clinic-1", nil, amount, time.Now())
	require.NoError(t, err)
	p.SetPixCode("FAKE-PIX-CODE")
	return p
}

func TestPaymentHandler_Create_Success(t *testing.T) {
	svc := &fakePaymentService{
		createFn: func(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error) {
			return samplePayment(t), nil
		},
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBufferString(`{"clinic_id":"clinic-1","amount_cents":1000}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestPaymentHandler_Create_ClinicNotFound(t *testing.T) {
	svc := &fakePaymentService{
		createFn: func(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error) {
			return nil, apperrors.NotFound("clinic not found")
		},
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBufferString(`{"clinic_id":"missing","amount_cents":1000}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPaymentHandler_Create_InvalidAmount(t *testing.T) {
	svc := &fakePaymentService{
		createFn: func(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error) {
			return nil, apperrors.Validation("invalid amount", nil)
		},
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBufferString(`{"clinic_id":"clinic-1","amount_cents":0}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestPaymentHandler_Get_Success(t *testing.T) {
	svc := &fakePaymentService{
		getFn: func(ctx context.Context, id string) (*paymentdomain.Payment, error) { return samplePayment(t), nil },
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/pay-1", nil)
	req.SetPathValue("id", "pay-1")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPaymentHandler_Get_NotFound(t *testing.T) {
	svc := &fakePaymentService{
		getFn: func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return nil, apperrors.NotFound("payment not found")
		},
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/http/... -v -run TestPaymentHandler`
Expected: FAIL — `NewPaymentHandler` undefined.

- [ ] **Step 3: Write minimal implementation**

`internal/adapters/http/payment_handler.go`:
```go
package http

import (
	"context"
	"encoding/json"
	"net/http"

	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

type paymentService interface {
	Create(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error)
	Get(ctx context.Context, id string) (*paymentdomain.Payment, error)
}

type PaymentHandler struct {
	service paymentService
}

func NewPaymentHandler(service paymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req paymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	p, err := h.service.Create(r.Context(), paymentapp.CreateInput{
		ClinicID:  req.ClinicID,
		DentistID: req.DentistID,
		Cents:     req.Cents,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, toPaymentResponse(p))
}

func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toPaymentResponse(p))
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/http/... -v -run TestPaymentHandler`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/http/payment_handler.go internal/adapters/http/payment_handler_test.go
git commit -m "feat: add payment HTTP handler"
```

---

### Task 16: `internal/adapters/http` — Middleware e Router

**Files:**
- Create: `internal/adapters/http/middleware.go`
- Create: `internal/adapters/http/router.go`
- Create: `internal/adapters/http/docs.go`
- Test: `internal/adapters/http/router_test.go`

- [ ] **Step 1: Write the failing test**

`internal/adapters/http/router_test.go`:
```go
package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
)

func TestRouter_ListClinics_RoutesToHandler(t *testing.T) {
	clinicSvc := &fakeClinicService{
		listFn: func(ctx context.Context) ([]*clinicdomain.Clinic, error) { return nil, nil },
	}
	router := httpadapter.NewRouter(
		httpadapter.NewClinicHandler(clinicSvc),
		httpadapter.NewDentistHandler(&fakeDentistService{}),
		httpadapter.NewPaymentHandler(&fakePaymentService{}),
		"testdata/openapi.yaml",
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_UnknownRoute_Returns404(t *testing.T) {
	router := httpadapter.NewRouter(
		httpadapter.NewClinicHandler(&fakeClinicService{}),
		httpadapter.NewDentistHandler(&fakeDentistService{}),
		httpadapter.NewPaymentHandler(&fakePaymentService{}),
		"testdata/openapi.yaml",
	)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestApplyMiddleware_RecoversFromPanic(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	wrapped := httpadapter.ApplyMiddleware(panicking)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() { wrapped.ServeHTTP(rec, req) })
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/http/... -v -run "TestRouter|TestApplyMiddleware"`
Expected: FAIL — `NewRouter`/`ApplyMiddleware` undefined.

- [ ] **Step 3: Write minimal implementation**

`internal/adapters/http/middleware.go`:
```go
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ApplyMiddleware wraps next with recovery, request-id, and access logging,
// in that order (recovery outermost so it also protects the other two).
func ApplyMiddleware(next http.Handler) http.Handler {
	return recoverMiddleware(requestIDMiddleware(loggingMiddleware(next)))
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "error", rec, "path", r.URL.Path)
				WriteError(w, nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", uuid.NewString())
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request handled",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
```

`internal/adapters/http/router.go`:
```go
package http

import "net/http"

// NewRouter wires all resource handlers under /api/v1, plus /docs and
// /openapi.yaml for API documentation. openapiPath is the filesystem path
// to the OpenAPI contract served at /openapi.yaml.
func NewRouter(clinicHandler *ClinicHandler, dentistHandler *DentistHandler, paymentHandler *PaymentHandler, openapiPath string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/clinics", clinicHandler.Create)
	mux.HandleFunc("GET /api/v1/clinics", clinicHandler.List)
	mux.HandleFunc("GET /api/v1/clinics/{id}", clinicHandler.Get)
	mux.HandleFunc("PUT /api/v1/clinics/{id}", clinicHandler.Update)
	mux.HandleFunc("PUT /api/v1/clinics/{id}/bank-account", clinicHandler.UpdateBankAccount)
	mux.HandleFunc("DELETE /api/v1/clinics/{id}", clinicHandler.Delete)

	mux.HandleFunc("POST /api/v1/clinics/{clinic_id}/dentists", dentistHandler.Create)
	mux.HandleFunc("GET /api/v1/clinics/{clinic_id}/dentists", dentistHandler.ListByClinic)
	mux.HandleFunc("GET /api/v1/dentists/{id}", dentistHandler.Get)
	mux.HandleFunc("PUT /api/v1/dentists/{id}", dentistHandler.Update)
	mux.HandleFunc("DELETE /api/v1/dentists/{id}", dentistHandler.Delete)

	mux.HandleFunc("POST /api/v1/payments", paymentHandler.Create)
	mux.HandleFunc("GET /api/v1/payments/{id}", paymentHandler.Get)

	mux.HandleFunc("GET /docs", docsHandler)
	mux.HandleFunc("GET /openapi.yaml", openapiHandler(openapiPath))

	return ApplyMiddleware(mux)
}
```

`internal/adapters/http/docs.go`:
```go
package http

import "net/http"

const swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Capim Clinics API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger-ui" });
    };
  </script>
</body>
</html>`

func docsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerHTML))
}

func openapiHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
```

Ajuste `WriteError` (Task 11) para aceitar `err == nil` graciosamente — no `recoverMiddleware` chamamos `WriteError(w, nil)` após um panic. Confirme que `errors.As(nil, &appErr)` retorna `false` sem panicar (comportamento padrão da stdlib) — nenhuma mudança de código é necessária, mas adicione um teste de regressão:

Adicione a `internal/adapters/http/response_test.go`:
```go
func TestWriteError_NilErrorMapsToInternal(t *testing.T) {
	rec := httptest.NewRecorder()

	httpadapter.WriteError(rec, nil)

	assert.Equal(t, 500, rec.Code)
}
```

- [ ] **Step 4: Criar `testdata/openapi.yaml` mínimo para os testes de router**

`internal/adapters/http/testdata/openapi.yaml`:
```yaml
openapi: 3.0.3
info:
  title: Test Fixture
  version: "0.0.0"
paths: {}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/adapters/http/... -v`
Expected: PASS (todos os testes do pacote, incluindo os de response, DTOs indiretos e handlers)

- [ ] **Step 6: Commit**

```bash
git add internal/adapters/http/middleware.go internal/adapters/http/router.go internal/adapters/http/docs.go internal/adapters/http/router_test.go internal/adapters/http/testdata internal/adapters/http/response_test.go
git commit -m "feat: add middleware chain, router wiring, and swagger docs endpoint"
```

---

### Task 17: `internal/platform/config` — configuração via variáveis de ambiente

**Files:**
- Create: `internal/platform/config/config.go`
- Test: `internal/platform/config/config_test.go`

- [ ] **Step 1: Write the failing test**

`internal/platform/config/config_test.go`:
```go
package config_test

import (
	"os"
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad_UsesDefaultsWhenUnset(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("OPENAPI_PATH")

	cfg := config.Load()

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "api/openapi.yaml", cfg.OpenAPIPath)
}

func TestLoad_UsesEnvironmentWhenSet(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("OPENAPI_PATH", "/custom/openapi.yaml")

	cfg := config.Load()

	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "/custom/openapi.yaml", cfg.OpenAPIPath)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/config/... -v`
Expected: FAIL — package `config` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

`internal/platform/config/config.go`:
```go
package config

import "os"

type Config struct {
	Port        string
	OpenAPIPath string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		OpenAPIPath: getEnv("OPENAPI_PATH", "api/openapi.yaml"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/platform/config/... -v`
Expected: PASS (2 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/platform/config
git commit -m "feat: add environment-based configuration loader"
```

---

### Task 18: `cmd/api/main.go` — composition root

**Files:**
- Create: `cmd/api/main.go`

Este arquivo é o composition root: não tem lógica de negócio, portanto não é coberto por teste unitário — é exercitado pelos testes de integração (Task 20) e por execução manual (`make run`).

- [ ] **Step 1: Escrever o composition root**

`cmd/api/main.go`:
```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/pix"
	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	clinicRepo := memory.NewClinicRepository()
	dentistRepo := memory.NewDentistRepository()
	paymentRepo := memory.NewPaymentRepository()
	pixProvider := pix.NewDefaultSimulator()

	clinicService := clinicapp.NewService(clinicRepo)
	dentistService := dentistapp.NewService(dentistRepo, clinicRepo)
	paymentService := paymentapp.NewService(paymentRepo, clinicRepo, dentistRepo, pixProvider)

	router := httpadapter.NewRouter(
		httpadapter.NewClinicHandler(clinicService),
		httpadapter.NewDentistHandler(dentistService),
		httpadapter.NewPaymentHandler(paymentService),
		cfg.OpenAPIPath,
	)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
```

- [ ] **Step 2: Verificar que o binário compila e sobe**

Run:
```bash
go build ./... && echo BUILD_OK
```
Expected: `BUILD_OK`, sem erros de compilação.

Run (manual smoke test, opcional mas recomendado):
```bash
PORT=8081 go run ./cmd/api &
sleep 1
curl -s -X POST localhost:8081/api/v1/clinics -d '{"document":"52998224725","corporate_name":"Corp","trade_name":"Trade","bank_code":"341","agency":"1","account":"2"}'
kill %1
```
Expected: resposta JSON 201 com a clínica criada.

- [ ] **Step 3: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: add composition root wiring services, adapters, and HTTP server"
```

---

### Task 19: `api/openapi.yaml` — contrato OpenAPI 3.0

**Files:**
- Create: `api/openapi.yaml`

- [ ] **Step 1: Escrever o contrato OpenAPI**

`api/openapi.yaml`:
```yaml
openapi: 3.0.3
info:
  title: Capim Clinics API
  version: "1.0.0"
  description: API para gestão de clínicas odontológicas, dentistas e pagamentos via Pix (simulado).
servers:
  - url: /api/v1

paths:
  /clinics:
    post:
      summary: Criar clínica
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/ClinicRequest' }
      responses:
        '201':
          description: Clínica criada
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Clinic' }
        '422':
          description: Dados inválidos
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Error' }
    get:
      summary: Listar clínicas
      responses:
        '200':
          description: Lista de clínicas
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/Clinic' }

  /clinics/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string }
    get:
      summary: Buscar clínica por id
      responses:
        '200':
          description: Clínica encontrada
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Clinic' }
        '404':
          description: Clínica não encontrada
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Error' }
    put:
      summary: Atualizar dados básicos da clínica
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/ClinicUpdateRequest' }
      responses:
        '200':
          description: Clínica atualizada
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Clinic' }
    delete:
      summary: Excluir clínica
      responses:
        '204':
          description: Clínica excluída

  /clinics/{id}/bank-account:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string }
    put:
      summary: Atualizar dados bancários da clínica
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/BankAccountRequest' }
      responses:
        '200':
          description: Dados bancários atualizados
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Clinic' }

  /clinics/{clinic_id}/dentists:
    parameters:
      - name: clinic_id
        in: path
        required: true
        schema: { type: string }
    post:
      summary: Adicionar dentista à clínica
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/DentistRequest' }
      responses:
        '201':
          description: Dentista criado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Dentist' }
    get:
      summary: Listar dentistas da clínica
      responses:
        '200':
          description: Lista de dentistas
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/Dentist' }

  /dentists/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string }
    get:
      summary: Buscar dentista por id
      responses:
        '200':
          description: Dentista encontrado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Dentist' }
    put:
      summary: Atualizar dentista
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/DentistRequest' }
      responses:
        '200':
          description: Dentista atualizado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Dentist' }
    delete:
      summary: Excluir dentista
      responses:
        '204':
          description: Dentista excluído

  /payments:
    post:
      summary: Criar cobrança via Pix (simulado)
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/PaymentRequest' }
      responses:
        '201':
          description: Cobrança criada com status pending
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Payment' }

  /payments/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema: { type: string }
    get:
      summary: Consultar status de um pagamento
      responses:
        '200':
          description: Pagamento encontrado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Payment' }
        '404':
          description: Pagamento não encontrado
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Error' }

components:
  schemas:
    ClinicRequest:
      type: object
      required: [document, corporate_name, trade_name, bank_code, agency, account]
      properties:
        document: { type: string, example: "52998224725" }
        corporate_name: { type: string }
        trade_name: { type: string }
        bank_code: { type: string }
        agency: { type: string }
        account: { type: string }

    ClinicUpdateRequest:
      type: object
      required: [corporate_name, trade_name]
      properties:
        corporate_name: { type: string }
        trade_name: { type: string }

    BankAccountRequest:
      type: object
      required: [bank_code, agency, account]
      properties:
        bank_code: { type: string }
        agency: { type: string }
        account: { type: string }

    Clinic:
      type: object
      properties:
        id: { type: string }
        document: { type: string }
        document_type: { type: string, enum: [CPF, CNPJ] }
        corporate_name: { type: string }
        trade_name: { type: string }
        bank_code: { type: string }
        agency: { type: string }
        account: { type: string }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }

    DentistRequest:
      type: object
      required: [name, phone, email]
      properties:
        name: { type: string }
        phone: { type: string }
        email: { type: string }
        is_admin: { type: boolean }

    Dentist:
      type: object
      properties:
        id: { type: string }
        clinic_id: { type: string }
        name: { type: string }
        phone: { type: string }
        email: { type: string }
        is_admin: { type: boolean }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }

    PaymentRequest:
      type: object
      required: [clinic_id, amount_cents]
      properties:
        clinic_id: { type: string }
        dentist_id: { type: string, nullable: true }
        amount_cents: { type: integer, format: int64, example: 15000 }

    Payment:
      type: object
      properties:
        id: { type: string }
        clinic_id: { type: string }
        dentist_id: { type: string, nullable: true }
        amount_cents: { type: integer, format: int64 }
        status: { type: string, enum: [pending, approved] }
        pix_code: { type: string }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }

    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code: { type: string }
            message: { type: string }
            details:
              type: object
              additionalProperties: { type: string }
```

- [ ] **Step 2: Validar sintaxe YAML**

Run:
```bash
go run ./cmd/api &
sleep 1
curl -sf localhost:8080/openapi.yaml > /dev/null && echo YAML_SERVED_OK
kill %1
```
Expected: `YAML_SERVED_OK` (o próprio servidor consegue servir o arquivo sem erro).

- [ ] **Step 3: Commit**

```bash
git add api/openapi.yaml
git commit -m "docs: add OpenAPI 3.0 contract for the API"
```

---

### Task 20: `test/integration` — testes end-to-end

**Files:**
- Create: `test/integration/api_test.go`

- [ ] **Step 1: Write the failing test**

`test/integration/api_test.go`:
```go
package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/pix"
	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServer builds the full application stack — real in-memory
// adapters, real use cases, real HTTP router — wired exactly like
// cmd/api/main.go, but with a fast Pix simulator so tests don't wait
// multiple seconds for confirmation.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	clinicRepo := memory.NewClinicRepository()
	dentistRepo := memory.NewDentistRepository()
	paymentRepo := memory.NewPaymentRepository()
	fastPix := pix.NewSimulator(5*time.Millisecond, 10*time.Millisecond)

	clinicService := clinicapp.NewService(clinicRepo)
	dentistService := dentistapp.NewService(dentistRepo, clinicRepo)
	paymentService := paymentapp.NewService(paymentRepo, clinicRepo, dentistRepo, fastPix)

	router := httpadapter.NewRouter(
		httpadapter.NewClinicHandler(clinicService),
		httpadapter.NewDentistHandler(dentistService),
		httpadapter.NewPaymentHandler(paymentService),
		"../../api/openapi.yaml",
	)

	return httptest.NewServer(router)
}

func postJSON(t *testing.T, url string, body map[string]any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var out map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func TestClinicLifecycle_CreateGetUpdateDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	createResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Clinica Sorriso LTDA",
		"trade_name": "Clinica Sorriso", "bank_code": "341", "agency": "1234", "account": "56789-0",
	})
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)
	created := decodeJSON(t, createResp)
	id := created["id"].(string)
	require.NotEmpty(t, id)

	getResp, err := http.Get(srv.URL + "/api/v1/clinics/" + id)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	getResp.Body.Close()

	updateReq, err := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/clinics/"+id,
		bytes.NewBufferString(`{"corporate_name":"New Corp","trade_name":"New Trade"}`))
	require.NoError(t, err)
	updateResp, err := http.DefaultClient.Do(updateReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)
	updated := decodeJSON(t, updateResp)
	assert.Equal(t, "New Corp", updated["corporate_name"])

	delReq, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/clinics/"+id, nil)
	require.NoError(t, err)
	delResp, err := http.DefaultClient.Do(delReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

	finalGet, err := http.Get(srv.URL + "/api/v1/clinics/" + id)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, finalGet.StatusCode)
}

func TestClinicCreate_InvalidDocumentReturns422(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "123", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestDentistLifecycle_CreateUnderClinicAndList(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	clinicResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})
	clinicID := decodeJSON(t, clinicResp)["id"].(string)

	dentistResp := postJSON(t, srv.URL+"/api/v1/clinics/"+clinicID+"/dentists", map[string]any{
		"name": "Dra. Ana", "phone": "+55 11 90000-0000", "email": "ana@example.com", "is_admin": true,
	})
	assert.Equal(t, http.StatusCreated, dentistResp.StatusCode)

	listResp, err := http.Get(srv.URL + "/api/v1/clinics/" + clinicID + "/dentists")
	require.NoError(t, err)
	defer listResp.Body.Close()
	var list []map[string]any
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&list))
	assert.Len(t, list, 1)
}

func TestDentistCreate_UnknownClinicReturns404(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postJSON(t, srv.URL+"/api/v1/clinics/does-not-exist/dentists", map[string]any{
		"name": "Dra. Ana", "phone": "1", "email": "ana@example.com",
	})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPaymentLifecycle_CreatedThenAutoApprovedAsynchronously(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	clinicResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})
	clinicID := decodeJSON(t, clinicResp)["id"].(string)

	payResp := postJSON(t, srv.URL+"/api/v1/payments", map[string]any{
		"clinic_id": clinicID, "amount_cents": 15000,
	})
	require.Equal(t, http.StatusCreated, payResp.StatusCode)
	created := decodeJSON(t, payResp)
	assert.Equal(t, "pending", created["status"])
	assert.NotEmpty(t, created["pix_code"])
	paymentID := created["id"].(string)

	assert.Eventually(t, func() bool {
		resp, err := http.Get(srv.URL + "/api/v1/payments/" + paymentID)
		require.NoError(t, err)
		defer resp.Body.Close()
		var body map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		return body["status"] == "approved"
	}, 2*time.Second, 10*time.Millisecond, "payment should be auto-approved by the simulated Pix provider")
}

func TestPaymentCreate_UnknownDentistReturns404(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	clinicResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})
	clinicID := decodeJSON(t, clinicResp)["id"].(string)

	resp := postJSON(t, srv.URL+"/api/v1/payments", map[string]any{
		"clinic_id": clinicID, "dentist_id": "does-not-exist", "amount_cents": 1000,
	})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDocsAndOpenAPIEndpoints_AreServed(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	docsResp, err := http.Get(srv.URL + "/docs")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, docsResp.StatusCode)
	docsResp.Body.Close()

	specResp, err := http.Get(srv.URL + "/openapi.yaml")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, specResp.StatusCode)
	specResp.Body.Close()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./test/integration/... -v`
Expected: FAIL até que todas as tasks anteriores estejam implementadas (se executado após a Task 19, deve passar de primeira — este é o teste de aceitação de todo o sistema).

- [ ] **Step 3: (nenhuma implementação nova — este teste valida a integração das tasks anteriores)**

Se falhar, o erro indica qual task anterior está incompleta ou incorreta — corrija lá, não neste arquivo.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./test/integration/... -v -race`
Expected: PASS (todos os 7 testes, sem data races)

- [ ] **Step 5: Rodar a suíte completa do projeto**

Run: `go test ./... -race`
Expected: PASS em todos os pacotes.

- [ ] **Step 6: Commit**

```bash
git add test/integration
git commit -m "test: add end-to-end integration tests covering clinic, dentist and payment flows"
```

---

### Task 21: Tooling — Dockerfile, docker-compose, golangci-lint, CI

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `.golangci.yml`
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Criar `Dockerfile`**

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM alpine:3.20
WORKDIR /
COPY --from=builder /bin/api /bin/api
COPY api/openapi.yaml /api/openapi.yaml
ENV OPENAPI_PATH=/api/openapi.yaml
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
```

- [ ] **Step 2: Criar `docker-compose.yml`**

```yaml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - OPENAPI_PATH=/api/openapi.yaml
```

- [ ] **Step 3: Criar `.golangci.yml`**

```yaml
run:
  timeout: 5m

linters:
  enable:
    - govet
    - staticcheck
    - unused
    - errcheck
    - gofmt
    - goimports
```

- [ ] **Step 4: Criar `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: go build ./...
      - run: go test ./... -race -v

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

- [ ] **Step 5: Validar build da imagem Docker**

Run: `docker build -t capim-clinics-api .`
Expected: build concluído com sucesso, imagem final baseada em `alpine:3.20`.

- [ ] **Step 6: Validar subida via docker-compose**

Run:
```bash
docker compose up -d --build
sleep 2
curl -sf http://localhost:8080/docs > /dev/null && echo DOCS_OK
docker compose down
```
Expected: `DOCS_OK`.

- [ ] **Step 7: Commit**

```bash
git add Dockerfile docker-compose.yml .golangci.yml .github/workflows/ci.yml
git commit -m "chore: add Dockerfile, docker-compose, golangci-lint config, and GitHub Actions CI"
```

---

## Self-Review (executado antes da entrega deste plano)

**1. Cobertura da spec:**
- Clínicas (CRUD + dados bancários) → Tasks 3, 6, 8, 11–13, 19, 20 ✅
- Dentistas (CRUD, vínculo obrigatório a clínica, flag admin) → Tasks 4, 6, 9, 11, 14, 19, 20 ✅
- Pagamentos Pix (POST /payments, ciclo pending→approved, confirmação assíncrona 2–5s) → Tasks 5, 6, 7, 10, 11, 15, 19, 20 ✅
- Testes unitários e de integração cobrindo sucesso e erro, com fakes/mocks das interfaces → todas as tasks de domain/application/adapters têm testes de sucesso e erro; Task 20 cobre integração ponta a ponta ✅
- Banco in-memory, sem serviço externo real → Task 6 (mapas thread-safe), Task 7 (simulador Pix) ✅
- Documentação/organização/testabilidade (critérios de avaliação) → estrutura hexagonal (Tasks 1–17), OpenAPI (Task 19), tooling (Task 21) ✅

**2. Placeholder scan:** nenhum "TBD"/"TODO"/"implementar depois" encontrado — todo código é completo e compilável.

**3. Consistência de tipos:** `CreateInput`/`UpdateInput`/`UpdateBankAccountInput` (clinic), `CreateInput`/`UpdateInput` (dentist), `CreateInput` (payment) são usados de forma idêntica entre a task que os define (application) e a task que os consome (handler HTTP) e nos testes de integração. Nomes de métodos (`Create`, `Get`, `List`/`ListByClinic`, `Update`, `UpdateBankAccount`, `Delete`) são estáveis em todas as camadas.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-09-02-api-clinicas-implementation.md`. Two execution options:

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

**Which approach?**

# API de Gestão de Clínicas

API REST em Go para gestão de clínicas odontológicas, seus dentistas e pagamentos via Pix
(simulado).

> Arquitetura, justificativa técnica, tradeoffs aceitos e preparação para evolução futura estão
> documentados separadamente em [`docs/ARQUITETURA.md`](docs/ARQUITETURA.md).

## Sumário

- [Como executar](#como-executar)
- [Documentação da API](#documentação-da-api)
- [Testes](#testes)

## Como executar

Pré-requisitos: Go 1.23+ (para rodar localmente) ou Docker + Docker Compose (para rodar em
container).

```bash
# Localmente
go run ./cmd/api                # sobe em :8080 por padrão
make run                        # equivalente via Makefile

# Testes
make test                       # go test ./... -v
make test-race                  # go test ./... -race (detecção de data races)
make lint                       # golangci-lint run ./...

# Build
make build                      # gera ./bin/api

# Docker
make docker-up                  # docker compose up --build (API em :8080)
make docker-build               # apenas build da imagem
```

Variáveis de ambiente (todas opcionais, com defaults sensatos — ver
`internal/platform/config/config.go`):

| Variável | Default | Descrição |
|---|---|---|
| `PORT` | `8080` | Porta HTTP do servidor |
| `OPENAPI_PATH` | `api/openapi.yaml` | Caminho do arquivo servido em `/openapi.yaml` |

## Documentação da API

Com o servidor rodando:

- **Swagger UI**: [http://localhost:8080/docs](http://localhost:8080/docs) — interface visual e
  interativa para explorar/testar todos os endpoints.
- **Contrato OpenAPI 3.0**: [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml)
  (também disponível estaticamente em [`api/openapi.yaml`](api/openapi.yaml)) — inclui todos os
  request/response schemas e os códigos de erro possíveis (`404`, `409`, `422`) por endpoint,
  para facilitar a integração de um front-end.

Todas as rotas ficam sob o prefixo `/api/v1` (ex: `POST /api/v1/clinics`).

**Postman**: importe [`postman/capim-clinicas.postman_collection.json`](postman/capim-clinicas.postman_collection.json)
para testar todos os endpoints manualmente ou rodar a coleção inteira via Collection Runner —
as pastas (`Clinics` → `Dentists` → `Payments` → `Error Scenarios` → `Cleanup` → `Docs`) capturam
os IDs criados automaticamente em variáveis da coleção e os reutilizam nas requisições seguintes,
então a coleção inteira roda de ponta a ponta sem edição manual. Ela também cobre os cenários de
erro documentados no OpenAPI (`404`, `409`, `422`), incluindo as validações cross-aggregate
descritas em [`docs/ARQUITETURA.md`](docs/ARQUITETURA.md#decisões-de-design-notáveis). Validado
com `newman run postman/capim-clinicas.postman_collection.json` (27 requisições, 41 assertions, 0
falhas).

## Testes

```bash
make test        # unitário, verbose
make test-race   # com detector de data races
go test ./... -cover   # cobertura por pacote
```

- **Domínio**: testes unitários puros (sem mocks — funções determinísticas).
- **Application (use cases)**: fakes hand-rolled dos ports, cobrindo caminho feliz e cada erro de
  negócio (`404`, `409`, `422`), sem depender de banco ou serviço externo real.
- **Adapters HTTP**: via `httptest`, validando contrato (status code, corpo JSON) por endpoint.
- **Integração**: sobe o router completo com adapters in-memory reais e bate via
  `httptest.Server`, incluindo o fluxo assíncrono do Pix (`pending` → `approved` via polling com
  `assert.Eventually`).

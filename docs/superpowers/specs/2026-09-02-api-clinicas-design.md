# Design: API de Gestão de Clínicas Odontológicas (Capim Challenge)

- **Data:** 2026-09-02
- **Status:** Aprovado para planejamento de implementação
- **Escopo do desafio:** ver `README.md` na raiz do repositório

## 1. Contexto e objetivo

O desafio pede uma API em Go para a Capim, permitindo:

1. **Clínicas** — CRUD completo, com documento (CPF/CNPJ), razão social, nome fantasia, e dados
   bancários (banco, conta, agência) editáveis.
2. **Dentistas** — atrelados obrigatoriamente a uma clínica, com nome, telefone, email; uma
   clínica pode ter um ou mais dentistas como administrador/responsável legal.
3. **Pagamentos via Pix (incluído no escopo, tratado com engenharia completa)** — `POST /payments`
   recebe `clinic_id`, `valor` e `dentist_id` (opcional), retorna identificador de cobrança e
   código Pix copia-e-cola simulado. Ciclo de status `pending → approved`, confirmado por um
   processo em background após 2–5s simulados.

Requisitos técnicos obrigatórios do desafio que **guiam diretamente as decisões de arquitetura**:

- Go, com pacotes livres à escolha.
- Código bem documentado, boas práticas.
- Testes unitários e de integração cobrindo sucesso e erro.
- **"As camadas de serviço/negócio devem usar implementações fake/mock das interfaces, sem
  depender de um banco ou serviço externo real rodando."**
- Banco de dados in-memory (sem serviço externo real).

Critérios de avaliação declarados: Funcionalidade, Organização do Código (separação lógica/API,
naming, estrutura de arquivos), Testes, Interface (intuitividade/facilidade de integração
front-end), Justificativa Técnica.

## 2. Decisão de arquitetura: Hexagonal / Clean Architecture (Ports & Adapters)

### 2.1 Comparação de alternativas

| Critério | Camadas simples (handler→service→repo) | **Hexagonal / Clean (escolhida)** | DDD tático completo |
|---|---|---|---|
| Separação lógica ↔ interface (critério do desafio) | Parcial — services acessam infra diretamente | Total — domínio não conhece HTTP nem storage | Total, mas com camadas extras (bounded contexts) |
| Uso de fakes/mocks nas camadas de negócio (requisito obrigatório) | Possível, mas interfaces tendem a vazar detalhes de infra | Natural — ports já são a interface a ser mockada | Natural, porém com mais interfaces do que o necessário |
| Complexidade / boilerplate para o escopo (3 entidades + 1 fluxo) | Baixa | Média | Alta |
| Trocar storage in-memory → Postgres depois | Requer refatoração em vários pontos | Troca só o adapter, domínio intocado | Troca só o adapter, domínio intocado |
| Risco de over-engineering vs. escopo pedido | Baixo | Baixo-médio (mitigado ao manter "light") | Alto — agregados/VOs/eventos para um CRUD |
| Comunica maturidade técnica (critério "Justificativa Técnica") | Limitado | Alto — decisão clássica e defensável | Alto, mas desproporcional ao problema |

### 2.2 Justificativa

O requisito obrigatório *"as camadas de serviço/negócio devem usar implementações fake/mock das
interfaces, sem depender de um banco ou serviço externo real"* já é, na prática, uma descrição de
**Ports & Adapters**: o domínio define interfaces (ports) que tanto implementações reais quanto
fakes podem satisfazer. Uma arquitetura em camadas simples consegue emular isso, mas tende a
deixar o contrato implícito e a acoplar regra de negócio a detalhes de infraestrutura ao longo do
tempo. DDD tático resolveria o mesmo problema, mas com aparato (agregados, value objects
extensivos, domain events, bounded contexts) desproporcional a um domínio de 3 entidades e um
fluxo de pagamento simples — risco real de over-engineering, que pesa negativamente em avaliação
de código.

**Hexagonal/Clean "light"** — usando elementos pontuais de DDD (value objects simples para
CPF/CNPJ e `Money`, uma pequena máquina de estados para `Payment`) sem ir até bounded
contexts/agregados formais — atende o requisito de forma direta, é o padrão mais defensável
tecnicamente para o escopo, e maximiza o critério "Organização do Código".

### 2.3 Fluxo de dependências

```
adapters/http  ──┐
                 ├──▶  application (use cases)  ──▶  domain (entidades + regras)  ◀── adapters/memory
adapters/pix   ──┘                                                                 ◀── adapters/pix
```

Todas as dependências apontam para dentro, em direção ao `domain`. HTTP, storage e o provedor Pix
são adapters substituíveis sem tocar a regra de negócio.

## 3. Decisões técnicas (registro de decisões desta sessão)

| Decisão | Escolha | Motivo resumido |
|---|---|---|
| Escopo do Pix | Incluído, com engenharia completa (concorrência, worker em background, state machine) | Mostra profundidade técnica além do CRUD obrigatório |
| Framework HTTP | `net/http` da stdlib (Go 1.23+, novo `ServeMux` com path params) | Zero dependências externas, demonstra domínio da stdlib |
| Estilo arquitetural | Hexagonal / Clean Architecture | Atende diretamente o requisito de fakes/mocks nas camadas de negócio |
| Tooling de suporte | Makefile, Dockerfile, docker-compose, golangci-lint, GitHub Actions CI | Facilita avaliação (`make run`/`make test`) e demonstra maturidade de entrega |
| Documentação da API | OpenAPI 3.0 (`api/openapi.yaml`) + Swagger UI servido pela própria API | Padrão de mercado, facilita integração front-end (critério "Interface") |
| Versionamento | Prefixo de URL `/api/v1` | Simples, explícito, permite evolução futura sem quebrar clientes |
| Bibliotecas de teste | `testing` (stdlib) + `testify` (assert/require/mock) | Legibilidade e padrão de mercado em Go |
| Estratégia de IDs | UUID v4 gerado na aplicação | Desacoplado do storage, sem colisão, padrão para sistemas distribuídos |
| Versão do Go | 1.23+ | Mais recente estável, compatível com o novo roteamento da stdlib |

## 4. Estrutura de pastas

```
capim-challenge-clinicas/
├── cmd/
│   └── api/
│       └── main.go                # composition root: wiring manual de dependências, start do servidor HTTP
│
├── internal/
│   ├── domain/                     # regras de negócio puras — zero I/O, zero framework
│   │   ├── clinic/
│   │   │   ├── clinic.go           # entidade Clinic + invariantes de negócio
│   │   │   ├── document.go         # Value Object: CPF/CNPJ (validação de formato/dígito verificador)
│   │   │   ├── bank_account.go     # Value Object: dados bancários (banco, conta, agência)
│   │   │   └── repository.go       # PORT: interface ClinicRepository
│   │   ├── dentist/
│   │   │   ├── dentist.go          # entidade Dentist (nome, telefone, email, clinic_id, is_admin)
│   │   │   └── repository.go       # PORT: interface DentistRepository
│   │   └── payment/
│   │       ├── payment.go          # entidade Payment + máquina de estados (pending → approved)
│   │       ├── money.go            # Value Object: Money (evita float para valores monetários)
│   │       ├── repository.go       # PORT: interface PaymentRepository
│   │       └── pix_provider.go     # PORT: interface PixProvider (gera código copia-e-cola, agenda confirmação)
│   │
│   ├── application/                # use cases — orquestram o domínio, dependem só de ports
│   │   ├── clinic/                 # create, update, get, list, delete, update_bank_account
│   │   ├── dentist/                # create, update, get, list, delete
│   │   └── payment/                # create (inicia cobrança Pix), get (consulta status)
│   │
│   ├── adapters/                   # implementações concretas dos ports
│   │   ├── http/                   # driving adapter — entra na aplicação
│   │   │   ├── router.go           # registra rotas em /api/v1/*
│   │   │   ├── clinic_handler.go
│   │   │   ├── dentist_handler.go
│   │   │   ├── payment_handler.go
│   │   │   ├── dto/                # request/response — nunca expõem entidades de domínio
│   │   │   ├── middleware/         # logging, recover, request-id
│   │   │   └── response.go         # envelope padronizado de sucesso/erro
│   │   ├── memory/                 # driven adapter — repositórios in-memory thread-safe (mutex)
│   │   │   ├── clinic_repository.go
│   │   │   ├── dentist_repository.go
│   │   │   └── payment_repository.go
│   │   └── pix/                    # driven adapter — simulador do provedor Pix
│   │       └── simulator.go        # goroutine com delay aleatório (2–5s) até approved
│   │
│   └── platform/                   # infraestrutura transversal
│       ├── config/                 # variáveis de ambiente / flags
│       ├── logger/                 # logger estruturado
│       └── apperrors/              # erros de domínio tipados → mapeamento para status HTTP
│
├── api/
│   └── openapi.yaml                # contrato OpenAPI 3.0 — fonte da verdade da API
│
├── test/
│   └── integration/                # sobe o router com adapters in-memory reais, bate via httptest
│
├── Makefile                        # make run / test / lint / build / docker-build
├── Dockerfile
├── docker-compose.yml              # API + Swagger UI
├── .golangci.yml
├── .github/workflows/ci.yml        # lint + test em cada push/PR
└── go.mod
```

Cada pasta responde uma pergunta sozinha: `domain` é "quais são as regras?", `application` é "o
que o sistema faz?", `adapters` é "como isso conversa com o mundo externo?". Isso maximiza
diretamente o critério "Organização do Código".

## 5. Fluxos de dados

### 5.1 Criar Clínica (exemplo representativo do CRUD)

1. `POST /api/v1/clinics` chega no `clinic_handler.go`, que decodifica o JSON num DTO e valida
   apenas formato (campos obrigatórios presentes).
2. Handler chama o use case `CreateClinic` (injetado via construtor no `main.go`).
3. Use case constrói a entidade `Clinic`, que valida as regras de negócio reais (CPF/CNPJ válido
   via Value Object `Document`, dados bancários consistentes).
4. Use case chama a port `ClinicRepository.Save(ctx, clinic)`.
5. O adapter `adapters/memory.ClinicRepository` implementa isso com um `map` protegido por mutex.
6. Use case retorna a entidade; handler mapeia para o DTO de resposta com o status HTTP correto
   (`201 Created`).

Os demais fluxos de CRUD (Dentista, atualização de dados bancários, exclusão, listagem) seguem o
mesmo padrão handler → use case → domínio → repository port.

### 5.2 Pagamento Pix

1. `POST /api/v1/payments` recebe `clinic_id`, `amount`, `dentist_id` (opcional).
2. Use case `CreatePayment` valida que a clínica existe (e o dentista, se informado, pertence a
   ela), cria o `Payment` com status `pending`, e pede à port `PixProvider` um código copia-e-cola
   simulado.
3. O adapter `adapters/pix.Simulator` dispara uma goroutine que aguarda um delay aleatório
   (2–5s) e então, via callback injetado (o domínio não sabe que existe uma goroutine — isso é
   detalhe do adapter), atualiza o status do pagamento para `approved` através do
   `PaymentRepository`.
4. `GET /api/v1/payments/{id}` sempre reflete o status atual.

O mecanismo de concorrência fica isolado no adapter, não no domínio — se um dia o simulador for
trocado por uma integração Pix real (webhook), só o adapter muda.

## 6. Tratamento de erros

- Erros de domínio são tipados em `internal/platform/apperrors` (ex: `ErrNotFound`,
  `ErrValidation`, `ErrConflict`), carregando código + mensagem + detalhes opcionais.
- O adapter HTTP é o único lugar que sabe mapear esses tipos para status HTTP (`404`, `422`,
  `409`, `500`).
- Envelope de resposta de erro consistente em toda a API:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "documento inválido",
    "details": { "field": "document", "reason": "cnpj com dígito verificador incorreto" }
  }
}
```

## 7. Estratégia de testes

| Camada | Tipo de teste | Ferramenta |
|---|---|---|
| `domain` | Unitário puro (sem mocks — funções/métodos determinísticos) | `testing` + `testify/assert` |
| `application` (use cases) | Unitário com fakes/mocks dos ports, cobrindo caminho feliz e cada erro de negócio | `testify/mock` |
| `adapters/http` | Unitário via `httptest`, use cases reais ou fakeados, valida contrato HTTP | `testify` |
| `test/integration` | Ponta a ponta: sobe o router completo com adapters in-memory reais | `httptest.Server` |

Isso satisfaz diretamente os requisitos "incluir testes unitários e de integração, cobrindo
cenários de sucesso e erro" e "testes cobrindo... usando mocks/fakes das interfaces de
persistência e do provedor de pagamento".

## 8. Como isso maximiza os critérios de avaliação do desafio

- **Funcionalidade** — CRUD completo de Clínicas e Dentistas, fluxo de Pix ponta a ponta com
  transição de estado assíncrona real.
- **Organização do Código** — separação estrita domain/application/adapters, naming alinhado ao
  vocabulário do desafio.
- **Testes** — cobertura de sucesso e erro em cada camada, sem dependência de infraestrutura real.
- **Interface** — contrato OpenAPI versionado, DTOs estáveis e desacoplados das entidades
  internas, facilitando integração com qualquer front-end.
- **Justificativa Técnica** — cada decisão deste documento é rastreável a um requisito do desafio
  ou critério de avaliação explícito.

## 9. Fora de escopo (não será implementado)

- Autenticação/autorização (não mencionada no desafio).
- Persistência real (Postgres, etc.) — banco in-memory é requisito explícito.
- Integração real com provedor Pix — simulação é requisito explícito.
- Multi-tenancy, internacionalização, rate limiting avançado.

## 10. Próximos passos

Este documento serve de base para o plano de implementação detalhado (via skill
`writing-plans`), que vai quebrar o trabalho em etapas incrementais e testáveis.

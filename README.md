Capim, trabalhamos com clínicas odontológicas e seus respectivos dentistas. A proposta deste
desafio é desenvolver uma aplicação em Go que permita o gerenciamento de registros relacionados
a esses conceitos.

Nesse contexto, seu objetivo é criar uma API para a gestão de clínicas (que chamaremos de Clinics),
possibilitando a criação de uma ou mais clínicas, bem como a administração de seus dentistas,
contas bancárias associadas e pagamentos recebidos.

1. Clínicas
• Deve ser possível criar, alterar, visualizar e excluir uma clínica.
• Informações da clínica: documento (CPF/CNPJ), razão social, nome fantasia.
• Para além das informações básicas, uma clínica tem informações bancárias (banco, conta,
agência). Essas informações também devem ser passíveis de alteração.

2. Dentistas
• Dentistas são atrelados, necessariamente, a uma clínica.
• Informações do dentista: nome, telefone, email.
• Uma clínica pode ter um ou mais dentistas como administrador e responsável legal da
clínica.

3. Pagamentos (Pix) - OPCIONAL
• Uma clínica (Clinic) deve poder receber pagamentos via Pix.
• Endpoint: POST /payments — recebe clinic_id, valor e dentist_id (opcional). Retorna um
identificador de cobrança e um código Pix copia-e-cola simulado.
• O pagamento deve ter um ciclo de status: pending → approved.
• Não é necessária integração com um provedor real. Simule a confirmação: por exemplo,
um processo em background que, após um tempo aleatório (ex: 2–5s), atualiza o status do
pagamento para approved.

Requisitos Técnicos (Obrigatório)
• O desafio deve ser executado usando Go, usando quaisquer pacotes que você julgar
úteis/necessários.
• O código deve ser bem documentado e seguir boas práticas de programação.
• Incluir testes unitários e de integração, cobrindo cenários de sucesso e erro.
• As camadas de serviço/negócio devem usar implementações fake/mock das interfaces, sem
depender de um banco ou serviço externo real rodando.
• Utilize banco de dados in-memory, não há necessidade de ter uma implementação real com
algum serviço de banco de dados.

Critérios de Avaliação
Critério O que avaliamos
Funcionalidade O sistema funciona? A lógica está correta? É possível criar e alterar

informações de uma clínica? E dos membros?

Organização do
Código

Separação entre lógica e interface da API. Naming claro. Estrutura
coerente de arquivos e módulos.

Testes Testes cobrindo cenários de sucesso e erro, usando mocks/fakes das

interfaces de persistência e do provedor de pagamento.
Interface A interface da API é intuitiva? Seria fácil integrar essa API em um

front-end?

Justificativa Técnica Capacidade de explicar e defender as decisões tomadas, tanto
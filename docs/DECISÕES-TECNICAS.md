# API de Gestão de Clínicas — Decisões Técnicas

1. As decisões técnicas

A primeira decisão grande foi de arquitetura. Optei por Hexagonal (Ports & Adapters): o domínio
(`Clinic`, `Dentist`, `Payment` e as regras de negócio em volta deles) não conhece nada sobre HTTP
ou sobre como os dados são guardados — ele só define interfaces, como `ClinicRepository` ou
`PixProvider`. Quem implementa essas interfaces são os adapters (`adapters/http` para a API,
`adapters/memory` para o storage, `adapters/pix` para o simulador de pagamento), e tudo é ligado
na inicialização da aplicação. Cheguei nessa decisão porque o próprio enunciado do desafio já
pedia isso implicitamente: a camada de negócio precisava ser testável com mocks, sem depender de
banco ou serviço externo de verdade. Também considerei ir direto para camadas mais simples
(handler → service → repositório acessando o banco direto) ou para um DDD tático mais completo,
com agregados e domain events. 
A primeira opção mistura responsabilidades e dificulta o
isolamento para teste; a segunda seria peso desproporcional para três entidades e um fluxo de
pagamento. Hexagonal ficou no meio: separação limpa, sem o aparato todo do DDD. Na prática, isso
significa que hoje, se eu quiser trocar o storage in-memory por Postgres ou o simulador de Pix por
uma integração real, eu troco só o adapter correspondente — não precisei tocar em domínio ou
regra de negócio nenhuma vez durante todo o desenvolvimento, mesmo depois de adicionar validações
novas.

A segunda decisão foi levar o fluxo de pagamento Pix a sério como problema de concorrência, e não
só como um CRUD com um status a mais. Implementei o fluxo completo, incluindo o worker
assíncrono que aprova o pagamento depois de alguns segundos, simulando o comportamento real de um
provedor de pagamento. Isso expôs um data race genuíno: o repositório em memória devolvia o
ponteiro vivo do mapa interno, então uma goroutine aprovando um pagamento em background e uma
requisição HTTP lendo esse mesmo pagamento ao mesmo tempo corrompiam o estado um do outro.
Corrigi fazendo com que o repositório sempre trabalhe sobre cópias defensivas dos dados — nunca
sobre o ponteiro compartilhado — e escrevi testes específicos para isso, além de rodar
`go test -race` na esteira de CI para não deixar isso voltar. Escolhi investir tempo nesse ponto
porque era onde dava para mostrar profundidade técnica de verdade: máquina de estados do
pagamento, concorrência, worker assíncrono — coisas que, se eu tivesse simplificado, teriam
passado batido.

A terceira decisão foi tratar integridade entre entidades como regra de domínio, não como
detalhe. Duas coisas específicas: primeiro, `POST /payments` valida que o dentista informado
realmente pertence à clínica informada — sem essa validação, era possível criar uma cobrança
associando um dentista de outra clínica, o que não fazia sentido nenhum de negócio. Segundo,
`DELETE /clinics/{id}` bloqueia a exclusão (com um erro claro) se ainda existem dentistas
vinculados à clínica, em vez de cascatear a exclusão e apagar tudo relacionado automaticamente, ou
de permitir e deixar dentistas órfãos no sistema. Encontrei os dois casos numa revisão final que
fiz olhando o sistema como um todo, me perguntando "que estado inválido esse conjunto de regras
permite hoje?" — não porque um teste tinha falhado. Escolhi bloquear em vez de cascatear porque
perda de dados é irreversível e prefiro um erro explícito, que a pessoa que está operando o
sistema pode decidir o que fazer, do que uma exclusão silenciosa que já apaga tudo.



2. O que eu faria diferente com mais tempo

Com mais tempo, o primeiro ponto que eu atacaria é autenticação e autorização, com
isolamento real por clínica (multi-tenancy). Hoje qualquer requisição enxerga todas as clínicas —
não existe conceito de "quem está chamando" a API. O domínio já está preparado para isso (a
clínica já é uma chave estrangeira em dentista e em pagamento, e já existe busca por clínica nos
repositórios), então a extensão ficaria isolada na camada de autorização. Não simulei isso de
propósito: preferia deixar de fora e documentado como decisão consciente do que fingir um
isolamento que não seria real.

Também trocaria a persistência em memória por Postgres de verdade, usando o mesmo contrato de
repositório que já existe — bastaria escrever um novo adapter, sem tocar em domínio ou regra de
negócio, e adicionar migrations e transações reais.

Um terceiro ponto é controle de concorrência otimista, com um campo de versão ou ETag. Hoje o
mutex evita corrupção de memória, mas duas requisições de atualização concorrentes no mesmo
recurso ainda podem se sobrescrever uma à outra — é um tradeoff que documentei como aceito, não
como resolvido.

Também adicionaria idempotência em `POST /payments`, com uma chave de idempotência enviada pelo
cliente. Num sistema de pagamento de verdade, um retry de rede não pode gerar duas cobranças para
o mesmo pedido.

E por último, paginação nas listagens — `GET /clinics` e `GET /clinics/{id}/dentists` hoje
retornam a lista inteira, o que é aceitável no escopo do desafio, mas não seria em produção com
volume real de dados.



3. Como a IA ajudou e onde fiz diferente do que ela sugeriu

Usei o Cursor como apoio principamente nos pontos abaixo:
- Manter consistência em código repetitivo por volume: o CRUD das 3 entidades segue exatamente o
  mesmo padrão de handler/DTO/service/repositório em todos os arquivos, sem nenhuma deriva de
  convenção de um pro outro.

- Numa revisão de qualidade de código, me ajudou a encontrar um problema de data race real —pois não apareceu em nenhum dos testes, Confirmei depois e cobri com teste de regressão.

- Trabalho mecânico e de alto volume, mas de baixo risco quando bem delimitado: gerou a coleção
  Postman inteira (27 requisições, com encadeamento automático de IDs, 0 falhas) e traduziu os
  comentários de ~30 arquivos Go pra português em paralelo, cada um validando `build`/`test`/`lint`
  antes de reportar como concluído.

- As tabelas de comparação de arquitetura foram um ponto de partida pra eu validar e ajustar, não
  uma decisão final.

- Revisão holística no final, olhando o projeto inteiro em busca de problemas que só aparecem
  quando várias partes interagem (foi aí que achei os dois bugs de integridade cross-aggregate)


Pontos em que fui por um caminho diferente do sugerido:

- Arquitetura sugerida: IA estava sugerindo partir para uma clean architecture mais completo e robusto estilo DDD, pela necessidade do uso de mocks e interfaces, mas por ser tratar de uma api menor com apenas 3 entidades entendi que uma arquitetura hexagonal que ja nos permitia atender todos os requisitos de ports/interfaces, mocks e separação de responsabilidades por camada com bom desacoplamento

- Compile-time interface assertions (`var _ Interface = (*Type)(nil)`): Adicionei essa assertion mais pra deixar registrado que conheço o padrão do que por uma necessidade real no nosso caso: com poucos adapters e wiring direto em main.go, o compilador já pegaria qualquer incompatibilidade de qualquer jeito — aqui ela funciona como documentação viva, não como proteção.

Onde ela vira indispensável de verdade é num cenário que a gente não tem hoje, mas que é bem realista conforme a API cresce: qualquer ponto onde o adapter passa a ser tratado via any/interface{} ou guardado num map — por exemplo um registry que escolhe a implementação por uma string vinda de config/env, ou um DI baseado em reflection. Nesses casos, sem a assertion, uma implementação incompleta compila sem erro nenhum; o Go só vai reclamar no exato momento do type assertion, em runtime — e isso pode acontecer direto em produção, na forma de panic. Não é o nosso cenário agora (tudo aqui é tipado e concreto, resolvido em compile-time), mas preferi deixar o padrão já estabelecido: no dia que precisar de algo assim, essa linha deixa de ser só documentação e passa a evitar um panic real.

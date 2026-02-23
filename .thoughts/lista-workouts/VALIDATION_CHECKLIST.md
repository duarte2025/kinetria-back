# Validation Checklist — Feature: lista-workouts

## ✅ Artefatos Obrigatórios (RPI Plan Phase)

- [x] **plan.md** — Plano completo de implementação
  - [x] Seção 1: Inputs usados
  - [x] Seção 2: AS-IS (resumo)
  - [x] Seção 3: TO-BE (proposta)
  - [x] Seção 4: Decisões e Assunções
  - [x] Seção 5: Riscos / Edge Cases
  - [x] Seção 6: Rollout / Compatibilidade
  - [x] Seção 7: Checklist de "Definition of Done"
  - [x] Seção 8: Próximos Passos

- [x] **test-scenarios.feature** — Cenários BDD em Gherkin
  - [x] Background (pré-condições)
  - [x] Happy paths (≥3 cenários)
  - [x] Edge cases (≥5 cenários)
  - [x] Sad paths - Validação (≥3 cenários)
  - [x] Sad paths - Autenticação (≥2 cenários)
  - [x] Sad paths - Infraestrutura (≥1 cenário)
  - [x] Observabilidade (≥1 cenário)
  - [x] Performance (≥1 cenário)

- [x] **tasks.md** — Backlog de tarefas executáveis
  - [x] Tarefas de Domain Layer (ports, use cases)
  - [x] Tarefas de Gateway Layer (repository, handler, router)
  - [x] Tarefas de Testes (unitários, integração)
  - [x] Tarefas de Documentação
  - [x] Cada tarefa com:
    - [x] Objetivo claro
    - [x] Arquivos/pacotes afetados
    - [x] Passos de implementação
    - [x] Critério de aceite (com testes)

## ✅ Artefatos Complementares

- [x] **README.md** — Guia de navegação
  - [x] Sumário executivo
  - [x] Links para artefatos
  - [x] Dependências
  - [x] Quickstart
  - [x] Contrato OpenAPI
  - [x] Diagrama de arquitetura (texto)

- [x] **PLANNING_COMPLETE.txt** — Sumário visual
  - [x] Status de completude
  - [x] Estatísticas de artefatos
  - [x] Próximos passos

## ✅ Qualidade dos Artefatos

### plan.md
- [x] AS-IS documenta gaps explicitamente (sem migrations, sem entidades, sem use cases)
- [x] TO-BE descreve todas as camadas (domain, gateway, persistência)
- [x] Contratos de dados definidos (Workout, DTOs, queries SQL)
- [x] Decisões arquiteturais justificadas (agregação no handler, paginação obrigatória)
- [x] Riscos identificados com mitigação (tabela não existe, middleware ausente, N+1 query)
- [x] Edge cases documentados (usuário sem workouts, página além do total, campos opcionais)
- [x] Rollout strategy definida (pré-requisitos, ordem, validação)
- [x] Logs e métricas especificados

### test-scenarios.feature
- [x] Sintaxe Gherkin válida
- [x] Cenários cobrem contrato OpenAPI (200, 401, 422)
- [x] Happy paths testam funcionalidade core
- [x] Edge cases testam comportamento em situações limite
- [x] Sad paths testam validação e autenticação
- [x] Observabilidade valida logs estruturados (sem PII)
- [x] Performance define SLA (p95 < 200ms)
- [x] Isolamento de dados entre usuários testado

### tasks.md
- [x] 10 tarefas bem definidas (T01-T10)
- [x] Ordem de execução clara (dependências explícitas)
- [x] Cada tarefa é atômica e executável
- [x] Critérios de aceite incluem testes
- [x] Tarefas de documentação incluídas (T10)
- [x] Tarefas de testes incluídas (T08, T09)
- [x] Exemplos de código fornecidos
- [x] Providers fx documentados
- [x] Estimativa de esforço (8-12h)

## ✅ Alinhamento com Inputs de Research

- [x] Baseado em `.thoughts/mvp-userflow/api-contract.yaml`
  - [x] Endpoint correto: GET /api/v1/workouts
  - [x] Query params corretos: page, pageSize
  - [x] Response schema correto: ApiResponse + PaginationMeta
  - [x] Códigos de erro corretos: 401, 422, 500

- [x] Baseado em `.thoughts/mvp-userflow/backend-architecture-report.simplified.md`
  - [x] Arquitetura hexagonal respeitada
  - [x] CRUD + Audit Log (não Event Sourcing)
  - [x] Modelo de dados Workout alinhado
  - [x] Observabilidade (logs, métricas) incluída

- [x] Baseado em `.thoughts/mvp-userflow/bff-aggregation-strategy.md`
  - [x] Agregação no handler (Opção 1)
  - [x] Use cases atômicos no domain
  - [x] DTOs no gateway layer

## ✅ Alinhamento com Instruções de Arquitetura

- [x] Baseado em `.github/instructions/global.instructions.md`
  - [x] Estrutura de diretórios hexagonal
  - [x] Injeção de dependência via fx
  - [x] Providers registrados corretamente
  - [x] Nomenclatura de arquivos consistente
  - [x] Padrão de DTOs e error handling

## ✅ Dependências Explicitadas

- [x] Foundation-infrastructure identificada como bloquante
  - [x] Migrations da tabela workouts
  - [x] Entidade Workout de domínio
  - [x] Docker Compose com PostgreSQL

- [x] Feature AUTH identificada como bloquante
  - [x] Middleware de autenticação JWT
  - [x] Extração de userID do context

- [x] Instruções de como verificar dependências fornecidas

## ✅ Critérios de Completude (Meta)

### Obrigatórios
- [x] Todos os artefatos obrigatórios criados
- [x] Plan contém AS-IS, TO-BE, decisões, riscos, rollout
- [x] Test scenarios cobrem happy + sad paths + edge cases
- [x] Tasks são executáveis e incluem testes/docs
- [x] Dependências bloquantes identificadas
- [x] Alinhamento com research verificado

### Desejáveis
- [x] README.md de navegação criado
- [x] PLANNING_COMPLETE.txt como sumário visual
- [x] Validação de qualidade (este checklist)
- [x] Exemplos de código nas tasks
- [x] Diagramas/flows descritos em texto

## 📊 Estatísticas Finais

- **Artefatos criados**: 5 arquivos (README, plan, test-scenarios, tasks, PLANNING_COMPLETE)
- **Tamanho total**: ~75 KB
- **Cenários BDD**: 30+ cenários
- **Tarefas**: 10 tarefas atômicas
- **Estimativa**: 8-12 horas
- **Cobertura**: happy paths + edge cases + sad paths + observabilidade + performance

## ✅ VALIDAÇÃO FINAL

**Status**: ✅ APROVADO — Planejamento completo e pronto para fase IMPLEMENT

**Próxima ação**: Executar fase IMPLEMENT usando agente RPI Developer ou rpi_implement

---

Validado em: 2026-02-23  
Responsável: RPI Plan Agent  
Versão: 1.0

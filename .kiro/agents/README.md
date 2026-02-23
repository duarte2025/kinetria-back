# Kiro Agents

Este diretório contém agents especializados para o Kiro CLI, adaptados do workflow RPI (Research → Plan → Implement).

## 📁 Estrutura

```
.kiro/agents/
├── rpi-researcher.md          # Research: entrevista técnica + pesquisa no repo
├── rpi-planner.md             # Plan: AS-IS/TO-BE + BDD + backlog
├── rpi-implement.md           # Implement: orquestrador de execução
├── rpi-developer.md           # Developer: executor task-by-task
├── code-analyzer.md           # Análise de fluxo AS-IS
├── fix-developer.md           # Correção de bugs
├── architect-backend.md       # Arquitetura de backend
├── architect-database.md      # Arquitetura de banco de dados
├── architect-event-sourcing.md # Arquitetura de eventos
└── architect-docs.md          # Arquitetura de documentação
```

## 🔄 Workflow RPI

### 1. Research (rpi-researcher)

**Objetivo:** Entender o problema e mapear o código existente.

**Processo:**
1. Fazer perguntas técnicas ao desenvolvedor (máx 10 por rodada)
2. Pesquisar no projeto usando subagents especializados:
   - `code-analyzer`: mapear fluxo AS-IS
   - `architect-backend`: analisar boundaries e contratos
   - `architect-event-sourcing`: analisar eventos e mensageria
   - `architect-database`: analisar persistência
3. Produzir `research-report.md` em `.thoughts/<feature|topic>/`

**Output:** `.thoughts/<feature|topic>/research-report.md`

### 2. Plan (rpi-planner)

**Objetivo:** Criar plano de implementação detalhado.

**Processo:**
1. Ler artefatos do Research
2. Consolidar AS-IS
3. Propor TO-BE (contratos, camadas, persistência)
4. Escrever cenários BDD (Gherkin)
5. Criar backlog de tarefas atômicas

**Outputs:**
- `.thoughts/<feature|topic>/plan.md`
- `.thoughts/<feature|topic>/test-scenarios.feature`
- `.thoughts/<feature|topic>/tasks.md`

### 3. Implement (rpi-implement)

**Objetivo:** Executar o backlog task-by-task.

**Processo:**
1. Preparar branch (`<type>/<service>/<subject>`)
2. Para cada task:
   - Delegar ao `rpi-developer` via subagent
   - Rodar testes
   - Commitar (1 task = 1 commit)
3. Manter `execution-report.md`
4. Abrir Pull Request ao final

**Output:** `.thoughts/<feature|topic>/execution-report.md`

## 🎯 Agents Especializados

### code-analyzer

Analisa fluxo AS-IS:
- Entrypoints (cmd/)
- Call chain (handler → usecase → gateway)
- Side effects (DB, eventos, HTTP)
- Observabilidade
- Segurança

**Output:** `.thoughts/<feature|topic>/as-is-flow-report.md`

### fix-developer

Correção de bugs:
- Reproduzir e entender o problema
- Localizar causa raiz
- Aplicar fix mínimo
- Adicionar teste de regressão

**Output:** `.thoughts/<bug-id>/fix-report.md`

### architect-backend

Arquitetura de backend:
- Service boundaries
- Contratos e versionamento
- Resiliência e observabilidade
- Segurança

**Output:** `.thoughts/<feature|topic>/backend-architecture-report.md`

### architect-database

Arquitetura de banco de dados:
- Schema design
- Migrations
- Índices e performance
- Transações e consistência

**Output:** `.thoughts/<feature|topic>/database-architecture-report.md`

### architect-event-sourcing

Arquitetura de eventos:
- Agregados e eventos
- Projeções e read models
- Sagas e process managers
- Idempotência e ordenação

**Output:** `.thoughts/<feature|topic>/event-sourcing-architecture-report.md`

### architect-docs

Arquitetura de documentação:
- ADRs (Architecture Decision Records)
- Runbooks e playbooks
- Guias de desenvolvimento
- Documentação de APIs

**Output:** `.thoughts/<feature|topic>/docs-architecture-report.md`

## 🚀 Como Usar

### Workflow completo (Research → Plan → Implement)

```bash
# 1. Research
kiro chat "Quero implementar <feature>. Use o agent rpi-researcher."

# 2. Plan
kiro chat "Use o agent rpi-planner para criar o plano baseado no research."

# 3. Implement
kiro chat "Use o agent rpi-implement para executar o backlog."
```

### Análise de código existente

```bash
kiro chat "Use o agent code-analyzer para mapear o fluxo de <feature>."
```

### Correção de bug

```bash
kiro chat "Use o agent fix-developer para corrigir <bug>."
```

### Análise arquitetural

```bash
# Backend
kiro chat "Use o agent architect-backend para analisar <feature>."

# Database
kiro chat "Use o agent architect-database para analisar mudanças no schema."

# Event Sourcing
kiro chat "Use o agent architect-event-sourcing para analisar eventos."

# Docs
kiro chat "Use o agent architect-docs para propor documentação."
```

## 📝 Convenções

### Diretório de artefatos

Todos os artefatos são salvos em:
```
.thoughts/<feature|topic>/
```

Use nomes curtos e estáveis (ex: `invoice-paid`, `token-service`, `refund-webhook`).

### Git discipline

- **1 task = 1 commit**
- Mensagem: `<type>(<feature|topic>): Txx - <título curto>`
- Tipos: `feat`, `fix`, `chore`, `test`, `docs`

### Branch naming

Template: `<type>/<service>/<subject>`
- Ex: `feat/user-service/add-registration`
- Ex: `fix/payment-service/nil-pointer`

## ✅ Princípios

1. **Evidência sobre suposição**: sempre cite caminhos e símbolos do código
2. **Pergunte antes de inventar**: se faltar informação, pare e pergunte
3. **Mudanças mínimas**: evite refactors amplos não solicitados
4. **Testes obrigatórios**: toda task deve ter critério de aceite com testes
5. **Documentação junto**: docs e testes fazem parte da implementação

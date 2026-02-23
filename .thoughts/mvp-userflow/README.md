# 📋 MVP User Flow — Start Workout Session

> **Feature**: Iniciar Sessão de Treino  
> **Endpoint**: `POST /api/v1/sessions`  
> **Status**: ✅ Planejamento concluído — pronto para implementação

---

## 📂 Artefatos Disponíveis

### 🔍 Research (inputs)
1. **[api-contract.yaml](./api-contract.yaml)** — Contrato OpenAPI completo da API
2. **[backend-architecture-report.simplified.md](./backend-architecture-report.simplified.md)** — Arquitetura AS-IS/TO-BE, decisões (CRUD + Audit Log), modelo de dados
3. **[bff-aggregation-strategy.md](./bff-aggregation-strategy.md)** — Estratégia de agregação (use cases atômicos + agregação no handler)

### 📝 Plan (outputs)
4. **[plan.md](./plan.md)** — Plano completo de implementação da feature StartSession
5. **[test-scenarios.feature](./test-scenarios.feature)** — Cenários BDD (Gherkin) cobrindo happy path e sad paths
6. **[tasks.md](./tasks.md)** — Backlog de 13 tarefas executáveis com critérios de aceite

---

## 🎯 Resumo Executivo

### Escopo da Feature
Implementar **apenas** o endpoint `POST /api/v1/sessions` que permite ao usuário autenticado iniciar uma nova sessão de treino.

**Funcionalidades incluídas**:
- ✅ Validação de ownership (workout deve pertencer ao usuário)
- ✅ Prevenção de duplicação (apenas 1 sessão ativa por usuário)
- ✅ Registro de auditoria (audit log obrigatório)
- ✅ Autenticação JWT (userID extraído do token)

**Funcionalidades NÃO incluídas** (features separadas):
- ❌ Registrar séries (`POST /sessions/{id}/sets`)
- ❌ Finalizar sessão (`PATCH /sessions/{id}/finish`)
- ❌ Abandonar sessão (`PATCH /sessions/{id}/abandon`)
- ❌ Consultar sessão (`GET /sessions/{id}`)

---

## 📦 Dependências

### Obrigatórias (foundation-infrastructure)
A feature assume que o **foundation-infrastructure** já foi implementado:
- ✅ Migrations SQL (users, workouts, sessions, audit_log)
- ✅ Entidades básicas (User, Workout, Session, AuditLog)
- ✅ Autenticação JWT (middleware que injeta userID no contexto)
- ✅ Database pool (PostgreSQL + SQLC)
- ✅ Infraestrutura básica (Docker Compose, health check, config)

**Se essas dependências não existem**, implementar foundation-infrastructure **ANTES** desta feature.

---

## 🚀 Como Implementar

### Passo 1: Ler o Plano
📖 **[plan.md](./plan.md)** — Leia para entender:
- AS-IS: estado atual do código
- TO-BE: arquitetura proposta (fluxo, contratos, persistência, auditoria)
- Decisões arquiteturais
- Riscos e edge cases

### Passo 2: Entender os Cenários de Teste
🧪 **[test-scenarios.feature](./test-scenarios.feature)** — Revise os cenários BDD para entender:
- Happy path (sucesso)
- Sad paths (validações, ownership, duplicação)
- Edge cases (concorrência, auditoria, segurança)

### Passo 3: Executar as Tasks
✅ **[tasks.md](./tasks.md)** — Siga as 13 tarefas na ordem:

| Task | Título | Estimativa |
|------|--------|-----------|
| T01  | Criar entidades de domínio | 1h |
| T02  | Criar Value Objects | 30min |
| T03  | Criar erros customizados | 15min |
| T04  | Criar interfaces de repositório | 30min |
| T05  | Criar queries SQLC | 1h |
| T06  | Implementar Use Case | 2h |
| T07  | Testes unitários Use Case | 2h |
| T08  | Implementar Handler HTTP | 2h |
| T09  | Testes integração Handler | 2h |
| T10  | Documentar código (Godoc) | 1h |
| T11  | Documentar API (README) | 30min |
| T12  | Validar conformidade OpenAPI | 30min |
| T13  | Logs e métricas | 1h |

**Total estimado**: 3-5 dias (1 dev experiente)

---

## ✅ Critérios de Aceite (Definition of Done)

Antes de considerar a feature completa, verifique:

### Código
- [ ] `make build` compila sem erro
- [ ] `make lint` passa sem warnings
- [ ] `make test` passa com cobertura > 70%
- [ ] `make test-integration` passa

### Funcionalidade
- [ ] `POST /api/v1/sessions` retorna 201 Created (happy path)
- [ ] Retorna 401 sem token JWT
- [ ] Retorna 404 para workout de outro usuário
- [ ] Retorna 409 para sessão ativa duplicada
- [ ] Retorna 422 para request inválido
- [ ] Audit log criado em toda sessão

### Documentação
- [ ] Godoc em todas as funções/tipos exportados
- [ ] README da API atualizado
- [ ] Exemplos cURL funcionam

### Observabilidade
- [ ] Logs estruturados (zerolog) em JSON
- [ ] Métricas Prometheus funcionando
- [ ] `/metrics` expõe `sessions_started_total`

---

## 🧪 Testes

### Executar Testes Unitários
```bash
make test
# ou
go test ./internal/kinetria/domain/sessions/...
```

### Executar Testes de Integração
```bash
make test-integration
# ou
go test -tags=integration ./internal/kinetria/gateways/http/...
```

### Testar Endpoint Manualmente
```bash
# Subir ambiente local
docker-compose up -d

# Obter token JWT (assumindo endpoint de login existe)
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  | jq -r '.data.accessToken')

# Iniciar sessão
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"workoutId":"b2c3d4e5-f6a7-8901-bcde-f12345678901"}'
```

---

## 📊 Observabilidade

### Logs
```bash
# Ver logs da aplicação
docker-compose logs -f api

# Filtrar logs de sessões
docker-compose logs api | grep "session_created"
```

### Métricas
```bash
# Consultar métricas Prometheus
curl http://localhost:8080/metrics | grep sessions_started_total
curl http://localhost:8080/metrics | grep sessions_start_errors_total
```

---

## 🔄 Próximas Features

Após implementar StartSession, as próximas features são:

1. **RecordSet** — `POST /sessions/{id}/sets`
   - Registrar séries executadas (peso, reps, status)
   - Validações: sessão ativa, exercício pertence ao workout

2. **FinishSession** — `PATCH /sessions/{id}/finish`
   - Finalizar sessão com sucesso
   - Registrar timestamp de conclusão e notas opcionais

3. **AbandonSession** — `PATCH /sessions/{id}/abandon`
   - Abandonar sessão sem salvar progresso

4. **GetActiveSession** — `GET /sessions/active`
   - Consultar sessão ativa do usuário (se existir)

5. **Dashboard** — `GET /dashboard`
   - Agregação: sessão ativa + stats da semana + workouts recentes

---

## 📚 Referências

- **Arquitetura Hexagonal**: [README.md](../../README.md)
- **Padrões de Código**: [.github/instructions/global.instructions.md](../../../.github/instructions/global.instructions.md)
- **Contrato da API**: [api-contract.yaml](./api-contract.yaml)
- **Decisões Arquiteturais**: [backend-architecture-report.simplified.md](./backend-architecture-report.simplified.md)

---

## 📞 Suporte

Se encontrar dúvidas ou problemas durante a implementação:
1. Revise o [plan.md](./plan.md) — seção de Riscos e Edge Cases
2. Consulte os [cenários BDD](./test-scenarios.feature) para entender comportamento esperado
3. Verifique se dependências (foundation-infrastructure) estão implementadas

---

**Documento criado em**: 2026-02-23  
**Versão**: 1.0  
**Status**: ✅ Pronto para implementação

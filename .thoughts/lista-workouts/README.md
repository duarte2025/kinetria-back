# Feature: lista-workouts

## Sumário Executivo

Esta feature implementa o endpoint `GET /api/v1/workouts` para listar workouts do usuário autenticado com paginação.

**Status**: ✅ Planejamento completo — pronto para implementação

**Endpoint**: `GET /api/v1/workouts`

**Escopo**:
- Listar workouts do usuário autenticado
- Suporte a paginação (`page`, `pageSize`)
- Retornar `WorkoutSummary` (sem lista de exercises para otimização)
- Autenticação JWT obrigatória

---

## Artefatos

### 📋 [plan.md](./plan.md)
Plano completo de implementação contendo:
- AS-IS (estado atual do código)
- TO-BE (proposta de implementação)
- Decisões arquiteturais
- Riscos e edge cases
- Estratégia de rollout

### 🧪 [test-scenarios.feature](./test-scenarios.feature)
Cenários de teste BDD em Gherkin cobrindo:
- Happy paths (listagem com sucesso, paginação)
- Edge cases (usuário sem workouts, página além do total, campos opcionais)
- Sad paths (validação de input, autenticação, erros de infraestrutura)
- Observabilidade (logs estruturados)
- Performance (p95 < 200ms)

### 📝 [tasks.md](./tasks.md)
Backlog detalhado de implementação com 10 tarefas:
- **T01-T04**: Domain layer (ports, use case)
- **T05-T07**: Gateway layer (DTOs, handler, router)
- **T08-T09**: Testes (unitários, integração)
- **T10**: Documentação

**Estimativa**: 8-12 horas (1-2 dias)

---

## Dependências

### Bloquantes (devem existir antes de começar)

1. **Foundation-infrastructure** (`.thoughts/foundation-infrastructure/`)
   - Migrations da tabela `workouts`
   - Entidade de domínio `Workout`
   - Docker Compose com PostgreSQL

2. **Feature AUTH**
   - Middleware de autenticação JWT
   - Extração de `userID` do token e injeção no `context.Context`

### Verificar antes de iniciar

```bash
# Verificar se migrations existem
ls migrations/*workouts*.sql

# Verificar se entidade Workout existe
grep -r "type Workout struct" internal/kinetria/domain/entities/

# Verificar se middleware de auth existe
grep -r "AuthMiddleware" internal/kinetria/gateways/http/
```

---

## Quickstart (após dependências prontas)

### 1. Revisar o plano
```bash
cat .thoughts/lista-workouts/plan.md
```

### 2. Revisar cenários de teste
```bash
cat .thoughts/lista-workouts/test-scenarios.feature
```

### 3. Iniciar implementação
Seguir as tarefas em ordem:
```bash
# T01: Criar port WorkoutRepository
# T02: Implementar queries SQLC
# T03: Implementar WorkoutRepository
# ...
```

Veja detalhes de cada tarefa em [tasks.md](./tasks.md).

### 4. Validar implementação
```bash
# Testes unitários
go test ./internal/kinetria/domain/workouts/... -v
go test ./internal/kinetria/gateways/http/... -v

# Testes de integração
docker-compose -f docker-compose.test.yml up -d
go test ./internal/kinetria/tests/integration/... -v

# Lint
make lint

# Build
make build
```

### 5. Teste manual
```bash
# Obter token JWT (assumindo que AUTH está implementado)
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  | jq -r '.data.accessToken')

# Listar workouts
curl -X GET "http://localhost:8080/api/v1/workouts?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN" \
  | jq .
```

---

## Contrato OpenAPI

A especificação completa do endpoint está em:
- `.thoughts/mvp-userflow/api-contract.yaml` (linhas 559-595)

**Schema de resposta**:
```yaml
200:
  description: Paginated workout list
  content:
    application/json:
      schema:
        properties:
          data:
            type: array
            items:
              $ref: '#/components/schemas/WorkoutSummary'
          meta:
            $ref: '#/components/schemas/PaginationMeta'
```

---

## Arquitetura

### Fluxo de dados

```
HTTP Request
    ↓
[WorkoutsHandler] ← extrai userID do JWT context
    ↓
[ListWorkoutsUC] ← valida input, calcula offset
    ↓
[WorkoutRepository] ← executa queries SQLC
    ↓
PostgreSQL
```

### Camadas (Hexagonal)

**Domain** (`internal/kinetria/domain/`):
- `ports/workout_repository.go` — Interface do repositório
- `workouts/uc_list_workouts.go` — Caso de uso
- `entities/workout.go` — Entidade de domínio (foundation-infrastructure)

**Gateways** (`internal/kinetria/gateways/`):
- `repositories/workout_repository.go` — Adapter SQLC
- `repositories/queries/workouts.sql` — Queries SQL
- `http/handler_workouts.go` — HTTP handler
- `http/dto_workouts.go` — DTOs de resposta
- `http/router.go` — Registro de rotas

---

## Decisões Importantes

1. **Agregação no handler, não no domain**
   - Justificativa: seguir estratégia BFF (`.thoughts/mvp-userflow/bff-aggregation-strategy.md`)
   - Use case retorna entidade de domínio
   - Handler mapeia para DTO

2. **WorkoutSummary sem exercises**
   - Justificativa: otimização de performance
   - Detalhes (com exercises) em `GET /workouts/:id` (feature futura)

3. **Paginação obrigatória com defaults**
   - `page=1, pageSize=20`
   - Máximo: `pageSize=100`

4. **Campos opcionais como ponteiros no DTO**
   - Strings vazias no domínio → `nil` no JSON
   - Exemplo: `description: ""` → `"description": null`

---

## Próximos Passos (pós-implementação)

Após concluir esta feature:

1. **Feature: get-workout-details** (`GET /workouts/:id`)
   - Reutiliza mesma base (repository, entidades)
   - Adiciona join com `exercises`
   - Retorna `Workout` completo

2. **Feature: seed-workouts**
   - Popularar workouts de exemplo para testes e demo

3. **Feature: dashboard**
   - Usa `ListWorkoutsUC` para obter workouts recentes
   - Agrega com stats, sessions, user

---

## Contato / Dúvidas

Se houver dúvidas sobre este plano:
1. Revisar `.thoughts/mvp-userflow/backend-architecture-report.simplified.md`
2. Revisar `.github/instructions/global.instructions.md` (padrões de código)
3. Consultar OpenAPI spec: `.thoughts/mvp-userflow/api-contract.yaml`

---

**Criado em**: 2026-02-23  
**Versão**: 1.0  
**Fase**: Research → **Plan** → Implement

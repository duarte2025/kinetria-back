# 🔀 BFF Aggregation Strategy — Análise Arquitetural

## Contexto

No contexto de um **BFF (Backend for Frontend)** para Kinetria, surge a necessidade de agregar dados de múltiplas entidades (usuário, workouts, sessões, estatísticas) para reduzir o número de chamadas HTTP do cliente mobile/web.

**Questão central**: Onde realizar a agregação de dados?

---

## Opção 1: Agregação no Handler HTTP (Camada de Gateway)

### ✅ Arquitetura

```
┌─────────────────────────────────────────┐
│  HTTP Handler (gateways/http)           │
│  ┌─────────────────────────────────┐    │
│  │  GET /api/v1/dashboard          │    │
│  │  ┌─────────────────────────┐    │    │
│  │  │ 1. Call GetUserUC       │    │    │
│  │  │ 2. Call GetWorkoutsUC   │    │    │
│  │  │ 3. Call GetSessionsUC   │    │    │
│  │  │ 4. Call GetStatsUC      │    │    │
│  │  │ 5. Aggregate into DTO   │    │    │
│  │  └─────────────────────────┘    │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
           ↓        ↓        ↓        ↓
    ┌──────────┬──────────┬──────────┬──────────┐
    │ GetUser  │ GetWODs  │ GetSess  │ GetStats │  ← Use Cases (domain)
    │    UC    │    UC    │   UC     │    UC    │
    └──────────┴──────────┴──────────┴──────────┘
```

### Implementação (exemplo)

```go
// gateways/http/handler_dashboard.go
type DashboardHandler struct {
    getUserUC      domain.GetUserUC
    getWorkoutsUC  domain.GetWorkoutsUC
    getSessionsUC  domain.GetSessionsUC
    getStatsUC     domain.GetStatsUC
}

func (h DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := extractUserID(ctx)

    // Agregação no handler
    user, _ := h.getUserUC.Execute(ctx, domain.GetUserInput{ID: userID})
    workouts, _ := h.getWorkoutsUC.Execute(ctx, domain.GetWorkoutsInput{UserID: userID, Limit: 5})
    sessions, _ := h.getSessionsUC.Execute(ctx, domain.GetSessionsInput{UserID: userID, Limit: 10})
    stats, _ := h.getStatsUC.Execute(ctx, domain.GetStatsInput{UserID: userID})

    // DTO específico do cliente
    response := DashboardResponseDTO{
        User:              mapUserToDTO(user),
        RecentWorkouts:    mapWorkoutsToDTO(workouts),
        RecentSessions:    mapSessionsToDTO(sessions),
        Stats:             mapStatsToDTO(stats),
    }

    json.NewEncoder(w).Respond(response)
}
```

### ✅ Vantagens

| Dimensão | Benefício |
|----------|-----------|
| **Separação de responsabilidades** | ✅ Domain permanece **puro** e **reutilizável** (use cases atômicos servem múltiplos clientes) |
| **Testabilidade** | ✅ Use cases testados isoladamente; handlers testam apenas agregação |
| **Flexibilidade** | ✅ Diferentes clientes podem agregar **de forma diferente** (mobile vs web vs API externa) |
| **Performance** | ✅ Agregação pode ser feita em **paralelo** (goroutines) sem afetar domínio |
| **Evolução** | ✅ Se adicionar GraphQL/gRPC, pode reusar os mesmos use cases |
| **Complexidade** | ✅ Lógica de agregação **não vaza** para o domínio |

### ⚠️ Desvantagens

| Dimensão | Risco |
|----------|-------|
| **Transações** | ❌ Difícil garantir consistência transacional entre múltiplas chamadas |
| **Error handling** | ⚠️ Handler precisa orquestrar erros de múltiplos use cases |
| **Duplicação** | ⚠️ Se houver múltiplos handlers BFF (mobile, web), pode haver duplicação de lógica de agregação |
| **Responsabilidade do handler** | ⚠️ Handler fica "mais gordo", mas ainda é só orquestração |

---

## Opção 2: Agregação no Domain (Use Case Composto)

### ✅ Arquitetura

```
┌─────────────────────────────────────────┐
│  HTTP Handler (gateways/http)           │
│  ┌─────────────────────────────────┐    │
│  │  GET /api/v1/dashboard          │    │
│  │  ┌─────────────────────────┐    │    │
│  │  │ 1. Call GetDashboardUC  │    │    │
│  │  │ 2. Map to DTO           │    │    │
│  │  └─────────────────────────┘    │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
                    ↓
    ┌───────────────────────────────────┐
    │    GetDashboardUC (domain)        │  ← Use Case Composto
    │  ┌──────────────────────────┐     │
    │  │ 1. Call getUserUC        │     │
    │  │ 2. Call getWorkoutsUC    │     │
    │  │ 3. Call getSessionsUC    │     │
    │  │ 4. Call getStatsUC       │     │
    │  │ 5. Return aggregated     │     │
    │  └──────────────────────────┘     │
    └───────────────────────────────────┘
           ↓        ↓        ↓        ↓
    ┌──────────┬──────────┬──────────┬──────────┐
    │ GetUser  │ GetWODs  │ GetSess  │ GetStats │  ← Use Cases Atômicos
    │    UC    │    UC    │   UC     │    UC    │
    └──────────┴──────────┴──────────┴──────────┘
```

### Implementação (exemplo)

```go
// domain/dashboard/uc_get_dashboard.go
type GetDashboardUC struct {
    getUserUC      GetUserUC
    getWorkoutsUC  GetWorkoutsUC
    getSessionsUC  GetSessionsUC
    getStatsUC     GetStatsUC
}

type GetDashboardInput struct {
    UserID uuid.UUID
}

type GetDashboardOutput struct {
    User          User
    Workouts      []Workout
    Sessions      []Session
    Stats         Stats
}

func (uc GetDashboardUC) Execute(ctx context.Context, input GetDashboardInput) (GetDashboardOutput, error) {
    // Agregação no domain
    user, err := uc.getUserUC.Execute(ctx, GetUserInput{ID: input.UserID})
    if err != nil {
        return GetDashboardOutput{}, err
    }

    workouts, err := uc.getWorkoutsUC.Execute(ctx, GetWorkoutsInput{UserID: input.UserID, Limit: 5})
    if err != nil {
        return GetDashboardOutput{}, err
    }

    sessions, err := uc.getSessionsUC.Execute(ctx, GetSessionsInput{UserID: input.UserID, Limit: 10})
    if err != nil {
        return GetDashboardOutput{}, err
    }

    stats, err := uc.getStatsUC.Execute(ctx, GetStatsInput{UserID: input.UserID})
    if err != nil {
        return GetDashboardOutput{}, err
    }

    return GetDashboardOutput{
        User:     user,
        Workouts: workouts,
        Sessions: sessions,
        Stats:    stats,
    }, nil
}
```

```go
// gateways/http/handler_dashboard.go
func (h DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := extractUserID(ctx)

    // Handler só chama um use case
    output, err := h.getDashboardUC.Execute(ctx, domain.GetDashboardInput{UserID: userID})
    if err != nil {
        respondError(w, err)
        return
    }

    // Mapeamento simples
    response := DashboardResponseDTO{
        User:           mapUserToDTO(output.User),
        RecentWorkouts: mapWorkoutsToDTO(output.Workouts),
        RecentSessions: mapSessionsToDTO(output.Sessions),
        Stats:          mapStatsToDTO(output.Stats),
    }

    json.NewEncoder(w).Respond(response)
}
```

### ✅ Vantagens

| Dimensão | Benefício |
|----------|-----------|
| **Handler mais limpo** | ✅ Handler só chama **um use case** e mapeia para DTO |
| **Transações** | ✅ Mais fácil gerenciar transação se necessário (tudo no mesmo contexto de UC) |
| **Testes de integração** | ✅ Use case composto pode ser testado de forma mais integrada |
| **Clareza** | ✅ Intenção de "obter dashboard" está explícita no domain |

### ⚠️ Desvantagens

| Dimensão | Risco |
|----------|-------|
| **Acoplamento** | ❌ Use case composto **acopla** múltiplos use cases no domain |
| **Reutilização** | ❌ Se diferentes clientes precisam agregações diferentes, **cria-se múltiplos use cases compostos** (GetDashboardMobileUC, GetDashboardWebUC) |
| **Domínio "poluído"** | ❌ Domain passa a ter conhecimento de **necessidades específicas de cliente** |
| **Testabilidade** | ⚠️ Testes do use case composto precisam mockar múltiplos use cases |
| **Evolução** | ❌ Adicionar novo client com agregação diferente = novo use case no domain |
| **Violação SRP** | ❌ Use case composto tem **múltiplas razões para mudar** (qualquer mudança em user, workouts, sessions ou stats) |

---

## 🎯 Recomendação

### ✅ **OPÇÃO 1: Agregação no Handler HTTP** (camada de gateway)

**Justificativa arquitetural**:

1. **Arquitetura Hexagonal**: O domínio deve ser **agnóstico ao cliente**. Use cases devem representar **ações de negócio atômicas**, não necessidades de apresentação.

2. **Ports & Adapters**: O handler HTTP é um **adapter** (driving adapter). Sua responsabilidade é **adaptar** a necessidade do cliente (mobile/web) chamando as **portas** do domínio (use cases).

3. **Reutilização**: Os mesmos use cases atômicos podem ser usados por:
   - Handler HTTP mobile (agrega de uma forma)
   - Handler HTTP web (agrega de outra forma)
   - Handler GraphQL (resolve fields sob demanda)
   - Handler gRPC (streaming)
   - API pública para integrações

4. **Evolução**: Se no futuro houver necessidade de:
   - Dashboard diferente para coach vs atleta → novo handler, mesmo domain
   - Diferentes agregações para mobile vs web → handlers diferentes, mesmo domain
   - GraphQL que resolve apenas campos solicitados → usa os mesmos use cases

5. **Testabilidade**: Use cases atômicos têm **testes mais simples e focados**. Handlers BFF testam apenas a orquestração.

---

## 📋 Padrão Recomendado — Agregação Paralela

```go
// gateways/http/handler_dashboard.go
func (h DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := extractUserID(ctx)

    // Agregação paralela (melhor performance)
    type result struct {
        user     domain.GetUserOutput
        workouts domain.GetWorkoutsOutput
        sessions domain.GetSessionsOutput
        stats    domain.GetStatsOutput
        err      error
    }

    ch := make(chan result, 4)

    // Execute em paralelo
    go func() {
        out, err := h.getUserUC.Execute(ctx, domain.GetUserInput{ID: userID})
        ch <- result{user: out, err: err}
    }()

    go func() {
        out, err := h.getWorkoutsUC.Execute(ctx, domain.GetWorkoutsInput{UserID: userID, Limit: 5})
        ch <- result{workouts: out, err: err}
    }()

    go func() {
        out, err := h.getSessionsUC.Execute(ctx, domain.GetSessionsInput{UserID: userID, Limit: 10})
        ch <- result{sessions: out, err: err}
    }()

    go func() {
        out, err := h.getStatsUC.Execute(ctx, domain.GetStatsInput{UserID: userID})
        ch <- result{stats: out, err: err}
    }()

    // Collect results
    var res result
    for i := 0; i < 4; i++ {
        r := <-ch
        if r.err != nil {
            respondError(w, r.err)
            return
        }
        // Merge results
        if r.user.ID != uuid.Nil {
            res.user = r.user
        }
        // ... merge outros
    }

    response := DashboardResponseDTO{
        User:           mapUserToDTO(res.user),
        RecentWorkouts: mapWorkoutsToDTO(res.workouts),
        RecentSessions: mapSessionsToDTO(res.sessions),
        Stats:          mapStatsToDTO(res.stats),
    }

    json.NewEncoder(w).Respond(response)
}
```

**Benefícios adicionais**:
- ⚡ **Performance**: chamadas paralelas reduzem latência total
- 🔒 **Contexto compartilhado**: todas as chamadas compartilham o mesmo `ctx` (trace, timeout, cancelamento)
- 🧪 **Testável**: pode-se mockar cada use case independentemente

---

## 🚨 Quando Considerar Opção 2 (Use Case Composto)

Use cases compostos **podem fazer sentido** quando:

1. ✅ A agregação representa uma **regra de negócio** (não só apresentação)
   - Ex: `CreateOrderWithPaymentUC` (atomic transaction)
   
2. ✅ A agregação é **invariante** (sempre a mesma para todos os clientes)
   - Ex: `GenerateMonthlyReportUC` (sempre mesmos dados, independente do cliente)

3. ✅ Há necessidade de **transação atômica**
   - Ex: `TransferBalanceBetweenAccountsUC` (precisa ser all-or-nothing)

4. ✅ A agregação representa um **bounded context** ou **agregado DDD**
   - Ex: `GetOrderWithItemsAndShippingUC` (Order é agregado raiz)

**No caso de dashboard/home**: Não se aplica, pois é **apresentação/view**, não regra de negócio.

---

## 📐 Estrutura Final Recomendada

```
internal/kinetria/
├── domain/
│   ├── auth/
│   │   ├── uc_login.go              # Atômico
│   │   └── uc_register.go           # Atômico
│   ├── workouts/
│   │   ├── uc_create_workout.go     # Atômico
│   │   ├── uc_get_workout.go        # Atômico
│   │   └── uc_list_workouts.go      # Atômico
│   ├── sessions/
│   │   ├── uc_start_session.go      # Atômico
│   │   ├── uc_record_set.go         # Atômico
│   │   └── uc_finish_session.go     # Atômico
│   └── users/
│       ├── uc_get_user.go           # Atômico
│       └── uc_get_stats.go          # Atômico (calculado)
│
└── gateways/
    └── http/
        ├── handler_auth.go          # Endpoints auth
        ├── handler_workouts.go      # CRUD workouts
        ├── handler_sessions.go      # Tracking sessions
        └── handler_dashboard.go     # ⭐ BFF agregação mobile/web
            └── GetDashboard()       # Agrega: user + workouts + sessions + stats
```

**Nota**: Se web precisar agregação diferente → `handler_dashboard_web.go` separado.

---

## 🎯 Decisões para o Plan

1. ✅ **Implementar use cases atômicos** em `domain/`
2. ✅ **Agregação no handler BFF** (`gateways/http/handler_dashboard.go`)
3. ✅ **Agregação paralela** para melhor performance
4. ✅ **DTOs específicos de cliente** apenas no handler
5. ❌ **NÃO criar use cases compostos** para agregação de view

---

## 📚 Referências

- **Arquitetura Hexagonal**: Alistair Cockburn (Ports & Adapters)
- **Clean Architecture**: Robert C. Martin (Use Cases representam ações de negócio)
- **DDD**: Eric Evans (Agregados vs serviços de aplicação vs apresentação)
- **API Patterns**: API composition pattern (agregação no API Gateway/BFF, não no core)

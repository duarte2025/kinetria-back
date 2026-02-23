# Plan — mvp-userflow (Auth: Registro, Login e Tokens)

## 1) Inputs usados

- `.thoughts/mvp-userflow/api-contract.yaml` — Contrato OpenAPI 3.0 completo da API Kinetria
- `.thoughts/mvp-userflow/backend-architecture-report.simplified.md` — Relatório de arquitetura AS-IS/TO-BE
- `.thoughts/mvp-userflow/bff-aggregation-strategy.md` — Estratégia de agregação BFF
- Estrutura do repositório: `/home/runner/work/kinetria-back/kinetria-back`
- `.github/instructions/global.instructions.md` — Padrões de código e arquitetura

## 2) AS-IS (resumo)

### Estrutura atual do repositório

```
internal/kinetria/
├── domain/
│   ├── entities/entities.go   # Entidades vazias (template)
│   ├── errors/errors.go       # Erros base: ErrNotFound, ErrConflict, ErrMalformedParameters, ErrFailedDependency
│   ├── example/uc_example.go  # Exemplo de use case (xuc.UseCase[TInput, TOutput])
│   ├── ports/
│   │   ├── repositories.go    # Vazio (template)
│   │   └── services.go        # Vazio (template)
│   └── vos/vos.go             # VOs vazios (template)
├── gateways/
│   ├── config/config.go       # Config básica (APP_NAME, ENVIRONMENT, REQUEST_TIMEOUT)
│   ├── events/handlers/       # Template
│   ├── events/publishers/     # Template
│   ├── http/
│   │   ├── handler.go         # Handler vazio com validator.Validate
│   │   └── router.go          # Router vazio em /api/v1
│   └── repositories/
│       ├── queries/queries.sql # Vazio
│       └── repository.go      # Vazio (template)
migrations/                    # Vazio (.gitkeep apenas)
cmd/kinetria/api/main.go       # Main vazio com fx.New()
```

### Dependências já disponíveis (go.mod)

- ✅ `github.com/go-chi/chi/v5 v5.2.5` — HTTP router
- ✅ `github.com/go-playground/validator/v10 v10.30.1` — Validação
- ✅ `github.com/google/uuid v1.6.0` — UUID
- ✅ `github.com/jackc/pgx/v5 v5.8.0` — PostgreSQL driver
- ✅ `github.com/kelseyhightower/envconfig v1.4.0` — Config via env vars
- ✅ `go.uber.org/fx v1.24.0` — Dependency Injection
- ✅ `golang.org/x/crypto v0.46.0` — bcrypt disponível

### Dependências ausentes

- ❌ `github.com/golang-jwt/jwt/v5` — JWT (precisa adicionar)

### Estado atual

- Scaffolding hexagonal preparado (domain/gateways/cmd)
- Migrations vazias
- Config básica de Fx
- Sem implementação de endpoints
- Sem modelo de dados (entidades vazias)
- Sem repositories ou use cases funcionais

---

## 3) TO-BE (proposta de implementação para Auth)

### Escopo

Implementar **apenas** os 4 endpoints de autenticação:

1. `POST /auth/register` — Criar usuário + retornar tokens
2. `POST /auth/login` — Autenticar + retornar tokens
3. `POST /auth/refresh` — Renovar access token
4. `POST /auth/logout` — Revogar refresh token

### Interface HTTP (contratos)

**Endpoint**: `POST /api/v1/auth/register`
- **Request**: `{ "name": "Bruno Costa", "email": "user@example.com", "password": "s3cr3tP@ssw0rd" }`
- **Response 201**: `{ "data": { "accessToken": "...", "refreshToken": "...", "expiresIn": 3600 } }`
- **Response 409**: `{ "code": "EMAIL_ALREADY_EXISTS", "message": "..." }` (email duplicado)
- **Response 422**: `{ "code": "VALIDATION_ERROR", "message": "..." }` (validação falhou)

**Endpoint**: `POST /api/v1/auth/login`
- **Request**: `{ "email": "user@example.com", "password": "s3cr3tP@ssw0rd" }`
- **Response 200**: `{ "data": { "accessToken": "...", "refreshToken": "...", "expiresIn": 3600 } }`
- **Response 401**: `{ "code": "INVALID_CREDENTIALS", "message": "..." }` (credenciais inválidas)

**Endpoint**: `POST /api/v1/auth/refresh`
- **Request**: `{ "refreshToken": "..." }`
- **Response 200**: `{ "data": { "accessToken": "...", "refreshToken": "...", "expiresIn": 3600 } }`
- **Response 401**: `{ "code": "UNAUTHORIZED", "message": "..." }` (token inválido/expirado/revogado)

**Endpoint**: `POST /api/v1/auth/logout`
- **Request Headers**: `Authorization: Bearer <access_token>`
- **Request Body**: `{ "refreshToken": "..." }`
- **Response 204**: (sem body, logout bem-sucedido)
- **Response 401**: `{ "code": "UNAUTHORIZED", "message": "..." }` (token inválido)

### Modelo de dados (PostgreSQL)

**Tabela `users`**:
```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    profile_image_url VARCHAR(500) NOT NULL DEFAULT '/assets/avatars/default.png',
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Tabela `refresh_tokens`**:
```sql
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256 hash
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,  -- NULL = válido
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
```

### Entidades de domínio (Go)

```go
// internal/kinetria/domain/entities/entities.go

type User struct {
    ID              uuid.UUID
    Name            string
    Email           string
    PasswordHash    string
    ProfileImageURL string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type RefreshToken struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    TokenHash string
    ExpiresAt time.Time
    RevokedAt *time.Time  // NULL = válido
    CreatedAt time.Time
}
```

### Erros de domínio

```go
// internal/kinetria/domain/errors/errors.go

var (
    // ... erros existentes ...
    
    // Auth errors
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrTokenExpired       = errors.New("token expired")
    ErrTokenRevoked       = errors.New("token revoked")
    ErrTokenInvalid       = errors.New("token invalid")
)
```

### Ports (interfaces)

```go
// internal/kinetria/domain/ports/repositories.go

type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
    GetByEmail(ctx context.Context, email string) (*entities.User, error)
    GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
}

type RefreshTokenRepository interface {
    Create(ctx context.Context, token *entities.RefreshToken) error
    GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)
    RevokeByTokenHash(ctx context.Context, tokenHash string) error
    RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
}
```

### Use Cases

**1. RegisterUC**
- Input: `{ Name, Email, Password }`
- Output: `{ AccessToken, RefreshToken, ExpiresIn }`
- Fluxo:
  1. Validar email (formato válido)
  2. Validar senha (mínimo 8 chars)
  3. Verificar se email já existe (GetByEmail → ErrEmailAlreadyExists)
  4. Hashear senha (bcrypt cost 12)
  5. Criar User no DB
  6. Gerar refresh token (random bytes → SHA-256 hash)
  7. Salvar refresh token no DB
  8. Gerar JWT (HS256, 1h, claims: sub=user_id)
  9. Retornar tokens

**2. LoginUC**
- Input: `{ Email, Password }`
- Output: `{ AccessToken, RefreshToken, ExpiresIn }`
- Fluxo:
  1. Buscar usuário por email (GetByEmail → ErrInvalidCredentials se não encontrar)
  2. Comparar senha (bcrypt.Compare → ErrInvalidCredentials se falhar)
  3. Gerar refresh token (random bytes → SHA-256 hash)
  4. Salvar refresh token no DB
  5. Gerar JWT (HS256, 1h, claims: sub=user_id)
  6. Retornar tokens

**3. RefreshTokenUC**
- Input: `{ RefreshToken }`
- Output: `{ AccessToken, RefreshToken, ExpiresIn }`
- Fluxo:
  1. Hashear token recebido (SHA-256)
  2. Buscar token no DB (GetByTokenHash → ErrTokenInvalid se não encontrar)
  3. Validar se não está revogado (RevokedAt != NULL → ErrTokenRevoked)
  4. Validar se não expirou (ExpiresAt < Now → ErrTokenExpired)
  5. Revogar token antigo (RevokeByTokenHash)
  6. Gerar novo refresh token (random bytes → SHA-256 hash)
  7. Salvar novo refresh token no DB
  8. Gerar novo JWT (HS256, 1h, claims: sub=user_id)
  9. Retornar tokens

**4. LogoutUC**
- Input: `{ RefreshToken, UserID }` (UserID extraído do JWT)
- Output: void
- Fluxo:
  1. Hashear token recebido (SHA-256)
  2. Revogar token (RevokeByTokenHash)
  3. Retornar sucesso (204)

### Config (extensão)

Adicionar ao `gateways/config/config.go`:

```go
type Config struct {
    // ... existentes ...
    
    // Database
    DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
    
    // JWT
    JWTSecret           string        `envconfig:"JWT_SECRET" required:"true"`
    JWTExpiry           time.Duration `envconfig:"JWT_EXPIRY" default:"1h"`
    RefreshTokenExpiry  time.Duration `envconfig:"REFRESH_TOKEN_EXPIRY" default:"720h"` // 30 dias
}
```

### Repositórios (SQLC queries)

**Users**:
```sql
-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, profile_image_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;
```

**RefreshTokens**:
```sql
-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1 LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL;
```

### Handlers HTTP + Router

```go
// internal/kinetria/gateways/http/handler_auth.go

type AuthHandler struct {
    registerUC     domain.RegisterUC
    loginUC        domain.LoginUC
    refreshTokenUC domain.RefreshTokenUC
    logoutUC       domain.LogoutUC
    validate       *validator.Validate
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) { ... }
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) { ... }
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) { ... }
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) { ... }
```

```go
// internal/kinetria/gateways/http/router.go

func (s ServiceRouter) Router(router chi.Router) {
    router.Route("/auth", func(r chi.Router) {
        r.Post("/register", s.authHandler.Register)
        r.Post("/login", s.authHandler.Login)
        r.Post("/refresh", s.authHandler.RefreshToken)
        r.Post("/logout", s.authHandler.Logout) // Middleware de autenticação pendente
    })
}
```

### Persistência (PostgreSQL via SQLC)

- Driver: `pgx/v5` (já disponível)
- SQLC: gerar código a partir de `queries.sql`
- Conexão pool gerenciada via `fx.Provide`

### Observabilidade

**Logs estruturados** (JSON):
- Todas as requests HTTP (method, path, status, duration, user_id)
- Erros de domínio (código de erro, mensagem)
- **NUNCA logar**: senhas, tokens, PII sensível

**Métricas** (futuro):
- `http_requests_total{method, path, status}`
- `auth_login_attempts_total{status}` (success/failure)
- `auth_tokens_issued_total{type}` (access/refresh)

**Regras de log**:
```go
log.Info().
    Str("method", "POST").
    Str("path", "/auth/login").
    Str("user_id", userID).
    Int("status", 200).
    Msg("user_logged_in")
```

---

## 4) Decisões e Assunções

### Decisões de segurança

1. **Senha**: bcrypt cost 12 (balanceamento segurança vs performance)
2. **Access Token**: JWT HS256, 1h de validade (stateless, curta duração)
3. **Refresh Token**: 30 dias, armazenado como SHA-256 hash no DB (permite revogação)
4. **JWT Secret**: 256 bits mínimo, via env var `JWT_SECRET`
5. **Token rotation**: ao fazer refresh, revogar o token antigo e gerar um novo
6. **Default profile image**: `/assets/avatars/default.png`

### Decisões de implementação

1. **Arquitetura hexagonal**: domain puro, gateways adaptadores
2. **Use cases atômicos**: 1 use case = 1 ação de negócio
3. **Repository pattern**: SQLC para type-safe queries
4. **Validation**: `validator/v10` no handler HTTP, lógica de negócio nos use cases
5. **Error handling**: erros de domínio customizados, mapeados para HTTP status codes
6. **Response wrapper**: `{ "data": {...} }` em sucesso, `{ "code": "...", "message": "..." }` em erro

### Assunções

1. PostgreSQL 15+ disponível (via Docker Compose)
2. Variáveis de ambiente configuradas (`.env` local)
3. Migration runner disponível (ex: `migrate` ou script manual)
4. Middleware de autenticação JWT será implementado em feature futura (para /logout)
5. Rate limiting será implementado em feature futura

---

## 5) Riscos / Edge Cases

| Risco | Probabilidade | Impacto | Mitigação |
|-------|---------------|---------|-----------|
| **Email duplicado (race condition)** | Média | Médio | UNIQUE constraint no DB (users.email) |
| **Refresh token reuse** | Baixa | Alto | Revogar token ao fazer refresh (rotation) |
| **JWT secret vazado** | Baixa | Crítico | Rotação manual, TTL curto (1h), refresh token revogável |
| **Senha fraca** | Alta | Médio | Validação mínimo 8 chars (reforçar com regex futuro) |
| **Token expirado mas ainda aceito** | Baixa | Médio | Validar `exp` claim no JWT + ExpiresAt no DB |
| **Logout não revoga token imediatamente** | Média | Médio | JWT stateless (expira em 1h), refresh token revogado no DB |
| **Concorrência no refresh token** | Baixa | Médio | Transaction isolation + unique constraint (token_hash) |

### Edge Cases

1. **Usuário tenta registrar com email já existente**: retornar 409 Conflict
2. **Usuário faz login com senha incorreta**: retornar 401 Unauthorized (sem revelar se email existe)
3. **Usuário tenta usar refresh token expirado**: retornar 401 Unauthorized + código `TOKEN_EXPIRED`
4. **Usuário tenta usar refresh token revogado**: retornar 401 Unauthorized + código `TOKEN_REVOKED`
5. **Usuário faz logout com token já revogado**: retornar 204 (idempotente)
6. **Campo obrigatório ausente (ex: email)**: retornar 422 Unprocessable Entity + detalhes de validação

---

## 6) Rollout / Compatibilidade

### Estratégia de deploy

**MVP (primeira versão)**:
1. Aplicar migrations (users + refresh_tokens)
2. Deploy da aplicação (health check: `/health`)
3. Smoke test: `POST /auth/register` + `POST /auth/login` + `POST /auth/refresh` + `POST /auth/logout`

### Backward compatibility

- Não aplicável (primeira feature do MVP, sem versão anterior)
- Futuras mudanças no schema devem usar migrations expand/contract

### Ambiente de desenvolvimento

**Docker Compose** (sugestão):
```yaml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: kinetria
      POSTGRES_USER: kinetria
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"
```

**.env.example**:
```bash
APP_NAME=kinetria
ENVIRONMENT=development
DATABASE_URL=postgres://kinetria:secret@localhost:5432/kinetria?sslmode=disable
JWT_SECRET=your-256-bit-secret-here-use-openssl-rand-hex-32
JWT_EXPIRY=1h
REFRESH_TOKEN_EXPIRY=720h
REQUEST_TIMEOUT=5s
```

### Critérios de "done"

- ✅ Migrations aplicadas sem erro
- ✅ `make lint` sem warnings
- ✅ Testes unitários passando (cobertura > 70% dos use cases)
- ✅ Testes de integração HTTP passando (4 endpoints)
- ✅ Documentação (Godoc) nas funções exportadas
- ✅ PR reviewed e aprovado

---

## 7) Próximos Passos (pós-implementação Auth)

Após concluir a feature **mvp-userflow** (auth), as próximas features sugeridas são:

1. **Middleware de autenticação JWT** (para proteger rotas autenticadas)
2. **Feature WORKOUTS** (CRUD básico + seed data)
3. **Feature SESSIONS** (tracking de treino)
4. **Feature DASHBOARD** (agregação de dados para home screen)
5. **Audit Log** (rastreabilidade de ações do usuário)

---

**Documento gerado em**: 2026-02-23  
**Versão**: 1.0  
**Status**: ✅ Pronto para implementação  
**Próxima revisão**: após implementação completa + testes

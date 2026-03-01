# 🔎 Research Report — Profile Endpoints

## 1) Task Summary

### O que é
Implementar 3 endpoints de gerenciamento de perfil do usuário:
- **GET /api/v1/profile** — Obter dados do perfil (name, email, profileImageUrl, preferences)
- **PATCH /api/v1/profile** — Atualizar dados do perfil (name, preferences)
- **POST /api/v1/profile/avatar** — Upload de foto de perfil

### O que não é (fora de escopo)
- Alteração de email (requer verificação, fluxo separado)
- Alteração de senha (já existe endpoint separado)
- Exclusão de conta
- Integração com CDN/S3 (usar URL mock por enquanto)

---

## 2) Decisions Made

### Persistência
1. **Campo `preferences`:** Struct tipada: `{"theme": "dark"|"light", "language": "pt-BR"|"en-US"}`. Notificações ficam para v2.
2. **Validação de preferences:** Validar schema no backend (struct tipada). Rejeitar valores inválidos com 400.

### Interface / Contrato
3. **PATCH /profile:** Atualização parcial (apenas campos enviados). Usar ponteiros nos DTOs.
4. **Upload de avatar:** Adiar para v2. Por enquanto, permitir atualizar URL via PATCH /profile (string).
5. **Formatos futuros (v2):** JPEG/PNG/WebP, max 5MB, 100x100px a 2000x2000px.

### Regras de Negócio
6. **Validação de name:** 2-100 caracteres. Permitir letras, números, espaços, acentos. Não permitir apenas espaços.
7. **Concorrência:** Last-write-wins (sem optimistic locking na v1).

---

## 3) Facts from the Codebase

### Domínio(s) candidato(s)
- `internal/kinetria/domain/profile/` (novo, a criar)

### Entrypoints (cmd/)
- `cmd/kinetria/api/main.go` — Único entrypoint, usa Fx para DI

### Principais pacotes/símbolos envolvidos

**Entidades existentes:**
```go
// internal/kinetria/domain/entities/user.go
type User struct {
    ID              uuid.UUID
    Name            string
    Email           string
    PasswordHash    string
    ProfileImageURL *string  // Já existe
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

**Ports existentes:**
```go
// internal/kinetria/domain/ports/repositories.go
type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
    GetByEmail(ctx context.Context, email string) (*entities.User, error)
    // FALTA: Update(ctx context.Context, user *entities.User) error
}
```

**Gateways existentes:**
- `gateways/repositories/user_repository.go` — Implementação com SQLC
- `gateways/repositories/queries/users.sql` — Queries SQL tipadas
- `gateways/http/handler_auth.go` — Padrão de handler (referência)
- `gateways/http/middleware_auth.go` — JWT middleware (`extractUserIDFromJWT`)

---

## 4) Current Flow (AS-IS)

### Fluxo de autenticação (referência)
1. **HTTP Request** → Chi router (`router.go`)
2. **Auth Middleware** → Valida JWT, extrai userID, injeta no context
3. **Handler** → Extrai userID do context via `extractUserIDFromJWT(r)`
4. **Use Case** → Recebe userID, executa lógica
5. **Repository** → Acessa DB via SQLC
6. **Response** → Handler mapeia entity para DTO, retorna JSON

### Estrutura atual de User
- Tabela `users` tem: `id`, `name`, `email`, `password_hash`, `profile_image_url`, `created_at`, `updated_at`
- **FALTA:** Campo `preferences JSONB`

---

## 5) Change Points (prováveis pontos de alteração)

### 5.1) Migration

**Arquivo a criar:**
- `internal/kinetria/gateways/migrations/010_add_user_preferences.sql`

```sql
-- Adicionar coluna preferences
ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT '{}'::jsonb;

-- Índice para busca em preferences (opcional, se precisar filtrar)
CREATE INDEX idx_users_preferences ON users USING gin(preferences);
```

---

### 5.2) Domain Layer

**Arquivo a modificar:**
- `internal/kinetria/domain/entities/user.go`

Adicionar campo `Preferences`:
```go
type User struct {
    ID              uuid.UUID
    Name            string
    Email           string
    PasswordHash    string
    ProfileImageURL *string
    Preferences     map[string]interface{} // ou struct tipada
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

**Alternativa (struct tipada):**
```go
type UserPreferences struct {
    Theme    string `json:"theme"`    // "dark" | "light"
    Language string `json:"language"` // "pt-BR" | "en-US"
}
```

**Arquivos a criar:**
- `internal/kinetria/domain/profile/uc_get_profile.go`
- `internal/kinetria/domain/profile/uc_update_profile.go`

---

### 5.3) Ports

**Arquivo a modificar:**
- `internal/kinetria/domain/ports/repositories.go`

Adicionar método `Update`:
```go
type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
    GetByEmail(ctx context.Context, email string) (*entities.User, error)
    Update(ctx context.Context, user *entities.User) error // NOVO
}
```

---

### 5.4) Repository Layer

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/queries/users.sql`

Adicionar query `UpdateUser`:
```sql
-- name: UpdateUser :exec
UPDATE users
SET 
    name = $2,
    profile_image_url = $3,
    preferences = $4,
    updated_at = NOW()
WHERE id = $1;
```

**Arquivo a modificar:**
- `internal/kinetria/gateways/repositories/user_repository.go`

Implementar método `Update`:
```go
func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
    preferencesJSON, err := json.Marshal(user.Preferences)
    if err != nil {
        return fmt.Errorf("failed to marshal preferences: %w", err)
    }

    return r.queries.UpdateUser(ctx, queries.UpdateUserParams{
        ID:              user.ID,
        Name:            user.Name,
        ProfileImageUrl: user.ProfileImageURL,
        Preferences:     preferencesJSON,
    })
}
```

---

### 5.5) HTTP Layer

**Arquivo a criar:**
- `internal/kinetria/gateways/http/handler_profile.go`

Estrutura:
```go
type ProfileHandler struct {
    getProfileUC    *profile.GetProfileUC
    updateProfileUC *profile.UpdateProfileUC
}

// DTOs
type GetProfileResponse struct {
    ID              string           `json:"id"`
    Name            string           `json:"name"`
    Email           string           `json:"email"`
    ProfileImageURL *string          `json:"profileImageUrl"`
    Preferences     UserPreferences  `json:"preferences"`
}

type UserPreferences struct {
    Theme    string `json:"theme"`    // "dark" | "light"
    Language string `json:"language"` // "pt-BR" | "en-US"
}

type UpdateProfileRequest struct {
    Name            *string          `json:"name"`            // opcional
    ProfileImageURL *string          `json:"profileImageUrl"` // opcional
    Preferences     *UserPreferences `json:"preferences"`     // opcional
}
```

**Handlers:**
- `GET /api/v1/profile` → `HandleGetProfile()`
- `PATCH /api/v1/profile` → `HandleUpdateProfile()`

---

### 5.6) Router

**Arquivo a modificar:**
- `internal/kinetria/gateways/http/router.go`

Adicionar rotas protegidas:
```go
r.Route("/api/v1", func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    
    // Profile endpoints
    r.Get("/profile", profileHandler.HandleGetProfile)
    r.Patch("/profile", profileHandler.HandleUpdateProfile)
})
```

---

### 5.7) Dependency Injection

**Arquivo a modificar:**
- `cmd/kinetria/api/main.go`

Registrar use cases e handler:
```go
fx.Provide(
    // Use cases
    profile.NewGetProfileUC,
    profile.NewUpdateProfileUC,
    
    // Handler
    fx.Annotate(
        http.NewProfileHandler,
        fx.As(new(http.ProfileHandler)),
    ),
),
```

---

## 6) Risks / Edge Cases

### Concorrência
- **Race condition em PATCH /profile:** Se múltiplas requisições simultâneas, última sobrescreve
- **Mitigação:** Usar optimistic locking com `updated_at` (comparar versão antes de atualizar)

### Validações
- **Preferences schema:** Se aceitar JSON livre, pode ter dados inconsistentes
- **Mitigação:** Validar schema no backend (usar struct tipada `UserPreferences`)
- **Name vazio:** Validar tamanho mínimo (2 caracteres), máximo (100 caracteres)
- **Preferences muito grande:** Limitar tamanho do JSON (ex: 1KB)

### Performance
- **GET /profile:** Query simples, sem risco
- **PATCH /profile:** Update simples, sem risco

---

## 7) Suggested Implementation Strategy (alto nível, sem código)

### Etapa 1: Migration e Domain (30min)
1. Criar migration `010_add_user_preferences.sql`
2. Adicionar campo `Preferences` em `entities.User`
3. Usar struct tipada `UserPreferences` (theme, language)

### Etapa 2: Repository (30min)
1. Adicionar método `Update()` em `ports.UserRepository`
2. Criar query `UpdateUser` em `queries/users.sql`
3. Rodar `make sqlc` para gerar código
4. Implementar `Update()` em `user_repository.go`

### Etapa 3: Use Cases (45min)
1. Criar `uc_get_profile.go`:
   - Recebe userID do context
   - Chama `userRepo.GetByID()`
   - Retorna entity
2. Criar `uc_update_profile.go`:
   - Recebe userID + dados para atualizar
   - Valida inputs (name 2-100 chars, preferences válido)
   - Busca user atual
   - Atualiza campos modificados
   - Chama `userRepo.Update()`

### Etapa 4: HTTP Handler (1h)
1. Criar `handler_profile.go` com DTOs
2. Implementar `HandleGetProfile()`:
   - Extrai userID do JWT
   - Chama use case
   - Mapeia entity para DTO
   - Retorna JSON
3. Implementar `HandleUpdateProfile()`:
   - Extrai userID do JWT
   - Valida request body
   - Chama use case
   - Retorna JSON

### Etapa 5: Routing e DI (15min)
1. Registrar rotas em `router.go`
2. Registrar use cases e handler em `main.go` (Fx)

### Etapa 6: Testes (1h)
1. Unit tests para use cases (mock repository)
2. Integration tests para endpoints (DB real)
3. Edge cases: preferences inválido, name vazio

---

## 8) Handoff Notes to Plan

### Assunções feitas
- Campo `preferences` será JSONB com struct tipada `UserPreferences` (theme, language)
- PATCH /profile permite atualização parcial (apenas campos enviados)
- Upload de avatar adiado para v2 (permitir atualizar URL via PATCH por enquanto)
- Não haverá optimistic locking na v1 (aceitar last-write-wins)
- Validação de name: 2-100 caracteres

### Dependências
- **Decisões implementadas:**
  - Schema de preferences: struct tipada (theme, language)
  - Upload de avatar: adiado para v2
  - Atualização parcial via PATCH
  - Last-write-wins (sem optimistic locking)
- **Validações:**
  - Name: 2-100 caracteres
  - Preferences: validar valores de theme e language

### Recomendações para Plano de Testes

**Unit tests:**
- `GetProfileUC`: retorna user corretamente
- `UpdateProfileUC`: atualiza name, atualiza preferences, valida inputs inválidos

**Integration tests:**
- `GET /profile`: retorna 200 com dados corretos
- `PATCH /profile`: atualiza name, atualiza preferences, atualiza profileImageUrl, retorna 400 para inputs inválidos

**Edge cases:**
- Preferences com valores inválidos (theme="invalid")
- Name vazio ou muito longo (>100 chars)
- ProfileImageUrl com URL inválida

### Próximos passos
1. Responder perguntas da seção 2
2. Criar plano detalhado com tasks granulares
3. Implementar migration + domain + repository
4. Implementar use cases + handler
5. Testes

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

## 2) Clarifying Questions (para o dev)

### Persistência
1. **Campo `preferences`:** Qual schema JSON esperado? Sugestão: `{"theme": "dark|light", "language": "pt-BR|en-US", "notifications": {"email": bool, "push": bool}}`
2. **Validação de preferences:** Validar schema no backend ou aceitar qualquer JSON válido?

### Interface / Contrato
3. **PATCH /profile:** Permitir atualização parcial (apenas campos enviados) ou exigir todos os campos?
4. **Upload de avatar:** Aceitar apenas imagens (JPEG/PNG/WebP)? Tamanho máximo? Dimensões mínimas/máximas?
5. **URL de avatar:** Retornar URL completa ou path relativo? Onde armazenar por enquanto (filesystem local, S3, ou apenas mock)?

### Regras de Negócio
6. **Validação de name:** Tamanho mínimo/máximo? Permitir caracteres especiais?
7. **Concorrência:** Usar optimistic locking (`updated_at`) ou last-write-wins?

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
    Theme         string                 `json:"theme"`         // "dark" | "light"
    Language      string                 `json:"language"`      // "pt-BR" | "en-US"
    Notifications NotificationPreferences `json:"notifications"`
}

type NotificationPreferences struct {
    Email bool `json:"email"`
    Push  bool `json:"push"`
}
```

**Arquivos a criar:**
- `internal/kinetria/domain/profile/uc_get_profile.go`
- `internal/kinetria/domain/profile/uc_update_profile.go`
- `internal/kinetria/domain/profile/uc_upload_avatar.go` (opcional, se implementar upload)

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
    uploadAvatarUC  *profile.UploadAvatarUC // opcional
}

// DTOs
type GetProfileResponse struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Email           string                 `json:"email"`
    ProfileImageURL *string                `json:"profileImageUrl"`
    Preferences     map[string]interface{} `json:"preferences"`
}

type UpdateProfileRequest struct {
    Name        *string                `json:"name"`        // opcional
    Preferences map[string]interface{} `json:"preferences"` // opcional
}
```

**Handlers:**
- `GET /api/v1/profile` → `HandleGetProfile()`
- `PATCH /api/v1/profile` → `HandleUpdateProfile()`
- `POST /api/v1/profile/avatar` → `HandleUploadAvatar()` (opcional)

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
    r.Post("/profile/avatar", profileHandler.HandleUploadAvatar) // opcional
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
    profile.NewUploadAvatarUC, // opcional
    
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
- **Mitigação:** Validar schema no backend (usar struct tipada + validação)
- **Name vazio:** Validar tamanho mínimo (ex: 2 caracteres)
- **Preferences muito grande:** Limitar tamanho do JSON (ex: 10KB)

### Upload de Avatar
- **Tipo de arquivo:** Validar MIME type (image/jpeg, image/png, image/webp)
- **Tamanho:** Limitar a 5MB
- **Dimensões:** Validar mínimo 100x100px, máximo 2000x2000px
- **Storage:** Decisão pendente (S3, filesystem local, ou mock URL)
- **Cleanup:** Se trocar avatar, deletar arquivo antigo

### Performance
- **GET /profile:** Query simples, sem risco
- **PATCH /profile:** Update simples, sem risco
- **Índice GIN em preferences:** Só necessário se filtrar por preferences (improvável)

---

## 7) Suggested Implementation Strategy (alto nível, sem código)

### Etapa 1: Migration e Domain (30min)
1. Criar migration `010_add_user_preferences.sql`
2. Adicionar campo `Preferences` em `entities.User`
3. Decidir: `map[string]interface{}` ou struct tipada `UserPreferences`

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
   - Valida inputs (name não vazio, preferences válido)
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
3. Edge cases: preferences inválido, name vazio, concorrência

### Etapa 7 (Opcional): Upload de Avatar (2-3h)
1. Decidir storage (S3, filesystem, mock)
2. Criar gateway `storage/image_storage.go`
3. Implementar `uc_upload_avatar.go`
4. Implementar `HandleUploadAvatar()` com `multipart/form-data`
5. Validar tipo, tamanho, dimensões
6. Testes com mock do storage

---

## 8) Handoff Notes to Plan

### Assunções feitas
- Campo `preferences` será JSONB com schema flexível (ou struct tipada)
- PATCH /profile permite atualização parcial (apenas campos enviados)
- Upload de avatar será adiado ou usará URL mock (decisão pendente)
- Não haverá optimistic locking na v1 (aceitar last-write-wins)

### Dependências
- **Decisão de negócio:** Schema de preferences (livre ou tipado?)
- **Decisão técnica:** Storage para avatars (S3, filesystem, mock?)
- **Validações:** Tamanho mínimo/máximo de name, tamanho máximo de preferences

### Recomendações para Plano de Testes

**Unit tests:**
- `GetProfileUC`: retorna user corretamente
- `UpdateProfileUC`: atualiza name, atualiza preferences, valida inputs inválidos

**Integration tests:**
- `GET /profile`: retorna 200 com dados corretos
- `PATCH /profile`: atualiza name, atualiza preferences, retorna 400 para inputs inválidos
- `POST /profile/avatar`: (se implementar) valida tipo, tamanho, dimensões

**Edge cases:**
- Preferences com JSON inválido
- Name vazio ou muito longo
- Concorrência (2 PATCH simultâneos)
- Avatar com tipo inválido, tamanho > 5MB, dimensões inválidas

### Próximos passos
1. Responder perguntas da seção 2
2. Criar plano detalhado com tasks granulares
3. Implementar migration + domain + repository
4. Implementar use cases + handler
5. Testes

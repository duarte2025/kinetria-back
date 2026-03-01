# Plan — Profile Endpoints

## 1) Inputs usados
- `.thoughts/profile-endpoints/research-report.md`

## 2) AS-IS (resumo)

### Estrutura atual
- Tabela `users` possui: `id`, `name`, `email`, `password_hash`, `profile_image_url`, `created_at`, `updated_at`
- **FALTA:** Campo `preferences JSONB`
- Entity `User` não possui campo `Preferences`
- `UserRepository` não possui método `Update()`

### Fluxo de autenticação existente
1. HTTP Request → Chi router
2. Auth Middleware → Valida JWT, extrai userID, injeta no context
3. Handler → Extrai userID via `extractUserIDFromJWT(r)`
4. Use Case → Executa lógica de negócio
5. Repository → Acessa DB via SQLC
6. Response → Handler mapeia entity para DTO, retorna JSON

## 3) TO-BE (proposta)

### Interface HTTP
**Endpoints:**
- `GET /api/v1/profile` — Obter perfil do usuário autenticado
- `PATCH /api/v1/profile` — Atualizar perfil (atualização parcial)

**Autenticação:** JWT Bearer (middleware existente)

### Contratos

**GET /api/v1/profile**
```json
Response 200:
{
  "data": {
    "id": "uuid",
    "name": "string",
    "email": "string",
    "profileImageUrl": "string|null",
    "preferences": {
      "theme": "dark|light",
      "language": "pt-BR|en-US"
    }
  }
}
```

**PATCH /api/v1/profile**
```json
Request:
{
  "name": "string (opcional)",
  "profileImageUrl": "string|null (opcional)",
  "preferences": {
    "theme": "dark|light (opcional)",
    "language": "pt-BR|en-US (opcional)"
  }
}

Response 200:
{
  "data": {
    "id": "uuid",
    "name": "string",
    "email": "string",
    "profileImageUrl": "string|null",
    "preferences": {
      "theme": "dark|light",
      "language": "pt-BR|en-US"
    }
  }
}

Response 400:
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "string"
  }
}
```

### Persistência

**Migration:** `010_add_user_preferences.sql`
```sql
ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT '{}'::jsonb;
CREATE INDEX idx_users_preferences ON users USING gin(preferences);
```

**Query SQLC:** `UpdateUser`
```sql
UPDATE users
SET 
    name = $2,
    profile_image_url = $3,
    preferences = $4,
    updated_at = NOW()
WHERE id = $1;
```

### Domain Layer

**Struct tipada para preferences:**
```go
type UserPreferences struct {
    Theme    string `json:"theme"`    // "dark" | "light"
    Language string `json:"language"` // "pt-BR" | "en-US"
}
```

**Entity User (adicionar campo):**
```go
Preferences UserPreferences `json:"preferences"`
```

**Use Cases:**
- `GetProfileUC` — Busca user por ID
- `UpdateProfileUC` — Valida inputs, atualiza campos modificados

### Validações

**Name:**
- Mínimo: 2 caracteres
- Máximo: 100 caracteres
- Não permitir apenas espaços
- Permitir letras, números, espaços, acentos

**Preferences:**
- `theme`: apenas "dark" ou "light"
- `language`: apenas "pt-BR" ou "en-US"

**ProfileImageUrl:**
- Aceitar qualquer string (validação de URL real adiada para v2)

### Observabilidade
- Logs de erro em caso de falha no repository
- Retornar erros de domínio apropriados (validation, not found, internal)

## 4) Decisões e Assunções

1. **Preferences como struct tipada:** Usar `UserPreferences` com campos `theme` e `language` (validação no backend)
2. **PATCH parcial:** Usar ponteiros nos DTOs para distinguir "não enviado" de "null"
3. **Upload de avatar:** Adiado para v2. Por enquanto, aceitar URL via PATCH
4. **Concorrência:** Last-write-wins (sem optimistic locking na v1)
5. **Validação de name:** 2-100 caracteres, permitir acentos e espaços
6. **Default preferences:** `{"theme": "light", "language": "pt-BR"}` ao criar user

## 5) Riscos / Edge Cases

### Concorrência
- **Risco:** Múltiplas requisições simultâneas podem causar race condition
- **Mitigação v1:** Aceitar last-write-wins
- **Mitigação v2:** Implementar optimistic locking com `updated_at`

### Validações
- **Preferences inválido:** Validar schema no backend antes de persistir
- **Name vazio/muito longo:** Validar tamanho (2-100 chars)
- **Preferences muito grande:** Limitar tamanho do JSON (1KB)

### Compatibilidade
- **Users existentes sem preferences:** Migration define default `'{}'::jsonb`
- **Código existente:** Adicionar campo `Preferences` na entity não quebra código atual

## 6) Rollout / Compatibilidade

### Fase 1: Migration
1. Aplicar migration `010_add_user_preferences.sql`
2. Todos os users existentes recebem `preferences = '{}'::jsonb`

### Fase 2: Backend
1. Atualizar entity `User` com campo `Preferences`
2. Implementar `UserRepository.Update()`
3. Criar use cases `GetProfileUC` e `UpdateProfileUC`
4. Criar handler `ProfileHandler` com rotas

### Fase 3: Testes
1. Unit tests para use cases
2. Integration tests para endpoints
3. Edge cases (validações, concorrência)

### Compatibilidade
- ✅ Migration é aditiva (não quebra código existente)
- ✅ Campo `preferences` tem default (users antigos funcionam)
- ✅ Endpoints novos não afetam fluxos existentes

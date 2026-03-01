# Tasks — Profile Endpoints

## T01 — Criar migration para adicionar campo preferences
- **Objetivo:** Adicionar coluna `preferences JSONB` na tabela `users`
- **Arquivos:**
  - `internal/kinetria/gateways/migrations/010_add_user_preferences.sql`
- **Implementação:**
  1. Criar arquivo de migration com `ALTER TABLE users ADD COLUMN preferences JSONB DEFAULT '{}'::jsonb`
  2. Adicionar índice GIN para busca em preferences (opcional)
  3. Testar migration localmente com `docker-compose down -v && docker-compose up -d`
- **Critério de aceite:**
  - Migration aplicada com sucesso ao subir o ambiente
  - Coluna `preferences` existe na tabela `users`
  - Users existentes recebem `preferences = '{}'::jsonb`

---

## T02 — Adicionar struct UserPreferences no domain
- **Objetivo:** Criar struct tipada para preferences
- **Arquivos:**
  - `internal/kinetria/domain/vos/user_preferences.go` (novo)
- **Implementação:**
  1. Criar struct `UserPreferences` com campos `Theme` e `Language`
  2. Adicionar constantes para valores válidos (`ThemeDark`, `ThemeLight`, `LanguagePtBR`, `LanguageEnUS`)
  3. Adicionar método `Validate()` que retorna erro se valores inválidos
  4. Adicionar método `MarshalJSON()` e `UnmarshalJSON()` para serialização
- **Critério de aceite:**
  - Struct compila sem erros
  - `Validate()` rejeita valores inválidos
  - JSON serialization/deserialization funciona corretamente

---

## T03 — Atualizar entity User com campo Preferences
- **Objetivo:** Adicionar campo `Preferences` na entity `User`
- **Arquivos:**
  - `internal/kinetria/domain/entities/user.go`
- **Implementação:**
  1. Adicionar campo `Preferences vos.UserPreferences` na struct `User`
  2. Atualizar construtores/factories se necessário
- **Critério de aceite:**
  - Código compila sem erros
  - Testes existentes de `User` continuam passando

---

## T04 — Adicionar método Update no UserRepository port
- **Objetivo:** Definir contrato para atualizar usuário
- **Arquivos:**
  - `internal/kinetria/domain/ports/repositories.go`
- **Implementação:**
  1. Adicionar método `Update(ctx context.Context, user *entities.User) error` na interface `UserRepository`
- **Critério de aceite:**
  - Interface compila sem erros
  - Implementações existentes quebram (esperado, será corrigido em T05)

---

## T05 — Implementar query UpdateUser no SQLC
- **Objetivo:** Criar query SQL para atualizar usuário
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/queries/users.sql`
- **Implementação:**
  1. Adicionar query `UpdateUser` que atualiza `name`, `profile_image_url`, `preferences`, `updated_at`
  2. Rodar `make sqlc` para gerar código Go
- **Critério de aceite:**
  - Query compila sem erros no SQLC
  - Código Go gerado em `queries/users.sql.go`

---

## T06 — Implementar método Update no UserRepository
- **Objetivo:** Implementar lógica de atualização de usuário
- **Arquivos:**
  - `internal/kinetria/gateways/repositories/user_repository.go`
- **Implementação:**
  1. Implementar método `Update()` que:
     - Serializa `Preferences` para JSON
     - Chama `queries.UpdateUser()` com parâmetros corretos
     - Retorna erro se falhar
  2. Tratar erro de serialização JSON
- **Critério de aceite:**
  - Código compila sem erros
  - Método atualiza corretamente no banco (testar manualmente)

---

## T07 — Criar use case GetProfileUC
- **Objetivo:** Implementar lógica de negócio para obter perfil
- **Arquivos:**
  - `internal/kinetria/domain/profile/uc_get_profile.go` (novo)
- **Implementação:**
  1. Criar struct `GetProfileUC` com dependência `UserRepository`
  2. Implementar método `Execute(ctx context.Context, userID uuid.UUID) (*entities.User, error)`
  3. Chamar `userRepo.GetByID()`
  4. Retornar erro de domínio se user não encontrado
- **Critério de aceite:**
  - Use case retorna user corretamente
  - Retorna erro apropriado se user não existe

---

## T08 — Criar use case UpdateProfileUC
- **Objetivo:** Implementar lógica de negócio para atualizar perfil
- **Arquivos:**
  - `internal/kinetria/domain/profile/uc_update_profile.go` (novo)
- **Implementação:**
  1. Criar struct `UpdateProfileUC` com dependência `UserRepository`
  2. Criar struct `UpdateProfileInput` com campos opcionais (ponteiros)
  3. Implementar método `Execute(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*entities.User, error)`
  4. Validar inputs:
     - Name: 2-100 chars, não apenas espaços
     - Preferences: validar theme e language
  5. Buscar user atual via `userRepo.GetByID()`
  6. Atualizar apenas campos fornecidos
  7. Chamar `userRepo.Update()`
  8. Retornar user atualizado
- **Critério de aceite:**
  - Use case atualiza campos corretamente
  - Validações funcionam (rejeitar inputs inválidos)
  - Retorna erro apropriado se user não existe

---

## T09 — Criar DTOs para ProfileHandler
- **Objetivo:** Definir contratos de request/response
- **Arquivos:**
  - `internal/kinetria/gateways/http/dtos/profile.go` (novo)
- **Implementação:**
  1. Criar `GetProfileResponse` com campos: id, name, email, profileImageUrl, preferences
  2. Criar `UpdateProfileRequest` com campos opcionais (ponteiros): name, profileImageUrl, preferences
  3. Criar `UserPreferencesDTO` com campos: theme, language
  4. Adicionar tags JSON e validação
- **Critério de aceite:**
  - DTOs compilam sem erros
  - Serialização JSON funciona corretamente

---

## T10 — Criar ProfileHandler
- **Objetivo:** Implementar handlers HTTP para endpoints de perfil
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_profile.go` (novo)
- **Implementação:**
  1. Criar struct `ProfileHandler` com dependências `GetProfileUC` e `UpdateProfileUC`
  2. Implementar `HandleGetProfile()`:
     - Extrair userID do JWT via `extractUserIDFromJWT(r)`
     - Chamar `getProfileUC.Execute()`
     - Mapear entity para DTO
     - Retornar JSON 200
     - Tratar erros (401, 404, 500)
  3. Implementar `HandleUpdateProfile()`:
     - Extrair userID do JWT
     - Decodificar request body
     - Validar que ao menos um campo foi enviado
     - Chamar `updateProfileUC.Execute()`
     - Mapear entity para DTO
     - Retornar JSON 200
     - Tratar erros (400, 401, 404, 500)
- **Critério de aceite:**
  - Handlers compilam sem erros
  - Retornam status codes corretos
  - Mapeiam erros de domínio para HTTP corretamente

---

## T11 — Registrar rotas de perfil no router
- **Objetivo:** Adicionar endpoints de perfil no Chi router
- **Arquivos:**
  - `internal/kinetria/gateways/http/router.go`
- **Implementação:**
  1. Adicionar rotas protegidas:
     - `GET /api/v1/profile` → `profileHandler.HandleGetProfile`
     - `PATCH /api/v1/profile` → `profileHandler.HandleUpdateProfile`
  2. Garantir que middleware de autenticação está aplicado
- **Critério de aceite:**
  - Rotas registradas corretamente
  - Middleware de autenticação aplicado
  - Endpoints acessíveis via HTTP

---

## T12 — Registrar dependências no Fx (main.go)
- **Objetivo:** Configurar injeção de dependências
- **Arquivos:**
  - `cmd/kinetria/api/main.go`
- **Implementação:**
  1. Adicionar `profile.NewGetProfileUC` no `fx.Provide`
  2. Adicionar `profile.NewUpdateProfileUC` no `fx.Provide`
  3. Adicionar `http.NewProfileHandler` no `fx.Provide`
  4. Garantir que handler é injetado no router
- **Critério de aceite:**
  - Aplicação inicia sem erros
  - Dependências resolvidas corretamente pelo Fx

---

## T13 — Criar testes unitários para GetProfileUC
- **Objetivo:** Testar lógica de negócio de obter perfil
- **Arquivos:**
  - `internal/kinetria/domain/profile/uc_get_profile_test.go` (novo)
- **Implementação:**
  1. Criar mock de `UserRepository`
  2. Testar cenários:
     - Happy path: retorna user corretamente
     - User não encontrado: retorna erro apropriado
     - Erro no repository: retorna erro apropriado
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case

---

## T14 — Criar testes unitários para UpdateProfileUC
- **Objetivo:** Testar lógica de negócio de atualizar perfil
- **Arquivos:**
  - `internal/kinetria/domain/profile/uc_update_profile_test.go` (novo)
- **Implementação:**
  1. Criar mock de `UserRepository`
  2. Testar cenários (table-driven):
     - Atualizar name
     - Atualizar preferences
     - Atualizar profileImageUrl
     - Atualizar múltiplos campos
     - Name inválido (vazio, muito longo, apenas espaços)
     - Preferences inválido (theme/language inválidos)
     - User não encontrado
     - Nenhum campo fornecido
- **Critério de aceite:**
  - Testes passam com `go test`
  - Cobertura > 80% no use case
  - Todos os cenários BDD cobertos

---

## T15 — Criar testes de integração para GET /profile
- **Objetivo:** Testar endpoint de obter perfil com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_profile_integration_test.go` (novo)
- **Implementação:**
  1. Setup: criar user no DB, gerar JWT válido
  2. Testar cenários:
     - GET /profile com JWT válido: retorna 200 com dados corretos
     - GET /profile sem JWT: retorna 401
     - GET /profile com JWT inválido: retorna 401
  3. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Cenários BDD cobertos

---

## T16 — Criar testes de integração para PATCH /profile
- **Objetivo:** Testar endpoint de atualizar perfil com DB real
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_profile_integration_test.go`
- **Implementação:**
  1. Setup: criar user no DB, gerar JWT válido
  2. Testar cenários (table-driven):
     - PATCH /profile com name válido: retorna 200, atualiza DB
     - PATCH /profile com preferences válido: retorna 200, atualiza DB
     - PATCH /profile com profileImageUrl: retorna 200, atualiza DB
     - PATCH /profile com múltiplos campos: retorna 200, atualiza DB
     - PATCH /profile com name inválido: retorna 400
     - PATCH /profile com preferences inválido: retorna 400
     - PATCH /profile sem campos: retorna 400
     - PATCH /profile sem JWT: retorna 401
  3. Teardown: limpar DB
- **Critério de aceite:**
  - Testes passam com `INTEGRATION_TEST=1 go test`
  - Todos os cenários BDD cobertos
  - DB atualizado corretamente

---

## T17 — Documentar endpoints no README
- **Objetivo:** Atualizar documentação da API
- **Arquivos:**
  - `README.md`
- **Implementação:**
  1. Adicionar seção "Profile" na tabela de endpoints
  2. Adicionar exemplos de request/response para GET e PATCH /profile
  3. Documentar validações e erros possíveis
  4. Adicionar exemplo de uso com curl
- **Critério de aceite:**
  - Documentação clara e completa
  - Exemplos funcionam corretamente
  - Alinhada com comportamento implementado

---

## T18 — Adicionar comentários Godoc nos use cases
- **Objetivo:** Documentar código para geração de docs
- **Arquivos:**
  - `internal/kinetria/domain/profile/uc_get_profile.go`
  - `internal/kinetria/domain/profile/uc_update_profile.go`
- **Implementação:**
  1. Adicionar comentário de pacote em `doc.go`
  2. Adicionar comentários Godoc em structs e métodos públicos
  3. Documentar parâmetros e retornos
  4. Documentar erros possíveis
- **Critério de aceite:**
  - Comentários seguem padrão Godoc
  - `go doc` exibe documentação corretamente

---

## T19 — Atualizar documentação Swagger
- **Objetivo:** Gerar documentação interativa da API
- **Arquivos:**
  - `internal/kinetria/gateways/http/handler_profile.go`
- **Implementação:**
  1. Adicionar annotations Swagger nos handlers:
     - `@Summary`, `@Description`, `@Tags`
     - `@Accept`, `@Produce`
     - `@Param`, `@Success`, `@Failure`
     - `@Security` (JWT)
  2. Rodar `make swagger` para regenerar docs
  3. Testar endpoints no Swagger UI
- **Critério de aceite:**
  - Swagger UI exibe endpoints de perfil
  - Exemplos de request/response corretos
  - Autenticação JWT funciona no Swagger UI

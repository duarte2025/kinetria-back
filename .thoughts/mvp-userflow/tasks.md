# Tasks — mvp-userflow (Auth: Registro, Login e Tokens)

## Ordem de execução sugerida

As tarefas estão numeradas na ordem recomendada de implementação. Cada tarefa deve ser executada de forma atômica, com testes e validação antes de prosseguir.

---

## T01 — Adicionar dependência JWT ao go.mod

**Objetivo**: Adicionar a biblioteca JWT necessária para geração e validação de tokens.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/go.mod`

**Implementação**:
1. Executar: `go get github.com/golang-jwt/jwt/v5`
2. Executar: `go mod tidy`
3. Verificar que a dependência foi adicionada corretamente

**Critério de aceite**:
- ✅ `go.mod` contém `github.com/golang-jwt/jwt/v5` nas dependências
- ✅ `go mod tidy` roda sem erros
- ✅ Build da aplicação funciona: `go build ./cmd/kinetria/api`

---

## T02 — Criar migrations SQL para tabelas users e refresh_tokens

**Objetivo**: Criar as migrations SQL para as tabelas `users` e `refresh_tokens`.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/migrations/001_create_users.sql`
- `/home/runner/work/kinetria-back/kinetria-back/migrations/002_create_refresh_tokens.sql`

**Implementação**:

**001_create_users.sql**:
```sql
-- +migrate Up
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

-- +migrate Down
DROP TABLE IF EXISTS users;
```

**002_create_refresh_tokens.sql**:
```sql
-- +migrate Up
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS refresh_tokens;
```

**Critério de aceite**:
- ✅ Arquivos de migration criados em `/migrations/`
- ✅ Migration `001_create_users.sql` contém tabela `users` com todos os campos
- ✅ Migration `002_create_refresh_tokens.sql` contém tabela `refresh_tokens` com todos os campos
- ✅ Índices criados: `idx_users_email`, `idx_refresh_tokens_user_id`, `idx_refresh_tokens_token_hash`, `idx_refresh_tokens_expires_at`
- ✅ Constraint UNIQUE em `users.email` e `refresh_tokens.token_hash`
- ✅ Foreign key `refresh_tokens.user_id` → `users.id` com `ON DELETE CASCADE`

---

## T03 — Atualizar Config para adicionar variáveis de ambiente de JWT e Database

**Objetivo**: Estender `gateways/config/config.go` para incluir configurações de database e JWT.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/config/config.go`

**Implementação**:
```go
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName     string `envconfig:"APP_NAME" required:"true"`
	Environment string `envconfig:"ENVIRONMENT" required:"true"`

	RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`

	// Database
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// JWT
	JWTSecret          string        `envconfig:"JWT_SECRET" required:"true"`
	JWTExpiry          time.Duration `envconfig:"JWT_EXPIRY" default:"1h"`
	RefreshTokenExpiry time.Duration `envconfig:"REFRESH_TOKEN_EXPIRY" default:"720h"` // 30 dias
}

func ParseConfigFromEnv() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}
```

**Critério de aceite**:
- ✅ Arquivo `config.go` atualizado com campos `DatabaseURL`, `JWTSecret`, `JWTExpiry`, `RefreshTokenExpiry`
- ✅ Defaults corretos: `JWTExpiry=1h`, `RefreshTokenExpiry=720h`
- ✅ Build da aplicação funciona sem erros

---

## T04 — Criar entidades User e RefreshToken no domain

**Objetivo**: Definir as entidades de domínio `User` e `RefreshToken`.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/entities/entities.go`

**Implementação**:
```go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// User representa um usuário do sistema.
type User struct {
	ID              uuid.UUID
	Name            string
	Email           string
	PasswordHash    string
	ProfileImageURL string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RefreshToken representa um token de refresh para renovação de access tokens.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string        // SHA-256 hash do token (64 caracteres hex)
	ExpiresAt time.Time
	RevokedAt *time.Time    // NULL = válido
	CreatedAt time.Time
}
```

**Critério de aceite**:
- ✅ Arquivo `entities.go` contém struct `User` com todos os campos conforme modelo
- ✅ Arquivo `entities.go` contém struct `RefreshToken` com todos os campos conforme modelo
- ✅ Campo `RevokedAt` é um ponteiro (`*time.Time`)
- ✅ Comentários Godoc adicionados para as structs exportadas
- ✅ Build da aplicação funciona sem erros

---

## T05 — Adicionar erros de domínio para Auth

**Objetivo**: Definir erros específicos do domínio de autenticação.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/errors/errors.go`

**Implementação**:
```go
package errors

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("data conflict")
	ErrMalformedParameters = errors.New("malformed parameters")
	ErrFailedDependency    = errors.New("failed dependency")

	// Auth errors
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrTokenInvalid       = errors.New("token invalid")
)
```

**Critério de aceite**:
- ✅ Arquivo `errors.go` contém os 5 novos erros de auth
- ✅ Erros existentes não foram removidos
- ✅ Build da aplicação funciona sem erros

---

## T06 — Definir ports (interfaces) UserRepository e RefreshTokenRepository

**Objetivo**: Criar as interfaces de repositório para abstração da persistência.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/ports/repositories.go`

**Implementação**:
```go
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
)

// UserRepository define operações de persistência para usuários.
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
}

// RefreshTokenRepository define operações de persistência para refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
}
```

**Critério de aceite**:
- ✅ Arquivo `repositories.go` contém interface `UserRepository` com 3 métodos
- ✅ Arquivo `repositories.go` contém interface `RefreshTokenRepository` com 4 métodos
- ✅ Comentários Godoc adicionados para as interfaces exportadas
- ✅ Build da aplicação funciona sem erros

---

## T07 — Criar queries SQLC para UserRepository

**Objetivo**: Escrever queries SQL tipadas para operações de User usando SQLC.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/repositories/queries/users.sql`

**Implementação**:
```sql
-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, profile_image_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, profile_image_url, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT id, name, email, password_hash, profile_image_url, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;
```

**Critério de aceite**:
- ✅ Arquivo `users.sql` criado em `gateways/repositories/queries/`
- ✅ Query `CreateUser` retorna registro criado
- ✅ Query `GetUserByEmail` busca por email
- ✅ Query `GetUserByID` busca por ID
- ✅ Executar `sqlc generate` (verificar que código é gerado sem erros)

---

## T08 — Criar queries SQLC para RefreshTokenRepository

**Objetivo**: Escrever queries SQL tipadas para operações de RefreshToken usando SQLC.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/repositories/queries/refresh_tokens.sql`

**Implementação**:
```sql
-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE token_hash = $1
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;
```

**Critério de aceite**:
- ✅ Arquivo `refresh_tokens.sql` criado em `gateways/repositories/queries/`
- ✅ Query `CreateRefreshToken` retorna registro criado
- ✅ Query `GetRefreshTokenByHash` busca por hash
- ✅ Query `RevokeRefreshToken` revoga por hash
- ✅ Query `RevokeAllUserTokens` revoga todos os tokens ativos de um usuário
- ✅ Executar `sqlc generate` (verificar que código é gerado sem erros)

---

## T09 — Implementar UserRepository com SQLC (PostgreSQL)

**Objetivo**: Implementar a interface `UserRepository` usando queries geradas pelo SQLC.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/repositories/user_repository.go`

**Implementação**:
```go
package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/sqlc" // path gerado pelo SQLC
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		queries: sqlc.New(db),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	_, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		ProfileImageUrl: user.ProfileImageURL,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	})
	if err != nil {
		// Detectar constraint violation (email duplicado)
		// Ajustar conforme driver pgx retorna erros
		if errors.Is(err, sql.ErrNoRows) || /* check unique constraint */ false {
			return domainerrors.ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &entities.User{
		ID:              row.ID,
		Name:            row.Name,
		Email:           row.Email,
		PasswordHash:    row.PasswordHash,
		ProfileImageURL: row.ProfileImageUrl,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &entities.User{
		ID:              row.ID,
		Name:            row.Name,
		Email:           row.Email,
		PasswordHash:    row.PasswordHash,
		ProfileImageURL: row.ProfileImageUrl,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}
```

**Critério de aceite**:
- ✅ Arquivo `user_repository.go` criado
- ✅ Struct `UserRepository` implementa interface `ports.UserRepository`
- ✅ Erros de domínio são retornados corretamente (ErrNotFound, ErrEmailAlreadyExists)
- ✅ Detecção de unique constraint violation implementada (verificar com driver pgx)
- ✅ Build da aplicação funciona sem erros

---

## T10 — Implementar RefreshTokenRepository com SQLC (PostgreSQL)

**Objetivo**: Implementar a interface `RefreshTokenRepository` usando queries geradas pelo SQLC.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/repositories/refresh_token_repository.go`

**Implementação**:
```go
package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories/sqlc"
)

type RefreshTokenRepository struct {
	queries *sqlc.Queries
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		queries: sqlc.New(db),
	}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	_, err := r.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	})
	return err
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	row, err := r.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrTokenInvalid
		}
		return nil, err
	}
	
	var revokedAt *time.Time
	if row.RevokedAt.Valid {
		revokedAt = &row.RevokedAt.Time
	}
	
	return &entities.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		RevokedAt: revokedAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *RefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	return r.queries.RevokeRefreshToken(ctx, tokenHash)
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.RevokeAllUserTokens(ctx, userID)
}
```

**Critério de aceite**:
- ✅ Arquivo `refresh_token_repository.go` criado
- ✅ Struct `RefreshTokenRepository` implementa interface `ports.RefreshTokenRepository`
- ✅ Conversão correta de `sql.NullTime` para `*time.Time`
- ✅ Erros de domínio retornados corretamente (ErrTokenInvalid)
- ✅ Build da aplicação funciona sem erros

---

## T11 — Criar helper para geração e validação de JWT

**Objetivo**: Criar utilitário para geração e parsing de JWT tokens.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/auth/jwt.go`

**Implementação**:
```go
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secret []byte
	expiry time.Duration
}

func NewJWTManager(secret string, expiry time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// GenerateToken gera um JWT com o user ID no claim "sub".
func (m *JWTManager) GenerateToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken valida e extrai o user ID de um JWT.
func (m *JWTManager) ParseToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return m.secret, nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return uuid.Nil, err
		}
		return userID, nil
	}

	return uuid.Nil, errors.New("invalid token claims")
}
```

**Critério de aceite**:
- ✅ Arquivo `jwt.go` criado em `gateways/auth/`
- ✅ Struct `JWTManager` implementa `GenerateToken` e `ParseToken`
- ✅ JWT usa algoritmo HS256
- ✅ Claims incluem `sub`, `iat`, `exp`
- ✅ Comentários Godoc adicionados
- ✅ Build da aplicação funciona sem erros

---

## T12 — Criar helper para geração e hash de refresh tokens

**Objetivo**: Criar utilitário para geração de refresh tokens e hash SHA-256.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/auth/refresh_token.go`

**Implementação**:
```go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const refreshTokenLength = 32 // 32 bytes = 256 bits

// GenerateRefreshToken gera um token aleatório de 256 bits (base64 encoded).
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, refreshTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashRefreshToken retorna o hash SHA-256 de um refresh token (hex string).
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
```

**Critério de aceite**:
- ✅ Arquivo `refresh_token.go` criado em `gateways/auth/`
- ✅ Função `GenerateRefreshToken` gera token de 256 bits (base64 encoded)
- ✅ Função `HashRefreshToken` retorna SHA-256 hash em formato hexadecimal (64 chars)
- ✅ Comentários Godoc adicionados
- ✅ Build da aplicação funciona sem erros

---

## T13 — Implementar use case RegisterUC

**Objetivo**: Criar o use case de registro de usuário com validações de domínio.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/uc_register.go`

**Implementação**:
```go
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type RegisterOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // segundos
}

type RegisterUC struct {
	userRepo         ports.UserRepository
	refreshTokenRepo ports.RefreshTokenRepository
	jwtManager       *auth.JWTManager
	tokenExpiry      time.Duration
}

func NewRegisterUC(
	userRepo ports.UserRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtManager *auth.JWTManager,
	tokenExpiry time.Duration,
) *RegisterUC {
	return &RegisterUC{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		tokenExpiry:      tokenExpiry,
	}
}

// Execute registra um novo usuário e retorna tokens de autenticação.
func (uc *RegisterUC) Execute(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	// 1. Validações de domínio
	if len(input.Password) < 8 {
		return RegisterOutput{}, domainerrors.ErrMalformedParameters
	}

	// 2. Verificar se email já existe
	_, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err == nil {
		// Usuário encontrado = email duplicado
		return RegisterOutput{}, domainerrors.ErrEmailAlreadyExists
	}
	if err != domainerrors.ErrNotFound {
		// Erro inesperado
		return RegisterOutput{}, err
	}

	// 3. Hashear senha (bcrypt cost 12)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return RegisterOutput{}, err
	}

	// 4. Criar usuário
	now := time.Now()
	user := &entities.User{
		ID:              uuid.New(),
		Name:            input.Name,
		Email:           input.Email,
		PasswordHash:    string(passwordHash),
		ProfileImageURL: "/assets/avatars/default.png",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return RegisterOutput{}, err
	}

	// 5. Gerar refresh token
	refreshTokenPlain, err := auth.GenerateRefreshToken()
	if err != nil {
		return RegisterOutput{}, err
	}
	refreshTokenHash := auth.HashRefreshToken(refreshTokenPlain)

	refreshToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: now.Add(uc.tokenExpiry),
		CreatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return RegisterOutput{}, err
	}

	// 6. Gerar JWT
	accessToken, err := uc.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return RegisterOutput{}, err
	}

	return RegisterOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenPlain,
		ExpiresIn:    3600, // 1 hora
	}, nil
}
```

**Critério de aceite**:
- ✅ Arquivo `uc_register.go` criado em `domain/auth/`
- ✅ Use case valida senha mínima de 8 caracteres
- ✅ Use case verifica duplicação de email antes de criar
- ✅ Senha hasheada com bcrypt cost 12
- ✅ Refresh token gerado e hasheado (SHA-256) antes de salvar
- ✅ JWT gerado com user ID no claim `sub`
- ✅ Comentários Godoc adicionados
- ✅ Build da aplicação funciona sem erros

---

## T14 — Implementar use case LoginUC

**Objetivo**: Criar o use case de login com autenticação de credenciais.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/uc_login.go`

**Implementação**:
```go
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type LoginUC struct {
	userRepo         ports.UserRepository
	refreshTokenRepo ports.RefreshTokenRepository
	jwtManager       *auth.JWTManager
	tokenExpiry      time.Duration
}

func NewLoginUC(
	userRepo ports.UserRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtManager *auth.JWTManager,
	tokenExpiry time.Duration,
) *LoginUC {
	return &LoginUC{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		tokenExpiry:      tokenExpiry,
	}
}

// Execute autentica um usuário e retorna tokens.
func (uc *LoginUC) Execute(ctx context.Context, input LoginInput) (LoginOutput, error) {
	// 1. Buscar usuário por email
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if err == domainerrors.ErrNotFound {
			return LoginOutput{}, domainerrors.ErrInvalidCredentials
		}
		return LoginOutput{}, err
	}

	// 2. Comparar senha
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return LoginOutput{}, domainerrors.ErrInvalidCredentials
	}

	// 3. Gerar refresh token
	refreshTokenPlain, err := auth.GenerateRefreshToken()
	if err != nil {
		return LoginOutput{}, err
	}
	refreshTokenHash := auth.HashRefreshToken(refreshTokenPlain)

	now := time.Now()
	refreshToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: now.Add(uc.tokenExpiry),
		CreatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return LoginOutput{}, err
	}

	// 4. Gerar JWT
	accessToken, err := uc.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenPlain,
		ExpiresIn:    3600,
	}, nil
}
```

**Critério de aceite**:
- ✅ Arquivo `uc_login.go` criado em `domain/auth/`
- ✅ Use case busca usuário por email
- ✅ Use case retorna `ErrInvalidCredentials` se email não encontrado ou senha incorreta
- ✅ Senha comparada com bcrypt
- ✅ Refresh token gerado e hasheado antes de salvar
- ✅ JWT gerado com user ID
- ✅ Comentários Godoc adicionados
- ✅ Build da aplicação funciona sem erros

---

## T15 — Implementar use case RefreshTokenUC

**Objetivo**: Criar o use case de renovação de access token usando refresh token.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/uc_refresh_token.go`

**Implementação**:
```go
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

type RefreshTokenInput struct {
	RefreshToken string
}

type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type RefreshTokenUC struct {
	refreshTokenRepo ports.RefreshTokenRepository
	jwtManager       *auth.JWTManager
	tokenExpiry      time.Duration
}

func NewRefreshTokenUC(
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtManager *auth.JWTManager,
	tokenExpiry time.Duration,
) *RefreshTokenUC {
	return &RefreshTokenUC{
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		tokenExpiry:      tokenExpiry,
	}
}

// Execute valida o refresh token e gera novos tokens.
func (uc *RefreshTokenUC) Execute(ctx context.Context, input RefreshTokenInput) (RefreshTokenOutput, error) {
	// 1. Hashear token recebido
	tokenHash := auth.HashRefreshToken(input.RefreshToken)

	// 2. Buscar token no DB
	storedToken, err := uc.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return RefreshTokenOutput{}, err // ErrTokenInvalid se não encontrado
	}

	// 3. Validar se não está revogado
	if storedToken.RevokedAt != nil {
		return RefreshTokenOutput{}, domainerrors.ErrTokenRevoked
	}

	// 4. Validar se não expirou
	if time.Now().After(storedToken.ExpiresAt) {
		return RefreshTokenOutput{}, domainerrors.ErrTokenExpired
	}

	// 5. Revogar token antigo
	if err := uc.refreshTokenRepo.RevokeByTokenHash(ctx, tokenHash); err != nil {
		return RefreshTokenOutput{}, err
	}

	// 6. Gerar novo refresh token
	newRefreshTokenPlain, err := auth.GenerateRefreshToken()
	if err != nil {
		return RefreshTokenOutput{}, err
	}
	newRefreshTokenHash := auth.HashRefreshToken(newRefreshTokenPlain)

	now := time.Now()
	newRefreshToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    storedToken.UserID,
		TokenHash: newRefreshTokenHash,
		ExpiresAt: now.Add(uc.tokenExpiry),
		CreatedAt: now,
	}
	if err := uc.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return RefreshTokenOutput{}, err
	}

	// 7. Gerar novo JWT
	accessToken, err := uc.jwtManager.GenerateToken(storedToken.UserID)
	if err != nil {
		return RefreshTokenOutput{}, err
	}

	return RefreshTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenPlain,
		ExpiresIn:    3600,
	}, nil
}
```

**Critério de aceite**:
- ✅ Arquivo `uc_refresh_token.go` criado em `domain/auth/`
- ✅ Use case valida se token não está revogado
- ✅ Use case valida se token não expirou
- ✅ Use case revoga token antigo antes de gerar novo (rotation)
- ✅ Novo refresh token gerado e hasheado
- ✅ Novo JWT gerado
- ✅ Comentários Godoc adicionados
- ✅ Build da aplicação funciona sem erros

---

## T16 — Implementar use case LogoutUC

**Objetivo**: Criar o use case de logout que revoga o refresh token.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/uc_logout.go`

**Implementação**:
```go
package auth

import (
	"context"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

type LogoutInput struct {
	RefreshToken string
}

type LogoutOutput struct {
	// Sem output (204 No Content)
}

type LogoutUC struct {
	refreshTokenRepo ports.RefreshTokenRepository
}

func NewLogoutUC(refreshTokenRepo ports.RefreshTokenRepository) *LogoutUC {
	return &LogoutUC{
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Execute revoga o refresh token (logout).
func (uc *LogoutUC) Execute(ctx context.Context, input LogoutInput) (LogoutOutput, error) {
	// 1. Hashear token recebido
	tokenHash := auth.HashRefreshToken(input.RefreshToken)

	// 2. Revogar token (idempotente)
	if err := uc.refreshTokenRepo.RevokeByTokenHash(ctx, tokenHash); err != nil {
		return LogoutOutput{}, err
	}

	return LogoutOutput{}, nil
}
```

**Critério de aceite**:
- ✅ Arquivo `uc_logout.go` criado em `domain/auth/`
- ✅ Use case revoga refresh token
- ✅ Operação é idempotente (não retorna erro se token já revogado)
- ✅ Comentários Godoc adicionados
- ✅ Build da aplicação funciona sem erros

---

## T17 — Implementar AuthHandler (HTTP handlers para os 4 endpoints)

**Objetivo**: Criar os handlers HTTP para os endpoints de auth.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/http/handler_auth.go`

**Implementação**:
```go
package service

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
)

type AuthHandler struct {
	registerUC     *auth.RegisterUC
	loginUC        *auth.LoginUC
	refreshTokenUC *auth.RefreshTokenUC
	logoutUC       *auth.LogoutUC
	validate       *validator.Validate
}

func NewAuthHandler(
	registerUC *auth.RegisterUC,
	loginUC *auth.LoginUC,
	refreshTokenUC *auth.RefreshTokenUC,
	logoutUC *auth.LogoutUC,
	validate *validator.Validate,
) *AuthHandler {
	return &AuthHandler{
		registerUC:     registerUC,
		loginUC:        loginUC,
		refreshTokenUC: refreshTokenUC,
		logoutUC:       logoutUC,
		validate:       validate,
	}
}

// Register lida com POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	output, err := h.registerUC.Execute(r.Context(), auth.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		switch err {
		case domainerrors.ErrEmailAlreadyExists:
			respondError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "An account with this email already exists.")
		case domainerrors.ErrMalformedParameters:
			respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		default:
			respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	respondSuccess(w, http.StatusCreated, output)
}

// Login lida com POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	output, err := h.loginUC.Execute(r.Context(), auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		if err == domainerrors.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect.")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	respondSuccess(w, http.StatusOK, output)
}

// RefreshToken lida com POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	output, err := h.refreshTokenUC.Execute(r.Context(), auth.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		switch err {
		case domainerrors.ErrTokenInvalid, domainerrors.ErrTokenExpired, domainerrors.ErrTokenRevoked:
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token.")
		default:
			respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		}
		return
	}

	respondSuccess(w, http.StatusOK, output)
}

// Logout lida com POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// TODO: Extrair userID do JWT (middleware de autenticação pendente)
	
	var req struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Request body is invalid.")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		return
	}

	_, err := h.logoutUC.Execute(r.Context(), auth.LogoutInput{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions
func respondSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": data,
	})
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
	})
}
```

**Critério de aceite**:
- ✅ Arquivo `handler_auth.go` criado em `gateways/http/`
- ✅ Struct `AuthHandler` com os 4 métodos (Register, Login, RefreshToken, Logout)
- ✅ Validação de input com `validator/v10`
- ✅ Response wrapper correto: `{ "data": {...} }` em sucesso, `{ "code": "...", "message": "..." }` em erro
- ✅ Mapeamento de erros de domínio para HTTP status codes
- ✅ Status codes corretos: 201 (register), 200 (login/refresh), 204 (logout), 401/409/422 (erros)
- ✅ TODO adicionado para middleware de autenticação no Logout
- ✅ Build da aplicação funciona sem erros

---

## T18 — Atualizar Router para registrar rotas de Auth

**Objetivo**: Adicionar as rotas de auth ao router HTTP.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/http/router.go`

**Implementação**:
```go
package service

import (
	"github.com/go-chi/chi/v5"
)

type ServiceRouter struct {
	authHandler *AuthHandler
}

func NewServiceRouter(authHandler *AuthHandler) ServiceRouter {
	return ServiceRouter{
		authHandler: authHandler,
	}
}

func (s ServiceRouter) Pattern() string {
	return "/api/v1"
}

func (s ServiceRouter) Router(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", s.authHandler.Register)
		r.Post("/login", s.authHandler.Login)
		r.Post("/refresh", s.authHandler.RefreshToken)
		r.Post("/logout", s.authHandler.Logout) // TODO: adicionar middleware de autenticação
	})
}
```

**Critério de aceite**:
- ✅ Arquivo `router.go` atualizado
- ✅ Rotas registradas em `/api/v1/auth/*`
- ✅ TODO adicionado para middleware de autenticação no /logout
- ✅ Build da aplicação funciona sem erros

---

## T19 — Registrar dependências no main.go (fx.Provide)

**Objetivo**: Configurar DI para injetar todas as dependências via Fx.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/cmd/kinetria/api/main.go`

**Implementação**:
```go
package main

import (
	"database/sql"
	"log"

	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"go.uber.org/fx"

	"github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/config"
	gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
	httpservice "github.com/kinetria/kinetria-back/internal/kinetria/gateways/http"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/repositories"
)

var (
	AppName     = "kinetria"
	BuildCommit = "undefined"
	BuildTag    = "undefined"
	BuildTime   = "undefined"
)

func main() {
	fx.New(
		fx.Provide(
			// 1. Config
			config.ParseConfigFromEnv,

			// 2. Database connection pool
			func(cfg config.Config) (*sql.DB, error) {
				db, err := sql.Open("pgx", cfg.DatabaseURL)
				if err != nil {
					return nil, err
				}
				return db, db.Ping()
			},

			// 3. JWT Manager
			func(cfg config.Config) *gatewayauth.JWTManager {
				return gatewayauth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)
			},

			// 4. Repositories
			repositories.NewUserRepository,
			repositories.NewRefreshTokenRepository,

			// 5. Use Cases
			func(userRepo *repositories.UserRepository, refreshTokenRepo *repositories.RefreshTokenRepository, jwtMgr *gatewayauth.JWTManager, cfg config.Config) *auth.RegisterUC {
				return auth.NewRegisterUC(userRepo, refreshTokenRepo, jwtMgr, cfg.RefreshTokenExpiry)
			},
			func(userRepo *repositories.UserRepository, refreshTokenRepo *repositories.RefreshTokenRepository, jwtMgr *gatewayauth.JWTManager, cfg config.Config) *auth.LoginUC {
				return auth.NewLoginUC(userRepo, refreshTokenRepo, jwtMgr, cfg.RefreshTokenExpiry)
			},
			func(refreshTokenRepo *repositories.RefreshTokenRepository, jwtMgr *gatewayauth.JWTManager, cfg config.Config) *auth.RefreshTokenUC {
				return auth.NewRefreshTokenUC(refreshTokenRepo, jwtMgr, cfg.RefreshTokenExpiry)
			},
			auth.NewLogoutUC,

			// 6. Validator
			validator.New,

			// 7. HTTP Handlers
			httpservice.NewAuthHandler,

			// 8. HTTP Router
			httpservice.NewServiceRouter,
		),

		fx.Invoke(func(router httpservice.ServiceRouter) {
			// TODO: Inicializar servidor HTTP quando módulos xhttp estiverem disponíveis
			log.Printf("Router registrado: %s", router.Pattern())
		}),
	).Run()
}
```

**Critério de aceite**:
- ✅ Arquivo `main.go` atualizado com todos os `fx.Provide`
- ✅ Config injetado
- ✅ Database pool criado e conectado
- ✅ JWTManager injetado
- ✅ Repositories injetados
- ✅ Use cases injetados com dependências corretas
- ✅ Handler e Router injetados
- ✅ TODO adicionado para inicializar servidor HTTP
- ✅ Build da aplicação funciona sem erros: `go build ./cmd/kinetria/api`

---

## T20 — Criar testes unitários para RegisterUC (table-driven)

**Objetivo**: Criar testes unitários para o use case de registro.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/uc_register_test.go`

**Implementação**:
```go
package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/auth"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
	domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
	"github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

// Mock repositories
type mockUserRepo struct {
	users map[string]*entities.User
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error {
	if _, exists := m.users[user.Email]; exists {
		return domainerrors.ErrEmailAlreadyExists
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, domainerrors.ErrNotFound
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	return nil, domainerrors.ErrNotFound
}

type mockRefreshTokenRepo struct {
	tokens []*entities.RefreshToken
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *entities.RefreshToken) error {
	m.tokens = append(m.tokens, token)
	return nil
}

func (m *mockRefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	return nil, domainerrors.ErrTokenInvalid
}

func (m *mockRefreshTokenRepo) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	return nil
}

func (m *mockRefreshTokenRepo) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func TestRegisterUC_Execute(t *testing.T) {
	tests := []struct {
		name        string
		input       auth.RegisterInput
		setupMocks  func(*mockUserRepo)
		wantErr     error
		checkOutput func(t *testing.T, output auth.RegisterOutput)
	}{
		{
			name: "successful registration",
			input: auth.RegisterInput{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMocks: func(repo *mockUserRepo) {},
			wantErr:    nil,
			checkOutput: func(t *testing.T, output auth.RegisterOutput) {
				if output.AccessToken == "" {
					t.Error("AccessToken should not be empty")
				}
				if output.RefreshToken == "" {
					t.Error("RefreshToken should not be empty")
				}
				if output.ExpiresIn != 3600 {
					t.Errorf("ExpiresIn = %d, want 3600", output.ExpiresIn)
				}
			},
		},
		{
			name: "email already exists",
			input: auth.RegisterInput{
				Name:     "Jane Doe",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMocks: func(repo *mockUserRepo) {
				repo.users["existing@example.com"] = &entities.User{
					ID:    uuid.New(),
					Email: "existing@example.com",
				}
			},
			wantErr: domainerrors.ErrEmailAlreadyExists,
		},
		{
			name: "password too short",
			input: auth.RegisterInput{
				Name:     "Short Pass",
				Email:    "short@example.com",
				Password: "short",
			},
			setupMocks: func(repo *mockUserRepo) {},
			wantErr:    domainerrors.ErrMalformedParameters,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{users: make(map[string]*entities.User)}
			refreshTokenRepo := &mockRefreshTokenRepo{}
			jwtManager := gatewayauth.NewJWTManager("test-secret-key-256-bits-long", time.Hour)

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo)
			}

			uc := auth.NewRegisterUC(userRepo, refreshTokenRepo, jwtManager, 720*time.Hour)

			output, err := uc.Execute(context.Background(), tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkOutput != nil && err == nil {
				tt.checkOutput(t, output)
			}
		})
	}
}
```

**Critério de aceite**:
- ✅ Arquivo `uc_register_test.go` criado
- ✅ Testes table-driven cobrindo happy path e sad paths
- ✅ Mocks de repositórios implementados
- ✅ Testes validam: registro bem-sucedido, email duplicado, senha curta
- ✅ Testes passando: `go test ./internal/kinetria/domain/auth`

---

## T21 — Criar testes unitários para LoginUC (table-driven)

**Objetivo**: Criar testes unitários para o use case de login.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/uc_login_test.go`

**Implementação**: (similar à T20, mas testando login, senha incorreta, usuário não existe)

**Critério de aceite**:
- ✅ Arquivo `uc_login_test.go` criado
- ✅ Testes table-driven cobrindo: login bem-sucedido, senha incorreta, email não existe
- ✅ Testes passando: `go test ./internal/kinetria/domain/auth`

---

## T22 — Criar testes unitários para RefreshTokenUC (table-driven)

**Objetivo**: Criar testes unitários para o use case de refresh token.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/uc_refresh_token_test.go`

**Implementação**: (similar à T20, mas testando refresh bem-sucedido, token expirado, token revogado, token inválido)

**Critério de aceite**:
- ✅ Arquivo `uc_refresh_token_test.go` criado
- ✅ Testes table-driven cobrindo: refresh bem-sucedido, token expirado, token revogado, token inválido
- ✅ Testes validam rotação de token (token antigo revogado)
- ✅ Testes passando: `go test ./internal/kinetria/domain/auth`

---

## T23 — Criar testes unitários para LogoutUC (table-driven)

**Objetivo**: Criar testes unitários para o use case de logout.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/uc_logout_test.go`

**Implementação**: (similar à T20, mas testando logout bem-sucedido, idempotência)

**Critério de aceite**:
- ✅ Arquivo `uc_logout_test.go` criado
- ✅ Testes table-driven cobrindo: logout bem-sucedido, logout idempotente
- ✅ Testes passando: `go test ./internal/kinetria/domain/auth`

---

## T24 — Documentar API de Auth (comentários Godoc)

**Objetivo**: Adicionar comentários Godoc em todas as funções exportadas do domínio e handlers.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/domain/auth/*.go`
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/http/handler_auth.go`
- `/home/runner/work/kinetria-back/kinetria-back/internal/kinetria/gateways/auth/*.go`

**Implementação**:
- Adicionar comentários no formato Godoc acima de:
  - Structs exportadas (User, RefreshToken, RegisterInput, etc.)
  - Funções exportadas (Execute, GenerateToken, etc.)
  - Interfaces exportadas (UserRepository, etc.)

**Exemplo**:
```go
// RegisterUC implementa o caso de uso de registro de novo usuário.
// Valida os dados de entrada, hasheia a senha com bcrypt (cost 12),
// cria o usuário no banco de dados e retorna tokens de autenticação.
type RegisterUC struct { ... }

// Execute registra um novo usuário no sistema.
// Retorna ErrEmailAlreadyExists se o email já estiver cadastrado.
// Retorna ErrMalformedParameters se a senha tiver menos de 8 caracteres.
func (uc *RegisterUC) Execute(ctx context.Context, input RegisterInput) (RegisterOutput, error) { ... }
```

**Critério de aceite**:
- ✅ Todos os types, funções e métodos exportados têm comentários Godoc
- ✅ Comentários descrevem claramente: propósito, inputs, outputs, erros retornados
- ✅ `go doc` funciona corretamente: `go doc github.com/kinetria/kinetria-back/internal/kinetria/domain/auth.RegisterUC`

---

## T25 — Criar .env.example com variáveis de ambiente

**Objetivo**: Documentar as variáveis de ambiente necessárias.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/.env.example`

**Implementação**:
```bash
# Application
APP_NAME=kinetria
ENVIRONMENT=development

# Database
DATABASE_URL=postgres://kinetria:secret@localhost:5432/kinetria?sslmode=disable

# JWT
# Generate with: openssl rand -hex 32
JWT_SECRET=your-256-bit-secret-here-use-openssl-rand-hex-32
JWT_EXPIRY=1h
REFRESH_TOKEN_EXPIRY=720h

# Timeouts
REQUEST_TIMEOUT=5s
```

**Critério de aceite**:
- ✅ Arquivo `.env.example` criado na raiz do repositório
- ✅ Todas as variáveis de ambiente obrigatórias documentadas
- ✅ Comentários explicando como gerar valores (ex: JWT_SECRET)
- ✅ Valores de exemplo seguros (não usar secrets reais)

---

## T26 — Criar README ou documentação de setup local

**Objetivo**: Documentar como rodar a aplicação localmente para desenvolvimento.

**Arquivos/pacotes prováveis**:
- `/home/runner/work/kinetria-back/kinetria-back/.thoughts/mvp-userflow/DEV_SETUP.md`

**Implementação**:
```markdown
# Dev Setup — mvp-userflow (Auth)

## Pré-requisitos

- Go 1.25+
- PostgreSQL 15+
- Migrate CLI (opcional, para migrations)

## Setup

1. Clone o repositório:
   \`\`\`bash
   git clone <repo-url>
   cd kinetria-back
   \`\`\`

2. Copie o `.env.example` para `.env`:
   \`\`\`bash
   cp .env.example .env
   \`\`\`

3. Gere um secret JWT:
   \`\`\`bash
   openssl rand -hex 32
   \`\`\`
   Cole o valor em `.env` na variável `JWT_SECRET`.

4. Suba o PostgreSQL (Docker Compose sugerido):
   \`\`\`bash
   docker-compose up -d postgres
   \`\`\`

5. Aplique as migrations:
   \`\`\`bash
   migrate -path migrations -database "postgres://kinetria:secret@localhost:5432/kinetria?sslmode=disable" up
   \`\`\`

6. Rode a aplicação:
   \`\`\`bash
   go run ./cmd/kinetria/api
   \`\`\`

## Testes

Execute os testes:
\`\`\`bash
go test ./...
\`\`\`

Cobertura:
\`\`\`bash
go test -cover ./internal/kinetria/domain/auth
\`\`\`

## Endpoints

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
```

**Critério de aceite**:
- ✅ Arquivo `DEV_SETUP.md` criado em `.thoughts/mvp-userflow/`
- ✅ Instruções claras de setup local
- ✅ Comandos testados e funcionando

---

## Resumo de Tarefas

| Task | Título | Tipo |
|------|--------|------|
| T01 | Adicionar dependência JWT | Setup |
| T02 | Criar migrations SQL | Database |
| T03 | Atualizar Config | Config |
| T04 | Criar entidades User e RefreshToken | Domain |
| T05 | Adicionar erros de domínio | Domain |
| T06 | Definir ports (interfaces) | Domain |
| T07 | Criar queries SQLC para UserRepository | Database |
| T08 | Criar queries SQLC para RefreshTokenRepository | Database |
| T09 | Implementar UserRepository | Gateway |
| T10 | Implementar RefreshTokenRepository | Gateway |
| T11 | Criar helper JWT | Gateway |
| T12 | Criar helper refresh token | Gateway |
| T13 | Implementar RegisterUC | Domain |
| T14 | Implementar LoginUC | Domain |
| T15 | Implementar RefreshTokenUC | Domain |
| T16 | Implementar LogoutUC | Domain |
| T17 | Implementar AuthHandler | Gateway |
| T18 | Atualizar Router | Gateway |
| T19 | Registrar dependências no main.go | Setup |
| T20 | Testes RegisterUC | Tests |
| T21 | Testes LoginUC | Tests |
| T22 | Testes RefreshTokenUC | Tests |
| T23 | Testes LogoutUC | Tests |
| T24 | Documentar API (Godoc) | Docs |
| T25 | Criar .env.example | Docs |
| T26 | Criar README de setup | Docs |

---

**Total de tarefas**: 26  
**Estimativa**: 4-6 dias de implementação (1 dev full-time)  
**Ordem sugerida**: T01 → T02 → T03 → T04 → T05 → T06 → T07 → T08 → T09 → T10 → T11 → T12 → T13 → T14 → T15 → T16 → T17 → T18 → T19 → T20-T23 (paralelo) → T24 → T25 → T26

# Dev Setup — mvp-userflow (Auth)

## Pré-requisitos

- Go 1.25+
- PostgreSQL 15+
- Docker & Docker Compose (recomendado)

## Setup

1. Clone o repositório:
   ```bash
   git clone <repo-url>
   cd kinetria-back
   ```

2. Copie o `.env.example` para `.env`:
   ```bash
   cp .env.example .env
   ```

3. Gere um secret JWT seguro (mínimo 256 bits):
   ```bash
   openssl rand -hex 32
   ```
   Cole o valor em `.env` na variável `JWT_SECRET`.

4. Suba o PostgreSQL via Docker Compose:
   ```bash
   docker-compose up -d
   ```

5. Aplique as migrations:
   ```bash
   # Usando psql diretamente
   psql postgres://kinetria:secret@localhost:5432/kinetria?sslmode=disable -f migrations/001_create_users.sql
   psql postgres://kinetria:secret@localhost:5432/kinetria?sslmode=disable -f migrations/002_create_workouts.sql
   # ... demais migrations em ordem
   ```

6. Rode a aplicação:
   ```bash
   go run ./cmd/kinetria/api
   ```

## Testes

Execute os testes unitários:
```bash
go test ./internal/kinetria/domain/auth/...
```

Todos os testes:
```bash
go test ./...
```

Cobertura:
```bash
go test -cover ./internal/kinetria/domain/auth/...
```

Build:
```bash
go build ./cmd/kinetria/api
```

## Endpoints de Auth

| Método | Path | Descrição |
|--------|------|-----------|
| `POST` | `/api/v1/auth/register` | Criar conta + retornar tokens |
| `POST` | `/api/v1/auth/login` | Autenticar + retornar tokens |
| `POST` | `/api/v1/auth/refresh` | Renovar access token |
| `POST` | `/api/v1/auth/logout` | Revogar refresh token (requer `Authorization: Bearer <token>`) |

### Health check

```
GET /health
```

### Exemplo — Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Bruno Costa","email":"user@example.com","password":"s3cr3tP@ssw0rd"}'
```

Resposta 201:
```json
{
  "data": {
    "accessToken": "<JWT>",
    "refreshToken": "<token>",
    "expiresIn": 3600
  }
}
```

### Exemplo — Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"s3cr3tP@ssw0rd"}'
```

### Exemplo — Refresh

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken":"<refresh_token>"}'
```

### Exemplo — Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"refreshToken":"<refresh_token>"}'
```

## Decisões de Segurança

- **Senha**: hashed com bcrypt cost 12
- **Access Token**: JWT HS256, 1h de validade
- **Refresh Token**: 30 dias, armazenado como SHA-256 hash no banco
- **Token rotation**: ao fazer refresh, o token antigo é revogado

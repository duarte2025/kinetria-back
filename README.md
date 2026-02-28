# Kinetria Back

Backend do aplicativo Kinetria de acompanhamento de treinos.

## Pré-requisitos

- Go 1.25+
- Docker e Docker Compose
- Make

## Desenvolvimento com Docker

### Subir o ambiente

```bash
docker-compose up -d
```

### Verificar status

```bash
docker-compose ps
curl http://localhost:8080/health
```

### Ver logs

```bash
docker-compose logs -f app
```

### Conectar ao banco de dados

```bash
docker exec -it kinetria-postgres psql -U kinetria -d kinetria
```

### Parar o ambiente

```bash
docker-compose down
```

### Resetar banco de dados (apagar volumes)

```bash
docker-compose down -v
```

## Migrations

As migrations SQL são aplicadas automaticamente ao iniciar a aplicação, tanto em desenvolvimento quanto em produção.

Os arquivos SQL estão em `internal/kinetria/gateways/migrations/` e são embarcados no binário via `embed.FS`.

**Importante**: As migrations devem estar APENAS em `internal/kinetria/gateways/migrations/`. Não criar migrations em outros diretórios.

| Arquivo | Tabela | Descrição |
|---------|--------|-----------|
| `001_create_users.sql` | `users` | Usuários do sistema |
| `002_create_workouts.sql` | `workouts` | Planos de treino |
| `003_create_exercises.sql` | `exercises` | Exercícios do treino |
| `004_create_sessions.sql` | `sessions` | Sessões de treino ativas |
| `005_create_set_records.sql` | `set_records` | Registros de séries executadas |
| `006_create_refresh_tokens.sql` | `refresh_tokens` | Tokens para autenticação JWT |
| `007_create_audit_log.sql` | `audit_log` | Log de auditoria de ações |
| `008_add_sessions_dashboard_index.sql` | `sessions` | Índice de otimização dashboard |
| `009_refactor_exercises_to_shared_library.sql` | `exercises`, `workout_exercises` | Refatora exercises para biblioteca compartilhada com relacionamento N:N |

### Como funciona

- Ao iniciar, a aplicação cria a tabela `schema_migrations` se não existir
- Cada migration é executada apenas uma vez (controle via `schema_migrations`)
- Migrations são executadas em ordem alfabética (001, 002, 003...)
- Se uma migration falhar, a aplicação não inicia

### Resetar banco de dados

Para reaplicar todas as migrations do zero:

```bash
docker-compose down -v && docker-compose up -d
```

## Estrutura de Domínio

### Entidades

- `User` — Usuários do sistema
- `Workout` — Planos de treino personalizados
- `Exercise` — Exercícios compartilhados (biblioteca)
- `WorkoutExercise` — Configuração de um exercício dentro de um treino (N:N)
- `Session` — Sessão de treino ativa
- `SetRecord` — Registro de série executada
- `RefreshToken` — Tokens para renovação de autenticação
- `AuditLog` — Log de auditoria de ações

### Relacionamento Exercises → Workouts (N:N)

A partir da migration 009, exercises segue um modelo de **biblioteca compartilhada**:

- **`exercises`** — Tabela de exercícios base (compartilhados entre workouts)
  - `id`, `name`, `description`, `thumbnail_url`, `muscles`, timestamps
  - Múltiplos workouts podem referenciar o mesmo exercise

- **`workout_exercises`** — Tabela de junção (N:N)
  - Conecta `workout_id` com `exercise_id`
  - Armazena configurações específicas: `sets`, `reps`, `rest_time`, `weight`, `order_index`
  - Permite que o mesmo exercício seja usado em diferentes workouts com diferentes configurações

**Diagrama:**
```
workouts (1) ----< workout_exercises >---- (1) exercises
                     (configurações)
                  (sets, reps, weight, etc)
```

**Benefícios:**
- ✅ Reutilização de exercises entre workouts
- ✅ Economia de armazenamento (exercise definido uma vez)
- ✅ Facilita manutenção de biblioteca central
- ✅ Permite ajustar configurações por workout

### Value Objects

- `WorkoutType` — Tipo de treino: FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO
- `WorkoutIntensity` — Intensidade: BAIXA, MODERADA, ALTA
- `SessionStatus` — Status da sessão: active, completed, abandoned
- `SetRecordStatus` — Status do registro: completed, skipped

## API

### Documentação Interativa (Swagger)

Acesse a documentação completa e interativa da API em:

```
http://localhost:8080/api/v1/swagger/index.html
```

A documentação Swagger permite:
- ✅ Visualizar todos os endpoints disponíveis
- ✅ Testar requisições diretamente no navegador
- ✅ Ver exemplos de request/response
- ✅ Autenticar com JWT Bearer token
- ✅ Exportar especificação OpenAPI 3.0

Para regenerar a documentação após mudanças nos handlers:
```bash
make swagger
```

### Endpoints

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/health` | Health check da aplicação |
| POST | `/api/v1/auth/register` | Registrar novo usuário |
| POST | `/api/v1/auth/login` | Login de usuário |
| POST | `/api/v1/auth/refresh` | Renovar token de acesso |
| POST | `/api/v1/auth/logout` | Logout de usuário |
| GET | `/api/v1/dashboard` | Dashboard do usuário (requer autenticação) |
| GET | `/api/v1/workouts` | Listar workouts do usuário (requer autenticação) |
| POST | `/api/v1/sessions` | Iniciar sessão de treino (requer autenticação) |
| POST | `/api/v1/sessions/{id}/sets` | Registrar série executada (requer autenticação) |
| PATCH | `/api/v1/sessions/{id}/finish` | Finalizar sessão (requer autenticação) |
| PATCH | `/api/v1/sessions/{id}/abandon` | Abandonar sessão (requer autenticação) |

### Exemplo de resposta

```bash
curl http://localhost:8080/health
```

```json
{
  "status": "healthy",
  "service": "kinetria",
  "version": "undefined"
}
```

### Dashboard

Obter dados agregados do dashboard do usuário:

```bash
# Registrar usuário
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","password":"Password123!"}' \
  | jq -r '.data.accessToken')

# Obter dashboard
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/dashboard | jq
```

**Dica**: Use o Swagger UI em `http://localhost:8080/api/v1/swagger/index.html` para testar todos os endpoints interativamente!

Ver documentação completa em `internal/kinetria/domain/dashboard/README.md`.

### Workouts

Gerenciar e consultar planos de treino do usuário.

#### GET /api/v1/workouts

Lista todos os workouts do usuário autenticado com paginação.

**Autenticação**: Obrigatória (JWT Bearer)

**Query Parameters**:
- `page` (int, opcional, default: 1, min: 1) - Número da página
- `pageSize` (int, opcional, default: 20, min: 1, max: 100) - Itens por página

**Exemplo de Requisição**:
```bash
curl -X GET "http://localhost:8080/api/v1/workouts?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Exemplo de Resposta (200 OK)**:
```json
{
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "Treino de Peito e Tríceps",
      "description": "Foco em hipertrofia com exercícios compostos",
      "type": "FORÇA",
      "intensity": "Alta",
      "duration": 45,
      "imageUrl": "https://cdn.kinetria.app/workouts/chest.jpg"
    },
    {
      "id": "b2c3d4e5-f6g7-8901-bcde-f12345678901",
      "name": "Treino de Costas",
      "description": null,
      "type": "HIPERTROFIA",
      "intensity": "Moderada",
      "duration": 60,
      "imageUrl": null
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 10,
    "total": 25,
    "totalPages": 3
  }
}
```

**Possíveis Erros**:
- `401 Unauthorized` - Token JWT inválido ou ausente
- `422 Validation Error` - Parâmetros de paginação inválidos
- `500 Internal Error` - Erro interno do servidor

---

#### GET /api/v1/workouts/{id}

Retorna detalhes de um workout específico com seus exercises.

**Autenticação**: Obrigatória (JWT Bearer)

**Path Parameters**:
- `id` (uuid, obrigatório) - ID do workout

**Exemplo de Requisição**:
```bash
curl -X GET "http://localhost:8080/api/v1/workouts/a1b2c3d4-e5f6-7890-abcd-ef1234567890" \
  -H "Authorization: Bearer $TOKEN"
```

**Exemplo de Resposta (200 OK)**:
```json
{
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "name": "Treino de Peito e Tríceps",
    "description": "Foco em hipertrofia com exercícios compostos",
    "type": "FORÇA",
    "intensity": "Alta",
    "duration": 45,
    "imageUrl": "https://cdn.kinetria.app/workouts/chest.jpg",
    "exercises": [
      {
        "id": "ex-uuid-1",
        "name": "Supino Reto",
        "thumbnailUrl": "https://cdn.kinetria.app/exercises/bench-press.jpg",
        "sets": 4,
        "reps": "8-12",
        "muscles": ["Peito", "Tríceps", "Ombro"],
        "restTime": 90,
        "weight": 80000
      },
      {
        "id": "ex-uuid-2",
        "name": "Tríceps na Polia",
        "thumbnailUrl": "https://cdn.kinetria.app/exercises/tricep-pushdown.jpg",
        "sets": 3,
        "reps": "12-15",
        "muscles": ["Tríceps"],
        "restTime": 60,
        "weight": 40000
      }
    ]
  }
}
```

**Possíveis Erros**:
- `401 Unauthorized` - Token JWT inválido ou ausente
- `404 Not Found` - Workout não encontrado ou não pertence ao usuário
- `422 Validation Error` - ID inválido (não é UUID)
- `500 Internal Error` - Erro interno do servidor

**Notas**:
- `weight` é retornado em gramas (ex: 80000g = 80kg)
- Campos opcionais podem ser `null` (ex: `description`, `imageUrl`, `thumbnailUrl`, `weight`)
- `exercises` pode ser array vazio se o workout não tiver exercises cadastrados
- `reps` pode ser um número fixo ou range (ex: "8-12")
- `muscles` é uma lista de strings com os músculos trabalhados
- `restTime` é o tempo de descanso recomendado em segundos

## Testes

```bash
# Rodar todos os testes unitários
make test

# Rodar com cobertura
make test-coverage

# Rodar testes de integração (requer Docker Compose rodando)
INTEGRATION_TEST=1 make test
```

## Estrutura do Projeto

```
kinetria-back/
├── cmd/kinetria/api/       # Entrypoint da aplicação
├── internal/kinetria/
│   ├── domain/
│   │   ├── constants/      # Constantes de defaults e validação
│   │   ├── entities/       # Entidades de domínio
│   │   ├── errors/         # Erros de domínio
│   │   ├── ports/          # Interfaces (contratos)
│   │   └── vos/            # Value Objects
│   └── gateways/
│       ├── config/         # Configuração via variáveis de ambiente
│       ├── http/health/    # Handler de health check
│       └── repositories/   # Pool de conexão com banco de dados
└── migrations/             # Migrations SQL
```

## Variáveis de Ambiente

Ver `.env.example` para a lista completa de variáveis necessárias.

## Comandos Disponíveis

```bash
make help              # Mostra todos os comandos disponíveis
make run               # Executa a aplicação
make build             # Compila a aplicação
make test              # Executa os testes
make test-coverage     # Executa os testes com cobertura
make lint              # Executa o linter
make sqlc              # Gera código a partir das queries SQL
make mocks             # Gera mocks das interfaces
make tidy              # Organiza as dependências
make deps              # Instala as dependências
```

## Desenvolvimento Local

Este projeto segue as convenções definidas em `.kiro/instructions/golang-hexagonal.md`.

### Adicionando uma nova feature

1. Defina as entidades em `internal/kinetria/domain/entities/`
2. Crie as interfaces (ports) em `internal/kinetria/domain/ports/`
3. Implemente o use case em `internal/kinetria/domain/{feature}/`
4. Implemente os adaptadores em `internal/kinetria/gateways/`
5. Registre as dependências no `cmd/kinetria/api/main.go`

## Arquitetura

O projeto segue os princípios da arquitetura hexagonal (Ports and Adapters):

- **Domain**: Contém a lógica de negócio pura, independente de frameworks
- **Ports**: Interfaces que definem contratos entre camadas
- **Gateways**: Implementações concretas dos ports (HTTP, DB, eventos, etc)
- **Use Cases**: Orquestração da lógica de negócio

## Injeção de Dependências

Utilizamos `go.uber.org/fx` para injeção de dependências. Todas as dependências são registradas no `main.go`.

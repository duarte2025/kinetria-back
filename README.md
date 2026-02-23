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

As migrations SQL estão em `migrations/` e são aplicadas automaticamente quando o container PostgreSQL inicia pela primeira vez.

| Arquivo | Tabela | Descrição |
|---------|--------|-----------|
| `001_create_users.sql` | `users` | Usuários do sistema |
| `002_create_workouts.sql` | `workouts` | Planos de treino |
| `003_create_exercises.sql` | `exercises` | Exercícios do treino |
| `004_create_sessions.sql` | `sessions` | Sessões de treino ativas |
| `005_create_set_records.sql` | `set_records` | Registros de séries executadas |
| `006_create_refresh_tokens.sql` | `refresh_tokens` | Tokens para autenticação JWT |
| `007_create_audit_log.sql` | `audit_log` | Log de auditoria de ações |

Para reaplicar as migrations: `docker-compose down -v && docker-compose up -d`

## Estrutura de Domínio

### Entidades

- `User` — Usuários do sistema
- `Workout` — Planos de treino personalizados
- `Exercise` — Exercícios de um treino
- `Session` — Sessão de treino ativa
- `SetRecord` — Registro de série executada
- `RefreshToken` — Tokens para renovação de autenticação
- `AuditLog` — Log de auditoria de ações

### Value Objects

- `WorkoutType` — Tipo de treino: FORÇA, HIPERTROFIA, MOBILIDADE, CONDICIONAMENTO
- `WorkoutIntensity` — Intensidade: BAIXA, MODERADA, ALTA
- `SessionStatus` — Status da sessão: active, completed, abandoned
- `SetRecordStatus` — Status do registro: completed, skipped

## API

### Endpoints

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/health` | Health check da aplicação |

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

# Próximos Passos

O projeto foi inicializado com sucesso seguindo a arquitetura hexagonal! 🎉

## Estrutura Criada

✅ Estrutura de diretórios completa
✅ Configuração base (config, errors, ports)
✅ Main.go com Fx (injeção de dependências)
✅ Repositório base com suporte a SQLC
✅ HTTP handlers e routers base
✅ Event handlers e publishers base
✅ Makefile com comandos úteis
✅ Configuração do golangci-lint
✅ .gitignore e .env.example

## O que fazer agora?

### 1. Configurar variáveis de ambiente
```bash
cp .env.example .env
# Edite o .env com suas configurações
```

### 2. Adicionar módulos compartilhados (pkg/)

O projeto está configurado para usar módulos compartilhados que ainda não foram criados. Você precisará:

- `pkg/xfx` - Módulo base do Fx
- `pkg/xbuild` - Informações de build
- `pkg/xlog` - Logging estruturado
- `pkg/xtelemetry` - OpenTelemetry
- `pkg/xhttp` - HTTP server e utilitários
- `pkg/xhealth` - Health checks
- `pkg/xuc` - Interface UseCase[TInput, TOutput]
- `pkg/xapiclient` - Cliente HTTP base

Você pode criar esses pacotes ou usar bibliotecas existentes.

### 3. Criar sua primeira feature

Siga o fluxo recomendado:

1. **Defina a entidade** em `internal/kinetria/domain/entities/`
   ```go
   type User struct {
       ID        UserID
       Name      string
       Email     string
       CreatedAt time.Time
   }
   ```

2. **Crie os Value Objects** em `internal/kinetria/domain/vos/`
   ```go
   type UserStatus string
   const (
       UserStatusActive UserStatus = "active"
   )
   ```

3. **Defina as interfaces (ports)** em `internal/kinetria/domain/ports/`
   ```go
   //go:generate moq -stub -pkg mocks -out mocks/repositories.go . UserRepository
   type UserRepository interface {
       FindByID(ctx context.Context, id UserID) (entities.User, error)
   }
   ```

4. **Implemente o Use Case** em `internal/kinetria/domain/{feature}/`
   ```go
   type CreateUserUC struct {
       repository ports.UserRepository
   }
   ```

5. **Implemente o repositório** em `internal/kinetria/gateways/repositories/`
   - Adicione queries SQL em `queries/queries.sql`
   - Execute `make sqlc` para gerar o código
   - Implemente os métodos do port

6. **Crie o handler HTTP** em `internal/kinetria/gateways/http/`
   - Implemente o handler
   - Adicione a rota no router

7. **Registre no main.go**
   ```go
   fx.Provide(
       config.ParseConfigFromEnv,
       fx.Annotate(repository.NewRepository, fx.As(new(ports.UserRepository))),
       usecase.NewCreateUserUC,
       handler.NewHandler,
       xhttp.AsRouter(handler.NewRouter),
   )
   ```

### 4. Configurar banco de dados

Se for usar PostgreSQL:

1. Crie as migrations em `migrations/`
2. Configure a conexão no `config.go`
3. Adicione o pool de conexão no `main.go`

### 5. Gerar mocks

Quando tiver interfaces definidas:
```bash
make mocks
```

### 6. Executar testes

```bash
make test
make test-coverage
```

### 7. Executar a aplicação

```bash
make run
# ou
make build
./bin/kinetria
```

## Comandos Úteis

```bash
make help              # Ver todos os comandos
make deps              # Instalar dependências
make tidy              # Organizar dependências
make sqlc              # Gerar código SQLC
make mocks             # Gerar mocks
make lint              # Executar linter
make test              # Executar testes
make build             # Compilar
make run               # Executar
```

## Referências

- Convenções: `.kiro/instructions/golang-hexagonal.md`
- Exemplo de use case: `internal/kinetria/domain/example/uc_example.go`

## Dicas

- Sempre siga o fluxo: Domain → Ports → Use Cases → Gateways → Main
- Use `//go:generate moq` para gerar mocks automaticamente
- Mantenha a lógica de negócio no domain, sem dependências externas
- Use o padrão de erros de domínio para comunicação entre camadas
- Sempre adicione índices nas colunas usadas em WHERE das queries SQL

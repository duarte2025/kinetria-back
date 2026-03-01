# Testes de Integração

Testes de integração end-to-end usando **testcontainers** para simular todo o fluxo do backend com banco de dados PostgreSQL real.

## Estrutura

```
internal/kinetria/tests/
├── setup_test.go      # Configuração do servidor de testes com testcontainers
├── auth_test.go       # Testes de autenticação (register, login, refresh, logout)
└── dashboard_test.go  # Testes de dashboard
```

## Pré-requisitos

- Docker rodando (testcontainers precisa do Docker)
- Go 1.25+

## Executar Testes

### Todos os testes de integração

```bash
cd internal/kinetria/tests
go test -v
```

### Teste específico

```bash
go test -v -run TestAuthEndpoints
go test -v -run TestDashboardEndpoint
```

### Com timeout maior (para CI/CD)

```bash
go test -v -timeout 5m
```

## O que é testado

### Auth (`auth_test.go`)
- ✅ Registro de novo usuário
- ✅ Registro com email duplicado (409 Conflict)
- ✅ Login com credenciais válidas
- ✅ Login com credenciais inválidas (401 Unauthorized)
- ✅ Refresh token
- ✅ Logout (com autenticação JWT)

### Dashboard (`dashboard_test.go`)
- ✅ Obter dashboard completo
- ✅ Dashboard sem treino do dia (todayWorkout = null)
- ✅ Progresso da semana (7 dias)
- ✅ Estatísticas (calorias, tempo total)
- ✅ Acesso sem autenticação (401 Unauthorized)

## Como funciona

### Testcontainers

Cada teste cria um container PostgreSQL isolado:

```go
pgContainer, err := postgres.Run(ctx,
    "postgres:16-alpine",
    postgres.WithDatabase("kinetria_test"),
    postgres.WithUsername("test"),
    postgres.WithPassword("test"),
    testcontainers.WithWaitStrategy(
        wait.ForLog("database system is ready to accept connections").
            WithOccurrence(2).
            WithStartupTimeout(30*time.Second)),
)
```

### Setup do Servidor

`SetupTestServer()` cria:
- ✅ Container PostgreSQL
- ✅ Pool de conexões
- ✅ Migrations aplicadas
- ✅ Todos os use cases e handlers
- ✅ Router HTTP completo
- ✅ HTTP test server

### Cleanup

Cada teste limpa recursos automaticamente:

```go
defer ts.Cleanup(t)
```

Isso garante:
- ✅ Container PostgreSQL parado e removido
- ✅ Conexões fechadas
- ✅ HTTP server desligado

## Vantagens

### Testes Reais
- ✅ Banco de dados PostgreSQL real (não mocks)
- ✅ Migrations aplicadas como em produção
- ✅ HTTP requests reais
- ✅ JWT authentication real

### Isolamento
- ✅ Cada teste roda em container isolado
- ✅ Sem interferência entre testes
- ✅ Sem necessidade de limpar dados manualmente

### CI/CD Ready
- ✅ Funciona em qualquer ambiente com Docker
- ✅ Não precisa de banco de dados externo
- ✅ Containers são criados e destruídos automaticamente

### Confiança
- ✅ Testa o sistema completo (end-to-end)
- ✅ Detecta problemas de integração
- ✅ Valida fluxos reais de usuário

## Performance

Tempo médio de execução:
- Setup inicial (pull da imagem): ~10-30s (primeira vez)
- Cada teste: ~2-5s
- Total (todos os testes): ~30-60s

**Dica**: A imagem do PostgreSQL é cacheada após o primeiro pull.

## Troubleshooting

### Docker não está rodando
```
Error: Cannot connect to the Docker daemon
```
**Solução**: Inicie o Docker Desktop ou Docker daemon.

### Timeout ao criar container
```
Error: context deadline exceeded
```
**Solução**: Aumente o timeout ou verifique sua conexão com a internet.

### Porta já em uso
Testcontainers usa portas aleatórias, então isso raramente acontece. Se ocorrer, reinicie o Docker.

## Adicionar Novos Testes

1. Crie um novo arquivo `*_test.go` em `internal/kinetria/tests/`
2. Use `SetupTestServer(t)` para criar o ambiente
3. Faça requests HTTP usando `ts.URL("/endpoint")`
4. Use `ts.AuthHeader(token)` para requests autenticados
5. Limpe com `defer ts.Cleanup(t)`

Exemplo:

```go
func TestMyNewEndpoint(t *testing.T) {
    ts := SetupTestServer(t)
    defer ts.Cleanup(t)

    // Registrar usuário
    payload := map[string]string{
        "name":     "Test User",
        "email":    "test@example.com",
        "password": "Password123!",
    }
    body, _ := json.Marshal(payload)

    resp, err := http.Post(ts.URL("/auth/register"), "application/json", bytes.NewBuffer(body))
    require.NoError(t, err)
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    accessToken := result["data"].(map[string]interface{})["accessToken"].(string)

    // Fazer request autenticado
    req, _ := http.NewRequest("GET", ts.URL("/my-endpoint"), nil)
    req.Header = ts.AuthHeader(accessToken)

    resp, err = http.DefaultClient.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Integração com CI/CD

### GitHub Actions

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      - name: Run integration tests
        run: |
          cd internal/kinetria/tests
          go test -v -timeout 5m
```

### GitLab CI

```yaml
integration-tests:
  image: golang:1.25
  services:
    - docker:dind
  script:
    - cd internal/kinetria/tests
    - go test -v -timeout 5m
```

## Referências

- [Testcontainers Go](https://golang.testcontainers.org/)
- [Testcontainers Postgres Module](https://golang.testcontainers.org/modules/postgres/)
- [Stretchr Testify](https://github.com/stretchr/testify)

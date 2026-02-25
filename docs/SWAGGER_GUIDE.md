# Swagger/OpenAPI Documentation Guide

## Acessando a Documentação

Acesse a interface interativa do Swagger em:

```
http://localhost:8080/api/v1/swagger/index.html
```

## Como Usar

### 1. Visualizar Endpoints

- Todos os endpoints estão organizados por tags (auth, dashboard, workouts, sessions, health)
- Clique em qualquer endpoint para ver detalhes
- Veja exemplos de request/response

### 2. Testar Endpoints Públicos

Endpoints que não requerem autenticação (health, register, login):

1. Clique no endpoint desejado
2. Clique em "Try it out"
3. Preencha os parâmetros (se necessário)
4. Clique em "Execute"
5. Veja a resposta abaixo

**Exemplo: Registrar usuário**

```
POST /api/v1/auth/register
{
  "name": "Test User",
  "email": "test@example.com",
  "password": "Password123!"
}
```

### 3. Autenticar para Endpoints Protegidos

Para testar endpoints que requerem autenticação:

1. **Registre ou faça login** usando os endpoints de auth
2. **Copie o accessToken** da resposta
3. **Clique no botão "Authorize"** (cadeado no topo da página)
4. **Cole o token** no campo "Value" (sem o prefixo "Bearer")
5. **Clique em "Authorize"** e depois "Close"

Agora você pode testar todos os endpoints protegidos!

**Exemplo: Obter Dashboard**

```
GET /api/v1/dashboard
Authorization: Bearer <seu-token-aqui>
```

### 4. Endpoints Disponíveis

#### Health
- `GET /health` - Health check (público)

#### Auth
- `POST /api/v1/auth/register` - Registrar usuário (público)
- `POST /api/v1/auth/login` - Login (público)
- `POST /api/v1/auth/refresh` - Renovar token (público)
- `POST /api/v1/auth/logout` - Logout (público)

#### Dashboard
- `GET /api/v1/dashboard` - Obter dashboard agregado (protegido)

#### Workouts
- `GET /api/v1/workouts` - Listar workouts (protegido)

#### Sessions
- `POST /api/v1/sessions` - Iniciar sessão (protegido)
- `POST /api/v1/sessions/{sessionId}/sets` - Registrar série (protegido)
- `PATCH /api/v1/sessions/{sessionId}/finish` - Finalizar sessão (protegido)
- `PATCH /api/v1/sessions/{sessionId}/abandon` - Abandonar sessão (protegido)

## Regenerar Documentação

Após modificar handlers ou adicionar novos endpoints:

```bash
make swagger
```

Ou manualmente:

```bash
swag init -g cmd/kinetria/api/main.go -o docs --parseDependency --parseInternal
```

## Adicionar Documentação em Novos Endpoints

### 1. Adicione anotações no handler

```go
// CreateWorkout godoc
// @Summary Create a new workout
// @Description Create a new workout plan for the user
// @Tags workouts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateWorkoutRequest true "Workout details"
// @Success 201 {object} SuccessResponse{data=WorkoutResponse}
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 422 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/workouts [post]
func (h *WorkoutsHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

### 2. Crie DTOs em swagger_models.go

```go
// CreateWorkoutRequest represents the request to create a workout
type CreateWorkoutRequest struct {
    Name      string `json:"name" validate:"required" example:"Treino A"`
    Type      string `json:"type" validate:"required" example:"HIPERTROFIA"`
    Intensity string `json:"intensity" validate:"required" example:"ALTA"`
}

// WorkoutResponse represents a workout response
type WorkoutResponse struct {
    ID        string `json:"id" example:"uuid-here"`
    Name      string `json:"name" example:"Treino A"`
    Type      string `json:"type" example:"HIPERTROFIA"`
    Intensity string `json:"intensity" example:"ALTA"`
    CreatedAt string `json:"createdAt" example:"2026-02-25T15:30:00Z"`
}
```

### 3. Regenere a documentação

```bash
make swagger
```

## Anotações Swagger

### Tags Comuns

- `@Summary` - Resumo curto do endpoint
- `@Description` - Descrição detalhada
- `@Tags` - Categoria do endpoint (para agrupamento)
- `@Accept` - Content-Type aceito (json, xml, etc)
- `@Produce` - Content-Type da resposta
- `@Security` - Tipo de autenticação (BearerAuth para JWT)
- `@Param` - Parâmetros (body, query, path, header)
- `@Success` - Resposta de sucesso
- `@Failure` - Respostas de erro
- `@Router` - Rota e método HTTP

### Tipos de Parâmetros

```go
// Body parameter
// @Param request body CreateRequest true "Description"

// Query parameter
// @Param page query int false "Page number" default(1)

// Path parameter
// @Param id path string true "Resource ID"

// Header parameter
// @Param Authorization header string true "Bearer token"
```

### Respostas

```go
// Success with data
// @Success 200 {object} SuccessResponse{data=UserResponse}

// Success with array
// @Success 200 {object} SuccessResponse{data=[]UserResponse}

// Error
// @Failure 404 {object} ErrorResponse "Not found"
```

## Exportar Especificação OpenAPI

A especificação OpenAPI está disponível em:

- **JSON**: `http://localhost:8080/api/v1/swagger/doc.json`
- **YAML**: `docs/swagger.yaml` (arquivo local)

Você pode importar esses arquivos em ferramentas como:
- Postman
- Insomnia
- API testing tools
- Code generators

## Troubleshooting

### Swagger UI não carrega

1. Verifique se o serviço está rodando: `curl http://localhost:8080/health`
2. Verifique se a rota está registrada: `curl http://localhost:8080/api/v1/swagger/doc.json`
3. Veja os logs: `docker-compose logs -f app`

### Documentação desatualizada

1. Regenere: `make swagger`
2. Rebuild: `docker-compose up -d --build`
3. Limpe o cache do navegador

### Erro ao gerar documentação

1. Verifique se o swag está instalado: `swag --version`
2. Instale se necessário: `go install github.com/swaggo/swag/cmd/swag@latest`
3. Verifique sintaxe das anotações

## Referências

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)

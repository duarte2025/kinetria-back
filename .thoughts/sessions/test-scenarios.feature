Feature: Iniciar Sessão de Treino (Start Workout Session)
  Como um usuário autenticado
  Quero iniciar uma nova sessão de treino
  Para registrar meu progresso e séries executadas

  Background:
    Given o usuário "João Silva" está autenticado com ID "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And o usuário possui um workout "Treino de Peito" com ID "b2c3d4e5-f6a7-8901-bcde-f12345678901"
    And o usuário NÃO possui nenhuma sessão ativa

  # ────────────────────────────────────────────────────────────────
  # Happy Path
  # ────────────────────────────────────────────────────────────────

  Scenario: Iniciar sessão de treino com sucesso
    Given o workout "Treino de Peito" pertence ao usuário "João Silva"
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then a resposta deve ter status 201 Created
    And o response body deve conter:
      """json
      {
        "data": {
          "id": "<UUID>",
          "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
          "userId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
          "startedAt": "<ISO8601_TIMESTAMP>",
          "finishedAt": null,
          "status": "active"
        }
      }
      """
    And a sessão deve ser persistida no banco com status "active"
    And um registro de auditoria deve ser criado com:
      | campo        | valor                                        |
      | entity_type  | session                                      |
      | entity_id    | <session_id>                                 |
      | action       | created                                      |
      | user_id      | a1b2c3d4-e5f6-7890-abcd-ef1234567890         |
      | occurred_at  | <timestamp próximo de now>                   |

  # ────────────────────────────────────────────────────────────────
  # Sad Paths — Validação de Input
  # ────────────────────────────────────────────────────────────────

  Scenario: Falha ao iniciar sessão sem token JWT
    When o usuário envia POST /api/v1/sessions SEM header Authorization
    Then a resposta deve ter status 401 Unauthorized
    And o response body deve conter:
      """json
      {
        "code": "UNAUTHORIZED",
        "message": "Invalid or expired access token."
      }
      """
    And nenhuma sessão deve ser criada no banco
    And nenhum registro de auditoria deve ser criado

  Scenario: Falha ao iniciar sessão com token JWT expirado
    Given o token JWT do usuário está expirado
    When o usuário envia POST /api/v1/sessions com token expirado
    Then a resposta deve ter status 401 Unauthorized
    And o response body deve conter:
      """json
      {
        "code": "UNAUTHORIZED",
        "message": "Invalid or expired access token."
      }
      """

  Scenario: Falha ao iniciar sessão sem workoutId
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {}
      """
    Then a resposta deve ter status 422 Unprocessable Entity
    And o response body deve conter:
      """json
      {
        "code": "VALIDATION_ERROR",
        "message": "Request body is invalid.",
        "details": {
          "workoutId": "field is required"
        }
      }
      """
    And nenhuma sessão deve ser criada no banco

  Scenario: Falha ao iniciar sessão com workoutId inválido (não é UUID)
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "invalid-uuid-format"
      }
      """
    Then a resposta deve ter status 422 Unprocessable Entity
    And o response body deve conter:
      """json
      {
        "code": "VALIDATION_ERROR",
        "message": "Request body is invalid.",
        "details": {
          "workoutId": "must be a valid UUID"
        }
      }
      """
    And nenhuma sessão deve ser criada no banco

  # ────────────────────────────────────────────────────────────────
  # Sad Paths — Regras de Negócio
  # ────────────────────────────────────────────────────────────────

  Scenario: Falha ao iniciar sessão com workout inexistente
    Given NÃO existe workout com ID "99999999-0000-0000-0000-000000000000"
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "99999999-0000-0000-0000-000000000000"
      }
      """
    Then a resposta deve ter status 404 Not Found
    And o response body deve conter:
      """json
      {
        "code": "WORKOUT_NOT_FOUND",
        "message": "Workout with id '99999999-0000-0000-0000-000000000000' was not found."
      }
      """
    And nenhuma sessão deve ser criada no banco

  Scenario: Falha ao iniciar sessão com workout de outro usuário (ownership)
    Given existe um workout "Treino de Costas" com ID "c3d4e5f6-a7b8-9012-cdef-123456789012"
    And o workout "Treino de Costas" pertence ao usuário "Maria Santos" (ID diferente)
    When o usuário "João Silva" envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "c3d4e5f6-a7b8-9012-cdef-123456789012"
      }
      """
    Then a resposta deve ter status 404 Not Found
    And o response body deve conter:
      """json
      {
        "code": "WORKOUT_NOT_FOUND",
        "message": "Workout with id 'c3d4e5f6-a7b8-9012-cdef-123456789012' was not found."
      }
      """
    And nenhuma sessão deve ser criada no banco
    # Nota: retornamos 404 (não 403) para não vazar que o workout existe

  Scenario: Falha ao iniciar sessão quando já existe sessão ativa (duplicação)
    Given o usuário "João Silva" já possui uma sessão ativa com ID "d4e5f6a7-b8c9-0123-defa-234567890123"
    And a sessão ativa foi criada há 10 minutos
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then a resposta deve ter status 409 Conflict
    And o response body deve conter:
      """json
      {
        "code": "ACTIVE_SESSION_EXISTS",
        "message": "User already has an active session. Finish or abandon it before starting a new one."
      }
      """
    And nenhuma nova sessão deve ser criada no banco
    And a sessão ativa anterior deve permanecer inalterada

  # ────────────────────────────────────────────────────────────────
  # Edge Cases — Concorrência
  # ────────────────────────────────────────────────────────────────

  Scenario: Race condition - dois requests simultâneos para criar sessão
    Given o usuário "João Silva" NÃO possui sessão ativa
    When o usuário envia 2 requests POST /api/v1/sessions SIMULTANEAMENTE com:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then UM dos requests deve retornar 201 Created
    And o OUTRO request deve retornar 409 Conflict
    And apenas UMA sessão deve existir no banco
    And a sessão criada deve ter status "active"

  # ────────────────────────────────────────────────────────────────
  # Edge Cases — Integridade de Dados
  # ────────────────────────────────────────────────────────────────

  Scenario: Sessão criada deve ter timestamps corretos
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then a resposta deve ter status 201 Created
    And o campo "startedAt" deve ser um timestamp ISO8601 válido
    And o campo "startedAt" deve ser próximo de now (diferença < 5 segundos)
    And o campo "finishedAt" deve ser null
    And o campo "createdAt" deve ser próximo de now (no banco)
    And o campo "updatedAt" deve ser igual a "createdAt" (no banco)

  Scenario: Sessão criada deve ter status "active"
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then a resposta deve ter status 201 Created
    And o campo "status" deve ser "active"
    And no banco, o campo "status" deve ser "active"

  Scenario: Sessão criada deve ter ID único (UUID v4)
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then a resposta deve ter status 201 Created
    And o campo "id" deve ser um UUID v4 válido
    And o ID deve ser diferente de qualquer sessão anterior

  # ────────────────────────────────────────────────────────────────
  # Edge Cases — Auditoria
  # ────────────────────────────────────────────────────────────────

  Scenario: Audit log deve ser criado com todos os campos obrigatórios
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then a resposta deve ter status 201 Created
    And um registro de auditoria deve existir com:
      | campo         | validação                     |
      | id            | UUID v4 válido                |
      | user_id       | a1b2c3d4-e5f6-7890-abcd-ef1234567890 |
      | entity_type   | "session"                     |
      | entity_id     | <id da sessão criada>         |
      | action        | "created"                     |
      | action_data   | JSON contendo workoutId e startedAt |
      | occurred_at   | timestamp próximo de now      |
      | ip_address    | IP do request                 |
      | user_agent    | User-Agent do request         |

  Scenario: Audit log deve conter action_data estruturado
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then a resposta deve ter status 201 Created
    And o audit log deve ter action_data contendo:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
        "startedAt": "<ISO8601_TIMESTAMP>"
      }
      """

  # ────────────────────────────────────────────────────────────────
  # Edge Cases — Idempotência (comportamento esperado)
  # ────────────────────────────────────────────────────────────────

  Scenario: Retry de request após erro de rede (não é idempotente)
    Given o usuário enviou POST /api/v1/sessions e recebeu 201 Created
    And a sessão foi criada com sucesso
    When o usuário reenvia o MESMO request (retry)
    Then a resposta deve ser 409 Conflict (sessão ativa já existe)
    # Nota: endpoint NÃO é idempotente por design
    # Client deve verificar se sessão ativa existe antes de retry

  Scenario: Client pode consultar sessão ativa após 409 Conflict
    Given o usuário já possui sessão ativa com ID "d4e5f6a7-b8c9-0123-defa-234567890123"
    When o usuário envia POST /api/v1/sessions
    And recebe 409 Conflict
    Then o client pode consultar GET /api/v1/sessions/active (endpoint futuro)
    And obter a sessão ativa existente
    # Nota: este cenário assume endpoint GetActiveSession existe (feature futura)

  # ────────────────────────────────────────────────────────────────
  # Performance / Observabilidade
  # ────────────────────────────────────────────────────────────────

  Scenario: Request bem-sucedido deve ser logado com métricas
    When o usuário envia POST /api/v1/sessions com body:
      """json
      {
        "workoutId": "b2c3d4e5-f6a7-8901-bcde-f12345678901"
      }
      """
    Then a resposta deve ter status 201 Created
    And deve existir log estruturado contendo:
      | campo       | valor esperado                           |
      | level       | info                                     |
      | method      | POST                                     |
      | path        | /api/v1/sessions                         |
      | user_id     | a1b2c3d4-e5f6-7890-abcd-ef1234567890     |
      | workout_id  | b2c3d4e5-f6a7-8901-bcde-f12345678901     |
      | session_id  | <UUID da sessão criada>                  |
      | status      | 201                                      |
      | duration_ms | <número positivo>                        |
    And a métrica "sessions_started_total" deve incrementar em 1

  Scenario: Request com erro deve ser logado com tipo de erro
    Given o usuário já possui sessão ativa
    When o usuário envia POST /api/v1/sessions
    Then a resposta deve ter status 409 Conflict
    And deve existir log estruturado contendo:
      | campo       | valor esperado                           |
      | level       | warn ou error                            |
      | method      | POST                                     |
      | path        | /api/v1/sessions                         |
      | user_id     | a1b2c3d4-e5f6-7890-abcd-ef1234567890     |
      | status      | 409                                      |
      | error       | active session exists                    |
    And a métrica "sessions_start_errors_total{error_type=conflict}" deve incrementar em 1

  # ────────────────────────────────────────────────────────────────
  # Segurança
  # ────────────────────────────────────────────────────────────────

  Scenario: Logs não devem conter dados sensíveis
    When o usuário envia POST /api/v1/sessions
    Then os logs NÃO devem conter:
      | dado sensível    |
      | password         |
      | JWT token        |
      | refresh token    |
    And os logs PODEM conter:
      | dado permitido   |
      | user_id (UUID)   |
      | workout_id       |
      | session_id       |
      | IP address       |

  Scenario: Response não deve vazar informações de outros usuários
    Given existe workout "Treino X" de outro usuário
    When o usuário tenta iniciar sessão com workout de outro usuário
    Then a resposta deve ser 404 Not Found
    And a mensagem NÃO deve indicar que o workout existe
    And a mensagem NÃO deve revelar o owner do workout

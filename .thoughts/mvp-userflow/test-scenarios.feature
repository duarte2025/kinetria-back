Feature: MVP UserFlow — Kinetria Backend

  # ─────────────────────────────────────────────────────────────
  # AUTH
  # ─────────────────────────────────────────────────────────────

  Scenario: Registro bem-sucedido de novo usuário
    Given que o sistema está disponível
    When eu envio POST /auth/register com name "Bruno Costa", email "bruno@example.com" e password "s3cr3tP@ss"
    Then o sistema retorna 201 Created
    And o body contém accessToken, refreshToken e expiresIn
    And um registro de usuário é criado no banco com email "bruno@example.com"

  Scenario: Registro com email já cadastrado
    Given que já existe um usuário com email "bruno@example.com"
    When eu envio POST /auth/register com email "bruno@example.com"
    Then o sistema retorna 409 Conflict
    And o body contém code "EMAIL_ALREADY_EXISTS"

  Scenario: Registro com payload inválido (campos obrigatórios ausentes)
    Given que o sistema está disponível
    When eu envio POST /auth/register sem o campo "email"
    Then o sistema retorna 422 Unprocessable Entity
    And o body contém code "VALIDATION_ERROR"

  Scenario: Login bem-sucedido
    Given que existe um usuário com email "bruno@example.com" e senha "s3cr3tP@ss"
    When eu envio POST /auth/login com email "bruno@example.com" e password "s3cr3tP@ss"
    Then o sistema retorna 200 OK
    And o body contém accessToken, refreshToken e expiresIn

  Scenario: Login com credenciais inválidas
    Given que existe um usuário com email "bruno@example.com"
    When eu envio POST /auth/login com password incorreta
    Then o sistema retorna 401 Unauthorized
    And o body contém code "INVALID_CREDENTIALS"

  Scenario: Refresh de token bem-sucedido
    Given que possuo um refreshToken válido
    When eu envio POST /auth/refresh com o refreshToken
    Then o sistema retorna 200 OK
    And o body contém novos accessToken, refreshToken e expiresIn
    And o refreshToken antigo é revogado no banco

  Scenario: Refresh com token expirado ou revogado
    Given que o refreshToken está revogado ou expirado
    When eu envio POST /auth/refresh com esse refreshToken
    Then o sistema retorna 401 Unauthorized
    And o body contém code "UNAUTHORIZED"

  Scenario: Logout bem-sucedido
    Given que estou autenticado com um access token válido e possuo um refreshToken
    When eu envio POST /auth/logout com o refreshToken
    Then o sistema retorna 204 No Content
    And o refreshToken é marcado como revogado no banco

  Scenario: Logout sem autenticação
    Given que não possuo access token
    When eu envio POST /auth/logout
    Then o sistema retorna 401 Unauthorized

  # ─────────────────────────────────────────────────────────────
  # DASHBOARD
  # ─────────────────────────────────────────────────────────────

  Scenario: Busca de dados do dashboard para usuário autenticado
    Given que estou autenticado como "bruno@example.com"
    And o usuário possui workouts cadastrados
    And o usuário realizou sessões na semana atual
    When eu envio GET /dashboard
    Then o sistema retorna 200 OK
    And o body contém user com id, name, email
    And o body contém weekProgress com exatamente 7 dias
    And o body contém stats com calories e totalTimeMinutes
    And cada dia de weekProgress possui status "completed", "missed" ou "future"

  Scenario: Dashboard quando usuário não tem workout agendado para hoje
    Given que estou autenticado como "bruno@example.com"
    And não há workout agendado para hoje
    When eu envio GET /dashboard
    Then o sistema retorna 200 OK
    And o campo todayWorkout é null

  Scenario: Dashboard sem autenticação
    Given que não possuo access token
    When eu envio GET /dashboard
    Then o sistema retorna 401 Unauthorized

  # ─────────────────────────────────────────────────────────────
  # WORKOUTS
  # ─────────────────────────────────────────────────────────────

  Scenario: Listar workouts do usuário autenticado com paginação
    Given que estou autenticado como "bruno@example.com"
    And o usuário possui 25 workouts cadastrados
    When eu envio GET /workouts?page=1&pageSize=20
    Then o sistema retorna 200 OK
    And o body contém uma lista com 20 workouts
    And o body contém meta com page=1, pageSize=20, total=25, totalPages=2

  Scenario: Listar workouts de outro usuário não retorna dados
    Given que estou autenticado como "bruno@example.com" (userID A)
    And existe outro usuário (userID B) com workouts cadastrados
    When eu envio GET /workouts
    Then o sistema retorna 200 OK
    And a lista não contém workouts de userID B

  Scenario: Buscar detalhe de workout existente
    Given que estou autenticado como "bruno@example.com"
    And existe um workout com id "wod-uuid" pertencente ao usuário
    When eu envio GET /workouts/wod-uuid
    Then o sistema retorna 200 OK
    And o body contém o workout com seus exercises

  Scenario: Buscar workout inexistente
    Given que estou autenticado como "bruno@example.com"
    When eu envio GET /workouts/uuid-inexistente
    Then o sistema retorna 404 Not Found
    And o body contém code "NOT_FOUND"

  Scenario: Buscar workout de outro usuário
    Given que estou autenticado como "bruno@example.com" (userID A)
    And existe um workout com id "wod-uuid" pertencente a userID B
    When eu envio GET /workouts/wod-uuid
    Then o sistema retorna 404 Not Found

  # ─────────────────────────────────────────────────────────────
  # SESSIONS — Start
  # ─────────────────────────────────────────────────────────────

  Scenario: Iniciar sessão de treino com sucesso
    Given que estou autenticado como "bruno@example.com"
    And não há sessão ativa para o usuário
    And existe um workout com id "wod-uuid" pertencente ao usuário
    When eu envio POST /sessions com workoutId "wod-uuid"
    Then o sistema retorna 201 Created
    And o body contém session com status "active" e startedAt preenchido
    And um registro de audit log é criado com action "created"

  Scenario: Iniciar sessão com sessão ativa existente
    Given que estou autenticado como "bruno@example.com"
    And já existe uma sessão com status "active" para o usuário
    When eu envio POST /sessions com workoutId "wod-uuid"
    Then o sistema retorna 409 Conflict
    And o body contém code "SESSION_ALREADY_ACTIVE"

  Scenario: Iniciar sessão com workoutId inválido
    Given que estou autenticado como "bruno@example.com"
    When eu envio POST /sessions com workoutId "uuid-invalido"
    Then o sistema retorna 404 Not Found
    And o body contém code "NOT_FOUND"

  Scenario: Iniciar sessão com payload inválido
    Given que estou autenticado como "bruno@example.com"
    When eu envio POST /sessions sem o campo workoutId
    Then o sistema retorna 422 Unprocessable Entity

  # ─────────────────────────────────────────────────────────────
  # SESSIONS — Record Set
  # ─────────────────────────────────────────────────────────────

  Scenario: Registrar série com sucesso
    Given que estou autenticado como "bruno@example.com"
    And existe uma sessão ativa "session-uuid" pertencente ao usuário
    And "exercise-uuid" é um exercício do workout da sessão
    When eu envio POST /sessions/session-uuid/sets com exerciseId "exercise-uuid", setNumber 1, weight 80.0, reps 10, status "completed"
    Then o sistema retorna 201 Created
    And o body contém SetRecord com id, sessionId, exerciseId, setNumber, weight, reps, status e recordedAt
    And um registro de audit log é criado com action "created"

  Scenario: Registrar série duplicada (mesmo setNumber para mesmo exercício na mesma sessão)
    Given que estou autenticado como "bruno@example.com"
    And já existe um SetRecord para (session-uuid, exercise-uuid, setNumber=1)
    When eu envio POST /sessions/session-uuid/sets com setNumber 1 novamente
    Then o sistema retorna 409 Conflict
    And o body contém code "SET_ALREADY_RECORDED"

  Scenario: Registrar série em sessão inexistente
    Given que estou autenticado como "bruno@example.com"
    When eu envio POST /sessions/uuid-inexistente/sets com payload válido
    Then o sistema retorna 404 Not Found

  Scenario: Registrar série em sessão de outro usuário
    Given que estou autenticado como "bruno@example.com" (userID A)
    And existe uma sessão "session-uuid" pertencente a userID B
    When eu envio POST /sessions/session-uuid/sets
    Then o sistema retorna 404 Not Found

  Scenario: Registrar série em sessão já finalizada
    Given que estou autenticado como "bruno@example.com"
    And a sessão "session-uuid" possui status "completed"
    When eu envio POST /sessions/session-uuid/sets com payload válido
    Then o sistema retorna 409 Conflict
    And o body contém code "SESSION_ALREADY_CLOSED"

  Scenario: Registrar série pulada (status skipped)
    Given que estou autenticado como "bruno@example.com"
    And existe uma sessão ativa "session-uuid"
    When eu envio POST /sessions/session-uuid/sets com status "skipped" e reps 0
    Then o sistema retorna 201 Created
    And o SetRecord retornado possui status "skipped"

  Scenario: Registrar série com payload inválido (weight negativo)
    Given que estou autenticado como "bruno@example.com"
    And existe uma sessão ativa "session-uuid"
    When eu envio POST /sessions/session-uuid/sets com weight -5.0
    Then o sistema retorna 422 Unprocessable Entity

  # ─────────────────────────────────────────────────────────────
  # SESSIONS — Finish
  # ─────────────────────────────────────────────────────────────

  Scenario: Finalizar sessão ativa com sucesso
    Given que estou autenticado como "bruno@example.com"
    And existe uma sessão ativa "session-uuid" pertencente ao usuário
    When eu envio PATCH /sessions/session-uuid/finish com notes "Ótimo treino!"
    Then o sistema retorna 200 OK
    And o body contém session com status "completed" e finishedAt preenchido
    And um registro de audit log é criado com action "completed"

  Scenario: Finalizar sessão sem notes (campo opcional)
    Given que estou autenticado como "bruno@example.com"
    And existe uma sessão ativa "session-uuid"
    When eu envio PATCH /sessions/session-uuid/finish sem body
    Then o sistema retorna 200 OK
    And o body contém session com status "completed"

  Scenario: Finalizar sessão já finalizada
    Given que estou autenticado como "bruno@example.com"
    And a sessão "session-uuid" possui status "completed"
    When eu envio PATCH /sessions/session-uuid/finish
    Then o sistema retorna 409 Conflict
    And o body contém code "SESSION_ALREADY_CLOSED"

  Scenario: Finalizar sessão abandonada
    Given que estou autenticado como "bruno@example.com"
    And a sessão "session-uuid" possui status "abandoned"
    When eu envio PATCH /sessions/session-uuid/finish
    Then o sistema retorna 409 Conflict
    And o body contém code "SESSION_ALREADY_CLOSED"

  Scenario: Finalizar sessão inexistente
    Given que estou autenticado como "bruno@example.com"
    When eu envio PATCH /sessions/uuid-inexistente/finish
    Then o sistema retorna 404 Not Found

  Scenario: Finalizar sessão de outro usuário
    Given que estou autenticado como "bruno@example.com" (userID A)
    And existe uma sessão ativa "session-uuid" pertencente a userID B
    When eu envio PATCH /sessions/session-uuid/finish
    Then o sistema retorna 404 Not Found

  # ─────────────────────────────────────────────────────────────
  # SESSIONS — Abandon
  # ─────────────────────────────────────────────────────────────

  Scenario: Abandonar sessão ativa com sucesso
    Given que estou autenticado como "bruno@example.com"
    And existe uma sessão ativa "session-uuid" pertencente ao usuário
    When eu envio PATCH /sessions/session-uuid/abandon
    Then o sistema retorna 200 OK
    And o body contém session com status "abandoned" e finishedAt preenchido
    And um registro de audit log é criado com action "abandoned"

  Scenario: Abandonar sessão já fechada
    Given que estou autenticado como "bruno@example.com"
    And a sessão "session-uuid" possui status "completed" ou "abandoned"
    When eu envio PATCH /sessions/session-uuid/abandon
    Then o sistema retorna 409 Conflict
    And o body contém code "SESSION_ALREADY_CLOSED"

  Scenario: Abandonar sessão inexistente
    Given que estou autenticado como "bruno@example.com"
    When eu envio PATCH /sessions/uuid-inexistente/abandon
    Then o sistema retorna 404 Not Found

  # ─────────────────────────────────────────────────────────────
  # SECURITY / CROSS-CUTTING
  # ─────────────────────────────────────────────────────────────

  Scenario: Acesso a endpoint protegido com token expirado
    Given que possuo um accessToken expirado
    When eu envio qualquer requisição a endpoint protegido (ex: GET /dashboard)
    Then o sistema retorna 401 Unauthorized
    And o body contém code "UNAUTHORIZED"

  Scenario: Rate limiting em /auth/login excedido
    Given que o IP "1.2.3.4" realizou 20 tentativas de login no último minuto
    When o IP "1.2.3.4" tenta nova requisição POST /auth/login
    Then o sistema retorna 429 Too Many Requests

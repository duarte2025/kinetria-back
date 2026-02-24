Feature: Dashboard — Dados agregados do usuário

  Como usuário autenticado do Kinetria
  Quero visualizar meu dashboard personalizado
  Para ter uma visão geral do meu progresso semanal e treino de hoje

  Background:
    Given o serviço está rodando
    And o banco de dados está limpo
    And existe um usuário "user@example.com" com senha "Password123!"

  # ═══════════════════════════════════════════════════════════════
  # Happy Paths
  # ═══════════════════════════════════════════════════════════════

  Scenario: Carregar dashboard com dados completos
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 3 workouts cadastrados:
      | name                  | type  | intensity | duration |
      | Treino Peito/Tríceps  | FORÇA | Alta      | 45       |
      | Treino Costas/Bíceps  | FORÇA | Alta      | 50       |
      | HIIT Cardio           | CARDIO| Muito Alta| 30       |
    And o usuário completou as seguintes sessões nos últimos 7 dias:
      | workout               | started_at          | finished_at         | status    |
      | Treino Peito/Tríceps  | 2026-02-17 08:00:00 | 2026-02-17 08:50:00 | completed |
      | HIIT Cardio           | 2026-02-18 18:00:00 | 2026-02-18 18:35:00 | completed |
      | Treino Costas/Bíceps  | 2026-02-20 09:00:00 | 2026-02-20 10:00:00 | completed |
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And a resposta deve conter:
      """
      {
        "data": {
          "user": {
            "id": "<uuid>",
            "name": "<string>",
            "email": "user@example.com",
            "profileImageUrl": null
          },
          "todayWorkout": {
            "id": "<uuid>",
            "name": "Treino Peito/Tríceps",
            "description": "",
            "type": "FORÇA",
            "intensity": "Alta",
            "duration": 45,
            "imageUrl": ""
          },
          "weekProgress": [
            { "day": "S", "date": "2026-02-17", "status": "completed" },
            { "day": "T", "date": "2026-02-18", "status": "completed" },
            { "day": "Q", "date": "2026-02-19", "status": "missed" },
            { "day": "Q", "date": "2026-02-20", "status": "completed" },
            { "day": "S", "date": "2026-02-21", "status": "missed" },
            { "day": "S", "date": "2026-02-22", "status": "missed" },
            { "day": "D", "date": "2026-02-23", "status": "missed" }
          ],
          "stats": {
            "calories": 1225,
            "totalTimeMinutes": 175
          }
        }
      }
      """

  Scenario: Dashboard com dias futuros (segunda-feira visualizando semana)
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And o usuário completou 1 sessão em "2026-02-17"
    And hoje é "2026-02-17"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And o campo "weekProgress" deve ter 7 itens
    And "weekProgress[0].status" deve ser "missed" # 2026-02-11
    And "weekProgress[6].status" deve ser "completed" # 2026-02-17 (hoje)

  Scenario: Dashboard com semana completamente zerada
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And o usuário NÃO completou nenhuma sessão nos últimos 7 dias
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And o campo "weekProgress" deve ter 7 itens
    And todos os itens de "weekProgress" devem ter "status" = "missed"
    And "stats.calories" deve ser 0
    And "stats.totalTimeMinutes" deve ser 0

  # ═══════════════════════════════════════════════════════════════
  # Edge Cases — Dados Vazios
  # ═══════════════════════════════════════════════════════════════

  Scenario: Usuário sem workouts cadastrados
    Given o usuário "user@example.com" está autenticado
    And o usuário NÃO tem nenhum workout cadastrado
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "todayWorkout" deve ser null
    And o campo "weekProgress" deve ter 7 itens
    And todos os itens de "weekProgress" devem ter "status" = "missed"
    And "stats.calories" deve ser 0
    And "stats.totalTimeMinutes" deve ser 0

  Scenario: Usuário com workouts mas sem sessões completadas
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 2 workouts cadastrados
    And o usuário NÃO completou nenhuma sessão
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "todayWorkout" NÃO deve ser null
    And "todayWorkout.id" deve existir
    And o campo "weekProgress" deve ter 7 itens
    And todos os itens de "weekProgress" devem ter "status" = "missed"
    And "stats.calories" deve ser 0
    And "stats.totalTimeMinutes" deve ser 0

  # ═══════════════════════════════════════════════════════════════
  # Edge Cases — Sessões com Status Diferentes
  # ═══════════════════════════════════════════════════════════════

  Scenario: Sessões ativas NÃO contam no dashboard
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And o usuário tem uma sessão ATIVA iniciada hoje às 10:00
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "weekProgress[6].status" deve ser "missed" # hoje ainda não completed
    And "stats.calories" deve ser 0
    And "stats.totalTimeMinutes" deve ser 0

  Scenario: Sessões abandonadas NÃO contam no dashboard
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And o usuário tem as seguintes sessões nos últimos 7 dias:
      | workout      | started_at          | finished_at         | status    |
      | Treino Força | 2026-02-17 08:00:00 | 2026-02-17 08:20:00 | abandoned |
      | Treino Força | 2026-02-18 08:00:00 | 2026-02-18 08:45:00 | completed |
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "weekProgress[0].status" deve ser "missed" # 2026-02-17 (abandoned não conta)
    And "weekProgress[1].status" deve ser "completed" # 2026-02-18
    And "stats.totalTimeMinutes" deve ser 45 # apenas sessão completed

  Scenario: Múltiplas sessões no mesmo dia contam como 1 dia completed
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 2 workouts cadastrados
    And o usuário completou as seguintes sessões:
      | workout        | started_at          | finished_at         | status    |
      | Treino Manhã   | 2026-02-17 08:00:00 | 2026-02-17 08:40:00 | completed |
      | Treino Tarde   | 2026-02-17 16:00:00 | 2026-02-17 16:30:00 | completed |
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "weekProgress[0].status" deve ser "completed" # 2026-02-17
    And "stats.totalTimeMinutes" deve ser 70 # 40 + 30
    And "stats.calories" deve ser 490 # 70 * 7

  # ═══════════════════════════════════════════════════════════════
  # Edge Cases — Cálculos de Tempo e Calorias
  # ═══════════════════════════════════════════════════════════════

  Scenario: Sessão com duração < 1 minuto (arredondamento)
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And o usuário completou uma sessão de 45 segundos em "2026-02-17"
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "stats.totalTimeMinutes" deve ser 0 ou 1 # decisão: truncar ou arredondar?
    And "stats.calories" deve ser 0 ou 7 # depende do arredondamento

  Scenario: Cálculo correto de calorias com múltiplas sessões
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And o usuário completou as seguintes sessões nos últimos 7 dias:
      | started_at          | finished_at         |
      | 2026-02-17 08:00:00 | 2026-02-17 09:00:00 | # 60 min
      | 2026-02-18 08:00:00 | 2026-02-18 08:30:00 | # 30 min
      | 2026-02-19 08:00:00 | 2026-02-19 08:45:00 | # 45 min
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "stats.totalTimeMinutes" deve ser 135
    And "stats.calories" deve ser 945 # 135 * 7

  # ═══════════════════════════════════════════════════════════════
  # Edge Cases — Dias da Semana e Datas
  # ═══════════════════════════════════════════════════════════════

  Scenario: WeekProgress sempre tem 7 dias (de D-6 até hoje)
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And hoje é "2026-02-23" # domingo
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And o campo "weekProgress" deve ter exatamente 7 itens
    And "weekProgress[0].date" deve ser "2026-02-17" # seg
    And "weekProgress[1].date" deve ser "2026-02-18" # ter
    And "weekProgress[2].date" deve ser "2026-02-19" # qua
    And "weekProgress[3].date" deve ser "2026-02-20" # qui
    And "weekProgress[4].date" deve ser "2026-02-21" # sex
    And "weekProgress[5].date" deve ser "2026-02-22" # sáb
    And "weekProgress[6].date" deve ser "2026-02-23" # dom (hoje)

  Scenario: WeekProgress labels em português (abreviado)
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "weekProgress[0].day" deve ser "S" # segunda
    And "weekProgress[1].day" deve ser "T" # terça
    And "weekProgress[2].day" deve ser "Q" # quarta
    And "weekProgress[3].day" deve ser "Q" # quinta
    And "weekProgress[4].day" deve ser "S" # sexta
    And "weekProgress[5].day" deve ser "S" # sábado
    And "weekProgress[6].day" deve ser "D" # domingo

  Scenario: Sessão iniciada em um dia e terminada em outro (conta pelo dia de início)
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado
    And o usuário completou uma sessão:
      | started_at          | finished_at         | status    |
      | 2026-02-17 23:50:00 | 2026-02-18 00:10:00 | completed |
    And hoje é "2026-02-23"
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "weekProgress[0].status" deve ser "completed" # 2026-02-17 (dia de início)
    And "weekProgress[1].status" deve ser "missed" # 2026-02-18 (não conta)
    And "stats.totalTimeMinutes" deve ser 20

  # ═══════════════════════════════════════════════════════════════
  # Sad Paths — Autenticação e Autorização
  # ═══════════════════════════════════════════════════════════════

  Scenario: Acesso sem token JWT
    When um cliente faz GET "/api/v1/dashboard" sem autenticação
    Then a resposta deve ter status 401
    And a resposta deve conter:
      """
      {
        "code": "UNAUTHORIZED",
        "message": "Missing or invalid authentication token"
      }
      """

  Scenario: Acesso com token JWT expirado
    Given o usuário "user@example.com" estava autenticado
    And o token JWT expirou
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 401
    And a resposta deve conter "code" = "TOKEN_EXPIRED"

  Scenario: Acesso com token JWT inválido
    When um cliente faz GET "/api/v1/dashboard" com token "invalid.jwt.token"
    Then a resposta deve ter status 401
    And a resposta deve conter "code" = "INVALID_TOKEN"

  # ═══════════════════════════════════════════════════════════════
  # Sad Paths — Erros de Sistema
  # ═══════════════════════════════════════════════════════════════

  Scenario: Erro ao buscar dados do usuário (database down)
    Given o usuário "user@example.com" está autenticado
    And o banco de dados está offline
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 500
    And a resposta deve conter:
      """
      {
        "code": "INTERNAL_ERROR",
        "message": "Failed to load dashboard data"
      }
      """

  Scenario: Timeout em uma das agregações paralelas
    Given o usuário "user@example.com" está autenticado
    And o use case "GetWeekProgressUC" demora mais de 5 segundos
    When o usuário faz GET "/api/v1/dashboard" com timeout de 3 segundos
    Then a resposta deve ter status 500
    And a resposta deve conter "code" = "INTERNAL_ERROR"

  # ═══════════════════════════════════════════════════════════════
  # Performance e Observabilidade
  # ═══════════════════════════════════════════════════════════════

  Scenario: Dashboard deve carregar em menos de 500ms (agregação paralela)
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 10 workouts cadastrados
    And o usuário completou 30 sessões nos últimos 7 dias
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And o tempo de resposta deve ser menor que 500ms

  Scenario: Tracing e logs estruturados
    Given o usuário "user@example.com" está autenticado
    And o sistema está configurado para tracing
    When o usuário faz GET "/api/v1/dashboard"
    Then deve haver um span "GET /dashboard"
    And o span deve ter atributos:
      | key                              | value              |
      | user.id                          | <uuid>             |
      | dashboard.today_workout.found    | true               |
      | dashboard.week_progress.days_completed | <int>          |
    And deve haver logs estruturados com campos:
      | field                  | tipo   |
      | user_id                | string |
      | today_workout_found    | bool   |
      | week_days_completed    | int    |
      | duration_ms            | int    |

  # ═══════════════════════════════════════════════════════════════
  # Idempotência e Cache (futuro)
  # ═══════════════════════════════════════════════════════════════

  Scenario: Múltiplas chamadas retornam mesmos dados (idempotente)
    Given o usuário "user@example.com" está autenticado
    And o usuário tem dados no dashboard
    When o usuário faz GET "/api/v1/dashboard" 3 vezes seguidas
    Then todas as respostas devem ter status 200
    And todas as respostas devem ser idênticas

  Scenario: Dashboard é read-only (não gera audit log)
    Given o usuário "user@example.com" está autenticado
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And NÃO deve haver registro na tabela "audit_log" para esta operação

  # ═══════════════════════════════════════════════════════════════
  # Compatibilidade com ProfileImageURL
  # ═══════════════════════════════════════════════════════════════

  Scenario: Usuário com imagem de perfil personalizada
    Given o usuário "user@example.com" está autenticado
    And o usuário tem profileImageUrl = "https://cdn.kinetria.app/avatars/abc123.jpg"
    And o usuário tem 1 workout cadastrado
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "user.profileImageUrl" deve ser "https://cdn.kinetria.app/avatars/abc123.jpg"

  Scenario: Usuário sem imagem de perfil (null)
    Given o usuário "user@example.com" está autenticado
    And o usuário NÃO tem profileImageUrl configurado
    And o usuário tem 1 workout cadastrado
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "user.profileImageUrl" deve ser null ou string vazia

  # ═══════════════════════════════════════════════════════════════
  # Compatibilidade com Workout.Type e Workout.Intensity
  # ═══════════════════════════════════════════════════════════════

  Scenario: TodayWorkout com todos os campos preenchidos
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado:
      | name           | description                | type  | intensity  | duration | imageUrl                                |
      | Treino Completo| Treino intenso de hipertrofia | FORÇA | Muito Alta | 60       | https://cdn.kinetria.app/workouts/x.jpg |
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "todayWorkout.name" deve ser "Treino Completo"
    And "todayWorkout.description" deve ser "Treino intenso de hipertrofia"
    And "todayWorkout.type" deve ser "FORÇA"
    And "todayWorkout.intensity" deve ser "Muito Alta"
    And "todayWorkout.duration" deve ser 60
    And "todayWorkout.imageUrl" deve ser "https://cdn.kinetria.app/workouts/x.jpg"

  Scenario: TodayWorkout com campos opcionais vazios
    Given o usuário "user@example.com" está autenticado
    And o usuário tem 1 workout cadastrado:
      | name        | description | type | intensity | duration | imageUrl |
      | Treino Básico |             |      |           | 30       |          |
    When o usuário faz GET "/api/v1/dashboard"
    Then a resposta deve ter status 200
    And "todayWorkout.name" deve ser "Treino Básico"
    And "todayWorkout.description" deve ser string vazia ou null
    And "todayWorkout.type" deve ser string vazia ou null
    And "todayWorkout.intensity" deve ser string vazia ou null
    And "todayWorkout.duration" deve ser 30
    And "todayWorkout.imageUrl" deve ser string vazia ou null

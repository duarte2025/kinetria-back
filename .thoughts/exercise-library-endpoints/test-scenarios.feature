Feature: Exercise Library Endpoints

  Scenario: Listar todos os exercícios sem filtros
    Given a biblioteca possui 30 exercícios
    When o usuário faz GET /api/v1/exercises
    Then a resposta deve ser 200 OK
    And o body deve conter 20 exercícios (pageSize default)
    And o meta deve conter total=30, totalPages=2

  Scenario: Listar exercícios com paginação customizada
    Given a biblioteca possui 50 exercícios
    When o usuário faz GET /api/v1/exercises?page=2&pageSize=10
    Then a resposta deve ser 200 OK
    And o body deve conter 10 exercícios
    And o meta deve conter page=2, pageSize=10, total=50, totalPages=5

  Scenario: Filtrar exercícios por grupo muscular
    Given a biblioteca possui exercícios de "Peito", "Costas", "Pernas"
    When o usuário faz GET /api/v1/exercises?muscleGroup=Peito
    Then a resposta deve ser 200 OK
    And todos os exercícios retornados devem ter "Peito" em muscles

  Scenario: Filtrar exercícios por equipamento
    Given a biblioteca possui exercícios com "Barra", "Halteres", "Peso corporal"
    When o usuário faz GET /api/v1/exercises?equipment=Barra
    Then a resposta deve ser 200 OK
    And todos os exercícios retornados devem ter equipment="Barra"

  Scenario: Filtrar exercícios por dificuldade
    Given a biblioteca possui exercícios "Iniciante", "Intermediário", "Avançado"
    When o usuário faz GET /api/v1/exercises?difficulty=Intermediário
    Then a resposta deve ser 200 OK
    And todos os exercícios retornados devem ter difficulty="Intermediário"

  Scenario: Buscar exercícios por nome
    Given a biblioteca possui "Supino Reto", "Supino Inclinado", "Agachamento"
    When o usuário faz GET /api/v1/exercises?search=supino
    Then a resposta deve ser 200 OK
    And os exercícios retornados devem conter "Supino" no nome

  Scenario: Combinar múltiplos filtros
    Given a biblioteca possui diversos exercícios
    When o usuário faz GET /api/v1/exercises?muscleGroup=Peito&equipment=Barra&difficulty=Intermediário
    Then a resposta deve ser 200 OK
    And todos os exercícios devem atender aos 3 filtros

  Scenario: Listar exercícios com biblioteca vazia
    Given a biblioteca não possui exercícios (seed não aplicado)
    When o usuário faz GET /api/v1/exercises
    Then a resposta deve ser 200 OK
    And o body deve conter array vazio
    And o meta deve conter total=0

  Scenario: Obter detalhes de exercício sem autenticação
    Given um exercício "Supino Reto" existe na biblioteca
    When o usuário faz GET /api/v1/exercises/{id} sem JWT
    Then a resposta deve ser 200 OK
    And o body deve conter id, name, description, instructions, tips, difficulty, equipment, thumbnailUrl, videoUrl, muscles
    And o body NÃO deve conter userStats

  Scenario: Obter detalhes de exercício com autenticação
    Given um usuário autenticado que já executou "Supino Reto" 5 vezes
    And a melhor carga foi 80kg
    And a última execução foi em 2026-02-28
    When o usuário faz GET /api/v1/exercises/{id} com JWT válido
    Then a resposta deve ser 200 OK
    And o body deve conter userStats com:
      | lastPerformed  | 2026-02-28T10:30:00Z |
      | bestWeight     | 80000                |
      | timesPerformed | 5                    |
      | averageWeight  | 75000.5              |

  Scenario: Obter detalhes de exercício que usuário nunca executou
    Given um usuário autenticado que nunca executou "Supino Reto"
    When o usuário faz GET /api/v1/exercises/{id} com JWT válido
    Then a resposta deve ser 200 OK
    And o body deve conter userStats com:
      | lastPerformed  | null |
      | bestWeight     | null |
      | timesPerformed | 0    |
      | averageWeight  | null |

  Scenario: Obter detalhes de exercício inexistente
    Given um exerciseID que não existe no banco
    When o usuário faz GET /api/v1/exercises/{id}
    Then a resposta deve ser 404 Not Found
    And o erro deve conter "exercise not found"

  Scenario: Obter histórico de exercício
    Given um usuário autenticado que executou "Supino Reto" em 3 sessões
    And cada sessão teve 4 sets
    When o usuário faz GET /api/v1/exercises/{id}/history com JWT válido
    Then a resposta deve ser 200 OK
    And o body deve conter 3 entradas de histórico
    And cada entrada deve ter sessionId, workoutName, performedAt, sets[]
    And os sets devem estar ordenados por setNumber

  Scenario: Obter histórico com paginação
    Given um usuário autenticado que executou "Supino Reto" em 50 sessões
    When o usuário faz GET /api/v1/exercises/{id}/history?page=2&pageSize=10
    Then a resposta deve ser 200 OK
    And o body deve conter 10 entradas
    And o meta deve conter page=2, pageSize=10, total=50, totalPages=5

  Scenario: Obter histórico ordenado por mais recente
    Given um usuário executou "Supino Reto" em:
      | data       | workout  |
      | 2026-03-01 | Treino A |
      | 2026-02-28 | Treino B |
      | 2026-02-25 | Treino A |
    When o usuário faz GET /api/v1/exercises/{id}/history
    Then a resposta deve ser 200 OK
    And a primeira entrada deve ser de 2026-03-01
    And a última entrada deve ser de 2026-02-25

  Scenario: Obter histórico de exercício nunca executado
    Given um usuário autenticado que nunca executou "Supino Reto"
    When o usuário faz GET /api/v1/exercises/{id}/history
    Then a resposta deve ser 200 OK
    And o body deve conter array vazio
    And o meta deve conter total=0

  Scenario: Obter histórico sem autenticação
    Given um usuário não autenticado
    When o usuário faz GET /api/v1/exercises/{id}/history sem JWT
    Then a resposta deve ser 401 Unauthorized

  Scenario: Obter histórico com JWT inválido
    Given um JWT expirado
    When o usuário faz GET /api/v1/exercises/{id}/history com JWT inválido
    Then a resposta deve ser 401 Unauthorized

  Scenario: Obter histórico de exercício inexistente
    Given um exerciseID que não existe no banco
    When o usuário faz GET /api/v1/exercises/{id}/history com JWT válido
    Then a resposta deve ser 404 Not Found
    And o erro deve conter "exercise not found"

  Scenario: Validar paginação inválida em listagem
    Given a biblioteca possui exercícios
    When o usuário faz GET /api/v1/exercises?page=0
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "page must be >= 1"

  Scenario: Validar pageSize acima do limite
    Given a biblioteca possui exercícios
    When o usuário faz GET /api/v1/exercises?pageSize=200
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "pageSize must be <= 100"

  Scenario: Validar exerciseID inválido (não UUID)
    Given um exerciseID que não é UUID válido
    When o usuário faz GET /api/v1/exercises/invalid-id
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "invalid exercise ID"

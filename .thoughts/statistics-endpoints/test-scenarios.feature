Feature: Statistics Endpoints

  Scenario: Obter overview de estatísticas com período padrão
    Given um usuário autenticado que completou 10 workouts nos últimos 30 dias
    And o usuário executou 120 sets, 1440 reps, volume total de 115200kg
    And o usuário tem streak atual de 5 dias e longest streak de 10 dias
    When o usuário faz GET /api/v1/stats/overview
    Then a resposta deve ser 200 OK
    And o body deve conter:
      | totalWorkouts  | 10         |
      | totalSets      | 120        |
      | totalReps      | 1440       |
      | totalVolume    | 115200000  |
      | currentStreak  | 5          |
      | longestStreak  | 10         |
      | averagePerWeek | 2.3        |

  Scenario: Obter overview com período customizado
    Given um usuário autenticado
    When o usuário faz GET /api/v1/stats/overview?startDate=2026-01-01&endDate=2026-01-31
    Then a resposta deve ser 200 OK
    And o body deve conter stats do período especificado

  Scenario: Obter overview de usuário sem treinos
    Given um usuário autenticado que nunca treinou
    When o usuário faz GET /api/v1/stats/overview
    Then a resposta deve ser 200 OK
    And todos os valores devem ser zero

  Scenario: Obter overview com período inválido
    Given um usuário autenticado
    When o usuário faz GET /api/v1/stats/overview?startDate=2026-03-01&endDate=2026-02-01
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "startDate must be before endDate"

  Scenario: Obter overview com período muito longo
    Given um usuário autenticado
    When o usuário faz GET /api/v1/stats/overview?startDate=2020-01-01&endDate=2026-03-01
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "period cannot exceed 2 years"

  Scenario: Calcular streak corretamente
    Given um usuário treinou nos dias:
      | 2026-02-25 |
      | 2026-02-26 |
      | 2026-02-27 |
      | 2026-02-28 |
      | 2026-03-01 |
    When o usuário faz GET /api/v1/stats/overview
    Then currentStreak deve ser 5

  Scenario: Streak quebrado por dia sem treino
    Given um usuário treinou nos dias:
      | 2026-02-25 |
      | 2026-02-26 |
      | 2026-02-28 |
      | 2026-03-01 |
    When o usuário faz GET /api/v1/stats/overview
    Then currentStreak deve ser 2

  Scenario: Obter progressão geral (todos exercícios)
    Given um usuário autenticado com histórico de treinos
    When o usuário faz GET /api/v1/stats/progression?startDate=2026-02-01&endDate=2026-02-28
    Then a resposta deve ser 200 OK
    And o body deve conter dataPoints com date, value, change
    And exerciseId e exerciseName devem ser null

  Scenario: Obter progressão de exercício específico
    Given um usuário executou "Supino Reto" em 10 sessões
    When o usuário faz GET /api/v1/stats/progression?exerciseId={id}
    Then a resposta deve ser 200 OK
    And exerciseId deve ser o ID do Supino Reto
    And exerciseName deve ser "Supino Reto"
    And dataPoints devem conter apenas dados desse exercício

  Scenario: Calcular % de mudança na progressão
    Given um usuário executou exercício com volumes:
      | data       | volume |
      | 2026-02-01 | 10000  |
      | 2026-02-08 | 10500  |
      | 2026-02-15 | 11000  |
    When o usuário faz GET /api/v1/stats/progression
    Then o dataPoint de 2026-02-08 deve ter change=5.0
    And o dataPoint de 2026-02-15 deve ter change=4.76

  Scenario: Obter progressão de período sem treinos
    Given um usuário autenticado
    When o usuário faz GET /api/v1/stats/progression?startDate=2025-01-01&endDate=2025-01-31
    Then a resposta deve ser 200 OK
    And dataPoints deve ser array vazio

  Scenario: Obter progressão com exerciseId inválido
    Given um exerciseID que não é UUID válido
    When o usuário faz GET /api/v1/stats/progression?exerciseId=invalid
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "invalid exercise ID"

  Scenario: Obter personal records
    Given um usuário executou diversos exercícios
    And os recordes são:
      | exercício      | peso  | reps | data       |
      | Supino Reto    | 80kg  | 12   | 2026-02-28 |
      | Agachamento    | 120kg | 10   | 2026-02-25 |
      | Levantamento   | 150kg | 8    | 2026-02-20 |
    When o usuário faz GET /api/v1/stats/personal-records
    Then a resposta deve ser 200 OK
    And o body deve conter 3 records
    And cada record deve ter exerciseId, exerciseName, weight, reps, volume, achievedAt

  Scenario: Personal records ordenados por peso
    Given um usuário tem PRs de:
      | exercício   | peso  |
      | Supino      | 80kg  |
      | Agachamento | 120kg |
      | Rosca       | 40kg  |
    When o usuário faz GET /api/v1/stats/personal-records
    Then o primeiro record deve ser Agachamento (120kg)
    And o último record deve ser Rosca (40kg)

  Scenario: Personal records com apenas exercício mais usado por grupo muscular
    Given um usuário executou:
      | exercício         | grupo muscular | vezes usado | melhor peso |
      | Supino Reto       | Peito          | 20          | 80kg        |
      | Supino Inclinado  | Peito          | 10          | 75kg        |
      | Crucifixo         | Peito          | 5           | 30kg        |
    When o usuário faz GET /api/v1/stats/personal-records
    Then apenas "Supino Reto" deve aparecer (mais usado do grupo Peito)

  Scenario: Personal records limitado a top 15
    Given um usuário tem PRs de 20 exercícios diferentes
    When o usuário faz GET /api/v1/stats/personal-records
    Then a resposta deve conter no máximo 15 records

  Scenario: Desempate em personal record
    Given um usuário executou "Supino Reto" com:
      | peso | reps | data       |
      | 80kg | 12   | 2026-02-28 |
      | 80kg | 10   | 2026-02-25 |
      | 80kg | 12   | 2026-02-20 |
    When o usuário faz GET /api/v1/stats/personal-records
    Then o PR deve ser de 2026-02-28 (mais reps, depois mais recente)

  Scenario: Personal records de usuário sem treinos
    Given um usuário autenticado que nunca treinou
    When o usuário faz GET /api/v1/stats/personal-records
    Then a resposta deve ser 200 OK
    And records deve ser array vazio

  Scenario: Obter frequência de treinos (365 dias)
    Given um usuário treinou em:
      | data       | workouts |
      | 2026-03-01 | 1        |
      | 2026-02-28 | 2        |
      | 2026-02-27 | 0        |
      | 2026-02-26 | 1        |
    When o usuário faz GET /api/v1/stats/frequency
    Then a resposta deve ser 200 OK
    And o body deve conter 365 entradas
    And cada entrada deve ter date e count

  Scenario: Frequência com dias sem treino preenchidos com zero
    Given um usuário treinou apenas em 2026-03-01
    When o usuário faz GET /api/v1/stats/frequency
    Then a resposta deve conter 365 entradas
    And 364 entradas devem ter count=0
    And 1 entrada (2026-03-01) deve ter count=1

  Scenario: Frequência com período customizado
    Given um usuário autenticado
    When o usuário faz GET /api/v1/stats/frequency?startDate=2026-02-01&endDate=2026-02-28
    Then a resposta deve ser 200 OK
    And o body deve conter 28 entradas (dias de fevereiro)

  Scenario: Frequência de usuário sem treinos
    Given um usuário autenticado que nunca treinou
    When o usuário faz GET /api/v1/stats/frequency
    Then a resposta deve ser 200 OK
    And todas as 365 entradas devem ter count=0

  Scenario: Endpoints de stats sem autenticação
    Given um usuário não autenticado
    When o usuário faz GET /api/v1/stats/overview sem JWT
    Then a resposta deve ser 401 Unauthorized

  Scenario: Endpoints de stats com JWT inválido
    Given um JWT expirado
    When o usuário faz GET /api/v1/stats/overview com JWT inválido
    Then a resposta deve ser 401 Unauthorized

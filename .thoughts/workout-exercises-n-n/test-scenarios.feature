Feature: Workout Exercises N:N Relationship

  Background:
    Given a biblioteca de exercises contém os seguintes exercises:
      | id                                   | name        | description       | thumbnail_url           | muscles         |
      | e1111111-1111-1111-1111-111111111111 | Supino Reto | Exercício de peito | /assets/supino.png     | ["Peito"]       |
      | e2222222-2222-2222-2222-222222222222 | Agachamento | Exercício de pernas | /assets/agachamento.png | ["Quadríceps"] |
      | e3333333-3333-3333-3333-333333333333 | Rosca Direta | Exercício de braço | /assets/rosca.png      | ["Bíceps"]      |
    And existe um usuário "user-123" com token válido
    And o usuário possui um workout "Treino A" com id "w1111111-1111-1111-1111-111111111111"

  # ==========================================
  # CENÁRIOS: CRUD de Exercises (Biblioteca)
  # ==========================================

  Scenario: Criar um novo exercise na biblioteca
    When o sistema cria um novo exercise com os dados:
      | name          | description     | thumbnail_url         | muscles    |
      | Leg Press     | Exercício de perna | /assets/leg-press.png | ["Quadríceps"] |
    Then o exercise é criado na tabela "exercises" com sucesso
    And o exercise possui um id único
    And o exercise não está vinculado a nenhum workout

  Scenario: Buscar exercise por ID
    Given o exercise "Supino Reto" existe na biblioteca com id "e1111111-1111-1111-1111-111111111111"
    When o sistema busca o exercise pelo id "e1111111-1111-1111-1111-111111111111"
    Then o exercise "Supino Reto" é retornado
    And os campos name, description, thumbnail_url, muscles estão preenchidos

  Scenario: Atualizar metadata de um exercise na biblioteca
    Given o exercise "Supino Reto" existe na biblioteca
    When o sistema atualiza o exercise com os dados:
      | description            | thumbnail_url          |
      | Exercício de peito superior | /assets/supino-v2.png |
    Then o exercise é atualizado na tabela "exercises"
    And todos os workouts que usam esse exercise refletem a nova metadata

  Scenario: Deletar exercise da biblioteca (sem uso)
    Given o exercise "Leg Press" existe na biblioteca
    And nenhum workout usa o exercise "Leg Press"
    When o sistema deleta o exercise "Leg Press"
    Then o exercise é removido da tabela "exercises"

  Scenario: Deletar exercise da biblioteca (com uso — fail)
    Given o exercise "Supino Reto" existe na biblioteca
    And o workout "Treino A" usa o exercise "Supino Reto"
    When o sistema tenta deletar o exercise "Supino Reto"
    Then o sistema retorna erro "FK constraint violation"
    And o exercise "Supino Reto" permanece na biblioteca

  # ==========================================
  # CENÁRIOS: CRUD de WorkoutExercises (Vínculo)
  # ==========================================

  Scenario: Vincular exercise a um workout com configurações específicas
    Given o workout "Treino A" não possui exercises vinculados
    And o exercise "Supino Reto" existe na biblioteca
    When o sistema vincula o exercise "Supino Reto" ao workout "Treino A" com as configurações:
      | sets | reps  | rest_time | weight | order_index |
      | 4    | 8-12  | 90        | 80000  | 1           |
    Then um registro é criado na tabela "workout_exercises" com:
      | workout_id | exercise_id                          | sets | reps | rest_time | weight | order_index |
      | w1111111-1111-1111-1111-111111111111 | e1111111-1111-1111-1111-111111111111 | 4    | 8-12 | 90        | 80000  | 1           |
    And o registro possui um id único (workout_exercise_id)

  Scenario: Buscar exercises de um workout (com JOIN)
    Given o workout "Treino A" possui os seguintes exercises vinculados:
      | exercise_name | sets | reps  | rest_time | weight | order_index |
      | Supino Reto   | 4    | 8-12  | 90        | 80000  | 1           |
      | Agachamento   | 3    | 10-15 | 60        | 100000 | 2           |
    When o sistema busca os exercises do workout "Treino A"
    Then o sistema retorna 2 workout_exercises
    And cada workout_exercise contém:
      | campo          | origem             |
      | id             | workout_exercises  |
      | exerciseId     | workout_exercises  |
      | name           | exercises (JOIN)   |
      | description    | exercises (JOIN)   |
      | muscles        | exercises (JOIN)   |
      | sets           | workout_exercises  |
      | reps           | workout_exercises  |
      | restTime       | workout_exercises  |
      | weight         | workout_exercises  |
      | orderIndex     | workout_exercises  |
    And os exercises estão ordenados por order_index ASC

  Scenario: Atualizar configurações de um workout_exercise
    Given o workout "Treino A" possui o exercise "Supino Reto" vinculado com:
      | sets | reps | rest_time | weight |
      | 4    | 8-12 | 90        | 80000  |
    When o sistema atualiza o workout_exercise para:
      | sets | reps  | rest_time | weight |
      | 5    | 6-10  | 120       | 90000  |
    Then o registro em "workout_exercises" é atualizado
    And o exercise "Supino Reto" na biblioteca não é modificado
    And o campo updated_at é atualizado

  Scenario: Remover exercise de um workout (sem set_records)
    Given o workout "Treino A" possui o exercise "Supino Reto" vinculado
    And não existem set_records para esse workout_exercise
    When o sistema remove o workout_exercise
    Then o registro é deletado da tabela "workout_exercises"
    And o exercise "Supino Reto" permanece na biblioteca

  Scenario: Remover exercise de um workout (com set_records — fail)
    Given o workout "Treino A" possui o exercise "Supino Reto" vinculado
    And existem set_records para esse workout_exercise
    When o sistema tenta remover o workout_exercise
    Then o sistema retorna erro "FK constraint violation (ON DELETE RESTRICT)"
    And o workout_exercise permanece na tabela

  Scenario: Tentar vincular o mesmo exercise 2x ao mesmo workout (fail)
    Given o workout "Treino A" possui o exercise "Supino Reto" vinculado
    When o sistema tenta vincular novamente o exercise "Supino Reto" ao workout "Treino A"
    Then o sistema retorna erro "UNIQUE constraint violation (workout_id, exercise_id)"
    And nenhum novo registro é criado

  # ==========================================
  # CENÁRIOS: API - GET /workouts/{id}
  # ==========================================

  Scenario: Buscar workout com exercises (contrato API)
    Given o workout "Treino A" possui os seguintes exercises vinculados:
      | exercise_name | sets | reps  | rest_time | weight | order_index |
      | Supino Reto   | 4    | 8-12  | 90        | 80000  | 1           |
      | Agachamento   | 3    | 10-15 | 60        | 100000 | 2           |
    When o cliente faz GET /api/v1/workouts/w1111111-1111-1111-1111-111111111111
    Then o sistema retorna 200 OK
    And o payload contém:
      """json
      {
        "data": {
          "id": "w1111111-1111-1111-1111-111111111111",
          "name": "Treino A",
          "exercises": [
            {
              "id": "<workout_exercise_id>",
              "exerciseId": "e1111111-1111-1111-1111-111111111111",
              "name": "Supino Reto",
              "description": "Exercício de peito",
              "sets": 4,
              "reps": "8-12",
              "restTime": 90,
              "weight": 80000,
              "muscles": ["Peito"],
              "thumbnailUrl": "/assets/supino.png"
            },
            {
              "id": "<workout_exercise_id>",
              "exerciseId": "e2222222-2222-2222-2222-222222222222",
              "name": "Agachamento",
              "description": "Exercício de pernas",
              "sets": 3,
              "reps": "10-15",
              "restTime": 60,
              "weight": 100000,
              "muscles": ["Quadríceps"],
              "thumbnailUrl": "/assets/agachamento.png"
            }
          ]
        }
      }
      """

  # ==========================================
  # CENÁRIOS: Set Records (com workout_exercise_id)
  # ==========================================

  Scenario: Criar set_record usando workout_exercise_id
    Given o usuário iniciou uma session do workout "Treino A" com id "s1111111-1111-1111-1111-111111111111"
    And o workout "Treino A" possui o exercise "Supino Reto" vinculado com workout_exercise_id "we111111-1111-1111-1111-111111111111"
    When o cliente faz POST /api/v1/sessions/s1111111-1111-1111-1111-111111111111/set-records com:
      """json
      {
        "workoutExerciseId": "we111111-1111-1111-1111-111111111111",
        "setNumber": 1,
        "weight": 80000,
        "reps": 10,
        "status": "completed"
      }
      """
    Then o sistema cria um set_record em "set_records" com:
      | session_id                           | workout_exercise_id                  | set_number | weight | reps | status    |
      | s1111111-1111-1111-1111-111111111111 | we111111-1111-1111-1111-111111111111 | 1          | 80000  | 10   | completed |
    And o sistema retorna 201 Created

  Scenario: Validar que workout_exercise pertence ao workout da session
    Given o usuário iniciou uma session do workout "Treino A" com id "s1111111-1111-1111-1111-111111111111"
    And existe um workout "Treino B" com um exercise vinculado de workout_exercise_id "we222222-2222-2222-2222-222222222222"
    When o cliente tenta criar set_record com workoutExerciseId "we222222-2222-2222-2222-222222222222" na session de "Treino A"
    Then o sistema retorna 400 Bad Request
    And a mensagem de erro é "workout_exercise não pertence ao workout desta sessão"

  Scenario: Buscar set_records por workout_exercise_id
    Given existem set_records para o workout_exercise_id "we111111-1111-1111-1111-111111111111":
      | set_number | weight | reps | status    |
      | 1          | 80000  | 10   | completed |
      | 2          | 80000  | 9    | completed |
      | 3          | 80000  | 8    | completed |
    When o sistema busca set_records da session filtrados por workout_exercise_id
    Then o sistema retorna 3 set_records
    And todos possuem workout_exercise_id = "we111111-1111-1111-1111-111111111111"

  Scenario: Unique constraint em set_records (session + workout_exercise + set_number)
    Given existe um set_record com:
      | session_id | workout_exercise_id | set_number |
      | s1111111-1111-1111-1111-111111111111 | we111111-1111-1111-1111-111111111111 | 1 |
    When o sistema tenta criar outro set_record com os mesmos valores
    Then o sistema retorna erro "UNIQUE constraint violation"

  # ==========================================
  # CENÁRIOS: Migration 009 (data migration)
  # ==========================================

  Scenario: Migração de exercises existentes para biblioteca compartilhada
    Given as seguintes entradas existem na tabela "exercises" (antes da migration):
      | id  | workout_id | name        | sets | reps | rest_time | weight | muscles   |
      | ex1 | w1         | Supino Reto | 4    | 8-12 | 90        | 80000  | ["Peito"] |
      | ex2 | w2         | Supino Reto | 3    | 10   | 60        | 70000  | ["Peito"] |
    When a migration 009 é executada
    Then um único exercise "Supino Reto" é criado na biblioteca (deduplicado por name + thumbnail_url)
    And dois registros são criados em "workout_exercises":
      | workout_id | exercise_id     | sets | reps | rest_time | weight |
      | w1         | <supino_id>     | 4    | 8-12 | 90        | 80000  |
      | w2         | <supino_id>     | 3    | 10   | 60        | 70000  |

  Scenario: Migração de set_records para referenciar workout_exercises
    Given antes da migration existem set_records com:
      | session_id | exercise_id | set_number |
      | s1         | ex1         | 1          |
    And a session "s1" referencia o workout "w1"
    And o exercise "ex1" pertencia ao workout "w1"
    When a migration 009 é executada
    Then o set_record é atualizado para referenciar workout_exercise_id (vínculo w1 + ex1)
    And a coluna "exercise_id" é removida de "set_records"

  Scenario: Falha na migration se houver set_records órfãos
    Given existem set_records com exercise_id que não existe na tabela "exercises"
    When a migration 009 é executada
    Then a migration falha com erro "FK constraint violation"
    And nenhuma alteração é aplicada ao banco (rollback)

  # ==========================================
  # CENÁRIOS: Rollback / Edge Cases
  # ==========================================

  Scenario: Session ativa durante delete de workout_exercise (bloqueado)
    Given o usuário tem uma session ativa referenciando o workout "Treino A"
    And existem set_records dessa session para o workout_exercise "we111111"
    When o sistema tenta deletar o workout_exercise "we111111"
    Then o delete é bloqueado pelo constraint "ON DELETE RESTRICT"
    And a mensagem de erro indica que existem set_records vinculados

  Scenario: Editar workout_exercise não afeta set_records históricos
    Given uma session foi completada com set_records para o workout_exercise "we111111"
    And o workout_exercise tinha configuração: sets=4, reps="8-12", weight=80000
    When o sistema atualiza o workout_exercise para: sets=5, reps="6-10", weight=90000
    Then os set_records históricos mantêm a referência para "we111111"
    And os valores armazenados nos set_records (weight, reps) não são alterados
    And futuras sessions usarão a nova configuração

  Scenario: Buscar exercises com metadata desatualizada (exercise foi editado)
    Given o workout "Treino A" possui o exercise "Supino Reto" vinculado
    And o exercise "Supino Reto" tem thumbnail_url = "/assets/supino-v1.png"
    When o exercise "Supino Reto" é atualizado na biblioteca com thumbnail_url = "/assets/supino-v2.png"
    And o cliente busca o workout "Treino A"
    Then o exercise retornado possui thumbnail_url = "/assets/supino-v2.png" (metadata atualizada via JOIN)

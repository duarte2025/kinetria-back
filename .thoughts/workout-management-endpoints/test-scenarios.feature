Feature: Workout Management Endpoints

  Scenario: Criar workout customizado com sucesso
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com:
      | name        | Treino A - Peito      |
      | type        | HIPERTROFIA           |
      | intensity   | ALTA                  |
      | duration    | 60                    |
      | exercises   | 3 exercícios válidos  |
    Then a resposta deve ser 201 Created
    And o body deve conter o workout criado com id
    And o workout deve ter created_by = userID

  Scenario: Criar workout sem exercises
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com exercises vazio
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "at least one exercise is required"

  Scenario: Criar workout com exerciseID inválido
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com exerciseID que não existe
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "exercise not found"

  Scenario: Criar workout com name muito curto
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com name="AB"
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "name must be at least 3 characters"

  Scenario: Criar workout com type inválido
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com type="INVALID"
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "invalid workout type"

  Scenario: Criar workout com sets fora do range
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com exercise.sets=15
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "sets must be between 1 and 10"

  Scenario: Criar workout com restTime negativo
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com exercise.restTime=-10
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "restTime must be between 0 and 600"

  Scenario: Criar workout com orderIndex duplicado
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com 2 exercises com orderIndex=1
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "orderIndex must be unique"

  Scenario: Criar workout com mais de 20 exercises
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts com 21 exercises
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "maximum 20 exercises allowed"

  Scenario: Atualizar workout customizado com sucesso
    Given um usuário autenticado que criou um workout
    When o usuário faz PUT /api/v1/workouts/{id} com:
      | name      | Treino A - Atualizado |
      | exercises | 4 exercícios válidos  |
    Then a resposta deve ser 200 OK
    And o body deve conter o workout atualizado
    And os exercises antigos devem ser substituídos pelos novos

  Scenario: Atualizar workout de outro usuário
    Given um usuário autenticado
    And existe um workout criado por outro usuário
    When o usuário faz PUT /api/v1/workouts/{id}
    Then a resposta deve ser 403 Forbidden
    And o erro deve conter "you can only update your own workouts"

  Scenario: Atualizar workout template (created_by = NULL)
    Given um usuário autenticado
    And existe um workout template (created_by = NULL)
    When o usuário faz PUT /api/v1/workouts/{id}
    Then a resposta deve ser 403 Forbidden
    And o erro deve conter "cannot update template workouts"

  Scenario: Atualizar workout inexistente
    Given um usuário autenticado
    And um workoutID que não existe
    When o usuário faz PUT /api/v1/workouts/{id}
    Then a resposta deve ser 404 Not Found
    And o erro deve conter "workout not found"

  Scenario: Atualizar workout com validações inválidas
    Given um usuário autenticado que criou um workout
    When o usuário faz PUT /api/v1/workouts/{id} com exercises vazio
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "at least one exercise is required"

  Scenario: Deletar workout customizado com sucesso
    Given um usuário autenticado que criou um workout
    And o workout não tem sessions ativas
    When o usuário faz DELETE /api/v1/workouts/{id}
    Then a resposta deve ser 204 No Content
    And o workout deve ter deleted_at preenchido (soft delete)

  Scenario: Deletar workout de outro usuário
    Given um usuário autenticado
    And existe um workout criado por outro usuário
    When o usuário faz DELETE /api/v1/workouts/{id}
    Then a resposta deve ser 403 Forbidden
    And o erro deve conter "you can only delete your own workouts"

  Scenario: Deletar workout template
    Given um usuário autenticado
    And existe um workout template (created_by = NULL)
    When o usuário faz DELETE /api/v1/workouts/{id}
    Then a resposta deve ser 403 Forbidden
    And o erro deve conter "cannot delete template workouts"

  Scenario: Deletar workout com session ativa
    Given um usuário autenticado que criou um workout
    And o workout tem uma session com status='active'
    When o usuário faz DELETE /api/v1/workouts/{id}
    Then a resposta deve ser 409 Conflict
    And o erro deve conter "cannot delete workout with active sessions"

  Scenario: Deletar workout com sessions completadas
    Given um usuário autenticado que criou um workout
    And o workout tem apenas sessions com status='completed'
    When o usuário faz DELETE /api/v1/workouts/{id}
    Then a resposta deve ser 204 No Content
    And o workout deve ser deletado (soft delete)

  Scenario: Deletar workout inexistente
    Given um usuário autenticado
    And um workoutID que não existe
    When o usuário faz DELETE /api/v1/workouts/{id}
    Then a resposta deve ser 404 Not Found
    And o erro deve conter "workout not found"

  Scenario: GET /workouts retorna templates e workouts customizados
    Given um usuário autenticado
    And existem 5 workouts template (created_by = NULL)
    And o usuário criou 3 workouts customizados
    When o usuário faz GET /api/v1/workouts
    Then a resposta deve ser 200 OK
    And o body deve conter 8 workouts (5 templates + 3 customizados)

  Scenario: GET /workouts não retorna workouts deletados
    Given um usuário autenticado que criou 3 workouts
    And 1 workout foi deletado (deleted_at preenchido)
    When o usuário faz GET /api/v1/workouts
    Then a resposta deve ser 200 OK
    And o body deve conter apenas 2 workouts (não deletados)

  Scenario: GET /workouts não retorna workouts de outros usuários
    Given um usuário autenticado
    And outro usuário criou 5 workouts customizados
    When o usuário faz GET /api/v1/workouts
    Then a resposta deve ser 200 OK
    And o body NÃO deve conter os workouts do outro usuário

  Scenario: Transação rollback ao falhar criação de workout_exercises
    Given um usuário autenticado
    When o usuário faz POST /api/v1/workouts
    And a criação do workout sucede
    And a criação de workout_exercises falha
    Then a transação deve fazer rollback
    And o workout NÃO deve ser criado no banco

  Scenario: Criar workout sem autenticação
    Given um usuário não autenticado
    When o usuário faz POST /api/v1/workouts sem JWT
    Then a resposta deve ser 401 Unauthorized

  Scenario: Atualizar workout sem autenticação
    Given um usuário não autenticado
    When o usuário faz PUT /api/v1/workouts/{id} sem JWT
    Then a resposta deve ser 401 Unauthorized

  Scenario: Deletar workout sem autenticação
    Given um usuário não autenticado
    When o usuário faz DELETE /api/v1/workouts/{id} sem JWT
    Then a resposta deve ser 401 Unauthorized

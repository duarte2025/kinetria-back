Feature: Profile Endpoints

  Scenario: Obter perfil do usuário autenticado
    Given um usuário autenticado com ID "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And o usuário possui name "João Silva", email "joao@example.com"
    And o usuário possui preferences {"theme": "dark", "language": "pt-BR"}
    When o usuário faz GET /api/v1/profile com JWT válido
    Then a resposta deve ser 200 OK
    And o body deve conter id, name, email, profileImageUrl, preferences

  Scenario: Obter perfil sem autenticação
    Given um usuário não autenticado
    When o usuário faz GET /api/v1/profile sem JWT
    Then a resposta deve ser 401 Unauthorized

  Scenario: Atualizar name do perfil
    Given um usuário autenticado com name "João Silva"
    When o usuário faz PATCH /api/v1/profile com {"name": "João Pedro Silva"}
    Then a resposta deve ser 200 OK
    And o name do usuário deve ser "João Pedro Silva"
    And o campo updated_at deve ser atualizado

  Scenario: Atualizar preferences do perfil
    Given um usuário autenticado com preferences {"theme": "light", "language": "pt-BR"}
    When o usuário faz PATCH /api/v1/profile com {"preferences": {"theme": "dark"}}
    Then a resposta deve ser 200 OK
    And o theme do usuário deve ser "dark"
    And o language do usuário deve permanecer "pt-BR"

  Scenario: Atualizar profileImageUrl
    Given um usuário autenticado sem profileImageUrl
    When o usuário faz PATCH /api/v1/profile com {"profileImageUrl": "https://cdn.example.com/avatar.jpg"}
    Then a resposta deve ser 200 OK
    And o profileImageUrl do usuário deve ser "https://cdn.example.com/avatar.jpg"

  Scenario: Atualizar múltiplos campos simultaneamente
    Given um usuário autenticado
    When o usuário faz PATCH /api/v1/profile com {"name": "Maria", "preferences": {"theme": "dark"}}
    Then a resposta deve ser 200 OK
    And o name do usuário deve ser "Maria"
    And o theme do usuário deve ser "dark"

  Scenario: Atualizar perfil com name vazio
    Given um usuário autenticado
    When o usuário faz PATCH /api/v1/profile com {"name": ""}
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "name must be between 2 and 100 characters"

  Scenario: Atualizar perfil com name muito longo
    Given um usuário autenticado
    When o usuário faz PATCH /api/v1/profile com name de 101 caracteres
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "name must be between 2 and 100 characters"

  Scenario: Atualizar perfil com name apenas espaços
    Given um usuário autenticado
    When o usuário faz PATCH /api/v1/profile com {"name": "   "}
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "name cannot be only whitespace"

  Scenario: Atualizar preferences com theme inválido
    Given um usuário autenticado
    When o usuário faz PATCH /api/v1/profile com {"preferences": {"theme": "invalid"}}
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "theme must be 'dark' or 'light'"

  Scenario: Atualizar preferences com language inválido
    Given um usuário autenticado
    When o usuário faz PATCH /api/v1/profile com {"preferences": {"language": "invalid"}}
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "language must be 'pt-BR' or 'en-US'"

  Scenario: Atualizar perfil sem enviar nenhum campo
    Given um usuário autenticado
    When o usuário faz PATCH /api/v1/profile com body vazio {}
    Then a resposta deve ser 400 Bad Request
    And o erro deve conter "at least one field must be provided"

  Scenario: Atualizar perfil com JWT inválido
    Given um usuário com JWT expirado
    When o usuário faz PATCH /api/v1/profile com {"name": "João"}
    Then a resposta deve ser 401 Unauthorized

  Scenario: Atualizar perfil de usuário inexistente
    Given um JWT válido com userID que não existe no banco
    When o usuário faz PATCH /api/v1/profile com {"name": "João"}
    Then a resposta deve ser 404 Not Found
    And o erro deve conter "user not found"

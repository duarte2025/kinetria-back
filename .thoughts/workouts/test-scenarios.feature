Feature: Lista de Workouts do Usuário
  Como um usuário autenticado da plataforma Kinetria
  Eu quero listar meus workouts cadastrados
  Para poder visualizar e escolher qual treino executar

  Background:
    Given que a tabela "workouts" existe no banco de dados
    And que o middleware de autenticação está configurado

  # ──────────────────────────────────────────────────────────────
  # Happy Path
  # ──────────────────────────────────────────────────────────────

  Scenario: Listar workouts com sucesso (primeira página)
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem 5 workouts cadastrados para este usuário
    When eu faço uma requisição GET para "/api/v1/workouts"
    Then o status da resposta deve ser 200
    And o corpo da resposta deve conter um array "data" com 5 elementos
    And cada elemento deve ter os campos:
      | campo       | tipo   | obrigatório |
      | id          | uuid   | sim         |
      | name        | string | sim         |
      | description | string | não         |
      | type        | string | não         |
      | intensity   | string | não         |
      | duration    | int    | sim         |
      | imageUrl    | string | não         |
    And o campo "meta" deve conter:
      | campo      | valor |
      | page       | 1     |
      | pageSize   | 20    |
      | total      | 5     |
      | totalPages | 1     |

  Scenario: Listar workouts com paginação customizada
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem 25 workouts cadastrados para este usuário
    When eu faço uma requisição GET para "/api/v1/workouts?page=2&pageSize=10"
    Then o status da resposta deve ser 200
    And o array "data" deve conter 10 elementos
    And o campo "meta" deve conter:
      | campo      | valor |
      | page       | 2     |
      | pageSize   | 10    |
      | total      | 25    |
      | totalPages | 3     |

  Scenario: Listar workouts com valores default de paginação
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem 50 workouts cadastrados para este usuário
    When eu faço uma requisição GET para "/api/v1/workouts"
    Then o status da resposta deve ser 200
    And o array "data" deve conter 20 elementos
    And o campo "meta.page" deve ser 1
    And o campo "meta.pageSize" deve ser 20
    And o campo "meta.total" deve ser 50
    And o campo "meta.totalPages" deve ser 3

  Scenario: Workouts ordenados por data de criação (mais recente primeiro)
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem workouts cadastrados nas seguintes datas:
      | nome            | created_at          |
      | Treino A        | 2026-02-20 10:00:00 |
      | Treino B        | 2026-02-22 15:30:00 |
      | Treino C        | 2026-02-21 08:00:00 |
    When eu faço uma requisição GET para "/api/v1/workouts"
    Then o status da resposta deve ser 200
    And os workouts devem estar ordenados por:
      | posição | nome     |
      | 1       | Treino B |
      | 2       | Treino C |
      | 3       | Treino A |

  # ──────────────────────────────────────────────────────────────
  # Edge Cases
  # ──────────────────────────────────────────────────────────────

  Scenario: Usuário sem workouts cadastrados
    Given que estou autenticado como usuário "new-user-uuid"
    And que NÃO existem workouts cadastrados para este usuário
    When eu faço uma requisição GET para "/api/v1/workouts"
    Then o status da resposta deve ser 200
    And o array "data" deve estar vazio
    And o campo "meta" deve conter:
      | campo      | valor |
      | page       | 1     |
      | pageSize   | 20    |
      | total      | 0     |
      | totalPages | 0     |

  Scenario: Solicitar página além do total disponível
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem 5 workouts cadastrados para este usuário
    When eu faço uma requisição GET para "/api/v1/workouts?page=10&pageSize=20"
    Then o status da resposta deve ser 200
    And o array "data" deve estar vazio
    And o campo "meta" deve conter:
      | campo      | valor |
      | page       | 10    |
      | pageSize   | 20    |
      | total      | 5     |
      | totalPages | 1     |

  Scenario: PageSize maior que o total de workouts
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem 3 workouts cadastrados para este usuário
    When eu faço uma requisição GET para "/api/v1/workouts?pageSize=100"
    Then o status da resposta deve ser 200
    And o array "data" deve conter 3 elementos
    And o campo "meta" deve conter:
      | campo      | valor |
      | page       | 1     |
      | pageSize   | 100   |
      | total      | 3     |
      | totalPages | 1     |

  Scenario: Workout com campos opcionais vazios
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existe 1 workout com os seguintes dados:
      | campo       | valor                                |
      | id          | b1b2c3d4-e5f6-7890-abcd-ef1234567890 |
      | name        | Treino Minimalista                   |
      | description |                                      |
      | type        |                                      |
      | intensity   |                                      |
      | duration    | 30                                   |
      | imageUrl    |                                      |
    When eu faço uma requisição GET para "/api/v1/workouts"
    Then o status da resposta deve ser 200
    And o primeiro elemento do array "data" deve ter:
      | campo       | valor              | tipo |
      | id          | (uuid válido)      | uuid |
      | name        | Treino Minimalista | string |
      | description | null               | null |
      | type        | null               | null |
      | intensity   | null               | null |
      | duration    | 30                 | int  |
      | imageUrl    | null               | null |

  Scenario: Isolamento de dados entre usuários
    Given que existem os seguintes usuários e workouts:
      | user_id                              | nome_workout       |
      | user-a-uuid                          | Treino A do User A |
      | user-a-uuid                          | Treino B do User A |
      | user-b-uuid                          | Treino A do User B |
    And que estou autenticado como usuário "user-a-uuid"
    When eu faço uma requisição GET para "/api/v1/workouts"
    Then o status da resposta deve ser 200
    And o array "data" deve conter 2 elementos
    And todos os workouts retornados devem pertencer ao usuário "user-a-uuid"
    And nenhum workout do usuário "user-b-uuid" deve ser retornado

  # ──────────────────────────────────────────────────────────────
  # Sad Paths - Validação de Input
  # ──────────────────────────────────────────────────────────────

  Scenario: Page com valor zero
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    When eu faço uma requisição GET para "/api/v1/workouts?page=0"
    Then o status da resposta deve ser 422
    And o corpo da resposta deve conter:
      | campo   | valor                              |
      | code    | VALIDATION_ERROR                   |
      | message | page must be greater than or equal to 1 |

  Scenario: Page com valor negativo
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    When eu faço uma requisição GET para "/api/v1/workouts?page=-5"
    Then o status da resposta deve ser 422
    And o corpo da resposta deve conter:
      | campo   | valor                              |
      | code    | VALIDATION_ERROR                   |
      | message | page must be greater than or equal to 1 |

  Scenario: PageSize com valor zero
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    When eu faço uma requisição GET para "/api/v1/workouts?pageSize=0"
    Then o status da resposta deve ser 422
    And o corpo da resposta deve conter:
      | campo   | valor                                  |
      | code    | VALIDATION_ERROR                       |
      | message | pageSize must be between 1 and 100     |

  Scenario: PageSize acima do limite máximo
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    When eu faço uma requisição GET para "/api/v1/workouts?pageSize=101"
    Then o status da resposta deve ser 422
    And o corpo da resposta deve conter:
      | campo   | valor                                  |
      | code    | VALIDATION_ERROR                       |
      | message | pageSize must be between 1 and 100     |

  Scenario: Page com valor não numérico
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    When eu faço uma requisição GET para "/api/v1/workouts?page=abc"
    Then o status da resposta deve ser 422
    And o corpo da resposta deve conter:
      | campo   | valor                                  |
      | code    | VALIDATION_ERROR                       |
      | message | page must be a valid integer           |

  Scenario: PageSize com valor não numérico
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    When eu faço uma requisição GET para "/api/v1/workouts?pageSize=xyz"
    Then o status da resposta deve ser 422
    And o corpo da resposta deve conter:
      | campo   | valor                                  |
      | code    | VALIDATION_ERROR                       |
      | message | pageSize must be a valid integer       |

  # ──────────────────────────────────────────────────────────────
  # Sad Paths - Autenticação
  # ──────────────────────────────────────────────────────────────

  Scenario: Requisição sem token de autenticação
    When eu faço uma requisição GET para "/api/v1/workouts" sem o header "Authorization"
    Then o status da resposta deve ser 401
    And o corpo da resposta deve conter:
      | campo   | valor                                  |
      | code    | UNAUTHORIZED                           |
      | message | Missing authentication token           |

  Scenario: Requisição com token JWT inválido
    When eu faço uma requisição GET para "/api/v1/workouts" com o header:
      | campo         | valor             |
      | Authorization | Bearer token-invalido |
    Then o status da resposta deve ser 401
    And o corpo da resposta deve conter:
      | campo   | valor                                  |
      | code    | UNAUTHORIZED                           |
      | message | Invalid or expired access token        |

  Scenario: Requisição com token JWT expirado
    Given que existe um token JWT expirado para o usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    When eu faço uma requisição GET para "/api/v1/workouts" com este token expirado
    Then o status da resposta deve ser 401
    And o corpo da resposta deve conter:
      | campo   | valor                                  |
      | code    | UNAUTHORIZED                           |
      | message | Invalid or expired access token        |

  # ──────────────────────────────────────────────────────────────
  # Sad Paths - Infraestrutura
  # ──────────────────────────────────────────────────────────────

  Scenario: Erro ao consultar banco de dados
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que o banco de dados está indisponível
    When eu faço uma requisição GET para "/api/v1/workouts"
    Then o status da resposta deve ser 500
    And o corpo da resposta deve conter:
      | campo   | valor                                  |
      | code    | INTERNAL_ERROR                         |
      | message | An unexpected error occurred           |
    And o erro deve ser logado com nível "error"
    And nenhuma informação sensível deve ser exposta no response

  # ──────────────────────────────────────────────────────────────
  # Observabilidade
  # ──────────────────────────────────────────────────────────────

  Scenario: Logs estruturados são gerados para requisições bem-sucedidas
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem 10 workouts cadastrados para este usuário
    When eu faço uma requisição GET para "/api/v1/workouts?page=1&pageSize=5"
    Then o status da resposta deve ser 200
    And deve ser gerado um log estruturado com os campos:
      | campo       | tipo   | exemplo                              |
      | level       | string | info                                 |
      | message     | string | list_workouts_success                |
      | user_id     | uuid   | a1b2c3d4-e5f6-7890-abcd-ef1234567890 |
      | page        | int    | 1                                    |
      | page_size   | int    | 5                                    |
      | total       | int    | 10                                   |
      | duration_ms | int    | (tempo de execução)                  |

  Scenario: Logs de erro não expõem informações sensíveis
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que ocorre um erro de conexão com o banco de dados
    When eu faço uma requisição GET para "/api/v1/workouts"
    Then o status da resposta deve ser 500
    And o log de erro NÃO deve conter:
      | campo sensível    |
      | connection string |
      | passwords         |
      | tokens            |
      | user PII          |
    And o log deve conter:
      | campo     | valor              |
      | level     | error              |
      | message   | db_query_failed    |
      | user_id   | (uuid do usuário)  |
      | error     | (mensagem genérica)|

  # ──────────────────────────────────────────────────────────────
  # Performance (nice-to-have, não bloqueia MVP)
  # ──────────────────────────────────────────────────────────────

  @performance @skip-ci
  Scenario: Resposta deve ser retornada em menos de 200ms (p95)
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem 100 workouts cadastrados para este usuário
    When eu faço 100 requisições GET para "/api/v1/workouts?page=1&pageSize=20"
    Then 95% das requisições devem retornar em menos de 200ms
    And todas as requisições devem retornar status 200

  @performance @skip-ci
  Scenario: Query com LIMIT e OFFSET deve evitar full table scan
    Given que estou autenticado como usuário "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    And que existem 10000 workouts cadastrados para este usuário
    When eu faço uma requisição GET para "/api/v1/workouts?page=50&pageSize=20"
    Then a query executada no banco deve usar LIMIT e OFFSET
    And a query deve utilizar o índice em (user_id, created_at)
    And o tempo de resposta deve ser inferior a 500ms

---
name: Code Reviewer Instructions
description: Instruções para reviews de código (PR/changes) no projeto Kinetria.
applyTo: '**/*'
---

# Instruções — Code Reviewer (Kinetria)

Você está fazendo um **code review**. O objetivo é produzir feedback **acionável**, com **parecer claro** e **evidência**. O repositório segue arquitetura **hexagonal** e DI via **Fx**; respeite os padrões descritos em `.github/instructions/global.instructions.md`.

## Regras obrigatórias

1) **Comentários por seção (obrigatório):** você DEVE adicionar um comentário para **cada seção** do template abaixo, mesmo quando estiver tudo OK. Cada seção precisa conter:
- `Parecer:` (uma das opções: `OK`, `OK com ressalvas`, `Solicitar mudanças`)
- `Justificativa:` (1-3 frases objetivas)
- `Evidência:` (arquivo/símbolo/trecho ou comportamento observado; se não aplicável, escrever `N/A`)
- `Ação:` (o que deve mudar ou como validar; se não aplicável, escrever `N/A`)

2) **Idioma:** todos os comentários devem ser em **português do Brasil (PT-BR)**.

3) **Severidade (quando houver problema):** rotule cada achado como:
- `blocker`: impede merge (bug provável, risco de incidente, quebra de contrato, perda/corrupção de dados, falha de segurança)
- `high`: alto risco, deve ser resolvido antes do merge na maioria dos casos
- `medium`: importante, mas pode ser negociado (depende do contexto)
- `low`: melhoria incremental
- `nit`: detalhe/estilo; não deve bloquear

4) **Gates de aprovação:**
- Não aprovar com `blocker` sem correção.
- Se houver `high`, deixe explícito se é “solicitar mudanças” ou se existe mitigação/decisão registrada.
- Se houver trade-off, peça **decisão explícita** (quem decide e qual opção).

5) **Formato de achado (quando aplicável):**
- `Problema:`
- `Risco:`
- `Recomendação:`
- `Como validar:` (teste, comando, cenário, métrica/trace)

## Template obrigatório do review (copiar e preencher)

> Observação: Mesmo que o PR esteja excelente, escreva `Parecer: OK` em cada seção.

### 1) Resumo e parecer final
- Parecer final: (Aprovar / Aprovar com ressalvas / Solicitar mudanças)
- Contexto entendido: (1-2 frases)
- Principais riscos (se houver):

### 2) Correção funcional e regras de domínio
Checar também (eventos assíncronos): quando houver publish de eventos, valide se o estado relevante foi **persistido** antes do publish e se o fluxo evita condição de corrida (padrão outbox/transactional outbox, ou mecanismo equivalente). Considere ainda idempotência do consumidor e replay.
- Parecer:
- Justificativa:
- Evidência:
- Ação:

### 3) Contratos e compatibilidade (rollout)
Inclui mudanças de payload/rotas/eventos, migrações, backward compatibility e versionamento.
- Parecer:
- Justificativa:
- Evidência:
- Ação:

### 4) Segurança (authn/authz, validação, PII/secrets)
Checar: autorização por recurso, validação de input, risco de SSRF/injection, logs sem PII/secrets.
- Parecer:
- Justificativa:
- Evidência:
- Ação:

### 5) Persistência / DB (quando houver)
Checar: migrations seguras (expand/contract), locks, backfill, constraints, índices, N+1.

Obrigatório (queries): para cada query nova/alterada, verifique se as colunas usadas em `WHERE`, `JOIN` e `ORDER BY` possuem índice/constraint compatível (ex.: índice composto, unique). Se não existir índice declarado adequado, isso DEVE ser informado no review (com severidade e recomendação de criação/ajuste de índice).

Obrigatório (locks): para cada alteração relevante de acesso ao banco (queries novas/alteradas, transações, migrations, criação/remoção de índices/constraints), inclua um parecer explícito sobre a **possibilidade de locks** e impacto em concorrência/latência. Se houver risco de lock prolongado (ex.: alteração de coluna em tabela grande, backfill sem batch, índice sem estratégia compatível, transações longas), isso DEVE ser informado no review com recomendação de mitigação (expand/contract, batch/backfill, índices concorrentes quando aplicável, timeout, redução de escopo transacional).
- Parecer:
- Justificativa:
- Evidência:
- Ação:

### 6) Performance e escalabilidade
Checar: hot paths, alocações/GC, contenção de locks, fan-out de goroutines, roundtrips, timeouts.
- Parecer:
- Justificativa:
- Evidência:
- Ação:

### 7) Observabilidade (logs/métricas/tracing)
Checar: logs úteis e não ruidosos, métricas sem alta cardinalidade, tracing nas bordas (HTTP/DB).
- Parecer:
- Justificativa:
- Evidência:
- Ação:

### 8) Testes e validação
Checar: cobertura do fluxo principal, testes de erro, contratos, e como reproduzir/validar.

Obrigatório (formato): avalie se os testes (novos/alterados) estão no padrão **table-driven** (table tests) quando aplicável. Se não estiverem, registre no review o motivo/impacto e a ação sugerida (ex.: refatorar para table tests, ou justificar exceção quando não fizer sentido).
- Parecer:
- Justificativa:
- Evidência:
- Ação:

### 9) Manutenibilidade e design
Checar: separação de camadas (domain/ports/gateways), acoplamento, legibilidade, nomes, complexidade.
- Parecer:
- Justificativa:
- Evidência:
- Ação:

### 10) Estilo e consistência com o projeto
Checar: aderência aos padrões do repo e organização por feature/pacote.
- Parecer:
- Justificativa:
- Evidência:
- Ação:

## Dicas de postura (para evitar ping-pong)

- Um comentário = um tema; evite misturar múltiplos assuntos na mesma thread.
- Se algo estiver ambíguo, peça esclarecimento antes de sugerir refactor grande.
- Prefira recomendações incrementalmente testáveis.

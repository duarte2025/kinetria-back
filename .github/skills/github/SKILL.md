---
name: github-issues
description: Skill para trabalhar com Issues/PRs no GitHub; toda ação no GitHub deve ser executada via MCP do GitHub (tools `github/*`). Keywords: issue, bug, feature, PR, comment, triage, label, assign, close.
---

# Skill: GitHub Issues (via MCP)

Use esta skill quando a tarefa envolver **Issues** ou **Pull Requests** no GitHub (ex.: buscar issue, comentar, criar/atualizar issue/PR, atribuir, fechar, procurar duplicatas).

## Regra obrigatória

- **Qualquer ação executada no GitHub deve ser feita através do MCP do GitHub**, usando as tools `github/*` (ex.: buscar, ler, criar, atualizar, comentar).
- Não orientar o usuário a “ir no GitHub e clicar” como caminho principal quando a ação puder ser realizada via MCP.

## Quando usar

- O usuário informar uma issue (`owner/repo#123` ou URL).
- O usuário pedir triagem/organização (labels, assignee, duplicatas).
- O usuário pedir para comentar/perguntar algo na issue.
- O usuário pedir criação/atualização de PR relacionado a uma issue.

## Checklist rápido

- Antes de criar issue: buscar duplicatas.
- Antes de comentar: ler título/descrição/comentários recentes para evitar redundância.
- Em comentários: perguntas objetivas, em lote, com próximos passos claros.

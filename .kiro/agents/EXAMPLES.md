# Exemplo de Uso dos Agents

Este documento mostra exemplos práticos de como usar os agents do Kiro.

## Cenário 1: Implementar nova feature (workflow completo)

### Contexto
Você precisa implementar um endpoint para processar pagamentos via webhook.

### Passo 1: Research

```bash
kiro chat
```

Prompt:
```
Use o agent rpi-researcher para pesquisar como implementar um webhook de pagamento.

Contexto:
- Serviço: payment-service
- Objetivo: receber notificações de pagamento aprovado/rejeitado
- Provider: Stripe
- Preciso persistir o status e publicar evento
```

O agent vai:
1. Fazer perguntas técnicas (contrato, payload, idempotência, etc)
2. Delegar análise AS-IS para `code-analyzer`
3. Delegar análise de eventos para `architect-event-sourcing`
4. Delegar análise de DB para `architect-database`
5. Produzir `.thoughts/payment-webhook/research-report.md`

### Passo 2: Plan

Prompt:
```
Use o agent rpi-planner para criar o plano de implementação baseado no research em .thoughts/payment-webhook/
```

O agent vai:
1. Ler `research-report.md` e outros artefatos
2. Consolidar AS-IS
3. Propor TO-BE (endpoint, contrato, persistência, eventos)
4. Escrever cenários BDD em `test-scenarios.feature`
5. Criar backlog detalhado em `tasks.md`

### Passo 3: Implement

Prompt:
```
Use o agent rpi-implement para executar o backlog em .thoughts/payment-webhook/tasks.md
```

O agent vai:
1. Criar branch `feat/payment-service/webhook`
2. Para cada task:
   - Delegar ao `rpi-developer`
   - Rodar testes
   - Commitar (1 task = 1 commit)
3. Abrir Pull Request ao final

## Cenário 2: Analisar código existente

### Contexto
Você precisa entender como funciona o fluxo de autenticação.

### Comando

```bash
kiro chat
```

Prompt:
```
Use o agent code-analyzer para mapear o fluxo de autenticação.

Contexto:
- Serviço: auth-service
- Entrypoint: POST /api/v1/auth/login
```

O agent vai:
1. Localizar o handler em `internal/auth/`
2. Mapear call chain (handler → usecase → gateway)
3. Identificar side effects (DB, cache, tokens)
4. Documentar observabilidade
5. Produzir `.thoughts/auth-flow/as-is-flow-report.md`

## Cenário 3: Corrigir bug

### Contexto
Há um nil pointer panic no processamento de refund.

### Comando

```bash
kiro chat
```

Prompt:
```
Use o agent fix-developer para corrigir o bug de nil pointer no refund.

Contexto:
- Serviço: payment-service
- Erro: panic: runtime error: invalid memory address or nil pointer dereference
- Stack trace: internal/payment/usecase/refund.go:45
```

O agent vai:
1. Reproduzir o problema
2. Localizar causa raiz
3. Propor fix mínimo
4. Adicionar teste de regressão
5. Commitar: `fix(payment-service): handle nil payment in refund`
6. Produzir `.thoughts/bug-refund-nil/fix-report.md`

## Cenário 4: Análise arquitetural

### Contexto
Você precisa avaliar se deve criar um novo serviço ou estender um existente.

### Comando

```bash
kiro chat
```

Prompt:
```
Use o agent architect-backend para analisar se devo criar um novo serviço de notificações ou estender o user-service.

Contexto:
- Objetivo: enviar emails, SMS e push notifications
- Volumetria: ~10k notificações/dia
- Integrações: SendGrid, Twilio, FCM
```

O agent vai:
1. Analisar AS-IS (como notificações são enviadas hoje)
2. Propor TO-BE (service boundaries, contratos)
3. Avaliar trade-offs (novo serviço vs extensão)
4. Mapear riscos e dependências
5. Produzir `.thoughts/notification-service/backend-architecture-report.md`

## Cenário 5: Mudança de schema

### Contexto
Você precisa adicionar uma coluna e índice em uma tabela grande.

### Comando

```bash
kiro chat
```

Prompt:
```
Use o agent architect-database para analisar a adição de coluna 'status' na tabela 'orders'.

Contexto:
- Tabela: orders (~5M registros)
- Nova coluna: status (enum: pending, processing, completed, failed)
- Necessidade: filtrar por status + created_at
```

O agent vai:
1. Analisar schema atual
2. Propor migration
3. Sugerir índice composto (status, created_at)
4. Avaliar impacto em queries existentes
5. Estimar downtime
6. Produzir `.thoughts/orders-status/database-architecture-report.md`

## Cenário 6: Event sourcing

### Contexto
Você precisa implementar saga para processar pedido (order → payment → shipping).

### Comando

```bash
kiro chat
```

Prompt:
```
Use o agent architect-event-sourcing para analisar a implementação de saga para processamento de pedido.

Contexto:
- Agregados: Order, Payment, Shipping
- Eventos: OrderCreated, PaymentProcessed, ShippingScheduled
- Compensações: OrderCancelled, PaymentRefunded
```

O agent vai:
1. Analisar eventos atuais
2. Propor novos eventos e agregados
3. Desenhar saga/process manager
4. Mapear idempotência e ordenação
5. Avaliar outbox pattern
6. Produzir `.thoughts/order-saga/event-sourcing-architecture-report.md`

## Cenário 7: Documentação

### Contexto
Você precisa documentar decisões arquiteturais do projeto.

### Comando

```bash
kiro chat
```

Prompt:
```
Use o agent architect-docs para propor estrutura de ADRs e runbooks.

Contexto:
- Projeto: e-commerce platform
- Necessidade: documentar decisões de arquitetura e procedimentos operacionais
```

O agent vai:
1. Analisar documentação existente
2. Propor estrutura de ADRs
3. Sugerir runbooks para cenários comuns
4. Definir templates
5. Produzir `.thoughts/docs-structure/docs-architecture-report.md`

## Dicas

### Delegação de subagents

Os agents principais (`rpi-researcher`, `rpi-implement`) delegam automaticamente para agents especializados. Você não precisa chamar manualmente, mas pode se quiser análise isolada.

### Iteração

Você pode iterar sobre os artefatos:

```bash
# Refinar research
kiro chat "Use o agent rpi-researcher para adicionar análise de segurança ao research em .thoughts/payment-webhook/"

# Ajustar plano
kiro chat "Use o agent rpi-planner para adicionar task de migration ao plano em .thoughts/payment-webhook/"
```

### Combinação

Você pode combinar agents:

```bash
kiro chat "Use o agent code-analyzer para mapear o fluxo atual e depois o architect-backend para propor melhorias."
```

### Contexto

Sempre forneça contexto relevante:
- Serviço/domínio
- Objetivo
- Constraints (volumetria, latência, etc)
- Integrações
- Riscos conhecidos

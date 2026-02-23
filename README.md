# Kinetria Backend

Backend do projeto Kinetria seguindo arquitetura hexagonal (Ports and Adapters).

## Estrutura do Projeto

```
kinetria-back/
├── cmd/kinetria/api/        # Ponto de entrada da aplicação
├── internal/kinetria/       # Código fonte do serviço
│   ├── domain/              # Lógica de negócio pura
│   │   ├── entities/        # Entidades de domínio
│   │   ├── vos/             # Value Objects
│   │   ├── ports/           # Interfaces (contratos)
│   │   ├── errors/          # Erros de domínio
│   │   ├── validators/      # Validadores de domínio
│   │   └── services/        # Serviços de domínio
│   ├── gateways/            # Adaptadores externos
│   │   ├── http/            # Handlers HTTP
│   │   ├── events/          # Handlers e Publishers de eventos
│   │   ├── repositories/    # Repositórios (banco de dados)
│   │   └── config/          # Configurações do serviço
│   ├── extensions/          # Utilitários internos
│   ├── telemetry/           # Observabilidade customizada
│   └── tests/               # Testes de integração
└── pkg/                     # Pacotes compartilhados

```

## Requisitos

- Go 1.21+
- PostgreSQL (se usar banco de dados)
- Make

## Configuração

1. Copie o arquivo de exemplo de variáveis de ambiente:
```bash
cp .env.example .env
```

2. Ajuste as variáveis de ambiente conforme necessário

## Comandos Disponíveis

```bash
make help              # Mostra todos os comandos disponíveis
make run               # Executa a aplicação
make build             # Compila a aplicação
make test              # Executa os testes
make test-coverage     # Executa os testes com cobertura
make lint              # Executa o linter
make sqlc              # Gera código a partir das queries SQL
make mocks             # Gera mocks das interfaces
make tidy              # Organiza as dependências
make deps              # Instala as dependências
```

## Desenvolvimento

Este projeto segue as convenções definidas em `.kiro/instructions/golang-hexagonal.md`.

### Adicionando uma nova feature

1. Defina as entidades em `internal/kinetria/domain/entities/`
2. Crie as interfaces (ports) em `internal/kinetria/domain/ports/`
3. Implemente o use case em `internal/kinetria/domain/{feature}/`
4. Implemente os adaptadores em `internal/kinetria/gateways/`
5. Registre as dependências no `cmd/kinetria/api/main.go`

## Arquitetura

O projeto segue os princípios da arquitetura hexagonal:

- **Domain**: Contém a lógica de negócio pura, independente de frameworks
- **Ports**: Interfaces que definem contratos entre camadas
- **Gateways**: Implementações concretas dos ports (HTTP, DB, eventos, etc)
- **Use Cases**: Orquestração da lógica de negócio

## Injeção de Dependências

Utilizamos `go.uber.org/fx` para injeção de dependências. Todas as dependências são registradas no `main.go`.

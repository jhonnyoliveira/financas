# Finanças Pessoais

Sistema web local para controle financeiro pessoal, executável como serviço do Windows, usando SQLite e porta HTTP configurável.

## Objetivo

Entregar um MVP funcional, offline-first, para uso pessoal, com contas bancárias, saldos, receitas, despesas, transferências, cartões de crédito, faturas e dashboards. Este repositório também serve para validar desenvolvimento orientado a issues por agente de IA.

## Escopo do MVP

- Cadastro de contas bancárias (nome, tipo, saldo inicial, data de referência, status).
- Extrato, saldos atuais e saldos projetados por conta.
- Receitas, despesas, categorias, datas de competência e vencimento, e baixa de pagamentos/recebimentos.
- Transferências entre contas, sem contabilizar como receita ou despesa.
- Cadastro de cartões com limite opcional, dia de fechamento e dia de vencimento da fatura.
- Compras em cartão à vista ou parceladas, faturas por competência, pagamento total ou parcial da fatura por conta bancária.
- Dashboards de receitas, despesas, saldo, fluxo de caixa e gastos por categoria.
- Configuração de porta, endereço de escuta e caminho do banco.
- Serviço Windows com início automático; execução em modo console no Linux/Windows para desenvolvimento.
- Interface em português (pt-BR), moeda BRL e calendário local configurado para `America/Sao_Paulo`.

## Fora do escopo do MVP

Open Finance, sincronização bancária, multiusuário, autenticação remota, acesso público pela internet, anexos, OCR, investimentos, exportação contábil, notificações, recorrência automática e aplicativos móveis. Não implementar sem issue específica.

## Stack e arquitetura

- **Backend:** Go 1.25+ (ou versão estável suportada na implementação), `net/http` e módulos internos por domínio.
- **Banco:** SQLite via `database/sql` e `modernc.org/sqlite` (sem CGO); migrations versionadas.
- **Frontend:** HTML/CSS/JavaScript simples, sem build obrigatório; Chart.js local para gráficos.
- **Entrega:** frontend incorporado ao executável com `embed`; binário Windows amd64 único (exceto arquivos persistentes de configuração e dados).
- **Serviço:** integração Windows SCM com `golang.org/x/sys/windows/svc` e `svc/mgr`; modo console multiplataforma.
- **Configuração:** TOML com arquivo externo; flags opcionais `--config`, `--console`, `--version`, `--help`.
- **Testes:** `go test ./...`; testes de domínio, HTTP e integração SQLite em diretórios temporários.

Organização sugerida (adaptável sem criar dependências circulares):

```text
cmd/financas/                  # entrypoint e flags
internal/app/                 # composição, ciclo de vida
internal/config/              # config e validação
internal/http/                # router, handlers, middleware
internal/accounts/            # contas e saldos
internal/transactions/        # lançamentos, categorias, transferências
internal/cards/               # cartões, compras, faturas, pagamentos
internal/dashboard/           # consultas de relatórios
internal/storage/sqlite/      # migrations, transações, repositórios
internal/windowsservice/      # Windows SCM; arquivos _windows.go
web/                          # interface e ativos incorporados
configs/config.example.toml
migrations/
.github/workflows/
```

**Direção das dependências:** HTTP chama casos de uso; casos de uso dependem de interfaces de persistência; SQLite implementa essas interfaces. Regras financeiras pertencem ao domínio/aplicação e não ao JavaScript ou exclusivamente ao banco.

## Semântica financeira obrigatória

1. **Dinheiro:** valores em `int64` representando centavos de BRL. Nunca `float` para cálculos financeiros. Na API, usar centavos inteiros (`amount_cents`) e converter apenas na UI.
2. **Data e fuso:** datas civis `YYYY-MM-DD` para competência, vencimento e pagamento; timestamps técnicos UTC. Períodos apresentados no fuso `America/Sao_Paulo`. Não usar timestamps para representar somente datas.
3. **Saldo de conta:** saldo inicial na data-base mais fluxo confirmado de receitas recebidas, despesas pagas, transferências e pagamentos de fatura; lançamentos pendentes não alteram saldo disponível. Saldo projetado é calculado separadamente.
4. **Lançamentos:** `kind=income|expense`, `status=pending|settled|cancelled`, `competence_date`, `due_date` e `settled_date` conforme necessário. Baixa de receita/despesa exige conta vinculada e data de liquidação. Cancelamento não equivale a exclusão silenciosa de operação liquidada.
5. **Transferência:** uma operação atômica debitando origem e creditando destino; não entra em receitas/despesas nem em gráficos de gastos; não aceita mesma conta nos dois lados.
6. **Cartões:** compra no cartão gera obrigação na fatura, não movimenta conta bancária no instante da compra. Compras parceladas geram parcelas vinculadas à compra original, com soma exata em centavos.
7. **Faturas:** ciclo determinado pela data de fechamento; compras no próprio dia de fechamento **pertencem ao ciclo que fecha nesse dia**; após o fechamento, entram no ciclo seguinte. Vencimento deve ser posterior ao fechamento do ciclo, permitindo atravessar o mês. Meses curtos utilizam último dia do mês para dias configurados maiores que o mês. Mudanças de fechamento/vencimento não devem reclassificar retroativamente compras já registradas.
8. **Pagamento de fatura:** reduz saldo da conta pagadora e o saldo devido da fatura, sem registrar uma segunda despesa. Permitir pagamentos parciais, proibir valor maior que saldo devido; pagamentos e estornos precisam de transação SQLite atômica.
9. **Relatórios:** distinguir `competência` (receitas/despesas do período, inclusive compras no cartão conforme parcelas) de `caixa` (fluxos efetivamente liquidados). Pagamento de fatura é saída de caixa, mas não nova despesa de competência. Transferências internas não integram receitas/despesas.
10. **Integridade:** FKs habilitadas, constraints, índices e migrations; toda operação que altera dois ou mais registros relacionados deve ser transacional; evitar exclusão destrutiva de contas/cartões com histórico.

## Contrato HTTP

- Servir UI em `/` e API JSON em `/api/v1/`.
- Padrão de resposta bem-sucedida: objeto JSON ou lista; padrão de erro: `{ "error": { "code": "...", "message": "..." } }`.
- Erros de validação: 400; inexistente: 404; conflito de estado: 409; falha inesperada: 500, sem vazamento de stack ou SQL.
- Endpoints de escrita devem validar entrada no servidor; usar IDs opacos e parâmetros SQL, nunca interpolar SQL com entrada de usuário.
- Em listas, usar paginação com `limit` e `offset`, ordenação estável e filtros explícitos.
- `GET /healthz` para liveness sem detalhes sensíveis.

## Configuração e operação

Exemplo em `configs/config.example.toml`:

```toml
[server]
host = "127.0.0.1"
port = 8787

[storage]
database_path = "C:\\ProgramData\\FinancasPessoais\\financas.db"

[app]
timezone = "America/Sao_Paulo"
log_level = "info"
```

- Por padrão, ouvir apenas no loopback: **não expor o sistema sem autenticação**.
- O serviço não deve depender de diretório de trabalho ou sessão interativa; caminhos relativos são resolvidos a partir da localização do arquivo de configuração e documentados.
- Persistência recomendada em `%ProgramData%\FinancasPessoais`; conceder permissões mínimas à identidade do serviço.
- Implementar encerramento gracioso do HTTP e fechamento do banco em parada de serviço.
- SQLite: `foreign_keys=ON`, `busy_timeout` e WAL quando apropriado; migrations na inicialização, evitando operações destrutivas automáticas.
- Criar scripts documentados de instalação, início, parada e desinstalação do serviço com privilégios administrativos. Não iniciar serviço automaticamente durante testes.
- Backups consistentes com API SQLite de backup ou `VACUUM INTO` em caminho seguro; nunca copiar apenas `.db` enquanto WAL estiver ativo. Restaurar somente com serviço parado e backup validado.

## Segurança

- MVP estritamente local. `127.0.0.1` é default obrigatório; mudança de `host` para endereço externo deve ser rejeitada enquanto não existir autenticação e proteção apropriada.
- Não armazenar senhas bancárias, tokens de instituições ou números completos de cartões; cartão tem nome/apelido, bandeira opcional e últimos quatro dígitos opcionais.
- Não expor dados financeiros em logs nem respostas de erro.
- Usar header de segurança apropriado, validação de dados e parametrização de queries.

## Qualidade / definição de pronto

Uma issue só está concluída quando:

1. Implementa apenas o escopo descrito, sem alterar regras anteriores silenciosamente.
2. Tem testes automatizados adequados e `go test ./...` passa.
3. Compila para Windows (`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/financas`) e roda em modo console no Linux durante desenvolvimento (para código Windows, usar build tags).
4. Inclui migrations quando houver mudança de esquema, compatíveis com bases existentes.
5. Atualiza README/documentação ou contrato da API se necessário.
6. Produz PR com resumo, decisões, evidências de testes, riscos e referência à issue.
7. Não marca a issue como concluída antes de validar critérios de aceite.

## Instruções ao agente

- Tratar este README como contrato funcional e arquitetural. Se uma issue o contradizer, **não decidir silenciosamente**: registrar a divergência no PR e optar por preservar dados e regras existentes.
- Trabalhar **uma issue por branch e PR**, preferindo a menor alteração necessária. Convenção: `issue-XX-descricao-curta`.
- Verificar `Depends on` antes de iniciar; não implementar recurso dependente de PR não integrado.
- Começar por testes de regras financeiras críticas. Evitar frameworks, filas, ORMs, containers de produção, autenticação e serviços externos sem necessidade expressa.
- Não simular conclusão de testes. Distinguir o que foi executado do que não foi possível verificar.
- Não alterar histórico financeiro por migração sem plano de compatibilidade e teste.

## Sequência planejada

As issues estão em `issues/`, numeradas para sugerir ordem de execução. A ordem é um guia; dependências declaradas em cada issue prevalecem.

## Execução (após implementação)

```bash
go run ./cmd/financas --console --config ./configs/config.example.toml
go test ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o financas.exe ./cmd/financas
```

A instalação do serviço no Windows será documentada e implementada na issue específica.

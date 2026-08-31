# GerminaStack

Comunidade de estudos para turmas: alunos publicam dúvidas, discutem conteúdos por matéria e recebem notificações. O projeto entrega uma aplicação web completa em Go, PostgreSQL e JavaScript puro.

## O que já existe

- Cadastro institucional, login em duas etapas por e-mail, sessão por cookie e logout.
- Catálogo de turmas e matérias, posts com imagem/legenda, comentários e respostas.
- Reações, notificações por menção, perfil público, foto de perfil e preferências de acessibilidade.
- `@frok`: assistente educacional baseado em Groq que responde em posts, comentários e respostas. Ele lê o contexto relacionado, incluindo a descrição alternativa da imagem, e pode ter memória persistente por usuário no MongoDB.
- Frontend estático servido pelo próprio backend; não precisa de Node.js em produção.

## Stack

- Go 1.25, Gin e `pgx`.
- PostgreSQL para dados acadêmicos, usuários, discussões e autenticação.
- Groq Responses API para o Frok; MongoDB é opcional e usado apenas na memória de longo prazo dele.
- SMTP ou Gmail API para os códigos de login.

## Rodar localmente

Pré-requisitos: Go 1.25+, PostgreSQL acessível e as variáveis abaixo. Copie `.env.example` para `.env`, preencha os valores e exporte-os no terminal antes de iniciar.

```bash
set -a; source .env; set +a
go run .
```

O servidor inicia em `http://127.0.0.1:8080` por padrão e aplica as migrações automaticamente. Para desenvolvimento HTTP local, use `COOKIE_SECURE=false`.

## Variáveis de ambiente

| Variável | Obrigatória | Uso |
| --- | --- | --- |
| `DATABASE_URL` | Sim | URL PostgreSQL. |
| `JWT_SECRET` | Sim | Segredo de sessão, mínimo de 32 bytes. |
| `TWO_FACTOR_SECRET` | Sim | Segredo dos códigos 2FA, diferente do JWT e com mínimo de 32 bytes. |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM_NAME` | Sim, quando Gmail não é usado | Configuração SMTP. |
| `SMTP_FROM_ADDRESS` | Sim | Remetente dos códigos SMTP ou Gmail. |
| `COOKIE_SECURE` | Não | `true` em HTTPS; `false` apenas em HTTP local. |
| `HTTP_ADDR` | Não | Endereço HTTP; padrão `:8080`. |
| `AUTH_OPERATION_TIMEOUT` | Não | Prazo de operações de autenticação; padrão `15s`. |
| `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REFRESH_TOKEN` | Não | Se as três estiverem preenchidas, o envio usa Gmail API em vez de SMTP. |
| `GROQ_API_KEY` | Não | Ativa o `@frok`. Nunca expor no frontend. |
| `FROK_MODEL`, `FROK_TIMEOUT` | Não | Modelo e prazo do Frok; padrões `openai/gpt-oss-20b` e `30s`. |
| `FROK_MONGODB_URI`, `FROK_MONGODB_DATABASE` | Não | Memória persistente do Frok; banco padrão `germinastack`. |

Configure os três campos `GOOGLE_*` juntos para usar Gmail API; nesse caso, os campos de conexão SMTP não são necessários. Não versione valores reais em `.env`.

## Rotas principais

| Método | Rota | Descrição |
| --- | --- | --- |
| `POST` | `/api/users` | Cadastro pelo e-mail institucional. |
| `POST` | `/api/login` | Valida credenciais e envia o código 2FA. |
| `POST` | `/api/login/2fa` | Confirma o código e cria a sessão. |
| `POST` | `/api/logout` | Encerra a sessão. |
| `GET` | `/api/years`, `/api/subjects` | Catálogo acadêmico. |
| `GET/POST/PATCH/DELETE` | `/api/posts` e `/api/posts/:id` | Publicações. |
| `GET/POST/PATCH/DELETE` | Comentários e respostas sob `/api/posts/:id/comments`, `/api/comments/:id/replies` e `/api/replies/:id`. |
| `PUT` | `/api/posts/:id/reaction`, `/api/comments/:id/reaction`, `/api/replies/:id/reaction` | Reage a post, comentário ou resposta. |
| `GET` | `/api/notifications` | Lista notificações; `PATCH /api/notifications/read-all` marca todas como lidas. |
| `GET/PATCH` | `/api/me` | Perfil próprio; `GET /api/users/:username` expõe perfil público. |
| `GET/PATCH` | `/api/me/preferences` | Preferências de contraste, fonte e espaçamento. |

Todas as rotas, exceto cadastro, login e listagem de turmas, exigem sessão válida.

## Frok

Mencione `@frok` em uma publicação, comentário ou resposta. Ele responde na mesma conversa, marca apenas quem o chamou e dispara a notificação normal da plataforma. A persona é direta e provocadora, mas não ameaça ou ataca usuários.

Com `FROK_MONGODB_URI`, o Frok consulta e registra chamadas anteriores do mesmo usuário. Sem essa variável, ele continua funcionando sem memória persistente. A chave do Groq e a URI do Mongo ficam exclusivamente no servidor.

## Deploy no Discloud

O arquivo `discloud.config` já define o app como `site`, com `MAIN=main.go` e 512 MB de RAM. Configure as mesmas variáveis de ambiente do quadro acima no painel do Discloud, faça o deploy do repositório e acompanhe os logs de inicialização.

## Testes

```bash
go test ./...
go vet ./...
node --check static/js/api.js
```

## Estrutura

```text
auth/       autenticação, 2FA e provedores de e-mail
config/     leitura e validação das variáveis de ambiente
database/   PostgreSQL, migrações e repositórios
frok/       cliente Groq e memória Mongo opcional
handlers/   regras HTTP
routes/     registro das páginas e rotas API
static/     JavaScript, estilos e imagens
web/        páginas HTML
```

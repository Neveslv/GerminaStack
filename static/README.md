# static

Todo o código de front-end que não é HTML de página: JavaScript, CSS próprio e a cópia
local da lib `germinastack-ui-components`. Ver [web/README.md](../web/README.md) para
como as páginas em `web/` consomem o que está aqui.

Nada nesta pasta usa bundler, TypeScript ou framework. JavaScript é servido como módulos
ES nativos do navegador (`<script type="module">`), CSS é servido como está escrito.

```
static/
├── css/                          → vazia — os dois extras que já moraram aqui
│                                    (tema e popover de notificações) migraram
│                                    para dentro da lib, ver seção abaixo
├── js/
│   ├── api.js                    → única porta de saída para o back-end
│   ├── config.js                 → ambiente, URL base, timeout e rota de login
│   ├── tema-inicial.js           → aplica o tema salvo antes da 1ª pintura
│   ├── componentes/              → peças de UI reaproveitadas entre páginas
│   ├── utils/                    → funções puras sem estado de página
│   └── pages/                    → um arquivo por página em web/, orquestra tudo
└── vendor/
    └── germinastack/              → cópia sincronizada do pacote npm (não editar)
```

## `vendor/germinastack/` — a lib de componentes

Vem do pacote npm [`germinastack-ui-components`](https://www.npmjs.com/package/germinastack-ui-components),
mantido pelo colega responsável pelo kit de UI. **Nada dentro desta pasta deve ser
editado à mão** — qualquer ajuste feito aqui é apagado no próximo `npm install`, porque
o conteúdo é gerado, não versionado manualmente.

### Como ele chega até aqui

Este repositório **não usa mais git submodule** para consumir a lib (usava, até a
migração feita nesta branch). O fluxo atual:

1. A raiz do repositório tem um `package.json` com a dependência:
   ```json
   { "dependencies": { "germinastack-ui-components": "1.0.4" } }
   ```
2. Rodar `npm install` baixa o pacote para `node_modules/` e dispara automaticamente o
   script `postinstall`, definido no mesmo `package.json`:
   ```json
   { "scripts": { "postinstall": "node ./node_modules/germinastack-ui-components/scripts/copy-to-static.mjs ./static/vendor/germinastack" } }
   ```
3. Esse script (`copy-to-static.mjs`, publicado pelo próprio pacote) copia a pasta
   `dist/` de dentro do pacote para `static/vendor/germinastack/`, resultando em:

   ```
   static/vendor/germinastack/
   ├── css/germinastack.css
   ├── js/germinastack.js     (build UMD/CJS — funciona como <script> clássico)
   ├── js/germinastack.mjs    (build ESM — não usado hoje, HTML aqui carrega o CJS)
   ├── fonts/OpenDyslexic-Regular.woff2
   ├── themes.css             (a partir da v1.0.3 — ver seção abaixo)
   └── notifications.css      (a partir da v1.0.4 — ver seção abaixo)
   ```

### Por que o resultado é comitado no git (e `node_modules/` não é)

`node_modules/` está no `.gitignore`. `static/vendor/germinastack/` **não está** — vai
para o repositório como arquivo comum. Isso é intencional: o back-end em Go não precisa
de Node instalado para rodar ou servir estas páginas; só quem **atualiza a versão da
lib** localmente precisa rodar `npm install` uma vez e comitar o resultado. Reproduz a
mesma garantia que o submodule antigo dava (`git clone` sozinho já deixa o site
funcional), só trocando `git submodule update` por `npm install`.

### Como atualizar a versão da lib

```bash
npm install germinastack-ui-components@<nova-versao>
```

Isso atualiza `package.json`, `package-lock.json` e reexecuta o `postinstall`,
sincronizando `static/vendor/germinastack/` com a nova versão. Confira o diff resultante
antes de comitar — principalmente se algum nome de classe ou atributo `data-gs-*` mudou,
o que quebraria o HTML de `web/` silenciosamente.

A versão é travada exata no `package.json` (`"1.0.4"`, não `"^1.0.4"`), de propósito:
uma atualização só acontece quando alguém decide e comita o bump, nunca silenciosamente
num `npm install` de rotina.

## `vendor/germinastack/themes.css` — temas por usuário

A lib resolve acessibilidade que depende do **sistema operacional**
(`prefers-contrast`, `forced-colors`, `prefers-reduced-motion`), mas por si só não tinha
um conceito de tema **escolhido e salvo pelo usuário**. A tabela `preferences` do banco
exige exatamente isso:

```sql
contrast_theme text default 'normal' check (
    contrast_theme in ('normal', 'dark', 'high_contrast', 'black_yellow', 'yellow_black')
),
font_family text default 'normal' check (
    font_family in ('normal', 'arial', 'verdana', 'lexend', 'atkinson_hyperlegible', 'open_dyslexic')
)
```

**Histórico**: essa camada nasceu neste projeto como `static/css/temas.css`, mantida por
fora do kit. Uma cópia foi passada para o mantenedor da lib com a ideia de o mecanismo
migrar para dentro do próprio pacote, para virar reutilizável por qualquer consumidor,
não só este projeto. A partir da versão **1.0.3**, o pacote passou a publicar
`dist/themes.css` (sincronizado aqui como `static/vendor/germinastack/themes.css` pelo
mesmo `copy-to-static.mjs` do resto da lib) — então o arquivo local foi removido, sem
duplicar o que a lib já entrega.

O tema ativo entra via atributo no `<html>`:

```html
<html data-tema="dark" data-fonte="lexend">
```

O tema `normal` não escreve o atributo (ausência = padrão do kit vale sozinho, sem
nenhuma regra por cima). Ver `static/js/utils/preferencias.js` para a lista de valores
válidos e `static/js/tema-inicial.js`/`componentes/cabecalho.js` para quem escreve esse
atributo.

### Por que não bastava sobrescrever só os tokens `--gs-*`

A lib fixa branco (`rgba(255,255,255,...)`, `#fff`) diretamente em cerca de 30
seletores (`.gs-card`, `.gs-modal`, `.gs-topbar`, `.gs-input`, etc.), sem passar por
variável. Só trocar `--gs-page`/`--gs-ink` deixaria o fundo escuro com os cards ainda
brancos por cima. `themes.css` reescreve essas ~30 superfícies para usar variáveis
próprias (prefixo `--gsk-`, de "GerminaStack Kit"), com valor padrão **idêntico** ao que
o kit já usava — o tema `normal` fica pixel a pixel igual ao kit puro. Cada tema
(`dark`, `high_contrast`, `black_yellow`, `yellow_black`) só troca essas variáveis.

## `vendor/germinastack/notifications.css` — popover de notificações

O tamanho e a rolagem próprios do painel de notificações (o `.gs-menu-panel` do kit
sozinho tem só 240px de largura e nenhum teto de altura, pensado para itens de uma
linha como "Copiar texto" — notificação é frase inteira) vêm deste arquivo, publicado
pela lib a partir da v1.0.4. Ver
[web/README.md § Popover de notificações](../web/README.md) para o HTML que ele
estiliza e as classes (`gs-notif-*`).

**Histórico**: assim como `themes.css`, essa camada nasceu neste projeto como
`static/css/notificacoes.css`. Foi enviada para o mantenedor da lib junto com a correção
do negrito da OpenDyslexic (abaixo) e incorporada na mesma versão — não recrie esses
arquivos localmente; o objetivo de tê-los enviado foi justamente parar de duplicar
manutenção entre este projeto e a lib.

### Negrito distorcido em OpenDyslexic (corrigido na lib, v1.0.4)

A lib só embute o peso Regular (400) dessa fonte. Qualquer `<strong>`/`<b>` renderizado
nela pedia o peso 700, que não existe — o navegador sintetizava um "falso negrito"
esticando o traçado, o que numa fonte já desenhada com traços grossos produzia letras
visivelmente espaçadas e distorcidas (era o bug visto no rótulo de prévia "OpenDyslexic"
na tela de Acessibilidade). A correção —
`[data-fonte="open_dyslexic"] strong, b { font-weight: 400; }` — agora mora dentro de
`themes.css`, publicada pela lib.

## `js/config.js`

`config.js` define `API_BASE_URL` (vazio — front e API são servidos pela mesma origem,
tanto no localhost quanto na Discloud), `TIMEOUT_MS` (tempo máximo de espera por
resposta) e `ROTA_LOGIN` (para onde `api.js` redireciona em um 401).

## `js/api.js` — a única porta de saída

Nenhum outro arquivo do projeto chama `fetch` diretamente. Toda função aqui monta a
chamada via `requisitar()` (que já cuida de timeout, JSON, cookie de sessão e 401).

| Função | Rota | Tabela/função do banco |
|---|---|---|
| `listarPosts({ idSubject })` | `GET /api/posts` | `posts` (+ `JOIN` users, subjects) |
| `buscarPost(id)` | `GET /api/posts/:id` | `posts` |
| `criarPost(dados)` | `POST /api/posts` | `posts` (via `create_message('post', ...)`) |
| `listarMaterias()` | `GET /api/subjects` | `subjects` |
| `listarAnos()` | `GET /api/years` | `years` |
| `listarComentarios(idPost)` | `GET /api/posts/:id/comments` (+ `/api/comments/:id/replies`) | `comments` + `comments_on_comments` |
| `criarComentario({ tipo, idPai, content })` | `POST /api/posts/:id/comments` ou `POST /api/comments/:id/replies` | `comments` / `comments_on_comments` (via `create_message`) |
| `entrar(credenciais)` | `POST /api/login` | autenticação (JWT) |
| `completarLogin(dados)` | `POST /api/login/2fa` | autenticação (2º fator) |
| `sair()` | `POST /api/logout` | autenticação |
| `cadastrar(dados)` | `POST /api/users` | `users` |
| `buscarUsuario(username)` | `GET /api/users/:username` (ou `/api/me`) | `users` |
| `listarPostsDoAutor(idAutor)` | `GET /api/posts?author_id=` | `posts` filtrado por `id_user` |
| `listarNotificacoes()` | `GET /api/notifications` | `notifications` |
| `marcarNotificacoesComoLidas()` | `PATCH /api/notifications/read-all` | `mark_notifications_as_read(id_user)` |
| `reagir({ tipo, id, reacao })` | `PUT /api/{posts\|comments\|replies}/:id/reaction` | `reaction(id_user, id_message, message_type, reaction_type)` |
| `buscarMinhaReacao(tipo, id)` | — (cache em memória da sessão) | `reactions` (leitura) |
| `buscarPreferencias()` / `salvarPreferencias(prefs)` | `GET` / `PATCH /api/me/preferences` | `preferences` |

`ErroDeApi` é a única classe de erro que sobe até as páginas — sempre tem `.message`
(texto para mostrar) e `.status` (código HTTP, quando existir). Um 401 aciona
`encerrarSessao()` e redireciona para `ROTA_LOGIN` antes mesmo de a página tratar o erro.

### Convenção ao adicionar um endpoint novo

1. Adicione a função em `api.js`, montando a chamada com `requisitar(...)` e
   normalizando o retorno no formato que as páginas esperam.
2. Chame essa função só a partir de `js/pages/` ou `js/componentes/` — nunca duplique a
   URL/rota em outro lugar.

## `js/utils/` — funções sem estado de página

- **`dom.js`** — `criarElemento()` (cria elemento com `textContent`, nunca `innerHTML`,
  para dado externo nunca virar HTML por acidente), `comAtraso()` (debounce),
  `criarPainelDeEstado()` (controla um bloco `aria-live` de carregando/vazio/erro) e
  `inicializarKit()` (chama `GerminaStackUI.init()` num escopo já marcado como vinculado
  — ver o comentário no arquivo para o motivo exato de precisar disso).
- **`data.js`** — formatação de data (`formatarDataRelativa`, `formatarDataCompleta`),
  fixado no fuso `America/Sao_Paulo`.
- **`identidade.js`** — cor de avatar e tom de chip de matéria, derivados do `id` (não
  sorteados), para a mesma pessoa/matéria ter sempre a mesma cor em qualquer tela. A cor
  nunca é o único jeito de identificar algo — o nome/texto está sempre ao lado.
- **`preferencias.js`** — a lista `TEMAS`/`FONTES` (rótulo + descrição para a tela de
  preferências) e as funções `aplicarPreferencias()`/`guardarPreferenciasLocais()`/
  `lerPreferenciasLocais()`, usadas tanto por `tema-inicial.js` quanto pela página
  `preferencias.js`.

## `js/componentes/` — peças reaproveitadas

- **`cartao-post.js`** — monta o `<article class="gs-post">` completo (cabeçalho, chip
  de matéria, corpo, reações, rodapé). Usado no feed, na página de matéria, no perfil e
  na página de publicação (com `{ destaque: true }`, que troca o título de `<a>` para
  `<h1>`).
- **`reacoes.js`** — par de botões curtir/descurtir. Não usa o `data-gs-like` da lib de
  propósito: aquele atributo só soma 1 na tela para demonstração; aqui o número precisa
  vir do banco, então a chamada de API mora fora do kit, como o próprio kit recomenda.
  Implementa a mesma regra de alternância da função `reaction()` do banco (sem reação →
  insere; mesma reação → remove; reação diferente → troca).
- **`comentario.js`** — comentários e respostas de dois níveis. Uma resposta nunca ganha
  botão de responder — o banco não tem terceiro nível, e o front não oferece o que o
  banco não guarda.
- **`cabecalho.js`** — roda em toda página exceto login/cadastro. Sincroniza o tema
  salvo no servidor com o espelho local (`localStorage`) e busca as notificações uma
  única vez, usando o mesmo resultado para o badge (`[data-nao-lidas]`) e para o
  conteúdo do popover (`[data-notif-lista]`, `[data-notif-marcar-lidas]`) — ver
  [web/README.md § Popover de notificações](../web/README.md). O popover em si é o
  mesmo mecanismo `data-gs-menu` do botão "⋮" do post, já vindo pronto no HTML de cada
  página; como esse markup existe desde o carregamento inicial, `GerminaStackUI.init()`
  já rodou sobre ele antes de qualquer módulo executar — por isso a função só popula
  conteúdo e nunca chama `inicializarKit()` de novo (evita o registro duplicado de
  clique descrito no comentário de `inicializarKit()` em `utils/dom.js`).

## `js/pages/` — orquestração por página

Um arquivo por página em `web/`, mesmo nome. Cada um só faz três coisas: busca dados via
`api.js`, monta a tela com `componentes/` e `utils/dom.js`, e liga os eventos da própria
página. Nenhuma lógica de negócio mora aqui que deveria estar em `api.js` ou em um
componente — se dois `pages/*.js` precisam do mesmo pedaço de código, ele deveria ter
virado componente ou util antes de ser copiado.

## `js/tema-inicial.js` — por que não é um módulo

É o único script carregado como clássico e **sem `defer`**, direto no `<head>` de cada
página, antes de tudo (ver [web/README.md](../web/README.md)). Ele aplica o tema salvo
no `<html>` antes da primeira pintura da tela. Se essa aplicação fosse feita por um
módulo (sempre adiado até depois do parse do HTML), a página apareceria clara por um
instante antes de escurecer — e piscar a tela na cara de quem escolheu alto contraste
justamente por sensibilidade visual é o pior resultado possível. O custo é da ordem de
milissegundos de bloqueio, e o script não toca em nada além de dois atributos do
`<html>`.

A fonte da verdade continua sendo a tabela `preferences`, lida pela API; o
`localStorage` é só um espelho para esse carregamento inicial não depender de uma ida ao
servidor antes de pintar a tela.

## Convenções gerais

- **Idioma**: nomes de função, variável e comentário em português; nomes de classes CSS
  e atributos `data-gs-*` seguem o inglês da lib, sem tradução.
- **Módulos ES em toda parte, exceto `tema-inicial.js`** e o script da lib (que é
  carregado como clássico de propósito, ver [web/README.md](../web/README.md)).
- **`textContent`, nunca `innerHTML`, para dado vindo da API.** Só o próprio kit
  usa `innerHTML` internamente, e apenas para HTML que ele mesmo gera (toasts, popovers).
- **Cor nunca é o único sinal.** Todo indicador por cor (chip de matéria, badge de não
  lida, estado de curtida) tem também texto ou ícone com `aria-label`.
- **Um painel de estado (`criarPainelDeEstado`) por lista carregada assincronamente.**
  Nenhuma página escreve "Carregando..." fixo no HTML — o estado muda com a requisição.

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
│   ├── config.js                 → ambiente, URL base, flag de dados locais
│   ├── tema-inicial.js           → aplica o tema salvo antes da 1ª pintura
│   ├── componentes/              → peças de UI reaproveitadas entre páginas
│   ├── mock/                     → dados de exemplo, um arquivo por tabela do banco
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

## `js/config.js` e o modo de dados locais

```js
export const USAR_DADOS_LOCAIS = true;
```

Com essa flag ligada, toda função em `api.js` responde com dados de `js/mock/` em vez de
chamar o back-end. É a única chave que precisa mudar quando os endpoints do Gin
estiverem prontos — nenhuma página ou componente sabe que os dados são falsos.

`config.js` também resolve `API_BASE_URL` (local vs. produção) e `ROTA_LOGIN` (para onde
`api.js` redireciona em um 401).

## `js/api.js` — a única porta de saída

Nenhum outro arquivo do projeto chama `fetch` diretamente. Toda função aqui segue o
mesmo padrão: se `USAR_DADOS_LOCAIS`, resolve contra o mock correspondente; senão, monta
a chamada real via `requisitar()` (que já cuida de timeout, JSON, cookie de sessão e
401).

| Função | Tabela/função do banco | Mock usado |
|---|---|---|
| `listarPosts({ idSubject })` | `posts` (+ `JOIN` users, subjects) | `mock/posts.js` |
| `buscarPost(id)` | `posts` | `mock/posts.js` |
| `criarPost(dados)` | `posts` (via `create_message('post', ...)`) | `mock/posts.js` |
| `listarMaterias()` | `subjects` | `mock/subjects.js` |
| `listarAnos()` | `years` | `mock/years.js` |
| `listarComentarios(idPost)` | `comments` + `comments_on_comments` | `mock/comments.js` |
| `criarComentario({ tipo, idPai, content })` | `comments` / `comments_on_comments` (via `create_message`) | `mock/comments.js` |
| `entrar(credenciais)` | autenticação (JWT, futuro) | `mock/sessao.js` |
| `cadastrar(dados)` | `users` | `mock/users.js` |
| `buscarUsuario(username)` | `users` | `mock/users.js` |
| `listarPostsDoAutor(idAutor)` | `posts` filtrado por `id_user` | `mock/posts.js` |
| `listarNotificacoes()` | `notifications` | `mock/notifications.js` |
| `marcarNotificacoesComoLidas()` | `mark_notifications_as_read(id_user)` | `mock/notifications.js` |
| `reagir({ tipo, id, reacao })` | `reaction(id_user, id_message, message_type, reaction_type)` | `mock/reactions.js` |
| `buscarMinhaReacao(tipo, id)` | `reactions` (leitura) | `mock/reactions.js` |
| `buscarPreferencias()` / `salvarPreferencias(prefs)` | `preferences` | `mock/preferences.js` |

`ErroDeApi` é a única classe de erro que sobe até as páginas — sempre tem `.message`
(texto para mostrar) e `.status` (código HTTP, quando existir). Um 401 aciona
`encerrarSessao()` e redireciona para `ROTA_LOGIN` antes mesmo de a página tratar o erro.

### Convenção ao adicionar um endpoint novo

1. Crie (ou reaproveite) o mock em `js/mock/`, no formato exato que o endpoint real vai
   devolver — comentário no topo do mock explicando de qual tabela ele vem.
2. Adicione a função em `api.js`, seguindo o padrão `if (USAR_DADOS_LOCAIS) { ... } return requisitar(...)`.
3. Chame essa função só a partir de `js/pages/` ou `js/componentes/` — nunca duplique a
   URL/rota em outro lugar.

## `js/mock/` — um arquivo por tabela

Cada arquivo espelha o formato de uma tabela do banco (`GerminaStack.sql`), incluindo os
relacionamentos já resolvidos como o back-end deve devolver (autor e matéria aninhados
via `JOIN`, contagem de comentários já calculada). Os ids de autor são consistentes
entre arquivos — o mesmo `id: 4` é sempre "Ana Ribeiro" em `posts.js`, `comments.js`,
`users.js` e `sessao.js`, para o app se comportar como se fosse um `JOIN` de verdade.

| Arquivo | Tabela | Observação |
|---|---|---|
| `posts.js` | `posts` | Autor e matéria já aninhados; inclui `comments_count` |
| `comments.js` | `comments` + `comments_on_comments` | Estrutura de dois níveis (`replies`) |
| `subjects.js` | `subjects` | |
| `years.js` | `years` | `year` é `text`, não número (ex.: `"2º Tec"`) |
| `users.js` | `users` | Nunca inclui a coluna `password` |
| `sessao.js` | — | Usuário "logado" fixo, usado como autor de posts locais |
| `notifications.js` | `notifications` | `text_show` já vem pronto, como o trigger `notify_mentions` geraria |
| `preferences.js` | `preferences` | É `let`, não `const` — a API local escreve aqui ao salvar |
| `reactions.js` | `reactions` | Alternância `like`/`dislike`/remover espelha a função `reaction()` do banco |

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
- **`textContent`, nunca `innerHTML`, para dado vindo de API ou mock.** Só o próprio kit
  usa `innerHTML` internamente, e apenas para HTML que ele mesmo gera (toasts, popovers).
- **Cor nunca é o único sinal.** Todo indicador por cor (chip de matéria, badge de não
  lida, estado de curtida) tem também texto ou ícone com `aria-label`.
- **Um painel de estado (`criarPainelDeEstado`) por lista carregada assincronamente.**
  Nenhuma página escreve "Carregando..." fixo no HTML — o estado muda com a requisição.

# web

Páginas HTML do GerminaStack. Cada arquivo aqui é uma tela completa e independente —
não existe um roteador de front-end nem um layout compartilhado renderizado em runtime.
A "página mestra" é o próprio HTML, repetido em cada arquivo com a mesma estrutura de
cabeçalho e a mesma ordem de carregamento de scripts.

Este projeto não usa framework (React, Vue, etc.) nem bundler. Cada página carrega a
lib `germinastack-ui-components` e os módulos ES próprios diretamente via `<script>` e
`<link>`, com caminhos absolutos (`/static/...`). Veja [static/README.md](../static/README.md)
para a estrutura de `static/`.

## Inventário de páginas

| Arquivo | Rota prevista | Título | Script de página | Requer login\* |
|---|---|---|---|---|
| `index.html` | `/` | Feed | `static/js/pages/feed.js` | Não |
| `post.html` | `/post?id=<id>` | Publicação + respostas | `static/js/pages/post.js` | Não (responder, sim) |
| `publicar.html` | `/publicar` | Nova publicação | `static/js/pages/publicar.js` | Sim |
| `materias.html` | `/materias` ou `/materias?id=<id>` | Canais por matéria | `static/js/pages/materias.js` | Não |
| `perfil.html` | `/perfil` ou `/perfil?usuario=<username>` | Perfil próprio ou público | `static/js/pages/perfil.js` | Não (ver perfil de outros) |
| `preferencias.html` | `/preferencias` | Acessibilidade (contraste + fonte) | `static/js/pages/preferencias.js` | Sim |
| `login.html` | `/login` | Entrar | `static/js/pages/login.js` | — |
| `cadastro.html` | `/cadastro` | Criar conta | `static/js/pages/cadastro.js` | — |

\* "Requer login" depende da sessão JWT emitida pelo back-end: sem cookie válido, a API
devolve 401 e `static/js/api.js` redireciona para `ROTA_LOGIN`.

### O que cada página faz, em detalhe

**`index.html` — Feed**
Lista publicações (`GET /api/posts`), com busca por texto (client-side, com debounce)
e filtro por matéria na barra lateral. Mostra estatísticas agregadas no hero
(total de posts, matérias ativas, respostas). Os cartões de post vêm de
`static/js/componentes/cartao-post.js`.

**`post.html` — Publicação**
Mostra uma publicação em destaque (`<h1>`, não link) e a thread de comentários/respostas
em dois níveis, espelhando as tabelas `comments` e `comments_on_comments`. Tem formulário
de resposta embutido que chama `POST` via `criarComentario()`.

**`publicar.html` — Nova publicação**
Formulário com matéria (`<select>` populado por `GET /api/subjects`), título, conteúdo e
um campo opcional de link de imagem (valida que começa com `http(s)://`, já que a coluna
`image_url` aceita nulo).

**`materias.html` — Canais por matéria**
Grade de cartões, um por matéria, com contagem de publicações. Clicar em um cartão filtra
a lista abaixo e atualiza a URL (`?id=<id>`) via `history.pushState`, então o botão
"voltar" do navegador funciona e o link é compartilhável.

**Notificações** não é mais uma página — é o popover que abre a partir do item
"Notificações" da navegação, presente em todas as páginas listadas acima (exceto
login/cadastro). Ver seção [Popover de notificações](#popover-de-notificações) abaixo.

**`perfil.html` — Perfil**
Sem `?usuario=` na URL, mostra o próprio perfil (e-mail visível, atalhos para
"Ajustar acessibilidade" e "Nova publicação"). Com `?usuario=<username>`, mostra o
perfil público de outra pessoa (sem e-mail). Lista as publicações da pessoa e alguns
números agregados (curtidas recebidas, respostas recebidas, matérias em que publicou).

**`preferencias.html` — Acessibilidade**
Formulário de rádios (não botões com `aria-*`) para escolher `contrast_theme` e
`font_family` — os dois campos da tabela `preferences`, com os mesmos valores aceitos
pelos `CHECK` do schema. A pré-visualização é aplicada **antes** de salvar, para a pessoa
comparar visualmente. Ver [static/README.md § Acessibilidade](../static/README.md) para
como o tema é de fato aplicado.

**`login.html`** / **`cadastro.html`**
Formulários de autenticação. `entrar()` e `cadastrar()` em `static/js/api.js` chamam
`POST /api/login` e `POST /api/users`; o token de sessão volta em cookie definido pelo
servidor.

## Estrutura comum a toda página

Todo arquivo aqui segue o mesmo esqueleto. Ao criar uma página nova, comece copiando um
arquivo existente próximo do que você precisa (ex.: `materias.html` para uma listagem,
`perfil.html` para um formulário com estado carregado da API) em vez de escrever do zero.

```html
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Nome da página | GerminaStack</title>
    <meta name="description" content="..." />

    <link rel="stylesheet" href="/static/vendor/germinastack/css/germinastack.css" />
    <link rel="stylesheet" href="/static/vendor/germinastack/themes.css" />

    <!-- Sem defer de proposito: aplica o tema salvo antes da primeira pintura. -->
    <script src="/static/js/tema-inicial.js"></script>

    <script src="/static/vendor/germinastack/js/germinastack.js" defer></script>
    <script type="module" src="/static/js/componentes/cabecalho.js"></script>
    <script type="module" src="/static/js/pages/nome-da-pagina.js"></script>
</head>
<body>
    <a class="gs-skip-link" href="#conteudo-principal">Pular para o conteúdo principal</a>

    <header class="gs-topbar"> ... </header>

    <main class="gs-page" id="conteudo-principal">
        ...
    </main>
</body>
</html>
```

A ordem dos `<script>` no `<head>` importa e não é arbitrária:

1. **`tema-inicial.js` sem `defer`, antes de tudo.** Aplica o tema salvo no
   `<html>` antes da primeira pintura da página, para quem escolheu alto contraste ou
   tema escuro não ver um flash de tela clara. É a única exceção à regra de módulos —
   ver [static/README.md](../static/README.md) para o porquê.
2. **`germinastack.js` do kit, com `defer`.** Precisa rodar antes dos módulos da
   aplicação, porque eles chamam `window.GerminaStackUI.init(...)` em conteúdo que
   acabaram de montar.
3. **`cabecalho.js`** (exceto em `login.html`/`cadastro.html`). Sincroniza preferências
   com o servidor e preenche o popover de notificações (badge + lista + botão de marcar
   como lida).
4. **O script da própria página**, por último.

`login.html` e `cadastro.html` não incluem `cabecalho.js` porque a navegação deles não
tem o popover de notificações — não faz sentido carregar o script de um componente que
não existe na tela.

### Navegação

O bloco de navegação é repetido (não é um "include" — HTML puro não tem isso) em todas
as páginas exceto login/cadastro, que têm uma nav reduzida. O link da página atual leva
`aria-current="page"` e a classe `is-active` escritos direto no HTML, não calculados por
JavaScript — de propósito: quem usa leitor de tela ou tem JavaScript bloqueado ainda sabe
onde está.

```html
<nav class="gs-nav" aria-label="Navegação principal">
    <a class="gs-nav-link" href="/">Feed</a>
    <a class="gs-nav-link" href="/materias">Matérias</a>
    <a class="gs-nav-link" href="/publicar">Publicar</a>

    <div class="gs-menu" data-gs-menu>
        <button class="gs-nav-link" type="button" data-gs-menu-trigger aria-haspopup="menu"
            aria-expanded="false">
            Notificações
            <span class="gs-badge" data-nao-lidas hidden></span>
        </button>
        <div class="gs-menu-panel gs-notif-panel" data-gs-menu-panel hidden>
            <div class="gs-notif-head">
                <strong>Notificações</strong>
                <button class="gs-btn gs-btn-ghost" type="button" data-notif-marcar-lidas disabled>
                    Marcar todas como lidas
                </button>
            </div>
            <div class="gs-notif-lista" data-notif-lista></div>
        </div>
    </div>

    <a class="gs-nav-link" href="/preferencias">Acessibilidade</a>
    <a class="gs-nav-link" href="/perfil">Perfil</a>
</nav>
```

Se uma página nova incluir esse bloco tal como está, o popover já funciona sozinho —
nenhum código extra é necessário na página em si. Ver a seção seguinte para o contrato
completo.

### Popover de notificações

Não existe mais uma página dedicada a notificações — o bloco acima **é** a tela de
notificações, embutido na navegação de toda página que o inclui. A ideia é simples: o
conteúdo é curto (algumas frases), então uma navegação inteira para vê-lo era
desnecessária.

O popover reaproveita o mesmo mecanismo de menu contextual (`data-gs-menu` /
`data-gs-menu-trigger` / `data-gs-menu-panel`) que já existe no botão "⋮" do cartão de
post — ver [static/README.md § componentes/](../static/README.md). O HTML acima só
declara a estrutura; quem preenche o conteúdo é `static/js/componentes/cabecalho.js`:

| Atributo | Papel |
|---|---|
| `[data-nao-lidas]` | badge com a contagem, escondido quando não há pendência |
| `[data-notif-lista]` | recebe os itens de notificação renderizados |
| `[data-notif-marcar-lidas]` | botão que chama `marcarNotificacoesComoLidas()` |

O CSS específico do painel (largura maior que o menu padrão do kit, teto de altura com
rolagem) vem de `static/vendor/germinastack/notifications.css`, publicado pela lib a
partir da v1.0.4 — mesmo caminho que `themes.css` percorreu (ver
[static/README.md](../static/README.md)). O `<link>` fica logo depois de `themes.css`
em toda página que tem este bloco na nav.

### Convenções de acessibilidade obrigatórias

- **Link de pular para o conteúdo** (`gs-skip-link`) como primeiro elemento do `<body>`,
  apontando para `#conteudo-principal`.
- **Um `<main id="conteudo-principal">` só, com classe `gs-page`.**
- **Áreas de estado dinâmico** (carregando/vazio/erro) são um `<div role="status"
  aria-live="polite" hidden>` que o helper `criarPainelDeEstado` (em
  `static/js/utils/dom.js`) preenche e revela. Nunca escreva "Carregando..." direto no
  HTML estático — o texto muda conforme o estado real da requisição.
- **Todo `<input>`/`<select>`/`<textarea>` tem `<label>` associada** por `for`/`id`, e
  usa `aria-describedby` para ligar dica (`gs-form-hint`) e erro (`gs-form-error`,
  `role="alert"`) — não um `placeholder` como única explicação do campo.
- **Ícones puramente decorativos levam `aria-hidden="true"`.** Se um ícone é a única
  pista de uma ação (ex.: botão só com "⋮"), o botão tem `aria-label` explicando a ação.

## Como adicionar uma página nova

1. Copie o esqueleto de uma página parecida (seção acima).
2. Escreva o `<title>` e a `<meta name="description">` — são usados por leitores de
   tela e buscadores, não é decoração.
3. Adicione o link dela na `<nav>` de **todas** as páginas que devem exibi-la, com
   `aria-current="page"` só no arquivo que representa a própria página.
4. Crie o script de página correspondente em `static/js/pages/nome-da-pagina.js` (ver
   [static/README.md § pages/](../static/README.md)).
5. Se a página expõe uma tabela do banco que ainda não tem função em
   `static/js/api.js`, crie a função lá antes de montar a tela — assim a página já nasce
   puxando dados de um único lugar.

## Sobre rotas sem extensão

Os links usam caminhos limpos (`/materias`, `/post?id=1`), não `/materias.html`. Hoje,
sem um servidor rodando, **abrir estes arquivos direto (duplo clique) não funciona** —
os caminhos absolutos (`/static/...`) e as rotas limpas dependem de um servidor HTTP
servindo a partir da raiz do repositório. Quando o back-end (Gin) expuser essas rotas,
cada uma deve responder com o arquivo `.html` correspondente desta pasta, ver tabela no
topo deste documento.

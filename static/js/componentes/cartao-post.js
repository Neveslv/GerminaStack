import { criarElemento } from '../utils/dom.js';
import { aplicarImagemNoAvatar, corDoAutor, tomDaMateria, inicialDoNome } from '../utils/identidade.js';
import { formatarDataRelativa, formatarDataCompleta } from '../utils/data.js';
import { criarReacoes } from './reacoes.js';
import { montarTextoComMencoes } from './mencoes.js';

function montarChipDeMateria(subject) {
    const chip = criarElemento('a', {
        classe: `gs-chip ${tomDaMateria(subject.id)}`,
        atributos: {
            href: `/materias?id=${subject.id}`,
            'aria-label': `Ver publicações de ${subject.subject}`
        }
    });
    chip.append(
        criarElemento('span', { classe: 'gs-chip-dot', atributos: { 'aria-hidden': 'true' } }),
        criarElemento('span', { texto: subject.subject })
    );
    return chip;
}

function montarMenu(post) {
    const menu = criarElemento('div', { classe: 'gs-menu', atributos: { 'data-gs-menu': '' } });

    const gatilho = criarElemento('button', {
        classe: 'gs-icon-button',
        atributos: {
            type: 'button',
            'data-gs-menu-trigger': '',
            'aria-expanded': 'false',
            'aria-label': `Mais ações para a publicação de ${post.author.name}`
        }
    });
    gatilho.append(criarElemento('span', { texto: '⋮', atributos: { 'aria-hidden': 'true' } }));

    const painel = criarElemento('div', {
        classe: 'gs-menu-panel',
        atributos: { 'data-gs-menu-panel': '', hidden: '' }
    });
    painel.append(
        criarElemento('button', {
            classe: 'gs-menu-item',
            texto: 'Copiar texto',
            atributos: { type: 'button', 'data-action': 'copiar-post' }
        }),
        criarElemento('button', {
            classe: 'gs-menu-item is-danger',
            texto: 'Ocultar do feed',
            atributos: { type: 'button', 'data-action': 'ocultar-post' }
        })
    );

    menu.append(gatilho, painel);
    return menu;
}

function montarTitulo(post, destaque) {
    if (destaque) {
        return criarElemento('h1', { classe: 'gs-post-title', texto: post.title });
    }

    const titulo = criarElemento('h3', { classe: 'gs-post-title' });
    titulo.append(criarElemento('a', { texto: post.title, atributos: { href: `/post?id=${post.id}` } }));
    return titulo;
}

function montarCabecalho(post, destaque) {
    const meta = criarElemento('div', { classe: 'gs-meta' });
    meta.append(
        criarElemento(post.author.username ? 'a' : 'span', {
            texto: post.author.name,
            atributos: post.author.username ? { href: `/perfil?usuario=${post.author.username}` } : {}
        }),
        criarElemento('time', {
            texto: formatarDataRelativa(post.created_at),
            atributos: {
                datetime: post.created_at,
                title: formatarDataCompleta(post.created_at)
            }
        })
    );

    const chips = criarElemento('div', { classe: 'gs-cluster' });
    chips.append(montarChipDeMateria(post.subject));

    const usuario = criarElemento('div', { classe: 'gs-post-user' });
    usuario.append(
        meta,
        chips,
        montarTitulo(post, destaque),
        (() => {
            const texto = criarElemento('p', { classe: 'gs-post-copy' });
            texto.append(montarTextoComMencoes(post.content));
            return texto;
        })()
    );

    const avatar = criarElemento('span', {
        classe: 'gs-avatar',
        texto: inicialDoNome(post.author.name),
        atributos: { 'aria-hidden': 'true' }
    });
    if (!aplicarImagemNoAvatar(avatar, post.author)) avatar.style.background = corDoAutor(post.author.id);

    const cabecalho = criarElemento('div', { classe: 'gs-post-head' });
    cabecalho.append(avatar, usuario, montarMenu(post));
    return cabecalho;
}

function montarRodape(post, minhaReacao) {
    const rodape = criarElemento('div', { classe: 'gs-post-foot' });

    rodape.append(
        criarReacoes({
            tipo: 'post',
            id: post.id,
            likes: post.likes,
            dislikes: post.dislikes,
            minhaReacao
        })
    );

    const total = post.comments_count ?? 0;
    rodape.append(
        criarElemento('a', {
            classe: 'gs-action',
            texto: `${total} ${total === 1 ? 'resposta' : 'respostas'}`,
            atributos: {
                href: `/post?id=${post.id}`,
                'aria-label': `Ver as ${total} respostas de "${post.title}"`
            }
        })
    );

    return rodape;
}

export function criarCartaoDePost(post, { minhaReacao = null, destaque = false } = {}) {
    const artigo = criarElemento('article', {
        classe: 'gs-post',
        atributos: { 'data-post-id': String(post.id) }
    });

    artigo.style.overflow = 'visible';

    const corpo = criarElemento('div', { classe: 'gs-post-body' });
    corpo.append(montarCabecalho(post, destaque));

    if (post.image_url) {
        const descricao = post.image_description?.trim();
        const midia = criarElemento('figure', { classe: 'gs-post-media' });
        const imagem = criarElemento('img', {
            atributos: {
                src: post.image_url,
                alt: descricao ?? '',
                loading: 'lazy'
            }
        });
        const ampliar = criarElemento('a', {
            atributos: {
                href: post.image_url,
                target: '_blank',
                rel: 'noopener',
                'aria-label': 'Abrir imagem ou GIF em tamanho maior'
            }
        });
        ampliar.append(imagem);
        midia.append(ampliar);
        if (descricao) midia.append(criarElemento('figcaption', { texto: descricao }));
        corpo.append(midia);
    }

    artigo.append(corpo, montarRodape(post, minhaReacao));
    return artigo;
}

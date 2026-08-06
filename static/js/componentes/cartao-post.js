/** Monta o cartao de uma publicacao usando os blocos gs-post do kit. */

import { criarElemento } from '../utils/dom.js';
import { corDoAutor, tomDaMateria, inicialDoNome } from '../utils/identidade.js';
import { formatarDataRelativa, formatarDataCompleta } from '../utils/data.js';
import { criarReacoes } from './reacoes.js';

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
    // Na página da publicação o título é o h1 da tela e não leva a lugar nenhum:
    // um link apontando para a própria página só atrapalha quem navega por
    // lista de links no leitor de tela.
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
        criarElemento('p', { classe: 'gs-post-copy', texto: post.content })
    );

    const avatar = criarElemento('span', {
        classe: 'gs-avatar',
        texto: inicialDoNome(post.author.name),
        atributos: { 'aria-hidden': 'true' }
    });
    avatar.style.background = corDoAutor(post.author.id);

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

/**
 * Devolve o <article class="gs-post"> completo de uma publicacao.
 *
 * @param {object} post publicação no formato de GET /api/posts
 * @param {{ minhaReacao?: string|null, destaque?: boolean }} opcoes
 *        destaque=true monta a versão da página de detalhe, com <h1>.
 */
export function criarCartaoDePost(post, { minhaReacao = null, destaque = false } = {}) {
    const artigo = criarElemento('article', {
        classe: 'gs-post',
        atributos: { 'data-post-id': String(post.id) }
    });

    // O kit define .gs-post { overflow: hidden }, o que recorta o painel do menu
    // de acoes, que e position:absolute e abre para baixo do gatilho. Liberamos o
    // overflow so neste card para o menu ficar visivel.
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
        midia.append(imagem);
        if (descricao) midia.append(criarElemento('figcaption', { texto: descricao }));
        corpo.append(midia);
    }

    artigo.append(corpo, montarRodape(post, minhaReacao));
    return artigo;
}

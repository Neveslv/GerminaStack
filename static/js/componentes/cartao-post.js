/** Monta o cartao de uma publicacao usando os blocos gs-post do kit. */

import { criarElemento } from '../utils/dom.js';
import { formatarDataRelativa, formatarDataCompleta } from '../utils/data.js';

const TONS_DE_CHIP = ['is-blue', 'is-amber', 'is-mint', 'is-rose'];
const CORES_DE_AVATAR = ['#ffb347', '#4db8ff', '#5fd6a4', '#c9a0ff'];

/** Escolhe um tom estavel a partir do id, para a mesma materia ter sempre a mesma cor. */
function tomDaMateria(idMateria) {
    return TONS_DE_CHIP[idMateria % TONS_DE_CHIP.length];
}

function corDoAutor(idAutor) {
    return CORES_DE_AVATAR[idAutor % CORES_DE_AVATAR.length];
}

function montarChipDeMateria(subject) {
    const chip = criarElemento('span', { classe: `gs-chip ${tomDaMateria(subject.id)}` });
    chip.append(
        criarElemento('span', { classe: 'gs-chip-dot' }),
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

function montarCabecalho(post) {
    const meta = criarElemento('div', { classe: 'gs-meta' });
    meta.append(
        criarElemento('strong', { texto: post.author.name }),
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

    const titulo = criarElemento('h3', { classe: 'gs-post-title' });
    titulo.append(criarElemento('a', { texto: post.title, atributos: { href: `/post?id=${post.id}` } }));

    const usuario = criarElemento('div', { classe: 'gs-post-user' });
    usuario.append(meta, chips, titulo, criarElemento('p', { classe: 'gs-post-copy', texto: post.content }));

    const avatar = criarElemento('span', {
        classe: 'gs-avatar',
        texto: post.author.name.charAt(0).toUpperCase(),
        atributos: { 'aria-hidden': 'true' }
    });
    avatar.style.background = corDoAutor(post.author.id);

    const cabecalho = criarElemento('div', { classe: 'gs-post-head' });
    cabecalho.append(avatar, usuario, montarMenu(post));
    return cabecalho;
}

function montarRodape(post) {
    const rodape = criarElemento('div', { classe: 'gs-post-foot' });

    const curtidas = criarElemento('span', { classe: 'gs-action' });
    curtidas.append(
        criarElemento('span', { texto: `${post.likes} ${post.likes === 1 ? 'curtida' : 'curtidas'}` })
    );

    const respostas = criarElemento('a', {
        classe: 'gs-action',
        texto: `${post.comments_count} ${post.comments_count === 1 ? 'resposta' : 'respostas'}`,
        atributos: { href: `/post?id=${post.id}` }
    });

    rodape.append(curtidas, respostas);
    return rodape;
}

/** Devolve o <article class="gs-post"> completo de uma publicacao. */
export function criarCartaoDePost(post) {
    const artigo = criarElemento('article', {
        classe: 'gs-post',
        atributos: { 'data-post-id': String(post.id) }
    });

    const corpo = criarElemento('div', { classe: 'gs-post-body' });
    corpo.append(montarCabecalho(post));

    artigo.append(corpo, montarRodape(post));
    return artigo;
}
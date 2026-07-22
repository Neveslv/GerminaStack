/** Pagina de feed: lista publicacoes, filtra por materia e por busca. */

import { listarPosts, listarMaterias } from '../api.js';
import { criarCartaoDePost } from '../componentes/cartao-post.js';
import { criarElemento, comAtraso, criarPainelDeEstado } from '../utils/dom.js';

const feed = document.querySelector('#feed');
const listaMaterias = document.querySelector('#lista-materias');
const campoBusca = document.querySelector('#busca');
const estado = criarPainelDeEstado(document.querySelector('#estado-feed'));

const filtros = { idSubject: null, termo: '' };

let postsCarregados = [];

function aplicarBusca(posts) {
    const termo = filtros.termo.trim().toLowerCase();
    if (!termo) return posts;

    return posts.filter((post) =>
        `${post.title} ${post.content}`.toLowerCase().includes(termo)
    );
}

function renderizarFeed() {
    const visiveis = aplicarBusca(postsCarregados);

    feed.replaceChildren();

    if (visiveis.length === 0) {
        estado.vazio(
            'Nenhuma publicação encontrada',
            'Tente outra matéria ou outro termo de busca.'
        );
        return;
    }

    const fragmento = document.createDocumentFragment();
    visiveis.forEach((post) => fragmento.append(criarCartaoDePost(post)));
    feed.append(fragmento);

    estado.ocultar();
    window.GerminaStackUI?.init(feed);
    window.GerminaStackUI?.announceToScreenReader(`${visiveis.length} publicações listadas.`);
}

async function carregarPosts() {
    estado.carregando('Carregando publicações…');

    try {
        postsCarregados = await listarPosts({ idSubject: filtros.idSubject });
        renderizarFeed();
    } catch (erro) {
        feed.replaceChildren();
        estado.erro(erro.message, 'Recarregue a página para tentar de novo.');
    }
}

function marcarMateriaAtiva() {
    listaMaterias.querySelectorAll('button').forEach((botao) => {
        const id = botao.dataset.id ? Number(botao.dataset.id) : null;
        const ativa = id === filtros.idSubject;
        botao.setAttribute('aria-pressed', String(ativa));
        botao.classList.toggle('is-active', ativa);
    });
}

function criarLinhaDeMateria(rotulo, id, total) {
    const botao = criarElemento('button', {
        classe: 'gs-side-row',
        atributos: { type: 'button', 'aria-pressed': 'false' }
    });

    if (id !== null) botao.dataset.id = String(id);

    botao.append(
        criarElemento('span', { texto: rotulo }),
        criarElemento('span', { classe: 'gs-badge', texto: String(total) })
    );

    return botao;
}

function contarPorMateria(idMateria) {
    return postsCarregados.filter((post) => post.subject.id === idMateria).length;
}

function preencherEstatisticas(posts, materias) {
    const materiasComPost = new Set(posts.map((post) => post.subject.id));
    const totalRespostas = posts.reduce((soma, post) => soma + post.comments_count, 0);

    document.querySelector('#stat-posts').textContent = String(posts.length);
    document.querySelector('#stat-materias').textContent = String(
        materiasComPost.size || materias.length
    );
    document.querySelector('#stat-respostas').textContent = String(totalRespostas);
}

async function carregarMaterias() {
    try {
        const materias = await listarMaterias();
        const todos = await listarPosts();

        const fragmento = document.createDocumentFragment();
        fragmento.append(criarLinhaDeMateria('Todas', null, todos.length));

        materias.forEach((materia) => {
            const total = todos.filter((post) => post.subject.id === materia.id).length;
            fragmento.append(criarLinhaDeMateria(materia.subject, materia.id, total));
        });

        listaMaterias.append(fragmento);
        marcarMateriaAtiva();
        preencherEstatisticas(todos, materias);
    } catch {
        listaMaterias.append(criarLinhaDeMateria('Todas', null, 0));
        marcarMateriaAtiva();
    }
}

function copiarTextoDoPost(post) {
    navigator.clipboard?.writeText(post.content);
    window.GerminaStackUI?.showToast({
        title: 'Texto copiado',
        message: 'O conteúdo da publicação foi copiado.',
        tone: 'info'
    });
}

function ocultarPost(idPost) {
    postsCarregados = postsCarregados.filter((post) => post.id !== idPost);
    renderizarFeed();
    window.GerminaStackUI?.showToast({
        title: 'Publicação ocultada',
        message: 'Ela não vai mais aparecer no seu feed.',
        tone: 'warning'
    });
}

feed.addEventListener('click', (evento) => {
    const botao = evento.target.closest('[data-action]');
    if (!botao) return;

    const artigo = botao.closest('.gs-post');
    const idPost = Number(artigo?.dataset.postId);
    const post = postsCarregados.find((item) => item.id === idPost);
    if (!post) return;

    if (botao.dataset.action === 'copiar-post') copiarTextoDoPost(post);
    if (botao.dataset.action === 'ocultar-post') ocultarPost(idPost);
});

listaMaterias.addEventListener('click', (evento) => {
    const botao = evento.target.closest('button');
    if (!botao) return;

    filtros.idSubject = botao.dataset.id ? Number(botao.dataset.id) : null;
    marcarMateriaAtiva();
    carregarPosts();
});

campoBusca.addEventListener(
    'input',
    comAtraso((evento) => {
        filtros.termo = evento.target.value;
        renderizarFeed();
    }, 300)
);

carregarMaterias();
carregarPosts();
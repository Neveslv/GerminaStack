/** Pagina de detalhe: mostra a publicacao, a thread de respostas e o formulario. */

import {
    buscarPost,
    listarComentarios,
    criarComentario,
    buscarMinhaReacao
} from '../api.js';
import { criarPainelDeEstado, inicializarKit } from '../utils/dom.js';
import { criarCartaoDePost } from '../componentes/cartao-post.js';
import { criarThread, criarComentario as montarComentario } from '../componentes/comentario.js';

const publicacao = document.querySelector('#publicacao');
const thread = document.querySelector('#thread');
const formulario = document.querySelector('#form-resposta');
const campoResposta = document.querySelector('#nova-resposta');
const erroResposta = document.querySelector('#erro-resposta');
const estadoPost = criarPainelDeEstado(document.querySelector('#estado-post'));
const estadoRespostas = criarPainelDeEstado(document.querySelector('#estado-respostas'));

const idPost = new URLSearchParams(window.location.search).get('id');

function copiarTextoDoPost(post) {
    navigator.clipboard?.writeText(post.content);
    window.GerminaStackUI?.showToast({
        title: 'Texto copiado',
        message: 'O conteúdo da publicação foi copiado.',
        tone: 'info'
    });
}

function ligarMenuDaPublicacao(post) {
    publicacao.addEventListener('click', (evento) => {
        const botao = evento.target.closest('[data-action]');
        if (!botao) return;

        if (botao.dataset.action === 'copiar-post') copiarTextoDoPost(post);

        if (botao.dataset.action === 'ocultar-post') {
            window.GerminaStackUI?.showToast({
                title: 'Publicação ocultada',
                message: 'Você não vai mais ver essa publicação no feed.',
                tone: 'warning'
            });
            window.setTimeout(() => window.location.assign('/'), 900);
        }
    });
}

async function renderizarThread(comentarios) {
    if (comentarios.length === 0) {
        estadoRespostas.vazio('Ninguém respondeu ainda', 'Seja o primeiro a responder esta dúvida.');
        return;
    }

    thread.append(await criarThread(comentarios));

    estadoRespostas.ocultar();
    inicializarKit(thread);
}

function ligarFormularioDeResposta() {
    formulario.addEventListener('submit', async (evento) => {
        evento.preventDefault();

        erroResposta.textContent = '';
        campoResposta.setAttribute('aria-invalid', 'false');

        const texto = campoResposta.value.trim();

        if (!texto) {
            erroResposta.textContent = 'Escreva sua resposta antes de enviar.';
            campoResposta.setAttribute('aria-invalid', 'true');
            campoResposta.focus();
            return;
        }

        const botao = formulario.querySelector('button[type="submit"]');
        botao.disabled = true;

        try {
            const comentario = await criarComentario({
                tipo: 'comment',
                idPai: idPost,
                content: texto
            });

            // A thread pode estar mostrando o estado de "ninguém respondeu";
            // a primeira resposta precisa apagar esse aviso.
            estadoRespostas.ocultar();

            const novo = montarComentario(comentario);
            thread.append(novo);
            inicializarKit(novo);

            campoResposta.value = '';
            window.GerminaStackUI?.announceToScreenReader('Resposta publicada.', 'polite');
        } catch (erro) {
            erroResposta.textContent = erro.message;
            campoResposta.setAttribute('aria-invalid', 'true');
        } finally {
            botao.disabled = false;
        }
    });
}

async function carregarPagina() {
    if (!idPost) {
        estadoPost.erro('Publicação não informada', 'Volte ao feed e escolha uma publicação.');
        estadoRespostas.ocultar();
        formulario.hidden = true;
        return;
    }

    estadoPost.carregando('Carregando publicação…');
    estadoRespostas.carregando('Carregando respostas…');

    try {
        const post = await buscarPost(idPost);
        const minhaReacao = await buscarMinhaReacao('post', post.id);

        document.title = `${post.title} | GerminaStack`;

        publicacao.append(criarCartaoDePost(post, { minhaReacao, destaque: true }));
        inicializarKit(publicacao);
        estadoPost.ocultar();

        ligarMenuDaPublicacao(post);
        ligarFormularioDeResposta();
    } catch (erro) {
        estadoPost.erro(erro.message, 'Volte ao feed e tente outra publicação.');
        estadoRespostas.ocultar();
        formulario.hidden = true;
        return;
    }

    try {
        await renderizarThread(await listarComentarios(idPost));
    } catch (erro) {
        estadoRespostas.erro(erro.message, 'Recarregue a página para tentar de novo.');
    }
}

carregarPagina();

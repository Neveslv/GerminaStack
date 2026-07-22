/**
 * Ponto de entrada da página inicial.
 * Nesta etapa, controla apenas os estados do feed: carregando, erro e vazio.
 */

import { listarPosts } from './api.js';

const feed = document.querySelector('#feed');
const statusFeed = document.querySelector('#status-feed');

function exibirStatus(mensagem, ehErro = false) {
    statusFeed.textContent = mensagem;
    statusFeed.hidden = false;
    statusFeed.classList.toggle('gs-alert-error', ehErro);
}

function ocultarStatus() {
    statusFeed.hidden = true;
    statusFeed.textContent = '';
}

async function carregarFeed() {
    exibirStatus('Carregando publicações…');

    try {
        const posts = await listarPosts();

        if (posts.length === 0) {
            exibirStatus('Ainda não há publicações. Seja o primeiro a perguntar.');
            return;
        }

        ocultarStatus();
        feed.dataset.total = String(posts.length);
        exibirStatus(`${posts.length} publicações carregadas.`);
    } catch (erro) {
        exibirStatus(`${erro.message} Recarregue a página para tentar de novo.`, true);
    }
}

carregarFeed();
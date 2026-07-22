/**
 * Camada única de acesso à API.
 * Nenhum outro arquivo do projeto chama fetch diretamente.
 */

import {
    API_BASE_URL,
    ROTA_LOGIN,
    TIMEOUT_MS,
    USAR_DADOS_LOCAIS
} from './config.js';

import { POSTS_LOCAIS } from './posts-mockados.js';

/** Erro de rede ou de resposta da API, carregando o status HTTP quando existir. */
export class ErroDeApi extends Error {
    constructor(mensagem, status = 0) {
        super(mensagem);
        this.name = 'ErroDeApi';
        this.status = status;
    }
}

function encerrarSessao() {
    window.location.assign(ROTA_LOGIN);
}

async function lerMensagemDeErro(resposta) {
    try {
        const corpo = await resposta.json();
        return corpo.error || corpo.mensagem || `Erro ${resposta.status}.`;
    } catch {
        return `Erro ${resposta.status}.`;
    }
}

/**
 * Executa uma requisição à API e devolve o corpo já convertido de JSON.
 * @param {string} caminho rota iniciada por barra, ex.: '/api/posts'
 * @param {RequestInit} opcoes opções adicionais do fetch
 */
async function requisitar(caminho, opcoes = {}) {
    const controlador = new AbortController();
    const relogio = setTimeout(() => controlador.abort(), TIMEOUT_MS);

    try {
        const resposta = await fetch(`${API_BASE_URL}${caminho}`, {
            ...opcoes,
            signal: controlador.signal,
            credentials: 'include',
            headers: { 'Content-Type': 'application/json', ...opcoes.headers }
        });

        if (resposta.status === 401) {
            encerrarSessao();
            throw new ErroDeApi('Sessão expirada.', 401);
        }

        if (!resposta.ok) {
            throw new ErroDeApi(await lerMensagemDeErro(resposta), resposta.status);
        }

        return resposta.status === 204 ? null : await resposta.json();
    } catch (erro) {
        if (erro instanceof ErroDeApi) throw erro;
        if (erro.name === 'AbortError') {
            throw new ErroDeApi('O servidor demorou para responder.', 0);
        }
        throw new ErroDeApi('Não foi possível falar com o servidor.', 0);
    } finally {
        clearTimeout(relogio);
    }
}

/** Devolve a lista de publicações do feed. */
export async function listarPosts() {
    if (USAR_DADOS_LOCAIS) return POSTS_LOCAIS;
    return requisitar('/api/posts');
}

/** Devolve uma publicação pelo identificador. */
export async function buscarPost(id) {
    if (USAR_DADOS_LOCAIS) {
        const post = POSTS_LOCAIS.find((item) => item.id === Number(id));
        if (!post) throw new ErroDeApi('Publicação não encontrada.', 404);
        return post;
    }
    return requisitar(`/api/posts/${id}`);
}

/** Cria uma publicação e devolve o registro salvo. */
export async function criarPost(dados) {
    return requisitar('/api/posts', {
        method: 'POST',
        body: JSON.stringify(dados)
    });
}
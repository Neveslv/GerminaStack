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

import { POSTS_LOCAIS } from './mock/posts.js';
import { MATERIAS_LOCAIS } from './mock/subjects.js';
import { COMENTARIOS_LOCAIS } from './mock/comments.js';
import { USUARIO_ATUAL } from './mock/sessao.js';
import { ANOS_LOCAIS } from './mock/years.js';
import { USUARIOS_LOCAIS } from './mock/users.js';
import { NOTIFICACOES_LOCAIS } from './mock/notifications.js';
import { PREFERENCIAS_LOCAIS, gravarPreferenciasLocais } from './mock/preferences.js';
import { reacaoAtual, alternarReacao } from './mock/reactions.js';

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

        if (resposta.status === 401 && !caminho.startsWith('/api/login')) {
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

/**
 * Devolve a lista de publicações do feed.
 * @param {{ idSubject?: number }} filtros filtro opcional por matéria
 */
export async function listarPosts({ idSubject } = {}) {
    if (USAR_DADOS_LOCAIS) {
        if (!idSubject) return POSTS_LOCAIS;
        return POSTS_LOCAIS.filter((post) => post.subject.id === idSubject);
    }

    const query = idSubject ? `?subject_id=${idSubject}` : '';
    const [posts, materias, usuario] = await Promise.all([
        requisitar(`/api/posts${query}`),
        listarMaterias(),
        buscarMeuPerfil()
    ]);
    return posts.map((post) => normalizarPost(post, materias, usuario));
}

/** Devolve a lista de matérias disponíveis para filtro. */
export async function listarMaterias() {
    if (USAR_DADOS_LOCAIS) return MATERIAS_LOCAIS;
    return requisitar('/api/subjects');
}

/** Devolve uma publicação pelo identificador. */
export async function buscarPost(id) {
    if (USAR_DADOS_LOCAIS) {
        const post = POSTS_LOCAIS.find((item) => item.id === Number(id));
        if (!post) throw new ErroDeApi('Publicação não encontrada.', 404);
        return post;
    }
    const [post, materias, usuario] = await Promise.all([
        requisitar(`/api/posts/${id}`),
        listarMaterias(),
        buscarMeuPerfil()
    ]);
    return normalizarPost(post, materias, usuario);
}

/** Cria uma publicação e devolve o registro salvo. */
export async function criarPost(dados) {
    if (USAR_DADOS_LOCAIS) {
        const materia = MATERIAS_LOCAIS.find((item) => item.id === dados.id_subject);

        const novoPost = {
            id: Math.max(...POSTS_LOCAIS.map((post) => post.id)) + 1,
            title: dados.title,
            content: dados.content,
            image_url: dados.image_url ?? null,
            likes: 0,
            dislikes: 0,
            created_at: new Date().toISOString(),
            author: USUARIO_ATUAL,
            subject: materia ?? { id: dados.id_subject, subject: 'Matéria' },
            comments_count: 0
        };

        POSTS_LOCAIS.unshift(novoPost);
        return novoPost;
    }

    const post = await requisitar('/api/posts', {
        method: 'POST',
        body: JSON.stringify(dados)
    });
    return normalizarPost(post, await listarMaterias(), await buscarMeuPerfil());
}

/** Devolve os comentários de uma publicação, já com as respostas aninhadas. */
export async function listarComentarios(idPost) {
    if (USAR_DADOS_LOCAIS) return COMENTARIOS_LOCAIS[Number(idPost)] ?? [];
    const [comentarios, usuario] = await Promise.all([
        requisitar(`/api/posts/${idPost}/comments`),
        buscarMeuPerfil()
    ]);
    return Promise.all(comentarios.map(async (comentario) => ({
        ...normalizarMensagem(comentario, usuario),
        replies: (await requisitar(`/api/comments/${comentario.id}/replies`))
            .map((resposta) => normalizarMensagem(resposta, usuario))
    })));
}

/** Autentica o usuário. O token volta em cookie definido pelo servidor. */
export async function entrar(credenciais) {
    if (USAR_DADOS_LOCAIS) {
        if (!credenciais.username || !credenciais.password) {
            throw new ErroDeApi('Informe usuário e senha.', 400);
        }
        return { ...USUARIO_ATUAL, username: credenciais.username };
    }

    return requisitar('/api/login', {
        method: 'POST',
        body: JSON.stringify(credenciais)
    });
}

/** Cria uma conta. Espelha as colunas obrigatórias da tabela `users`. */
export async function cadastrar(dados) {
    if (USAR_DADOS_LOCAIS) {
        const jaExiste = USUARIOS_LOCAIS.some(
            (usuario) => usuario.username === dados.username || usuario.email === dados.email
        );

        // As colunas username e email são UNIQUE: o back devolve 409 nesse caso.
        if (jaExiste) throw new ErroDeApi('Já existe uma conta com esse usuário ou e-mail.', 409);

        const ano = ANOS_LOCAIS.find((item) => item.id === dados.id_year);
        const novoUsuario = {
            id: Math.max(...USUARIOS_LOCAIS.map((usuario) => usuario.id)) + 1,
            name: dados.name,
            username: dados.username,
            email: dados.email,
            year: ano ?? { id: dados.id_year, year: 'Turma' },
            created_at: new Date().toISOString()
        };

        USUARIOS_LOCAIS.push(novoUsuario);
        return novoUsuario;
    }

    return requisitar('/api/users', {
        method: 'POST',
        body: JSON.stringify({
            email: dados.email,
            year_id: dados.year_id ?? dados.id_year,
            password: dados.password,
            password_confirmation: dados.password_confirmation ?? dados.password
        })
    });
}

/** Devolve os anos/turmas cadastrados, usados no cadastro e no perfil. */
export async function listarAnos() {
    if (USAR_DADOS_LOCAIS) return ANOS_LOCAIS;
    return requisitar('/api/years');
}

/** Devolve o perfil público de um usuário pelo username. */
export async function buscarUsuario(username) {
    if (USAR_DADOS_LOCAIS) {
        const usuario = USUARIOS_LOCAIS.find((item) => item.username === username);
        if (!usuario) throw new ErroDeApi('Usuário não encontrado.', 404);
        return usuario;
    }
    const [meuPerfil, anos] = await Promise.all([buscarMeuPerfil(), listarAnos()]);
    const usuario = !username || username === meuPerfil.username
        ? meuPerfil
        : await requisitar(`/api/users/${encodeURIComponent(username)}`);
    return { ...usuario, year: anos.find((ano) => ano.id === usuario.id_year) };
}

/** Salva os dados editáveis do perfil do usuário logado. */
export async function salvarPerfil(dados) {
    return requisitar('/api/me', {
        method: 'PATCH',
        body: JSON.stringify(dados)
    });
}

/** Devolve as publicações escritas por um usuário. */
export async function listarPostsDoAutor(idAutor) {
    if (USAR_DADOS_LOCAIS) {
        return POSTS_LOCAIS.filter((post) => post.author.id === Number(idAutor));
    }
    const [posts, materias, usuario] = await Promise.all([
        requisitar(`/api/posts?author_id=${idAutor}`),
        listarMaterias(),
        buscarMeuPerfil()
    ]);
    return posts.map((post) => normalizarPost(post, materias, usuario));
}

/**
 * Devolve as notificações do usuário logado.
 * O `text_show` já vem montado pelo trigger `notify_mentions`.
 */
export async function listarNotificacoes() {
    if (USAR_DADOS_LOCAIS) return NOTIFICACOES_LOCAIS;
    return requisitar('/api/notifications');
}

/** Marca todas as notificações como lidas — equivale a mark_notifications_as_read(). */
export async function marcarNotificacoesComoLidas() {
    if (USAR_DADOS_LOCAIS) {
        NOTIFICACOES_LOCAIS.forEach((notificacao) => {
            notificacao.is_read = true;
        });
        return null;
    }

    return requisitar('/api/notifications/read-all', { method: 'PATCH' });
}

/**
 * Registra ou desfaz uma reação, espelhando a função `reaction()` do banco.
 *
 * @param {{ tipo: 'post'|'comment'|'comment_on_comment', id: number, reacao: 'like'|'dislike' }} alvo
 * @returns {Promise<{ reacao: string|null, likes: number, dislikes: number }>} deltas dos contadores
 */
export async function reagir({ tipo, id, reacao }) {
    if (USAR_DADOS_LOCAIS) {
        const deltas = alternarReacao(tipo, id, reacao);

        if (tipo === 'post') {
            const post = POSTS_LOCAIS.find((item) => item.id === Number(id));
            if (post) {
                post.likes = Math.max(0, post.likes + deltas.likes);
                post.dislikes = Math.max(0, post.dislikes + deltas.dislikes);
            }
        }

        return deltas;
    }

    const recurso = { post: 'posts', comment: 'comments', comment_on_comment: 'replies' }[tipo];
    const chave = `${tipo}:${id}`;
    const anterior = reacoesDaSessao.get(chave) ?? null;
    await requisitar(`/api/${recurso}/${id}/reaction`, {
        method: 'PUT',
        body: JSON.stringify({ reaction_type: reacao })
    });
    const atual = anterior === reacao ? null : reacao;
    reacoesDaSessao.set(chave, atual);
	const atualizado = await requisitar(`/api/${recurso}/${id}`);
    return {
        reacao: atual,
		likes: atualizado.likes,
		dislikes: atualizado.dislikes,
		absolutos: true
    };
}

/** Devolve a reação do usuário logado para um alvo, ou null. */
export async function buscarMinhaReacao(tipo, id) {
    if (USAR_DADOS_LOCAIS) return reacaoAtual(tipo, id);
    return reacoesDaSessao.get(`${tipo}:${id}`) ?? null;
}

/**
 * Cria um comentário ou uma resposta, espelhando `create_message()`.
 *
 * @param {{ tipo: 'comment'|'comment_on_comment', idPai: number, content: string }} dados
 *        idPai é o post quando tipo='comment' e o comentário quando é resposta.
 */
export async function criarComentario({ tipo, idPai, content }) {
    if (USAR_DADOS_LOCAIS) {
        const novo = {
            id: Math.floor(Math.random() * 100000) + 1000,
            content,
            likes: 0,
            dislikes: 0,
            created_at: new Date().toISOString(),
            author: USUARIO_ATUAL
        };

        if (tipo === 'comment') {
            const idPost = Number(idPai);
            novo.replies = [];
            COMENTARIOS_LOCAIS[idPost] = COMENTARIOS_LOCAIS[idPost] ?? [];
            COMENTARIOS_LOCAIS[idPost].push(novo);

            const post = POSTS_LOCAIS.find((item) => item.id === idPost);
            if (post) post.comments_count += 1;

            return novo;
        }

        for (const comentarios of Object.values(COMENTARIOS_LOCAIS)) {
            const pai = comentarios.find((item) => item.id === Number(idPai));
            if (pai) {
                pai.replies = pai.replies ?? [];
                pai.replies.push(novo);
                return novo;
            }
        }

        throw new ErroDeApi('Comentário pai não encontrado.', 404);
    }

    const rota = tipo === 'comment'
        ? `/api/posts/${idPai}/comments`
        : `/api/comments/${idPai}/replies`;

    const comentario = await requisitar(rota, { method: 'POST', body: JSON.stringify({ content }) });
    return normalizarMensagem(comentario, await buscarMeuPerfil());
}

/** Devolve as preferências de acessibilidade do usuário logado. */
export async function buscarPreferencias() {
    if (USAR_DADOS_LOCAIS) return PREFERENCIAS_LOCAIS;
    return requisitar('/api/me/preferences');
}

/** Salva as preferências de acessibilidade do usuário logado. */
export async function salvarPreferencias(preferencias) {
    if (USAR_DADOS_LOCAIS) return gravarPreferenciasLocais(preferencias);

    return requisitar('/api/me/preferences', {
        method: 'PATCH',
        body: JSON.stringify(preferencias)
    });
}

let perfilAtual;
const reacoesDaSessao = new Map();

async function buscarMeuPerfil() {
    perfilAtual ??= requisitar('/api/me').catch((erro) => {
        perfilAtual = null;
        throw erro;
    });
    return perfilAtual;
}

function normalizarAutor(id, usuario, name, username) {
    if (id === usuario.id) return usuario;
    return { id, name: name || `Usuário #${id}`, username: username || null };
}

function normalizarMensagem(item, usuario) {
    return { ...item, author: normalizarAutor(item.id_user, usuario, item.author_name, item.author_username) };
}

function normalizarPost(post, materias, usuario) {
    return {
        ...post,
        author: normalizarAutor(post.id_user, usuario, post.author_name, post.author_username),
        subject: materias.find((materia) => materia.id === post.id_subject) ?? {
            id: post.id_subject,
            subject: 'Matéria'
        },
        comments_count: post.comments_count ?? 0
    };
}

export async function completarLogin(dados) {
    if (USAR_DADOS_LOCAIS) return null;
    return requisitar('/api/login/2fa', {
        method: 'POST',
        body: JSON.stringify(dados)
    });
}

export async function sair() {
    if (USAR_DADOS_LOCAIS) return null;
    return requisitar('/api/logout', { method: 'POST' });
}

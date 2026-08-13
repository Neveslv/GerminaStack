import { API_BASE_URL, ROTA_LOGIN, TIMEOUT_MS, NO_CONTENT, NAO_AUTORIZADO } from './config.js';

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

        if (resposta.status === NAO_AUTORIZADO && !caminho.startsWith('/api/login')) {
            encerrarSessao();
            throw new ErroDeApi('Sessão expirada.', NAO_AUTORIZADO);
        }

        if (!resposta.ok) {
            throw new ErroDeApi(await lerMensagemDeErro(resposta), resposta.status);
        }

        return resposta.status === NO_CONTENT ? null : await resposta.json();
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

export async function listarPosts({ idSubject } = {}) {
    const query = idSubject ? `?subject_id=${idSubject}` : '';
    const [posts, materias, usuario] = await Promise.all([
        requisitar(`/api/posts${query}`),
        listarMaterias(),
        buscarMeuPerfil()
    ]);
    return Promise.all(posts.map((post) => normalizarPost(post, materias, usuario)));
}

export async function listarMaterias() {
    return requisitar('/api/subjects');
}

export async function buscarPost(id) {
    const [post, materias, usuario] = await Promise.all([
        requisitar(`/api/posts/${id}`),
        listarMaterias(),
        buscarMeuPerfil()
    ]);
    return normalizarPost(post, materias, usuario);
}

export async function criarPost(dados) {
    const post = await requisitar('/api/posts', {
        method: 'POST',
        body: JSON.stringify(dados)
    });
    return normalizarPost(post, await listarMaterias(), await buscarMeuPerfil());
}

export async function listarComentarios(idPost) {
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

export async function entrar(credenciais) {
    return requisitar('/api/login', {
        method: 'POST',
        body: JSON.stringify(credenciais)
    });
}

export async function cadastrar(dados) {
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

export async function listarAnos() {
    return requisitar('/api/years');
}

export async function buscarUsuario(username) {
    const [meuPerfil, anos] = await Promise.all([buscarMeuPerfil(), listarAnos()]);
    const usuario = !username || username === meuPerfil.username
        ? meuPerfil
        : await requisitar(`/api/users/${encodeURIComponent(username)}`);
    return { ...usuario, year: anos.find((ano) => ano.id === usuario.id_year) };
}

export async function salvarPerfil(dados) {
    return requisitar('/api/me', {
        method: 'PATCH',
        body: JSON.stringify(dados)
    });
}

export async function listarUsuariosAdmin({ page = 1, q = '' } = {}) {
    return requisitar(`/api/admin/users?page=${page}&q=${encodeURIComponent(q)}`);
}

export async function listarPostsAdmin({ page = 1, q = '' } = {}) {
    return requisitar(`/api/admin/posts?page=${page}&q=${encodeURIComponent(q)}`);
}
export async function banirUsuario(id, enabled = true) { 
    return requisitar(`/api/admin/users/${id}/ban`, { method: 'PATCH', body: JSON.stringify({ enabled }) }); 
}
export async function definirAdmin(id, enabled) { 
    return requisitar(`/api/admin/users/${id}/admin`, { method: 'PATCH', body: JSON.stringify({ enabled }) }); 
}

export async function excluirPost(id) { 
    return requisitar(`/api/admin/posts/${id}`, { method: 'DELETE' }); 
}

export async function listarPostsDoAutor(idAutor) {
    const [posts, materias, usuario] = await Promise.all([
        requisitar(`/api/posts?author_id=${idAutor}`),
        listarMaterias(),
        buscarMeuPerfil()
    ]);
    return Promise.all(posts.map((post) => normalizarPost(post, materias, usuario)));
}


export async function listarNotificacoes() {
    return requisitar('/api/notifications');
}

export async function marcarNotificacoesComoLidas() {
    return requisitar('/api/notifications/read-all', { method: 'PATCH' });
}

export async function reagir({ tipo, id, reacao }) {
    const recurso = { post: 'posts', comment: 'comments', comment_on_comment: 'replies' }[tipo];
    const chave = `${tipo}:${id}`;
    const anterior = reacoesDaSessao.get(chave) ?? null;
    const atualizado = await requisitar(`/api/${recurso}/${id}/reaction`, {
        method: 'PUT',
        body: JSON.stringify({ reaction_type: reacao })
    });
    const atual = anterior === reacao ? null : reacao;
    reacoesDaSessao.set(chave, atual);
    return {
        reacao: atual,
        likes: atualizado.likes,
        dislikes: atualizado.dislikes,
        absolutos: true
    };
}

export async function buscarMinhaReacao(tipo, id) {
    return reacoesDaSessao.get(`${tipo}:${id}`) ?? null;
}

export async function criarComentario({ tipo, idPai, content }) {
    const rota = tipo === 'comment'
        ? `/api/posts/${idPai}/comments`
        : `/api/comments/${idPai}/replies`;

    const comentario = await requisitar(rota, { method: 'POST', body: JSON.stringify({ content }) });
    return normalizarMensagem(comentario, await buscarMeuPerfil());
}

export async function buscarPreferencias() {
    return requisitar('/api/me/preferences');
}

export async function salvarPreferencias(preferencias) {
    return requisitar('/api/me/preferences', {
        method: 'PATCH',
        body: JSON.stringify(preferencias)
    });
}

let perfilAtual;
const reacoesDaSessao = new Map();
const perfisPublicos = new Map();

export async function buscarMeuPerfil() {
    perfilAtual ??= requisitar('/api/me').catch((erro) => {
        perfilAtual = null;
        throw erro;
    });
    return perfilAtual;
}

function normalizarAutor(id, usuario, name, username, imageUrl, imageDescription) {
    if (id === usuario.id) return usuario;
    return { id, name: name || `Usuário #${id}`, username: username || null, profile_image_url: imageUrl || null, profile_image_description: imageDescription || null };
}

function normalizarMensagem(item, usuario) {
    return { ...item, author: normalizarAutor(item.id_user, usuario, item.author_name, item.author_username) };
}

async function normalizarPost(post, materias, usuario) {
    const author = normalizarAutor(post.id_user, usuario, post.author_name, post.author_username, post.author_image_url, post.author_image_description);
    if (!author.profile_image_url && author.username) {
        perfisPublicos.set(author.username, perfisPublicos.get(author.username) ?? buscarUsuario(author.username).catch(() => null));
        const perfil = await perfisPublicos.get(author.username);
        if (perfil?.profile_image_url) {
            author.profile_image_url = perfil.profile_image_url;
            author.profile_image_description = perfil.profile_image_description;
        }
    }
    return {
        ...post,
        author,
        subject: materias.find((materia) => materia.id === post.id_subject) ?? {
            id: post.id_subject,
            subject: 'Matéria'
        },
        comments_count: post.comments_count ?? 0
    };
}

export async function completarLogin(dados) {
    return requisitar('/api/login/2fa', {
        method: 'POST',
        body: JSON.stringify(dados)
    });
}

export async function sair() {
    return requisitar('/api/logout', { method: 'POST' });
}

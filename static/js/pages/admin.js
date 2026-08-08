import { buscarMeuPerfil, excluirPost, listarPostsAdmin, listarUsuariosAdmin, banirUsuario, definirAdmin } from '../api.js';
import { criarElemento, criarPainelDeEstado } from '../utils/dom.js';

const usuarios = document.querySelector('#usuarios-admin');
const posts = document.querySelector('#posts-admin');
const estado = criarPainelDeEstado(document.querySelector('#estado-admin'));
const superAdmins = new Set(['nicolas.oliveira', 'matheus.fazan']);
const paginas = { usuarios: 1, posts: 1 };

function botao(texto, acao) { const item = criarElemento('button', { classe: 'gs-btn gs-btn-ghost', texto, atributos: { type: 'button' } }); item.addEventListener('click', acao); return item; }
function paginacao(alvo, tipo, total) {
    alvo.replaceChildren(); const atual = paginas[tipo]; const temAnterior = atual > 1; const temProxima = atual * 5 < total;
    if (temAnterior) alvo.append(botao('Anterior', () => { paginas[tipo]--; carregar(); }));
    alvo.append(criarElemento('span', { classe: 'gs-form-hint', texto: `Página ${atual} de ${Math.max(1, Math.ceil(total / 5))}` }));
    if (temProxima) alvo.append(botao('Próxima', () => { paginas[tipo]++; carregar(); }));
}
function renderizarUsuarios(lista, eu) {
    usuarios.replaceChildren(); lista.forEach((usuario) => {
        const linha = criarElemento('article', { classe: 'gs-list-row admin-user' }); const info = criarElemento('div');
        info.append(criarElemento('strong', { texto: usuario.name }), criarElemento('p', { classe: 'gs-form-hint', texto: `@${usuario.username}${usuario.is_banned ? ' · Banido' : usuario.is_admin ? ' · Admin' : ''}` })); linha.append(info);
        const acoes = criarElemento('div', { classe: 'gs-cluster' }); const protegido = usuario.id === eu.id || superAdmins.has(usuario.username) || (!superAdmins.has(eu.username) && usuario.is_admin);
        if (!protegido) acoes.append(botao(usuario.is_banned ? 'Desbanir' : 'Banir', async () => { await banirUsuario(usuario.id, !usuario.is_banned); carregar(); }));
        if (superAdmins.has(eu.username) && !superAdmins.has(usuario.username) && usuario.id !== eu.id) acoes.append(botao(usuario.is_admin ? 'Remover admin' : 'Tornar admin', async () => { await definirAdmin(usuario.id, !usuario.is_admin); carregar(); }));
        linha.append(acoes); usuarios.append(linha);
    });
}
function renderizarPosts(lista) { posts.replaceChildren(); lista.forEach((post) => { const linha = criarElemento('article', { classe: 'gs-list-row admin-post' }); const info = criarElemento('div'); info.append(criarElemento('strong', { texto: post.title }), criarElemento('p', { classe: 'gs-form-hint', texto: `Por ${post.author_name} (@${post.author_username})` })); linha.append(info, botao('Excluir', async () => { await excluirPost(post.id); carregar(); })); posts.append(linha); }); }
async function carregar() {
    try { estado.carregando('Carregando administração…'); const [eu, listaUsuarios, listaPosts] = await Promise.all([buscarMeuPerfil(), listarUsuariosAdmin({ page: paginas.usuarios, q: document.querySelector('#busca-usuarios').value.trim() }), listarPostsAdmin({ page: paginas.posts, q: document.querySelector('#busca-posts').value.trim() })]); renderizarUsuarios(listaUsuarios.items, eu); renderizarPosts(listaPosts.items); paginacao(document.querySelector('#paginacao-usuarios'), 'usuarios', listaUsuarios.total); paginacao(document.querySelector('#paginacao-posts'), 'posts', listaPosts.total); estado.ocultar(); } catch (erro) { estado.erro('Não foi possível carregar a administração.', erro.message); }
}
['usuarios', 'posts'].forEach((tipo) => document.querySelector(`#busca-${tipo}`).addEventListener('input', () => { paginas[tipo] = 1; carregar(); }));
carregar();
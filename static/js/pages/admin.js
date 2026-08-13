import {
    buscarMeuPerfil,
    excluirPost,
    listarPostsAdmin,
    listarUsuariosAdmin,
    banirUsuario,
    definirAdmin
} from '../api.js';
import { criarElemento, criarPainelDeEstado } from '../utils/dom.js';

const ITENS_POR_PAGINA = 5;

const usuarios = document.querySelector('#usuarios-admin');
const posts = document.querySelector('#posts-admin');
const estado = criarPainelDeEstado(document.querySelector('#estado-admin'));

const superAdmins = new Set(['nicolas.oliveira', 'matheus.fazan']);
const paginas = { usuarios: 1, posts: 1 };

function botao(texto, acao) {
    const item = criarElemento('button', {
        classe: 'gs-btn gs-btn-ghost',
        texto,
        atributos: { type: 'button' }
    });
    item.addEventListener('click', acao);
    return item;
}

function paginacao(alvo, tipo, total) {
    alvo.replaceChildren();

    const atual = paginas[tipo];
    const temAnterior = atual > 1;
    const temProxima = atual * ITENS_POR_PAGINA < total;

    if (temAnterior) {
        alvo.append(botao('Anterior', () => {
            paginas[tipo]--;
            carregar();
        }));
    }

    const totalDePaginas = Math.max(1, Math.ceil(total / ITENS_POR_PAGINA));
    alvo.append(criarElemento('span', {
        classe: 'gs-form-hint',
        texto: `Página ${atual} de ${totalDePaginas}`
    }));

    if (temProxima) {
        alvo.append(botao('Próxima', () => {
            paginas[tipo]++;
            carregar();
        }));
    }
}

function formatarDataCadastro(data) {
    if (!data) return 'Não informado';

    const dataConvertida = new Date(data);
    if (Number.isNaN(dataConvertida.getTime())) return 'Não informado';

    return dataConvertida.toLocaleDateString('pt-BR');
}

function criarDetalhesDoUsuario(usuario) {
    const detalhes = criarElemento('div', { classe: 'gs-stack' });

    detalhes.append(
        criarElemento('strong', { texto: usuario.name || 'Nome não informado' }),
        criarElemento('span', { classe: 'gs-form-hint', texto: `Username: @${usuario.username || 'não informado'}` }),
        criarElemento('span', { classe: 'gs-form-hint', texto: `E-mail: ${usuario.email || 'Não informado'}` }),
        criarElemento('span', { classe: 'gs-form-hint', texto: `Ano/turma: ${usuario.id_year ?? 'Não informado'}` }),
        criarElemento('span', { classe: 'gs-form-hint', texto: `Cadastro: ${formatarDataCadastro(usuario.created_at)}` }),
        criarElemento('span', { classe: 'gs-form-hint', texto: `Status: ${usuario.is_banned ? 'Banido' : 'Ativo'}` }),
        criarElemento('span', { classe: 'gs-form-hint', texto: `Administrador: ${usuario.is_admin ? 'Sim' : 'Não'}` })
    );

    return detalhes;
}

function renderizarUsuarios(lista, eu) {
    usuarios.replaceChildren();

    lista.forEach((usuario) => {
        const linha = criarElemento('article', { classe: 'gs-list-row admin-user' });

        const info = criarDetalhesDoUsuario(usuario);
        linha.append(info);

        const protegido = usuario.id === eu.id
            || superAdmins.has(usuario.username)
            || (!superAdmins.has(eu.username) && usuario.is_admin);

        const acoes = criarElemento('div', { classe: 'gs-cluster' });

        if (!protegido) {
            acoes.append(botao(usuario.is_banned ? 'Desbanir' : 'Banir', async () => {
                await banirUsuario(usuario.id, !usuario.is_banned);
                carregar();
            }));
        }

        const podePromover = superAdmins.has(eu.username)
            && !superAdmins.has(usuario.username)
            && usuario.id !== eu.id;

        if (podePromover) {
            acoes.append(botao(usuario.is_admin ? 'Remover admin' : 'Tornar admin', async () => {
                await definirAdmin(usuario.id, !usuario.is_admin);
                carregar();
            }));
        }

        linha.append(acoes);
        usuarios.append(linha);
    });
}

function renderizarPosts(lista) {
    posts.replaceChildren();

    lista.forEach((post) => {
        const linha = criarElemento('article', { classe: 'gs-list-row admin-post' });

        const info = criarElemento('div');
        info.append(
            criarElemento('strong', { texto: post.title }),
            criarElemento('p', {
                classe: 'gs-form-hint',
                texto: `Por ${post.author_name} (@${post.author_username})`
            })
        );

        const excluir = botao('Excluir', async () => {
            await excluirPost(post.id);
            carregar();
        });

        linha.append(info, excluir);
        posts.append(linha);
    });
}

async function carregar() {
    try {
        estado.carregando('Carregando administração…');

        const buscaUsuarios = document.querySelector('#busca-usuarios').value.trim();
        const buscaPosts = document.querySelector('#busca-posts').value.trim();

        const [eu, listaUsuarios, listaPosts] = await Promise.all([
            buscarMeuPerfil(),
            listarUsuariosAdmin({ page: paginas.usuarios, q: buscaUsuarios }),
            listarPostsAdmin({ page: paginas.posts, q: buscaPosts })
        ]);

        renderizarUsuarios(listaUsuarios.items, eu);
        renderizarPosts(listaPosts.items);
        paginacao(document.querySelector('#paginacao-usuarios'), 'usuarios', listaUsuarios.total);
        paginacao(document.querySelector('#paginacao-posts'), 'posts', listaPosts.total);

        estado.ocultar();
    } catch (erro) {
        estado.erro('Não foi possível carregar a administração.', erro.message);
    }
}

['usuarios', 'posts'].forEach((tipo) => {
    document.querySelector(`#busca-${tipo}`).addEventListener('input', () => {
        paginas[tipo] = 1;
        carregar();
    });
});

carregar();

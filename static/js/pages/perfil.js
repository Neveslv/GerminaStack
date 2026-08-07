/**
 * Página de perfil.
 *
 * Sem `?usuario=` na URL, mostra o perfil de quem está logado.
 * Com `?usuario=fulano`, mostra o perfil público daquela pessoa — é para lá
 * que apontam os links de autor no feed e nos comentários.
 */

import { buscarMeuPerfil, buscarUsuario, listarPostsDoAutor, buscarMinhaReacao, salvarPerfil, sair } from '../api.js';
import { criarCartaoDePost } from '../componentes/cartao-post.js';
import { criarElemento, criarPainelDeEstado, inicializarKit } from '../utils/dom.js';
import { corDoAutor, fotoDoAutor, inicialDoNome } from '../utils/identidade.js';
import { formatarDataCompleta } from '../utils/data.js';

const cartao = document.querySelector('#cartao-perfil');
const numeros = document.querySelector('#numeros-do-perfil');
const publicacoes = document.querySelector('#publicacoes-do-usuario');
const estadoPerfil = criarPainelDeEstado(document.querySelector('#estado-perfil'));
const estadoPublicacoes = criarPainelDeEstado(document.querySelector('#estado-publicacoes'));

const usuarioDaUrl = new URLSearchParams(window.location.search).get('usuario');
let ehMeuPerfil = false;

function montarCartao(usuario) {
    cartao.replaceChildren();

    const foto = fotoDoAutor(usuario);
    const avatar = foto
        ? criarElemento('img', {
            classe: 'gs-avatar',
            atributos: { src: foto, alt: usuario.profile_image_description || `Foto de ${usuario.name}` }
        })
        : criarElemento('span', {
            classe: 'gs-avatar', texto: inicialDoNome(usuario.name),
            atributos: { 'aria-hidden': 'true' }
        });
    if (foto) avatar.style.objectFit = 'cover';
    avatar.style.background = corDoAutor(usuario.id);

    const identificacao = criarElemento('div', { classe: 'gs-stack' });
    identificacao.append(
        criarElemento('h1', {
            classe: 'gs-card-title',
            texto: usuario.name,
            atributos: { id: 'nome-do-usuario' }
        }),
        criarElemento('p', { classe: 'gs-card-subtitle', texto: `@${usuario.username}` })
    );

    const chips = criarElemento('div', { classe: 'gs-cluster' });

    const turma = criarElemento('span', { classe: 'gs-chip is-blue' });
    turma.append(
        criarElemento('span', { classe: 'gs-chip-dot', atributos: { 'aria-hidden': 'true' } }),
        criarElemento('span', { texto: usuario.year?.year ?? 'Turma não informada' })
    );
    chips.append(turma);

    if (ehMeuPerfil) chips.append(criarElemento('span', { classe: 'gs-badge', texto: 'Você' }));

    identificacao.append(chips);

    // O e-mail é dado pessoal: só aparece no próprio perfil, nunca no de outra
    // pessoa, mesmo que a API mande o campo.
    if (ehMeuPerfil) {
        identificacao.append(
            criarElemento('p', { classe: 'gs-form-hint', texto: usuario.email })
        );
    }

    if (usuario.created_at) {
        identificacao.append(
            criarElemento('p', {
                classe: 'gs-form-hint',
                texto: `Na plataforma desde ${formatarDataCompleta(usuario.created_at)}`
            })
        );
    }

    const cabecalho = criarElemento('div', { classe: 'gs-post-head' });
    cabecalho.append(avatar, identificacao);
    cartao.append(cabecalho);

    if (ehMeuPerfil) {
        const formulario = criarElemento('form', { classe: 'gs-stack' });
        const url = criarElemento('input', {
            classe: 'gs-input', atributos: { type: 'url', name: 'profile_image_url', placeholder: 'https://exemplo.com/minha-foto.jpg', value: usuario.profile_image_url || '', maxlength: '2000' }
        });
        const descricao = criarElemento('input', {
            classe: 'gs-input', atributos: { type: 'text', name: 'profile_image_description', placeholder: 'Descreva sua foto', value: usuario.profile_image_description || '', maxlength: '200' }
        });
        const mensagem = criarElemento('span', { classe: 'gs-form-hint', texto: 'Use uma imagem pública. Deixe os dois campos vazios para remover a foto.' });
        const salvar = criarElemento('button', { classe: 'gs-btn gs-btn-secondary', texto: 'Salvar foto', atributos: { type: 'submit' } });
        formulario.append(
            criarElemento('label', { classe: 'gs-form-group', texto: 'Foto de perfil' }),
            url, descricao, mensagem, salvar
        );
        formulario.querySelector('label').append(url);
        formulario.addEventListener('submit', async (evento) => {
            evento.preventDefault();
            const imagem = url.value.trim();
            const texto = descricao.value.trim();
            salvar.disabled = true;
            try {
                await salvarPerfil({ profile_image_url: imagem || null, profile_image_description: imagem ? texto : null });
                window.location.reload();
            } catch (erro) {
                salvar.disabled = false;
                window.GerminaStackUI?.showToast({ title: 'Não foi possível salvar', message: erro.message, tone: 'danger' });
            }
        });
        cartao.append(formulario);
        const acoes = criarElemento('div', { classe: 'gs-cluster' });
        const botaoSair = criarElemento('button', {
            classe: 'gs-btn gs-btn-ghost',
            texto: 'Sair',
            atributos: { type: 'button' }
        });
        botaoSair.addEventListener('click', async () => {
            botaoSair.disabled = true;
            try {
                await sair();
                window.location.assign('/login');
            } catch (erro) {
                botaoSair.disabled = false;
                window.GerminaStackUI?.showToast({
                    title: 'Não foi possível sair',
                    message: erro.message,
                    tone: 'danger'
                });
            }
        });
        acoes.append(
            criarElemento('a', {
                classe: 'gs-btn gs-btn-secondary',
                texto: 'Ajustar acessibilidade',
                atributos: { href: '/preferencias' }
            }),
            criarElemento('a', {
                classe: 'gs-btn gs-btn-primary',
                texto: 'Nova publicação',
                atributos: { href: '/publicar' }
            }),
            botaoSair
        );
        cartao.append(acoes);
    }
}

function montarNumeros(posts) {
    numeros.replaceChildren();

    const totalCurtidas = posts.reduce((soma, post) => soma + post.likes, 0);
    const totalRespostas = posts.reduce((soma, post) => soma + (post.comments_count ?? 0), 0);
    const materias = new Set(posts.map((post) => post.subject.id)).size;

    const cartoes = [
        { rotulo: 'Publicações', valor: posts.length },
        { rotulo: 'Curtidas recebidas', valor: totalCurtidas },
        { rotulo: 'Respostas recebidas', valor: totalRespostas },
        { rotulo: 'Matérias em que publicou', valor: materias }
    ];

    const fragmento = document.createDocumentFragment();

    cartoes.forEach(({ rotulo, valor }) => {
        const item = criarElemento('article', { classe: 'gs-metric-card gs-stack' });
        item.append(
            criarElemento('span', { classe: 'gs-metric-label', texto: rotulo }),
            criarElemento('strong', { texto: String(valor) })
        );
        fragmento.append(item);
    });

    numeros.append(fragmento);
}

async function montarPublicacoes(usuario, posts) {
    publicacoes.replaceChildren();

    if (posts.length === 0) {
        estadoPublicacoes.vazio(
            ehMeuPerfil ? 'Você ainda não publicou' : `${usuario.name} ainda não publicou`,
            ehMeuPerfil ? 'Sua primeira dúvida pode ajudar mais gente do que você imagina.' : ''
        );
        return;
    }

    const reacoes = await Promise.all(posts.map((post) => buscarMinhaReacao('post', post.id)));

    const fragmento = document.createDocumentFragment();
    posts.forEach((post, indice) => {
        fragmento.append(criarCartaoDePost(post, { minhaReacao: reacoes[indice] }));
    });
    publicacoes.append(fragmento);

    estadoPublicacoes.ocultar();
    inicializarKit(publicacoes);
}

async function carregarPagina() {
    estadoPerfil.carregando('Carregando perfil…');
    estadoPublicacoes.carregando('Carregando publicações…');

    try {
        const [usuario, meuPerfil] = await Promise.all([buscarUsuario(usuarioDaUrl), buscarMeuPerfil()]);
        ehMeuPerfil = usuario.id === meuPerfil.id;

        document.title = `${usuario.name} | GerminaStack`;
        montarCartao(usuario);
        estadoPerfil.ocultar();

        const posts = await listarPostsDoAutor(usuario.id);
        montarNumeros(posts);
        await montarPublicacoes(usuario, posts);
    } catch (erro) {
        estadoPerfil.erro(erro.message, 'Confira o endereço ou volte ao feed.');
        estadoPublicacoes.ocultar();
    }
}

carregarPagina();

import { listarMaterias, listarPosts, buscarMinhaReacao } from '../api.js';
import { criarCartaoDePost } from '../componentes/cartao-post.js';
import { criarElemento, criarPainelDeEstado, inicializarKit } from '../utils/dom.js';
import { tomDaMateria } from '../utils/identidade.js';

const grade = document.querySelector('#grade-materias');
const resumo = document.querySelector('#resumo-materias');
const listaPublicacoes = document.querySelector('#publicacoes-da-materia');
const tituloPublicacoes = document.querySelector('#titulo-publicacoes');
const estadoMaterias = criarPainelDeEstado(document.querySelector('#estado-materias'));
const estadoPublicacoes = criarPainelDeEstado(document.querySelector('#estado-publicacoes'));

let materiasCarregadas = [];
let todosOsPosts = [];
let materiaAtiva = null;

function lerMateriaDaUrl() {
    const id = new URLSearchParams(window.location.search).get('id');
    return id ? Number(id) : null;
}

function totalDaMateria(idMateria) {
    return todosOsPosts.filter((post) => post.subject.id === idMateria).length;
}

function nomeDaMateria(idMateria) {
    return materiasCarregadas.find((materia) => materia.id === idMateria)?.subject ?? 'Matéria';
}

function criarCartaoDeMateria(materia) {
    const total = totalDaMateria(materia.id);
    const ativa = materia.id === materiaAtiva;

    const cartao = criarElemento('button', {
        classe: `gs-metric-card gs-stack${ativa ? ' is-active' : ''}`,
        atributos: {
            type: 'button',
            'aria-pressed': String(ativa),
            'data-id': String(materia.id),
            'aria-label': `${materia.subject}, ${total} ${total === 1 ? 'publicação' : 'publicações'}`
        }
    });

    const chip = criarElemento('span', { classe: `gs-chip ${tomDaMateria(materia.id)}` });
    chip.append(
        criarElemento('span', { classe: 'gs-chip-dot', atributos: { 'aria-hidden': 'true' } }),
        criarElemento('span', { texto: materia.subject })
    );

    cartao.append(
        chip,
        criarElemento('strong', { texto: String(total) }),
        criarElemento('span', {
            classe: 'gs-metric-label',
            texto: total === 1 ? 'publicação' : 'publicações'
        })
    );

    return cartao;
}

function renderizarMaterias() {
    grade.replaceChildren();

    if (materiasCarregadas.length === 0) {
        estadoMaterias.vazio('Nenhuma matéria cadastrada', 'As matérias vêm da tabela subjects.');
        return;
    }

    const fragmento = document.createDocumentFragment();
    materiasCarregadas.forEach((materia) => fragmento.append(criarCartaoDeMateria(materia)));
    grade.append(fragmento);

    estadoMaterias.ocultar();
}

function renderizarResumo() {
    resumo.replaceChildren();

    const fragmento = document.createDocumentFragment();

    const linhaTodas = criarElemento('a', {
        classe: 'gs-side-row',
        atributos: { href: '/materias' }
    });
    linhaTodas.append(
        criarElemento('span', { texto: 'Todas as matérias' }),
        criarElemento('span', { classe: 'gs-badge', texto: String(todosOsPosts.length) })
    );
    fragmento.append(linhaTodas);

    materiasCarregadas.forEach((materia) => {
        const linha = criarElemento('a', {
            classe: 'gs-side-row',
            atributos: { href: `/materias?id=${materia.id}` }
        });

        if (materia.id === materiaAtiva) linha.setAttribute('aria-current', 'true');

        linha.append(
            criarElemento('span', { texto: materia.subject }),
            criarElemento('span', { classe: 'gs-badge', texto: String(totalDaMateria(materia.id)) })
        );

        fragmento.append(linha);
    });

    resumo.append(fragmento);
}

async function renderizarPublicacoes() {
    listaPublicacoes.replaceChildren();

    if (materiaAtiva === null) {
        tituloPublicacoes.textContent = 'Publicações da matéria';
        estadoPublicacoes.vazio(
            'Escolha uma matéria',
            'Clique em um dos cartões acima para ver as publicações do canal.'
        );
        return;
    }

    const nome = nomeDaMateria(materiaAtiva);
    tituloPublicacoes.textContent = `Publicações de ${nome}`;

    estadoPublicacoes.carregando(`Carregando publicações de ${nome}…`);

    try {
        const posts = await listarPosts({ idSubject: materiaAtiva });

        if (posts.length === 0) {
            estadoPublicacoes.vazio(
                `Nada em ${nome} ainda`,
                'Seja a primeira pessoa a publicar uma dúvida nessa matéria.'
            );
            return;
        }

        const reacoes = await Promise.all(
            posts.map((post) => buscarMinhaReacao('post', post.id))
        );

        const fragmento = document.createDocumentFragment();
        posts.forEach((post, indice) => {
            fragmento.append(criarCartaoDePost(post, { minhaReacao: reacoes[indice] }));
        });

        listaPublicacoes.append(fragmento);

        estadoPublicacoes.ocultar();
        inicializarKit(listaPublicacoes);
        window.GerminaStackUI?.announceToScreenReader(
            `${posts.length} ${posts.length === 1 ? 'publicação' : 'publicações'} em ${nome}.`
        );
    } catch (erro) {
        estadoPublicacoes.erro(erro.message, 'Recarregue a página para tentar de novo.');
    }
}

function selecionarMateria(id) {
    materiaAtiva = id;

    const url = id === null ? '/materias' : `/materias?id=${id}`;
    window.history.pushState({ id }, '', url);

    renderizarMaterias();
    renderizarResumo();
    renderizarPublicacoes();
}

grade.addEventListener('click', (evento) => {
    const cartao = evento.target.closest('[data-id]');
    if (!cartao) return;

    const id = Number(cartao.dataset.id);
    selecionarMateria(id === materiaAtiva ? null : id);
    tituloPublicacoes.scrollIntoView({ behavior: 'smooth', block: 'start' });
});

window.addEventListener('popstate', () => {
    materiaAtiva = lerMateriaDaUrl();
    renderizarMaterias();
    renderizarResumo();
    renderizarPublicacoes();
});

async function carregarPagina() {
    estadoMaterias.carregando('Carregando matérias…');

    try {
        [materiasCarregadas, todosOsPosts] = await Promise.all([listarMaterias(), listarPosts()]);
        materiaAtiva = lerMateriaDaUrl();

        renderizarMaterias();
        renderizarResumo();
        renderizarPublicacoes();
    } catch (erro) {
        estadoMaterias.erro(erro.message, 'Recarregue a página para tentar de novo.');
        estadoPublicacoes.ocultar();
    }
}

carregarPagina();

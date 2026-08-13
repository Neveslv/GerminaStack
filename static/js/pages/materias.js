import { listarMaterias, listarPosts, buscarMinhaReacao } from '../api.js';
import { criarCartaoDePost } from '../componentes/cartao-post.js';
import { criarElemento, criarPainelDeEstado, inicializarKit } from '../utils/dom.js';
import { tomDaMateria } from '../utils/identidade.js';

const grade = document.querySelector('#grade-materias');
const listaPublicacoes = document.querySelector('#publicacoes-da-materia');
const tituloPublicacoes = document.querySelector('#titulo-publicacoes');
const estadoMaterias = criarPainelDeEstado(document.querySelector('#estado-materias'));
const estadoPublicacoes = criarPainelDeEstado(document.querySelector('#estado-publicacoes'));
const carregarMais = criarElemento('button', { classe: 'gs-btn gs-btn-ghost', texto: 'Carregar mais', atributos: { type: 'button', hidden: 'true' } });
listaPublicacoes.after(carregarMais);

let materiasCarregadas = [];
let todosOsPosts = [];
let materiaAtiva = null;
let paginaMateria = 1;
let haMaisMateria = false;
let carregandoMateria = false;

function lerMateriaDaUrl() {
    const id = new URLSearchParams(window.location.search).get('id');
    return id ? Number(id) : null;
}

function totalDaMateria(idMateria) {
    return materiasCarregadas.find((materia) => materia.id === idMateria)?.posts_count ?? 0;
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
        paginaMateria = 1;
        const pagina = await listarPosts({ idSubject: materiaAtiva, page: paginaMateria });
        const posts = pagina.items;
        haMaisMateria = pagina.has_more;
        carregarMais.hidden = !haMaisMateria;
        carregarMais.textContent = haMaisMateria ? 'Carregar mais' : 'Todas as publicaÃ§Ãµes foram carregadas';

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
        estadoPublicacoes.erro(erro.message, 'Tentar novamente', renderizarPublicacoes);
    }
}

carregarMais.addEventListener('click', async () => {
    if (carregandoMateria || !haMaisMateria || materiaAtiva === null) return;
    carregandoMateria = true;
    try {
        const pagina = await listarPosts({ idSubject: materiaAtiva, page: paginaMateria + 1 });
        paginaMateria += 1;
        haMaisMateria = pagina.has_more;
        const reacoes = await Promise.all(pagina.items.map((post) => buscarMinhaReacao('post', post.id)));
        pagina.items.forEach((post, indice) => listaPublicacoes.append(criarCartaoDePost(post, { minhaReacao: reacoes[indice] })));
        carregarMais.hidden = !haMaisMateria;
        carregarMais.textContent = haMaisMateria ? 'Carregar mais' : 'Todas as publicaÃ§Ãµes foram carregadas';
        inicializarKit(listaPublicacoes);
    } catch (erro) {
        estadoPublicacoes.erro(erro.message, 'Tentar novamente', () => carregarMais.click());
    } finally {
        carregandoMateria = false;
    }
});

if ('IntersectionObserver' in window) {
    new IntersectionObserver((entradas) => {
        if (entradas.some((entrada) => entrada.isIntersecting)) carregarMais.click();
    }, { rootMargin: '240px' }).observe(carregarMais);
}

function selecionarMateria(id) {
    materiaAtiva = id;

    // Troca a URL sem recarregar: o botão "voltar" do navegador continua
    // andando entre as matérias que a pessoa visitou.
    const url = id === null ? '/materias' : `/materias?id=${id}`;
    window.history.pushState({ id }, '', url);

    renderizarMaterias();
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
    renderizarPublicacoes();
});

async function carregarPagina() {
    estadoMaterias.carregando('Carregando matérias…');

    try {
        const pagina = await listarPosts();
        [materiasCarregadas, todosOsPosts] = await Promise.all([listarMaterias(), Promise.resolve(pagina.items)]);
        materiaAtiva = lerMateriaDaUrl();

        renderizarMaterias();
        renderizarPublicacoes();
    } catch (erro) {
        estadoMaterias.erro(erro.message, 'Tentar novamente', carregarPagina);
        estadoPublicacoes.ocultar();
    }
}

carregarPagina();

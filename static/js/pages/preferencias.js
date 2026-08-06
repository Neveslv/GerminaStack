import { buscarPreferencias, salvarPreferencias } from '../api.js';
import {
    aplicarPreferencias,
    guardarPreferenciasLocais,
    lerPreferenciasLocais,
    PREFERENCIAS_PADRAO
} from '../utils/preferencias.js';
import { criarPainelDeEstado } from '../utils/dom.js';

const CAMPOS = ['contrast_theme', 'font_family', 'font_spacing', 'font_size'];

const formulario = document.querySelector('#form-preferencias');
const botaoSalvar = document.querySelector('#salvar');
const botaoRestaurar = document.querySelector('#restaurar');
const estado = criarPainelDeEstado(document.querySelector('#estado-preferencias'));

let salvasNoServidor = { ...PREFERENCIAS_PADRAO };

function lerFormulario() {
    const dados = new FormData(formulario);
    return Object.fromEntries(CAMPOS.map((campo) => [campo, dados.get(campo) ?? 'normal']));
}

function marcarFormulario(preferencias) {
    CAMPOS.forEach((campo) => {
        formulario.querySelectorAll(`input[name="${campo}"]`).forEach((entrada) => {
            entrada.checked = entrada.value === preferencias[campo];
        });
    });
}

function temMudancaPendente() {
    const atual = lerFormulario();
    return CAMPOS.some((campo) => atual[campo] !== salvasNoServidor[campo]);
}

function atualizarBotaoSalvar() {
    const pendente = temMudancaPendente();

    botaoSalvar.disabled = !pendente;
    botaoSalvar.textContent = pendente ? 'Salvar preferências' : 'Preferências salvas';
}

function pre_visualizar() {
    const escolha = lerFormulario();

    aplicarPreferencias(escolha);

    guardarPreferenciasLocais(escolha);

    atualizarBotaoSalvar();
}

formulario.addEventListener('change', (evento) => {
    if (!CAMPOS.includes(evento.target.name)) return;
    pre_visualizar();
});

botaoRestaurar.addEventListener('click', () => {
    marcarFormulario(PREFERENCIAS_PADRAO);
    pre_visualizar();
    window.GerminaStackUI?.announceToScreenReader(
        'Contraste, fonte, espaçamento e tamanho voltaram ao padrão. Salve para manter.',
        'polite'
    );
});

formulario.addEventListener('submit', async (evento) => {
    evento.preventDefault();

    const escolha = lerFormulario();

    botaoSalvar.disabled = true;
    estado.carregando('Salvando…');

    try {
        await salvarPreferencias(escolha);

        salvasNoServidor = escolha;
        guardarPreferenciasLocais(escolha);
        estado.ocultar();
        atualizarBotaoSalvar();

        window.GerminaStackUI?.showToast({
            title: 'Preferências salvas',
            message: 'Elas valem em todas as páginas e continuam no próximo acesso.',
            tone: 'success'
        });
    } catch (erro) {
        estado.erro(erro.message, 'A escolha continua valendo neste navegador, mas não foi salva na sua conta.');
        botaoSalvar.disabled = false;
    }
});

async function carregarPagina() {
    marcarFormulario(lerPreferenciasLocais());

    try {
        const doServidor = await buscarPreferencias();

        salvasNoServidor = Object.fromEntries(CAMPOS.map((campo) => [campo, doServidor[campo] ?? 'normal']));

        marcarFormulario(salvasNoServidor);
        aplicarPreferencias(salvasNoServidor);
        guardarPreferenciasLocais(salvasNoServidor);
    } catch {
        estado.erro(
            'Não foi possível ler suas preferências salvas',
            'Você pode escolher mesmo assim: vale neste navegador até o servidor responder.'
        );
    }

    atualizarBotaoSalvar();
}

carregarPagina();

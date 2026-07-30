/**
 * Página de acessibilidade: escolhe contrast_theme e font_family.
 *
 * A escolha vale na hora, antes de salvar. É proposital: comparar contraste
 * de cabeça não funciona: quem precisa do recurso precisa VER o resultado
 * para decidir. O botão de salvar só grava o que já está na tela.
 */

import { buscarPreferencias, salvarPreferencias } from '../api.js';
import {
    aplicarPreferencias,
    guardarPreferenciasLocais,
    lerPreferenciasLocais,
    PREFERENCIAS_PADRAO
} from '../utils/preferencias.js';
import { criarPainelDeEstado } from '../utils/dom.js';

const formulario = document.querySelector('#form-preferencias');
const botaoSalvar = document.querySelector('#salvar');
const botaoRestaurar = document.querySelector('#restaurar');
const estado = criarPainelDeEstado(document.querySelector('#estado-preferencias'));

/** Preferências salvas no servidor — usadas para saber se há mudança pendente. */
let salvasNoServidor = { ...PREFERENCIAS_PADRAO };

function lerFormulario() {
    const dados = new FormData(formulario);
    return {
        contrast_theme: dados.get('contrast_theme') ?? 'normal',
        font_family: dados.get('font_family') ?? 'normal'
    };
}

function marcarFormulario({ contrast_theme, font_family }) {
    formulario.querySelectorAll('input[name="contrast_theme"]').forEach((entrada) => {
        entrada.checked = entrada.value === contrast_theme;
    });

    formulario.querySelectorAll('input[name="font_family"]').forEach((entrada) => {
        entrada.checked = entrada.value === font_family;
    });
}

function temMudancaPendente() {
    const atual = lerFormulario();
    return (
        atual.contrast_theme !== salvasNoServidor.contrast_theme ||
        atual.font_family !== salvasNoServidor.font_family
    );
}

function atualizarBotaoSalvar() {
    const pendente = temMudancaPendente();

    botaoSalvar.disabled = !pendente;
    botaoSalvar.textContent = pendente ? 'Salvar preferências' : 'Preferências salvas';
}

/** Pré-visualiza sem gravar: aplica no <html> e espelha no localStorage. */
function pre_visualizar() {
    const escolha = lerFormulario();

    aplicarPreferencias(escolha);

    // O espelho local é atualizado junto para a pessoa poder navegar para
    // outra página e continuar vendo o que escolheu, mesmo sem ter salvo.
    guardarPreferenciasLocais(escolha);

    atualizarBotaoSalvar();
}

formulario.addEventListener('change', (evento) => {
    if (evento.target.name !== 'contrast_theme' && evento.target.name !== 'font_family') return;
    pre_visualizar();
});

botaoRestaurar.addEventListener('click', () => {
    marcarFormulario(PREFERENCIAS_PADRAO);
    pre_visualizar();
    window.GerminaStackUI?.announceToScreenReader(
        'Contraste e fonte voltaram ao padrão. Salve para manter.',
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
    // Começa pelo espelho local para o formulário já abrir marcado igual ao
    // que a tela está mostrando — o tema-inicial.js aplicou esse valor.
    marcarFormulario(lerPreferenciasLocais());

    try {
        const doServidor = await buscarPreferencias();

        salvasNoServidor = {
            contrast_theme: doServidor.contrast_theme,
            font_family: doServidor.font_family
        };

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

/**
 * Página de acessibilidade: escolhe contrast_theme, font_family, font_spacing
 * e font_size — os quatro campos da tabela `preferences`.
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

/** Os quatro campos de `preferences`, na ordem em que aparecem no formulário. */
const CAMPOS = ['contrast_theme', 'font_family', 'font_spacing', 'font_size'];

const formulario = document.querySelector('#form-preferencias');
const botaoSalvar = document.querySelector('#salvar');
const botaoRestaurar = document.querySelector('#restaurar');
const estado = criarPainelDeEstado(document.querySelector('#estado-preferencias'));

/** Preferências salvas no servidor — usadas para saber se há mudança pendente. */
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
    // Começa pelo espelho local para o formulário já abrir marcado igual ao
    // que a tela está mostrando — o tema-inicial.js aplicou esse valor.
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

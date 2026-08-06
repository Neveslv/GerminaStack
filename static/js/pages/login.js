/** Página de login: valida credenciais e envia à API. */

import { entrar } from '../api.js';
import { criarPainelDeEstado } from '../utils/dom.js';

const formulario = document.querySelector('#form-login');
const estado = criarPainelDeEstado(document.querySelector('#estado-login'));

const REGRAS = [
    { campo: 'usuario', erro: 'erro-usuario', mensagem: 'Informe seu usuário.' },
    { campo: 'senha', erro: 'erro-senha', mensagem: 'Informe sua senha.' }
];

function limparErros() {
    REGRAS.forEach(({ campo, erro }) => {
        document.querySelector(`#${erro}`).textContent = '';
        document.querySelector(`#${campo}`).setAttribute('aria-invalid', 'false');
    });
}

function validar() {
    let primeiroInvalido = null;

    REGRAS.forEach(({ campo, erro, mensagem }) => {
        const elemento = document.querySelector(`#${campo}`);
        if (elemento.value.trim()) return;

        document.querySelector(`#${erro}`).textContent = mensagem;
        elemento.setAttribute('aria-invalid', 'true');
        if (!primeiroInvalido) primeiroInvalido = elemento;
    });

    return primeiroInvalido;
}

formulario.addEventListener('submit', async (evento) => {
    evento.preventDefault();
    limparErros();

    const invalido = validar();
    if (invalido) {
        estado.erro('Revise os campos destacados', 'Os campos com erro estão marcados abaixo.');
        invalido.focus();
        return;
    }

    const botao = formulario.querySelector('button[type="submit"]');
    botao.disabled = true;
    estado.carregando('Entrando…');

    try {
        await entrar({
            username: document.querySelector('#usuario').value.trim(),
            password: document.querySelector('#senha').value
        });

        window.GerminaStackUI?.showToast({
            title: 'Bem-vindo de volta',
            message: 'Você entrou na sua conta do Germinare.',
            tone: 'success'
        });

        window.setTimeout(() => window.location.assign('/'), 700);
    } catch (erro) {
        estado.erro(erro.message, 'Tente de novo em alguns instantes.');
        botao.disabled = false;
        document.querySelector('#senha').value = '';
    }
});
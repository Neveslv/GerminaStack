/** Página de login: valida credenciais e envia à API. */

import { completarLogin, entrar } from '../api.js';
import { criarPainelDeEstado } from '../utils/dom.js';

const formulario = document.querySelector('#form-login');
const estado = criarPainelDeEstado(document.querySelector('#estado-login'));

const REGRAS = [
    { campo: 'email', erro: 'erro-email', mensagem: 'Informe seu e-mail institucional.' },
    { campo: 'senha', erro: 'erro-senha', mensagem: 'Informe sua senha.' }
];

let desafio;

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
    document.querySelector('#erro-codigo').textContent = '';
    document.querySelector('#codigo').setAttribute('aria-invalid', 'false');

    const invalido = desafio
        ? (/^[0-9]{6}$/.test(document.querySelector('#codigo').value.trim()) ? null : document.querySelector('#codigo'))
        : validar();
    if (invalido) {
        if (desafio) document.querySelector('#erro-codigo').textContent = 'Informe o código de 6 dígitos.';
        estado.erro('Revise os campos destacados', 'Os campos com erro estão marcados abaixo.');
        invalido.setAttribute('aria-invalid', 'true');
        invalido.focus();
        return;
    }

    const botao = formulario.querySelector('button[type="submit"]');
    botao.disabled = true;
    estado.carregando('Entrando…');

    try {
        if (!desafio) {
            const resposta = await entrar({
                email: document.querySelector('#email').value.trim(),
                password: document.querySelector('#senha').value
            });
            desafio = resposta.challenge_id;
            document.querySelector('#grupo-codigo').hidden = false;
            document.querySelector('#email').disabled = true;
            document.querySelector('#senha').disabled = true;
            document.querySelector('#codigo').focus();
            botao.textContent = 'Confirmar código';
            botao.disabled = false;
            estado.carregando('Código enviado para seu e-mail.');
            return;
        }

        await completarLogin({
            challenge_id: desafio,
            code: document.querySelector('#codigo').value.trim()
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
        if (desafio) document.querySelector('#codigo').value = '';
        else document.querySelector('#senha').value = '';
    }
});

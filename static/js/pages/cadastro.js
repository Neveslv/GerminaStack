/** Página de cadastro: valida os dados e cria a conta na tabela `users`. */

import { cadastrar, listarAnos } from '../api.js';
import { criarElemento, criarPainelDeEstado } from '../utils/dom.js';

const formulario = document.querySelector('#form-cadastro');
const seletorAno = document.querySelector('#ano');
const estado = criarPainelDeEstado(document.querySelector('#estado-cadastro'));

const REGRAS = [
    {
        campo: 'email',
        erro: 'erro-email',
        validar: (valor) => {
            if (!valor) return 'Informe seu e-mail.';
            if (!/^[a-z]+\.[a-z]+@institutojef\.org\.br$/.test(valor)) {
                return 'Use seu e-mail institucional no formato nome.sobrenome@institutojef.org.br.';
            }
            return '';
        }
    },
    {
        campo: 'ano',
        erro: 'erro-ano',
        validar: (valor) => (valor ? '' : 'Selecione sua turma.')
    },
    {
        campo: 'senha',
        erro: 'erro-senha',
        validar: (valor) => {
            if (!valor) return 'Crie uma senha.';
            if (valor.length < 8) return 'A senha precisa ter pelo menos 8 caracteres.';
            return '';
        }
    },
    {
        campo: 'confirmar-senha',
        erro: 'erro-confirmar-senha',
        validar: (valor) => {
            if (!valor) return 'Repita a senha.';
            if (valor !== document.querySelector('#senha').value) return 'As senhas não são iguais.';
            return '';
        }
    }
];

function limparErros() {
    REGRAS.forEach(({ campo, erro }) => {
        document.querySelector(`#${erro}`).textContent = '';
        document.querySelector(`#${campo}`).setAttribute('aria-invalid', 'false');
    });
}

/** Valida tudo e devolve o primeiro campo inválido, ou null. */
function validar() {
    let primeiroInvalido = null;

    REGRAS.forEach(({ campo, erro, validar: checar }) => {
        const elemento = document.querySelector(`#${campo}`);

        // A senha não passa por trim: espaço no começo ou no fim é parte dela.
        const valor = campo.includes('senha') ? elemento.value : elemento.value.trim();
        const mensagem = checar(valor);
        if (!mensagem) return;

        document.querySelector(`#${erro}`).textContent = mensagem;
        elemento.setAttribute('aria-invalid', 'true');
        if (!primeiroInvalido) primeiroInvalido = elemento;
    });

    return primeiroInvalido;
}

async function carregarAnos() {
    try {
        const anos = await listarAnos();
        const fragmento = document.createDocumentFragment();

        anos.forEach((ano) => {
            fragmento.append(
                criarElemento('option', { texto: ano.year, atributos: { value: String(ano.id) } })
            );
        });

        seletorAno.append(fragmento);
    } catch {
        estado.erro('Não foi possível carregar as turmas', 'Recarregue a página para tentar de novo.');
    }
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
    estado.carregando('Criando sua conta…');

    try {
        await cadastrar({
            email: document.querySelector('#email').value.trim(),
            year_id: Number(seletorAno.value),
            password: document.querySelector('#senha').value,
            password_confirmation: document.querySelector('#confirmar-senha').value
        });

        window.GerminaStackUI?.showToast({
            title: 'Conta criada',
            message: 'Agora é só entrar com seu e-mail e senha.',
            tone: 'success'
        });

        window.setTimeout(() => window.location.assign('/login'), 900);
    } catch (erro) {
        estado.erro(erro.message, 'Confira os dados e tente de novo.');
        botao.disabled = false;

        if (erro.status === 409) {
            const email = document.querySelector('#email');
            document.querySelector('#erro-email').textContent = 'Esse e-mail já está em uso.';
            email.setAttribute('aria-invalid', 'true');
            email.focus();
        }
    }
});

carregarAnos();

/** Página de administração: cria novos perfis de usuário na tabela `users`. */

import { cadastrar, listarAnos } from '../api.js';
import { criarElemento, criarPainelDeEstado } from '../utils/dom.js';

const formulario = document.querySelector('#form-admin');
const seletorAno = document.querySelector('#ano');
const estado = criarPainelDeEstado(document.querySelector('#estado-admin'));

/**
 * O `username` vira menção com @ nos posts, e o trigger `notify_mentions` do
 * banco captura menções com o padrão `@([a-zA-Z0-9_]+)`. Mesma regra usada em
 * cadastro.js, para não abrir uma segunda porta de criação de conta com regras
 * diferentes.
 */
const PADRAO_USUARIO = /^[a-zA-Z0-9_]+$/;

const REGRAS = [
    {
        campo: 'nome',
        erro: 'erro-nome',
        validar: (valor) => (valor ? '' : 'Informe o nome completo.')
    },
    {
        campo: 'usuario',
        erro: 'erro-usuario',
        validar: (valor) => {
            if (!valor) return 'Escolha um nome de usuário.';
            if (valor.length < 3) return 'O usuário precisa ter pelo menos 3 caracteres.';
            if (!PADRAO_USUARIO.test(valor)) {
                return 'Use só letras sem acento, números e traço baixo (_).';
            }
            return '';
        }
    },
    {
        campo: 'email',
        erro: 'erro-email',
        validar: (valor) => {
            if (!valor) return 'Informe o e-mail.';
            if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(valor)) return 'Esse e-mail não parece válido.';
            return '';
        }
    },
    {
        campo: 'ano',
        erro: 'erro-ano',
        validar: (valor) => (valor ? '' : 'Selecione a turma.')
    },
    {
        campo: 'senha',
        erro: 'erro-senha',
        validar: (valor) => {
            if (!valor) return 'Crie uma senha provisória.';
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
    estado.carregando('Criando perfil…');

    try {
        const nome = document.querySelector('#nome').value.trim();

        await cadastrar({
            name: nome,
            username: document.querySelector('#usuario').value.trim(),
            email: document.querySelector('#email').value.trim(),
            id_year: Number(seletorAno.value),
            password: document.querySelector('#senha').value
        });

        window.GerminaStackUI?.showToast({
            title: 'Perfil criado',
            message: `A conta de ${nome} foi criada.`,
            tone: 'success'
        });

        formulario.reset();
        estado.ocultar();
        document.querySelector('#nome').focus();
    } catch (erro) {
        estado.erro(erro.message, 'Confira os dados e tente de novo.');

        // 409 é o conflito de UNIQUE em username/email: dá para apontar o campo.
        if (erro.status === 409) {
            const usuario = document.querySelector('#usuario');
            document.querySelector('#erro-usuario').textContent =
                'Esse usuário ou e-mail já está em uso.';
            usuario.setAttribute('aria-invalid', 'true');
            usuario.focus();
        }
    } finally {
        botao.disabled = false;
    }
});

carregarAnos();

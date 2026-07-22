/** Página de nova publicação: valida o formulário e envia à API. */

import { criarPost, listarMaterias } from '../api.js';
import { criarElemento, criarPainelDeEstado } from '../utils/dom.js';

const formulario = document.querySelector('#form-publicar');
const seletorMateria = document.querySelector('#materia');
const estado = criarPainelDeEstado(document.querySelector('#estado-formulario'));

const REGRAS = [
    { campo: 'materia', erro: 'erro-materia', mensagem: 'Escolha a matéria da publicação.' },
    { campo: 'titulo', erro: 'erro-titulo', mensagem: 'Escreva um título para a publicação.' },
    { campo: 'conteudo-post', erro: 'erro-conteudo', mensagem: 'Escreva o conteúdo da publicação.' }
];

function limparErros() {
    REGRAS.forEach(({ campo, erro }) => {
        document.querySelector(`#${erro}`).textContent = '';
        document.querySelector(`#${campo}`).setAttribute('aria-invalid', 'false');
    });
}

/** Valida os campos e devolve o primeiro elemento inválido, ou null. */
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

async function carregarMaterias() {
    try {
        const materias = await listarMaterias();
        const fragmento = document.createDocumentFragment();

        materias.forEach((materia) => {
            fragmento.append(
                criarElemento('option', { texto: materia.subject, atributos: { value: String(materia.id) } })
            );
        });

        seletorMateria.append(fragmento);
    } catch {
        estado.erro('Não foi possível carregar as matérias', 'Recarregue a página para tentar de novo.');
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
    estado.carregando('Publicando…');

    try {
        const post = await criarPost({
            id_subject: Number(seletorMateria.value),
            title: document.querySelector('#titulo').value.trim(),
            content: document.querySelector('#conteudo-post').value.trim()
        });

        window.GerminaStackUI?.showToast({
            title: 'Publicado',
            message: 'Sua publicação já está no feed.',
            tone: 'success'
        });

        window.setTimeout(() => window.location.assign(`/post?id=${post.id}`), 700);
    } catch (erro) {
        estado.erro(erro.message, 'Tente de novo em alguns instantes.');
        botao.disabled = false;
    }
});

carregarMaterias();
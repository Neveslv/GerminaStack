/** Página de nova publicação: valida o formulário e envia à API. */

import { criarPost, listarMaterias } from '../api.js';
import { criarElemento, criarPainelDeEstado } from '../utils/dom.js';

const formulario = document.querySelector('#form-publicar');
const seletorMateria = document.querySelector('#materia');
const campoImagem = document.querySelector('#imagem');
const campoDescricaoImagem = document.querySelector('#descricao-imagem');
const erroImagem = document.querySelector('#erro-imagem');
const erroDescricaoImagem = document.querySelector('#erro-descricao-imagem');
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

    erroImagem.textContent = '';
    campoImagem.setAttribute('aria-invalid', 'false');
    erroDescricaoImagem.textContent = '';
    campoDescricaoImagem.setAttribute('aria-invalid', 'false');
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

    // A coluna image_url aceita nulo, então o campo só é validado se foi preenchido.
    const imagem = campoImagem.value.trim();
    const descricaoImagem = campoDescricaoImagem.value.trim();
    if (imagem && !/^https?:\/\/.+/i.test(imagem)) {
        erroImagem.textContent = 'O endereço precisa começar com http:// ou https://.';
        campoImagem.setAttribute('aria-invalid', 'true');
        if (!primeiroInvalido) primeiroInvalido = campoImagem;
    }
    if (imagem && !descricaoImagem) {
        erroDescricaoImagem.textContent = 'Descreva a imagem para leitores de tela.';
        campoDescricaoImagem.setAttribute('aria-invalid', 'true');
        if (!primeiroInvalido) primeiroInvalido = campoDescricaoImagem;
    }
    if (!imagem && descricaoImagem) {
        erroDescricaoImagem.textContent = 'Inclua uma imagem ou remova a descrição.';
        campoDescricaoImagem.setAttribute('aria-invalid', 'true');
        if (!primeiroInvalido) primeiroInvalido = campoDescricaoImagem;
    }

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
        const imagem = campoImagem.value.trim();
        const descricaoImagem = campoDescricaoImagem.value.trim();

        const titulo = document.querySelector('#titulo').value.trim();
        const post = await criarPost({
            subject_id: Number(seletorMateria.value),
            title: titulo,
            content: document.querySelector('#conteudo-post').value.trim(),
            image_url: imagem || null,
            image_description: imagem ? descricaoImagem : null
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

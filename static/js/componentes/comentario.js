/**
 * Comentários e respostas da página de publicação.
 *
 * Os dois níveis espelham as tabelas `comments` e `comments_on_comments`.
 * O banco não tem terceiro nível, então uma resposta não ganha botão de
 * responder — o front não oferece o que o banco não guarda.
 */

import { criarElemento } from '../utils/dom.js';
import { aplicarImagemNoAvatar, corDoAutor, inicialDoNome } from '../utils/identidade.js';
import { formatarDataRelativa, formatarDataCompleta } from '../utils/data.js';
import { criarReacoes } from './reacoes.js';
import { criarComentario as enviarComentario, buscarMinhaReacao } from '../api.js';

function criarAvatar(autor) {
    const avatar = criarElemento('span', {
        classe: 'gs-avatar-xs',
        texto: inicialDoNome(autor.name),
        atributos: { 'aria-hidden': 'true' }
    });
    if (!aplicarImagemNoAvatar(avatar, autor)) avatar.style.background = corDoAutor(autor.id);
    return avatar;
}

function criarBolha(item, classeBolha) {
    const meta = criarElemento('div', { classe: 'gs-meta' });
    meta.append(
        criarElemento('strong', { texto: item.author.name }),
        criarElemento('time', {
            texto: formatarDataRelativa(item.created_at),
            atributos: {
                datetime: item.created_at,
                title: formatarDataCompleta(item.created_at)
            }
        })
    );

    const bolha = criarElemento('div', { classe: classeBolha });
    bolha.append(meta, criarElemento('p', { texto: item.content }));
    return bolha;
}

/** Monta uma resposta (tabela `comments_on_comments`). */
export function criarResposta(resposta, minhaReacao = null) {
    const conteudo = criarElemento('div', { classe: 'gs-reply-content' });
    conteudo.append(criarBolha(resposta, 'gs-reply-bubble'));

    const acoes = criarElemento('div', { classe: 'gs-comment-actions gs-cluster' });
    acoes.append(
        criarReacoes({
            tipo: 'comment_on_comment',
            id: resposta.id,
            likes: resposta.likes,
            dislikes: resposta.dislikes,
            minhaReacao
        })
    );
    conteudo.append(acoes);

    const linha = criarElemento('div', { classe: 'gs-reply-row' });
    linha.append(criarAvatar(resposta.author), conteudo);

    const item = criarElemento('div', {
        classe: 'gs-reply',
        atributos: { 'data-resposta-id': String(resposta.id) }
    });
    item.append(linha);
    return item;
}

/**
 * Formulário de resposta que abre embaixo de um comentário.
 * Fica escondido até a pessoa pedir, para a thread não virar uma parede
 * de campos de texto para quem navega por teclado ou leitor de tela.
 */
function criarFormularioDeResposta(idComentario, aoEnviar) {
    const formulario = criarElemento('form', {
        classe: 'gs-inline-form',
        atributos: { hidden: '' }
    });

    const idCampo = `resposta-${idComentario}`;

    const rotulo = criarElemento('label', {
        classe: 'gs-sr-only',
        texto: 'Escreva sua resposta',
        atributos: { for: idCampo }
    });

    const campo = criarElemento('input', {
        classe: 'gs-input',
        atributos: {
            type: 'text',
            id: idCampo,
            name: 'resposta',
            placeholder: 'Responder este comentário',
            autocomplete: 'off',
            maxlength: '500'
        }
    });

    const enviar = criarElemento('button', {
        classe: 'gs-btn gs-btn-primary',
        texto: 'Responder',
        atributos: { type: 'submit' }
    });

    formulario.append(rotulo, campo, enviar);

    formulario.addEventListener('submit', async (evento) => {
        evento.preventDefault();

        const texto = campo.value.trim();
        if (!texto) {
            campo.focus();
            return;
        }

        enviar.disabled = true;

        try {
            const resposta = await enviarComentario({
                tipo: 'comment_on_comment',
                idPai: idComentario,
                content: texto
            });

            aoEnviar(resposta);
            campo.value = '';

            window.GerminaStackUI?.announceToScreenReader('Resposta publicada.', 'polite');
        } catch (erro) {
            window.GerminaStackUI?.showToast({
                title: 'Não foi possível responder',
                message: erro.message,
                tone: 'danger'
            });
        } finally {
            enviar.disabled = false;
        }
    });

    return { formulario, campo };
}

/** Monta um comentário (tabela `comments`) com suas respostas. */
export function criarComentario(comentario, minhaReacao = null) {
    const item = criarElemento('div', {
        classe: 'gs-comment',
        atributos: { 'data-comentario-id': String(comentario.id) }
    });

    const conteudo = criarElemento('div', { classe: 'gs-comment-content' });
    conteudo.append(criarBolha(comentario, 'gs-comment-bubble'));

    const acoes = criarElemento('div', { classe: 'gs-comment-actions gs-cluster' });
    acoes.append(
        criarReacoes({
            tipo: 'comment',
            id: comentario.id,
            likes: comentario.likes,
            dislikes: comentario.dislikes,
            minhaReacao
        })
    );

    const botaoResponder = criarElemento('button', {
        classe: 'gs-action',
        texto: 'Responder',
        atributos: {
            type: 'button',
            'aria-expanded': 'false',
            'aria-label': `Responder o comentário de ${comentario.author.name}`
        }
    });

    acoes.append(botaoResponder);
    conteudo.append(acoes);

    const linha = criarElemento('div', { classe: 'gs-comment-row' });
    linha.append(criarAvatar(comentario.author), conteudo);
    item.append(linha);

    const respostas = criarElemento('div');
    comentario.replies?.forEach((resposta) => respostas.append(criarResposta(resposta)));
    item.append(respostas);

    const { formulario, campo } = criarFormularioDeResposta(comentario.id, (resposta) => {
        respostas.append(criarResposta(resposta));
    });

    item.append(formulario);

    botaoResponder.addEventListener('click', () => {
        const aberto = botaoResponder.getAttribute('aria-expanded') === 'true';

        botaoResponder.setAttribute('aria-expanded', String(!aberto));
        formulario.hidden = aberto;

        // Levar o foco para o campo recém-aberto evita que quem usa teclado
        // precise tabular de volta por toda a thread para alcançá-lo.
        if (!aberto) campo.focus();
    });

    return item;
}

/**
 * Monta a thread inteira já com a reação do usuário em cada item.
 * As reações são buscadas em paralelo para a thread não abrir em cascata.
 */
export async function criarThread(comentarios) {
    const fragmento = document.createDocumentFragment();

    const reacoes = await Promise.all(
        comentarios.map((comentario) => buscarMinhaReacao('comment', comentario.id))
    );

    comentarios.forEach((comentario, indice) => {
        fragmento.append(criarComentario(comentario, reacoes[indice]));
    });

    return fragmento;
}

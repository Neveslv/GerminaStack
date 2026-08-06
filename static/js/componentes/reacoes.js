/**
 * Botões de curtir e não curtir, ligados à tabela `reactions`.
 *
 * Não usa o `data-gs-like` do kit de propósito. Aquele atributo soma 1 no
 * contador da tela e para por aí — serve para demonstração. Aqui o número
 * precisa vir do banco, então a regra de negócio fica fora do kit, como o
 * próprio README do kit recomenda. O que continua vindo do kit é a aparência:
 * a classe `gs-action` e o estado `data-state="on"`, que já têm estilo pronto.
 */

import { criarElemento } from '../utils/dom.js';
import { reagir } from '../api.js';

const DESENHOS = {
    like: 'M7 22V10l5-8 1 1v6h6a2 2 0 0 1 2 2l-2 8a2 2 0 0 1-2 2H7Zm0 0H3V10h4',
    dislike: 'M17 2v12l-5 8-1-1v-6H5a2 2 0 0 1-2-2l2-8a2 2 0 0 1 2-2h10Zm0 0h4v12h-4'
};

function icone(nome) {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('width', '16');
    svg.setAttribute('height', '16');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('stroke', 'currentColor');
    svg.setAttribute('stroke-width', '2');
    svg.setAttribute('stroke-linecap', 'round');
    svg.setAttribute('stroke-linejoin', 'round');
    svg.setAttribute('aria-hidden', 'true');

    const caminho = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    caminho.setAttribute('d', DESENHOS[nome]);
    svg.append(caminho);

    return svg;
}

const ROTULOS = {
    like: { acao: 'Curtir', desfazer: 'Remover curtida', plural: 'curtidas', singular: 'curtida' },
    dislike: { acao: 'Não curtir', desfazer: 'Remover descurtida', plural: 'descurtidas', singular: 'descurtida' }
};

function descrever(reacao, total, ativo) {
    const rotulo = ROTULOS[reacao];
    const contagem = `${total} ${total === 1 ? rotulo.singular : rotulo.plural}`;
    return `${ativo ? rotulo.desfazer : rotulo.acao}. ${contagem}.`;
}

function criarBotao(reacao, total, ativo) {
    const botao = criarElemento('button', {
        classe: 'gs-action',
        atributos: {
            type: 'button',
            'data-reacao': reacao,
            'aria-pressed': String(ativo)
        }
    });

    botao.dataset.state = ativo ? 'on' : 'off';
    botao.dataset.total = String(total);
    if (reacao === 'like' && ativo) botao.classList.add('is-danger');

    const numero = criarElemento('span', { texto: String(total) });
    numero.dataset.numero = '';

    botao.append(icone(reacao), numero);
    botao.setAttribute('aria-label', descrever(reacao, total, ativo));

    return botao;
}

function atualizarBotao(botao, total, ativo) {
    const reacao = botao.dataset.reacao;

    botao.dataset.state = ativo ? 'on' : 'off';
    botao.dataset.total = String(total);
    botao.setAttribute('aria-pressed', String(ativo));
    botao.setAttribute('aria-label', descrever(reacao, total, ativo));
    botao.querySelector('[data-numero]').textContent = String(total);
    botao.classList.toggle('is-danger', reacao === 'like' && ativo);
}

/**
 * Monta o par de botões de reação de um post, comentário ou resposta.
 *
 * @param {{ tipo: 'post'|'comment'|'comment_on_comment', id: number,
 *           likes: number, dislikes: number, minhaReacao: string|null }} alvo
 * @returns {HTMLElement} um <div class="gs-cluster"> pronto para ser inserido
 */
export function criarReacoes({ tipo, id, likes, dislikes, minhaReacao }) {
    const grupo = criarElemento('div', {
        classe: 'gs-cluster',
        atributos: { role: 'group', 'aria-label': 'Reações' }
    });

    const botaoLike = criarBotao('like', likes, minhaReacao === 'like');
    const botaoDislike = criarBotao('dislike', dislikes, minhaReacao === 'dislike');

    grupo.append(botaoLike, botaoDislike);

    grupo.addEventListener('click', async (evento) => {
        const botao = evento.target.closest('[data-reacao]');
        if (!botao) return;

        const escolhida = botao.dataset.reacao;

        // Trava os dois botões: um duplo-clique rápido mandaria duas
        // alternâncias e o contador da tela sairia do valor do banco.
        botaoLike.disabled = true;
        botaoDislike.disabled = true;

        try {
            const deltas = await reagir({ tipo, id, reacao: escolhida });

            const novoLikes = Math.max(0, Number(botaoLike.dataset.total) + deltas.likes);
            const novoDislikes = Math.max(0, Number(botaoDislike.dataset.total) + deltas.dislikes);

            atualizarBotao(botaoLike, novoLikes, deltas.reacao === 'like');
            atualizarBotao(botaoDislike, novoDislikes, deltas.reacao === 'dislike');

            window.GerminaStackUI?.announceToScreenReader(
                deltas.reacao === null
                    ? 'Reação removida.'
                    : `${deltas.reacao === 'like' ? 'Curtido' : 'Não curtido'}.`
            );
        } catch (erro) {
            window.GerminaStackUI?.showToast({
                title: 'Não foi possível reagir',
                message: erro.message,
                tone: 'danger'
            });
        } finally {
            botaoLike.disabled = false;
            botaoDislike.disabled = false;
        }
    });

    return grupo;
}

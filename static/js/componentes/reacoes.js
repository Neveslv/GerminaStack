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

        botaoLike.disabled = true;
        botaoDislike.disabled = true;

        try {
            const deltas = await reagir({ tipo, id, reacao: escolhida });

            const novoLikes = deltas.absolutos
                ? deltas.likes
                : Math.max(0, Number(botaoLike.dataset.total) + deltas.likes);
            const novoDislikes = deltas.absolutos
                ? deltas.dislikes
                : Math.max(0, Number(botaoDislike.dataset.total) + deltas.dislikes);

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

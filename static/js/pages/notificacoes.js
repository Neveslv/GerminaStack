/**
 * Página de notificações.
 *
 * O texto de cada linha vem pronto do banco, no campo `text_show`, escrito
 * pelo trigger `notify_mentions`. O front não remonta a frase: se a regra da
 * menção mudar, ela muda em um lugar só.
 */

import { listarNotificacoes, marcarNotificacoesComoLidas } from '../api.js';
import { criarElemento, criarPainelDeEstado } from '../utils/dom.js';
import { formatarDataRelativa, formatarDataCompleta } from '../utils/data.js';

const lista = document.querySelector('#lista-notificacoes');
const contagem = document.querySelector('#contagem-notificacoes');
const botaoMarcar = document.querySelector('#marcar-lidas');
const estado = criarPainelDeEstado(document.querySelector('#estado-notificacoes'));

let notificacoes = [];

function criarLinha(notificacao) {
    const item = criarElemento('li');

    const cartao = criarElemento('article', {
        classe: `gs-card${notificacao.is_read ? '' : ' gs-banner'}`
    });

    const cabecalho = criarElemento('div', { classe: 'gs-meta' });

    // "Nova" é texto, não só uma cor de fundo: quem não distingue as cores
    // precisa conseguir saber o que já leu (WCAG 1.4.1).
    if (!notificacao.is_read) {
        cabecalho.append(criarElemento('span', { classe: 'gs-badge', texto: 'Nova' }));
    }

    cabecalho.append(
        criarElemento('time', {
            texto: formatarDataRelativa(notificacao.created_at),
            atributos: {
                datetime: notificacao.created_at,
                title: formatarDataCompleta(notificacao.created_at)
            }
        })
    );

    cartao.append(cabecalho, criarElemento('p', { texto: notificacao.text_show }));

    if (notificacao.id_post) {
        cartao.append(
            criarElemento('a', {
                classe: 'gs-btn gs-btn-ghost',
                texto: 'Abrir publicação',
                atributos: {
                    href: `/post?id=${notificacao.id_post}`,
                    'aria-label': `Abrir a publicação desta notificação de ${formatarDataRelativa(notificacao.created_at)}`
                }
            })
        );
    }

    item.append(cartao);
    return item;
}

function atualizarContagem() {
    const naoLidas = notificacoes.filter((notificacao) => !notificacao.is_read).length;

    contagem.textContent = naoLidas === 0
        ? `Tudo lido. ${notificacoes.length} ${notificacoes.length === 1 ? 'notificação' : 'notificações'} no total.`
        : `${naoLidas} não ${naoLidas === 1 ? 'lida' : 'lidas'} de ${notificacoes.length}.`;

    botaoMarcar.disabled = naoLidas === 0;
}

function renderizar() {
    lista.replaceChildren();

    if (notificacoes.length === 0) {
        estado.vazio(
            'Nenhuma notificação',
            'Quando alguém te marcar com @ em uma publicação, ela aparece aqui.'
        );
        contagem.textContent = 'Nenhuma notificação.';
        botaoMarcar.disabled = true;
        return;
    }

    const fragmento = document.createDocumentFragment();
    notificacoes.forEach((notificacao) => fragmento.append(criarLinha(notificacao)));
    lista.append(fragmento);

    estado.ocultar();
    atualizarContagem();
}

botaoMarcar.addEventListener('click', async () => {
    botaoMarcar.disabled = true;

    try {
        await marcarNotificacoesComoLidas();

        notificacoes = notificacoes.map((notificacao) => ({ ...notificacao, is_read: true }));
        renderizar();

        // O contador do topo é montado no carregamento da página; some com ele
        // aqui para a tela não continuar prometendo notificações que já foram lidas.
        const marcador = document.querySelector('[data-nao-lidas]');
        if (marcador) marcador.hidden = true;

        window.GerminaStackUI?.showToast({
            title: 'Notificações lidas',
            message: 'Todas foram marcadas como lidas.',
            tone: 'success'
        });
    } catch (erro) {
        botaoMarcar.disabled = false;
        window.GerminaStackUI?.showToast({
            title: 'Não foi possível marcar',
            message: erro.message,
            tone: 'danger'
        });
    }
});

async function carregarPagina() {
    estado.carregando('Carregando notificações…');

    try {
        notificacoes = await listarNotificacoes();
        renderizar();
    } catch (erro) {
        estado.erro(erro.message, 'Recarregue a página para tentar de novo.');
        contagem.textContent = '—';
        botaoMarcar.disabled = true;
    }
}

carregarPagina();

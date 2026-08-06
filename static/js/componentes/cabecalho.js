/**
 * Comportamento do topo, comum a todas as páginas (exceto login/cadastro).
 *
 * Faz três coisas:
 *   1. sincroniza as preferências de acessibilidade com a tabela `preferences`;
 *   2. busca as notificações uma única vez e usa o mesmo resultado tanto para
 *      o badge de não lidas quanto para o conteúdo do popover;
 *   3. liga o botão "marcar todas como lidas" dentro do popover.
 *
 * O popover reaproveita o mecanismo de menu contextual (`data-gs-menu`) que o
 * cartão de post já usa no botão "⋮" — não é uma página própria porque o
 * conteúdo (poucas linhas de texto) não justifica uma navegação inteira.
 *
 * Como o wrapper `.gs-menu` já está no HTML desde o carregamento da página, o
 * `GerminaStackUI.init(document.body)` do próprio kit (rodado antes destes
 * módulos, ver web/README.md) já cuidou de marcar `aria-haspopup`, esconder o
 * painel e ligar o clique de abrir/fechar. Chamar `inicializarKit()` de novo
 * aqui registraria um segundo listener de clique só neste trecho, e o clique
 * abriria e fecharia o menu na mesma interação — por isso esta função só
 * preenche conteúdo, nunca re-inicializa o kit.
 */

import {
    buscarPreferencias,
    listarNotificacoes,
    marcarNotificacoesComoLidas
} from '../api.js';
import {
    aplicarPreferencias,
    guardarPreferenciasLocais,
    lerPreferenciasLocais
} from '../utils/preferencias.js';
import { criarElemento } from '../utils/dom.js';
import { formatarDataRelativa, formatarDataCompleta } from '../utils/data.js';

async function sincronizarPreferencias() {
    try {
        const doServidor = await buscarPreferencias();
        const local = lerPreferenciasLocais();

        const mudou =
            doServidor.contrast_theme !== local.contrast_theme ||
            doServidor.font_family !== local.font_family ||
            doServidor.font_spacing !== local.font_spacing ||
            doServidor.font_size !== local.font_size;

        if (!mudou) return;

        guardarPreferenciasLocais(doServidor);
        aplicarPreferencias(doServidor);
    } catch {
        // Sem resposta do servidor a tela continua com o tema local. Falhar aqui
        // não pode derrubar a página: preferência é acessório, conteúdo não é.
    }
}

function criarItemDeNotificacao(notificacao) {
    const item = criarElemento('article', {
        classe: `gs-notif-item${notificacao.is_read ? '' : ' is-nao-lida'}`
    });

    const cabecalho = criarElemento('div', { classe: 'gs-meta' });

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

    item.append(cabecalho, criarElemento('p', { texto: notificacao.text_show }));

    if (notificacao.id_post) {
        item.append(
            criarElemento('a', {
                classe: 'gs-btn gs-btn-ghost',
                texto: 'Abrir publicação',
                atributos: { href: `/post?id=${notificacao.id_post}` }
            })
        );
    }

    return item;
}

async function carregarNotificacoes() {
    const marcador = document.querySelector('[data-nao-lidas]');
    const lista = document.querySelector('[data-notif-lista]');
    const botaoMarcar = document.querySelector('[data-notif-marcar-lidas]');
    if (!marcador || !lista || !botaoMarcar) return;

    const gatilho = marcador.closest('[data-gs-menu-trigger]');
    let notificacoes = [];

    function renderizar() {
        const naoLidas = notificacoes.filter((notificacao) => !notificacao.is_read).length;

        if (naoLidas === 0) {
            marcador.hidden = true;
            gatilho?.removeAttribute('aria-label');
        } else {
            marcador.textContent = String(naoLidas);
            marcador.hidden = false;
            gatilho?.setAttribute(
                'aria-label',
                `Notificações, ${naoLidas} não ${naoLidas === 1 ? 'lida' : 'lidas'}`
            );
        }

        botaoMarcar.disabled = naoLidas === 0;

        lista.replaceChildren();

        if (notificacoes.length === 0) {
            lista.append(
                criarElemento('p', {
                    classe: 'gs-notif-vazio',
                    texto: 'Nenhuma notificação. Quando alguém te marcar com @, aparece aqui.'
                })
            );
            return;
        }

        const fragmento = document.createDocumentFragment();
        notificacoes.forEach((notificacao) => fragmento.append(criarItemDeNotificacao(notificacao)));
        lista.append(fragmento);
    }

    try {
        notificacoes = await listarNotificacoes();
        renderizar();
    } catch {
        marcador.hidden = true;
        lista.replaceChildren(
            criarElemento('p', {
                classe: 'gs-notif-vazio',
                texto: 'Não foi possível carregar as notificações.'
            })
        );
        return;
    }

    botaoMarcar.addEventListener('click', async () => {
        botaoMarcar.disabled = true;

        try {
            await marcarNotificacoesComoLidas();
            notificacoes = notificacoes.map((notificacao) => ({ ...notificacao, is_read: true }));
            renderizar();

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
}

sincronizarPreferencias();
carregarNotificacoes();

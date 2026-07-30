/**
 * Comportamento do topo, comum a todas as páginas.
 *
 * Faz duas coisas e só:
 *   1. sincroniza as preferências de acessibilidade com a tabela `preferences`;
 *   2. mostra quantas notificações não lidas existem.
 *
 * O link ativo NÃO é marcado aqui. Ele já vem escrito no HTML de cada página
 * com `aria-current="page"`, para quem usa leitor de tela saber onde está mesmo
 * que o JavaScript falhe ou ainda não tenha carregado.
 */

import { buscarPreferencias, listarNotificacoes } from '../api.js';
import {
    aplicarPreferencias,
    guardarPreferenciasLocais,
    lerPreferenciasLocais
} from '../utils/preferencias.js';

/**
 * O tema-inicial.js já aplicou o que estava no localStorage antes da primeira
 * pintura. Aqui confirmamos com o banco: se o usuário mudou a preferência em
 * outro dispositivo, o valor do servidor vence e o espelho local é corrigido.
 */
async function sincronizarPreferencias() {
    try {
        const doServidor = await buscarPreferencias();
        const local = lerPreferenciasLocais();

        const mudou =
            doServidor.contrast_theme !== local.contrast_theme ||
            doServidor.font_family !== local.font_family;

        if (!mudou) return;

        guardarPreferenciasLocais(doServidor);
        aplicarPreferencias(doServidor);
    } catch {
        // Sem resposta do servidor a tela continua com o tema local. Falhar aqui
        // não pode derrubar a página: preferência é acessório, conteúdo não é.
    }
}

async function mostrarNaoLidas() {
    const marcador = document.querySelector('[data-nao-lidas]');
    if (!marcador) return;

    try {
        const notificacoes = await listarNotificacoes();
        const total = notificacoes.filter((notificacao) => !notificacao.is_read).length;

        if (total === 0) {
            marcador.hidden = true;
            return;
        }

        marcador.textContent = String(total);
        marcador.hidden = false;

        const link = marcador.closest('a');
        if (link) {
            link.setAttribute(
                'aria-label',
                `Notificações, ${total} não ${total === 1 ? 'lida' : 'lidas'}`
            );
        }
    } catch {
        marcador.hidden = true;
    }
}

sincronizarPreferencias();
mostrarNaoLidas();

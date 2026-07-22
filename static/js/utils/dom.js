/** Utilitários de manipulação do DOM usados por todas as páginas. */

/**
 * Cria um elemento já configurado.
 * O texto sempre entra por textContent, nunca por innerHTML.
 *
 * @param {string} tag nome da tag, ex.: 'li'
 * @param {{ classe?: string, texto?: string, atributos?: Record<string, string> }} opcoes
 */
export function criarElemento(tag, { classe, texto, atributos } = {}) {
    const elemento = document.createElement(tag);

    if (classe) elemento.className = classe;
    if (texto !== undefined) elemento.textContent = texto;

    if (atributos) {
        for (const [nome, valor] of Object.entries(atributos)) {
            elemento.setAttribute(nome, valor);
        }
    }

    return elemento;
}

/**
 * Devolve uma versão da função que só executa depois de `espera`
 * milissegundos sem novas chamadas.
 */
export function comAtraso(funcao, espera = 300) {
    let relogio;

    return (...argumentos) => {
        clearTimeout(relogio);
        relogio = setTimeout(() => funcao(...argumentos), espera);
    };
}

/**
 * Controla uma área de estado com aria-live, usando os blocos do kit:
 * gs-alert para erro e gs-empty-state para vazio ou carregamento.
 */
export function criarPainelDeEstado(elemento) {
    function desenhar(titulo, detalhe, classe) {
        elemento.replaceChildren();
        const caixa = criarElemento('div', { classe });
        caixa.append(criarElemento('strong', { texto: titulo }));
        if (detalhe) caixa.append(criarElemento('span', { texto: detalhe }));
        elemento.append(caixa);
        elemento.hidden = false;
    }

    return {
        carregando(mensagem) {
            desenhar(mensagem, '', 'gs-empty-state');
        },
        vazio(titulo, detalhe) {
            desenhar(titulo, detalhe, 'gs-empty-state');
        },
        erro(titulo, detalhe) {
            desenhar(titulo, detalhe, 'gs-alert');
        },
        ocultar() {
            elemento.replaceChildren();
            elemento.hidden = true;
        }
    };
}
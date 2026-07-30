/**
 * Preferências de acessibilidade do usuário.
 *
 * Os valores são exatamente os aceitos pelos CHECK da tabela `preferences`.
 * Se um valor novo entrar no banco, ele entra aqui e no static/css/temas.css —
 * em nenhum outro lugar.
 */

const CHAVE = 'germinastack:preferencias';

/** contrast_theme — mesmo domínio do CHECK da tabela `preferences`. */
export const TEMAS = [
    { valor: 'normal', rotulo: 'Padrão', descricao: 'Fundo claro com o azul e o laranja da identidade.' },
    { valor: 'dark', rotulo: 'Escuro', descricao: 'Fundo escuro, útil à noite e para quem tem fotofobia.' },
    { valor: 'high_contrast', rotulo: 'Alto contraste', descricao: 'Preto sobre branco com bordas reforçadas.' },
    { valor: 'black_yellow', rotulo: 'Preto e amarelo', descricao: 'Fundo preto com texto amarelo, contraste máximo.' },
    { valor: 'yellow_black', rotulo: 'Amarelo e preto', descricao: 'Fundo amarelo com texto preto, contraste máximo.' }
];

/** font_family — mesmo domínio do CHECK da tabela `preferences`. */
export const FONTES = [
    { valor: 'normal', rotulo: 'Padrão do sistema', descricao: 'A fonte que o kit já usa.' },
    { valor: 'arial', rotulo: 'Arial', descricao: 'Sem serifa, presente em qualquer computador.' },
    { valor: 'verdana', rotulo: 'Verdana', descricao: 'Letras mais largas e espaçadas.' },
    { valor: 'lexend', rotulo: 'Lexend', descricao: 'Desenhada para reduzir esforço de leitura.' },
    { valor: 'atkinson_hyperlegible', rotulo: 'Atkinson Hyperlegible', descricao: 'Criada para baixa visão; diferencia letras parecidas.' },
    { valor: 'open_dyslexic', rotulo: 'OpenDyslexic', descricao: 'Base pesada nas letras, pensada para dislexia.' }
];

const VALORES_DE_TEMA = TEMAS.map((tema) => tema.valor);
const VALORES_DE_FONTE = FONTES.map((fonte) => fonte.valor);

export const PREFERENCIAS_PADRAO = { contrast_theme: 'normal', font_family: 'normal' };

function normalizar(preferencias = {}) {
    return {
        contrast_theme: VALORES_DE_TEMA.includes(preferencias.contrast_theme)
            ? preferencias.contrast_theme
            : 'normal',
        font_family: VALORES_DE_FONTE.includes(preferencias.font_family)
            ? preferencias.font_family
            : 'normal'
    };
}

/** Lê o espelho local. O valor oficial vem da API. */
export function lerPreferenciasLocais() {
    try {
        return normalizar(JSON.parse(window.localStorage.getItem(CHAVE)) || {});
    } catch {
        return { ...PREFERENCIAS_PADRAO };
    }
}

/** Guarda o espelho local usado pelo tema-inicial.js no próximo carregamento. */
export function guardarPreferenciasLocais(preferencias) {
    window.localStorage.setItem(CHAVE, JSON.stringify(normalizar(preferencias)));
}

/**
 * Aplica o tema e a fonte no <html>.
 * O tema `normal` remove o atributo em vez de escrever "normal", para o CSS
 * do kit continuar valendo sem nenhuma regra nossa por cima.
 */
export function aplicarPreferencias(preferencias) {
    const { contrast_theme, font_family } = normalizar(preferencias);
    const raiz = document.documentElement;

    if (contrast_theme === 'normal') raiz.removeAttribute('data-tema');
    else raiz.setAttribute('data-tema', contrast_theme);

    if (font_family === 'normal') raiz.removeAttribute('data-fonte');
    else raiz.setAttribute('data-fonte', font_family);

    return { contrast_theme, font_family };
}

export { normalizar as normalizarPreferencias };

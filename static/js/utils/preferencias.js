/**
 * Preferências de acessibilidade do usuário.
 *
 * Os valores são exatamente os aceitos pelos CHECK da tabela `preferences`.
 * Se um valor novo entrar no banco, ele entra aqui e no CSS que trata os
 * atributos data-tema/data-fonte — hoje static/vendor/germinastack/themes.css,
 * publicado pela lib (ver static/README.md).
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

/** font_spacing — mesmo domínio do CHECK da tabela `preferences`. */
export const ESPACAMENTOS = [
    { valor: 'normal', rotulo: 'Padrão', descricao: 'O espaçamento que a fonte já usa.' },
    { valor: 'pequeno', rotulo: 'Pequeno', descricao: 'Letras um pouco mais próximas.' },
    { valor: 'grande', rotulo: 'Grande', descricao: 'Letras e palavras mais separadas, para ler com menos esforço.' }
];

/** font_size — mesmo domínio do CHECK da tabela `preferences`. */
export const TAMANHOS = [
    { valor: 'normal', rotulo: 'Padrão', descricao: 'O tamanho que o kit já usa.' },
    { valor: 'pequeno', rotulo: 'Pequeno', descricao: 'Cabe mais conteúdo na tela.' },
    { valor: 'grande', rotulo: 'Grande', descricao: 'Texto e botões maiores, para baixa visão.' }
];

const VALORES_DE_TEMA = TEMAS.map((tema) => tema.valor);
const VALORES_DE_FONTE = FONTES.map((fonte) => fonte.valor);
const VALORES_DE_ESPACAMENTO = ESPACAMENTOS.map((espacamento) => espacamento.valor);
const VALORES_DE_TAMANHO = TAMANHOS.map((tamanho) => tamanho.valor);

export const PREFERENCIAS_PADRAO = {
    contrast_theme: 'normal',
    font_family: 'normal',
    font_spacing: 'normal',
    font_size: 'normal'
};

function normalizar(preferencias = {}) {
    return {
        contrast_theme: VALORES_DE_TEMA.includes(preferencias.contrast_theme)
            ? preferencias.contrast_theme
            : 'normal',
        font_family: VALORES_DE_FONTE.includes(preferencias.font_family)
            ? preferencias.font_family
            : 'normal',
        font_spacing: VALORES_DE_ESPACAMENTO.includes(preferencias.font_spacing)
            ? preferencias.font_spacing
            : 'normal',
        font_size: VALORES_DE_TAMANHO.includes(preferencias.font_size)
            ? preferencias.font_size
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
    const { contrast_theme, font_family, font_spacing, font_size } = normalizar(preferencias);
    const raiz = document.documentElement;

    if (contrast_theme === 'normal') raiz.removeAttribute('data-tema');
    else raiz.setAttribute('data-tema', contrast_theme);

    if (font_family === 'normal') raiz.removeAttribute('data-fonte');
    else raiz.setAttribute('data-fonte', font_family);

    if (font_spacing === 'normal') raiz.removeAttribute('data-espacamento');
    else raiz.setAttribute('data-espacamento', font_spacing);

    if (font_size === 'normal') raiz.removeAttribute('data-tamanho');
    else raiz.setAttribute('data-tamanho', font_size);

    return { contrast_theme, font_family, font_spacing, font_size };
}

export { normalizar as normalizarPreferencias };

/**
 * Aplica o tema salvo ANTES da primeira pintura da tela.
 *
 * Este arquivo é carregado como script clássico e sem `defer` no <head>, então
 * ele roda enquanto o <body> ainda nem existe. É de propósito: se o tema fosse
 * aplicado por um módulo (que é sempre adiado), a página apareceria clara por
 * um instante antes de escurecer — e piscar a tela na cara de quem escolheu
 * alto contraste justamente por sensibilidade visual é o pior resultado
 * possível. O custo é ~1ms de bloqueio.
 *
 * A fonte da verdade continua sendo a tabela `preferences`; o localStorage é
 * só um espelho para o carregamento não depender de ida ao servidor.
 */

(function () {
    var CHAVE = 'germinastack:preferencias';

    var TEMAS = ['normal', 'dark', 'high_contrast', 'black_yellow', 'yellow_black'];
    var FONTES = ['normal', 'arial', 'verdana', 'lexend', 'atkinson_hyperlegible', 'open_dyslexic'];
    var ESCALA = ['normal', 'pequeno', 'grande'];

    var salvo = {};

    try {
        salvo = JSON.parse(window.localStorage.getItem(CHAVE)) || {};
    } catch (erro) {
        salvo = {};
    }

    var tema = TEMAS.indexOf(salvo.contrast_theme) >= 0 ? salvo.contrast_theme : 'normal';
    var fonte = FONTES.indexOf(salvo.font_family) >= 0 ? salvo.font_family : 'normal';
    var espacamento = ESCALA.indexOf(salvo.font_spacing) >= 0 ? salvo.font_spacing : 'normal';
    var tamanho = ESCALA.indexOf(salvo.font_size) >= 0 ? salvo.font_size : 'normal';

    // O valor `normal` é a ausência do atributo: assim o CSS do kit vale sozinho.
    if (tema !== 'normal') document.documentElement.setAttribute('data-tema', tema);
    if (fonte !== 'normal') document.documentElement.setAttribute('data-fonte', fonte);
    if (espacamento !== 'normal') document.documentElement.setAttribute('data-espacamento', espacamento);
    if (tamanho !== 'normal') document.documentElement.setAttribute('data-tamanho', tamanho);
})();

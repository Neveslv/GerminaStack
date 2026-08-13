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

    if (tema !== 'normal') document.documentElement.setAttribute('data-tema', tema);
    if (fonte !== 'normal') document.documentElement.setAttribute('data-fonte', fonte);
    if (espacamento !== 'normal') document.documentElement.setAttribute('data-espacamento', espacamento);
    if (tamanho !== 'normal') document.documentElement.setAttribute('data-tamanho', tamanho);
})();

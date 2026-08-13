import { buscarSugestoesDeUsuario } from '../api.js';
import { criarElemento } from '../utils/dom.js';

const MENTION = /@([A-Za-z0-9_]+(?:\.[A-Za-z0-9_]+)*)/g;
const TOKEN = /(^|[^A-Za-z0-9_])@([A-Za-z0-9_.]*)$/;

export function montarTextoComMencoes(texto) {
    const fragmento = document.createDocumentFragment();
    let inicio = 0;
    let match;
    while ((match = MENTION.exec(texto)) !== null) {
        const anterior = texto[match.index - 1];
        if (anterior && /[A-Za-z0-9_]/.test(anterior)) continue;
        fragmento.append(document.createTextNode(texto.slice(inicio, match.index)));
        fragmento.append(criarElemento('mark', {
            classe: 'gs-mention',
            texto: match[0],
            atributos: { 'data-username': match[1] }
        }));
        inicio = match.index + match[0].length;
    }
    fragmento.append(document.createTextNode(texto.slice(inicio)));
    return fragmento;
}

export function montarTextoComMencoesEGifs(texto) {
    const fragmento = document.createDocumentFragment();
    const partes = texto.split(/(https?:\/\/\S+\.gif(?:\?\S*)?)/gi);
    partes.forEach((parte) => {
        if (/^https?:\/\/\S+\.gif(?:\?\S*)?$/i.test(parte)) {
            fragmento.append(criarElemento('img', {
                atributos: { src: parte, alt: 'GIF enviado na resposta', loading: 'lazy' }
            }));
        } else {
            fragmento.append(montarTextoComMencoes(parte));
        }
    });
    return fragmento;
}

function esconder(lista) {
    lista.replaceChildren();
    lista.hidden = true;
}

export function ativarAutocompleteDeMencoes(campo) {
    const lista = criarElemento('div', {
        classe: 'gs-mention-suggestions',
        atributos: { role: 'listbox', hidden: '' }
    });
    campo.insertAdjacentElement('afterend', lista);
    let sugestoes = [];
    let indice = -1;
    let requisicao = 0;

    function tokenAtual() {
        return TOKEN.exec(campo.value.slice(0, campo.selectionStart ?? campo.value.length));
    }

    function escolher(usuario) {
        const token = tokenAtual();
        if (!token) return;
        const inicio = (campo.selectionStart ?? 0) - token[0].length + token[1].length;
        const fim = campo.selectionStart ?? campo.value.length;
        campo.setRangeText(`@${usuario.username} `, inicio, fim, 'end');
        esconder(lista);
        sugestoes = [];
        indice = -1;
        campo.focus();
    }

    function desenhar() {
        lista.replaceChildren();
        sugestoes.forEach((usuario, posicao) => {
            const botao = criarElemento('button', {
                classe: 'gs-mention-option',
                atributos: { type: 'button', role: 'option', 'aria-selected': String(posicao === indice) }
            });
            botao.append(
                criarElemento('strong', { texto: `@${usuario.username}` }),
                criarElemento('span', { texto: usuario.name })
            );
            botao.addEventListener('mousedown', (evento) => {
                evento.preventDefault();
                escolher(usuario);
            });
            lista.append(botao);
        });
        lista.hidden = sugestoes.length === 0;
    }

    async function atualizar() {
        const token = tokenAtual();
        const prefixo = token?.[2] ?? '';
        if (!token || prefixo.length === 0) {
            esconder(lista);
            sugestoes = [];
            return;
        }
        const atual = ++requisicao;
        try {
            const resultado = await buscarSugestoesDeUsuario(prefixo);
            if (atual !== requisicao || !tokenAtual()) return;
            sugestoes = resultado;
            indice = -1;
            desenhar();
        } catch {
            esconder(lista);
        }
    }

    campo.addEventListener('input', atualizar);
    campo.addEventListener('keydown', (evento) => {
        if (lista.hidden || sugestoes.length === 0) return;
        if (evento.key === 'ArrowDown') {
            evento.preventDefault();
            indice = (indice + 1) % sugestoes.length;
            desenhar();
        } else if (evento.key === 'ArrowUp') {
            evento.preventDefault();
            indice = (indice - 1 + sugestoes.length) % sugestoes.length;
            desenhar();
        } else if (evento.key === 'Enter' && indice >= 0) {
            evento.preventDefault();
            escolher(sugestoes[indice]);
        } else if (evento.key === 'Escape') esconder(lista);
    });
    campo.addEventListener('blur', () => setTimeout(() => esconder(lista), 120));
}


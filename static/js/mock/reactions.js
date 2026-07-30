/**
 * Reações do usuário logado, no formato da tabela `reactions`.
 *
 * A tabela tem um CHECK que obriga exatamente UM entre id_post, id_comment e
 * id_comment_on_comment a ser preenchido. Aqui isso vira uma chave
 * "tipo:id", que dá o mesmo efeito sem repetir três colunas nulas.
 *
 * O comportamento de alternância copia a função `reaction()` do banco:
 *   sem reação        -> insere
 *   mesma reação      -> apaga (desfaz)
 *   reação diferente  -> troca
 * Manter essa regra igual dos dois lados evita o front mostrar um estado que
 * o banco não confirmaria.
 */

const reacoesDoUsuario = new Map([
    ['post:3', 'like'],
    ['post:5', 'like'],
    ['comment:21', 'like']
]);

function chave(tipo, id) {
    return `${tipo}:${Number(id)}`;
}

/** Devolve 'like', 'dislike' ou null para o alvo informado. */
export function reacaoAtual(tipo, id) {
    return reacoesDoUsuario.get(chave(tipo, id)) ?? null;
}

/**
 * Aplica a alternância e devolve o quanto cada contador mudou.
 * @returns {{ reacao: string|null, likes: number, dislikes: number }} deltas
 */
export function alternarReacao(tipo, id, novaReacao) {
    const anterior = reacaoAtual(tipo, id);
    const deltas = { reacao: novaReacao, likes: 0, dislikes: 0 };

    if (anterior === novaReacao) {
        reacoesDoUsuario.delete(chave(tipo, id));
        deltas.reacao = null;
        deltas[anterior === 'like' ? 'likes' : 'dislikes'] = -1;
        return deltas;
    }

    reacoesDoUsuario.set(chave(tipo, id), novaReacao);
    deltas[novaReacao === 'like' ? 'likes' : 'dislikes'] = 1;
    if (anterior) deltas[anterior === 'like' ? 'likes' : 'dislikes'] = -1;

    return deltas;
}

/**
 * Preferências do usuário logado, no formato da tabela `preferences`.
 *
 * É `let` e não `const` porque a API local escreve aqui ao salvar, simulando
 * o UPDATE que o back-end vai fazer. Os valores são os DEFAULT do schema.
 */

export let PREFERENCIAS_LOCAIS = {
    id: 1,
    id_user: 4,
    contrast_theme: 'normal',
    font_family: 'normal',
    font_spacing: 'normal',
    font_size: 'normal',
    created_at: '2026-02-10T08:00:00-03:00'
};

export function gravarPreferenciasLocais(preferencias) {
    PREFERENCIAS_LOCAIS = { ...PREFERENCIAS_LOCAIS, ...preferencias };
    return PREFERENCIAS_LOCAIS;
}

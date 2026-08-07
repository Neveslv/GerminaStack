/**
 * Cor e inicial de avatar, e tom do chip de matéria.
 *
 * A escolha é derivada do id, nunca sorteada: o mesmo usuário precisa ter
 * sempre a mesma cor, no feed e na página da publicação, senão quem reconhece
 * a pessoa pela cor se perde a cada tela.
 *
 * A cor é só reforço visual — nunca é o único jeito de identificar alguém,
 * porque o nome está sempre escrito ao lado (WCAG 1.4.1, uso de cor).
 */

const CORES_DE_AVATAR = ['#ffb347', '#4db8ff', '#5fd6a4', '#c9a0ff'];
const TONS_DE_CHIP = ['is-blue', 'is-amber', 'is-mint', 'is-rose'];
const FOTO_DO_FROK = '/static/images/frok-profile.jpeg';

export function corDoAutor(idAutor) {
    return CORES_DE_AVATAR[idAutor % CORES_DE_AVATAR.length];
}

export function tomDaMateria(idMateria) {
    return TONS_DE_CHIP[idMateria % TONS_DE_CHIP.length];
}

export function inicialDoNome(nome) {
    return String(nome || '?').trim().charAt(0).toUpperCase();
}

export function fotoDoAutor(autor) {
    const identificador = String(autor?.username || autor?.name || '').trim().toLowerCase();
    return autor?.profile_image_url || (identificador === 'frok' ? FOTO_DO_FROK : null);
}

export function aplicarImagemNoAvatar(avatar, autor) {
    const url = fotoDoAutor(autor);
    if (!url) return false;

    const imagem = document.createElement('img');
    imagem.src = url;
    imagem.alt = '';
    imagem.style.cssText = 'width:100%;height:100%;object-fit:cover;border-radius:inherit;';
    avatar.replaceChildren(imagem);
    return true;
}

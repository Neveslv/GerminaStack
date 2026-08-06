const CORES_DE_AVATAR = ['#ffb347', '#4db8ff', '#5fd6a4', '#c9a0ff'];
const TONS_DE_CHIP = ['is-blue', 'is-amber', 'is-mint', 'is-rose'];

export function corDoAutor(idAutor) {
    return CORES_DE_AVATAR[idAutor % CORES_DE_AVATAR.length];
}

export function tomDaMateria(idMateria) {
    return TONS_DE_CHIP[idMateria % TONS_DE_CHIP.length];
}

export function inicialDoNome(nome) {
    return String(nome || '?').trim().charAt(0).toUpperCase();
}

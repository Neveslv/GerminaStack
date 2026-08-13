const formatadorLongo = new Intl.DateTimeFormat('pt-BR', {
    timeZone: 'America/Sao_Paulo',
    day: '2-digit',
    month: 'long',
    year: 'numeric'
});

const formatadorRelativo = new Intl.RelativeTimeFormat('pt-BR', { numeric: 'auto' });

const ESCALAS = [
    { unidade: 'year', segundos: 31536000 },
    { unidade: 'month', segundos: 2592000 },
    { unidade: 'day', segundos: 86400 },
    { unidade: 'hour', segundos: 3600 },
    { unidade: 'minute', segundos: 60 }
];

export function formatarDataCompleta(iso) {
    return formatadorLongo.format(new Date(iso));
}

export function formatarDataRelativa(iso) {
    const segundos = (new Date(iso).getTime() - Date.now()) / 1000;

    for (const escala of ESCALAS) {
        if (Math.abs(segundos) >= escala.segundos) {
            return formatadorRelativo.format(Math.round(segundos / escala.segundos), escala.unidade);
        }
    }

    return 'agora mesmo';
}
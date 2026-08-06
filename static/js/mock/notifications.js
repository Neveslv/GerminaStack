/**
 * Notificações de exemplo, no formato da tabela `notifications`.
 *
 * No banco elas não são criadas pelo back-end: quem insere é o trigger
 * `notify_mentions`, que varre o conteúdo de posts, comentários e respostas
 * atrás de `@usuario` e grava uma linha por menção. Por isso `text_show` já
 * chega pronto do banco — o front só exibe, nunca monta a frase.
 *
 * Todas pertencem ao usuário de mock/sessao.js (id 4).
 */

export const NOTIFICACOES_LOCAIS = [
    {
        id: 501,
        id_post: 5,
        text_show: 'elisa.tavares mencionou você no post "Experimento de queda livre deu errado, e eu descobri o motivo"',
        is_read: false,
        created_at: '2026-07-28T17:42:00-03:00'
    },
    {
        id: 502,
        id_post: 2,
        text_show: 'bruno.salles mencionou você no post "Dúvida sobre ligação iônica e covalente"',
        is_read: false,
        created_at: '2026-07-28T09:15:00-03:00'
    },
    {
        id: 503,
        id_post: 9,
        text_show: 'joao.alves mencionou você no post "Fichamento de “Vidas Secas” capítulo por capítulo"',
        is_read: false,
        created_at: '2026-07-27T20:03:00-03:00'
    },
    {
        id: 504,
        id_post: 3,
        text_show: 'carla.menezes mencionou você no post "Resumo da Revolução Industrial que fiz para a prova"',
        is_read: true,
        created_at: '2026-07-24T11:28:00-03:00'
    },
    {
        id: 505,
        id_post: 1,
        text_show: 'diego.farias mencionou você no post "Como resolver equação do 2º grau sem decorar Bhaskara"',
        is_read: true,
        created_at: '2026-07-21T15:50:00-03:00'
    }
];

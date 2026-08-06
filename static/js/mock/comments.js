/**
 * Comentários provisórios, no formato dos endpoints de comentários.
 * A estrutura em dois níveis espelha as tabelas `comments` e
 * `comments_on_comments`: uma resposta responde a um comentário,
 * e não existe terceiro nível.
 */

export const COMENTARIOS_LOCAIS = {
    1: [
        {
            id: 11,
            content: 'Completar quadrados é exatamente como Bhaskara é demonstrada. Bom ver alguém chegar nisso sozinho.',
            likes: 4,
            dislikes: 0,
            created_at: '2026-07-02T15:10:00-03:00',
            author: { id: 15, username: 'elisa.tavares', name: 'Elisa Tavares' },
            replies: [
                {
                    id: 111,
                    content: 'Tem um vídeo do professor no drive da turma mostrando a demonstração completa.',
                    likes: 2,
                    dislikes: 0,
                    created_at: '2026-07-02T16:02:00-03:00',
                    author: { id: 27, username: 'isabela.nunes', name: 'Isabela Nunes' }
                }
            ]
        },
        {
            id: 12,
            content: 'Fiz assim na prova e o professor aceitou. Só cuidado ao dividir os dois lados quando o coeficiente não é 1.',
            likes: 7,
            dislikes: 0,
            created_at: '2026-07-03T08:45:00-03:00',
            author: { id: 12, username: 'diego.farias', name: 'Diego Farias' },
            replies: []
        }
    ],
    2: [
        {
            id: 21,
            content: 'Dissolvido em água o NaCl se separa em Na+ e Cl-, e são esses íons livres que conduzem.',
            likes: 9,
            dislikes: 0,
            created_at: '2026-07-03T10:30:00-03:00',
            author: { id: 18, username: 'felipe.andrade', name: 'Felipe Andrade' },
            replies: []
        }
    ]
};
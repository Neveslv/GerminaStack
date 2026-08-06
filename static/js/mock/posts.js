/**
 * Dados provisórios do feed, usados enquanto os endpoints do back-end
 * não estão disponíveis.
 *
 * O formato aqui NÃO é inventado: reproduz exatamente o que o endpoint
 * GET /api/posts deve devolver, segundo o modelo do banco (tabela `posts`)
 * combinado com o contrato proposto ao time do back (autor e matéria
 * aninhados via JOIN, contagem de comentários já calculada).
 *
 * Se o formato combinado com o back mudar, é só ajustar aqui —
 * nenhum outro arquivo do front sabe que estes dados são falsos.
 */

export const POSTS_LOCAIS = [
    {
        id: 1,
        title: 'Como resolver equação do 2º grau sem decorar Bhaskara',
        content: 'Completar quadrados me ajudou a entender de onde a fórmula vem.',
        image_url: null,
        likes: 12,
        dislikes: 0,
        created_at: '2026-07-02T13:20:00-03:00',
        author: { id: 4, username: 'ana.ribeiro', name: 'Ana Ribeiro' },
        subject: { id: 1, subject: 'Matemática' },
        comments_count: 7
    },
    {
        id: 2,
        title: 'Dúvida sobre ligação iônica e covalente',
        content: 'Não entendi por que o NaCl conduz corrente dissolvido em água.',
        image_url: null,
        likes: 5,
        dislikes: 1,
        created_at: '2026-07-03T09:05:00-03:00',
        author: { id: 7, username: 'bruno.salles', name: 'Bruno Salles' },
        subject: { id: 2, subject: 'Química' },
        comments_count: 12
    },
    {
        id: 3,
        title: 'Resumo da Revolução Industrial que fiz para a prova',
        content: 'Dividi por fases e destaquei as causas sociais de cada uma.',
        image_url: null,
        likes: 20,
        dislikes: 0,
        created_at: '2026-07-05T18:40:00-03:00',
        author: { id: 9, username: 'carla.menezes', name: 'Carla Menezes' },
        subject: { id: 3, subject: 'História' },
        comments_count: 4
    },
    {
        id: 4,
        title: 'Alguém tem material bom de análise sintática?',
        content: 'Travo sempre na diferença entre objeto direto e indireto.',
        image_url: null,
        likes: 3,
        dislikes: 0,
        created_at: '2026-07-07T11:15:00-03:00',
        author: { id: 12, username: 'diego.farias', name: 'Diego Farias' },
        subject: { id: 4, subject: 'Português' },
        comments_count: 9
    },
    {
        id: 5,
        title: 'Experimento de queda livre deu errado, e eu descobri o motivo',
        content: 'A resistência do ar explicava a diferença que achávamos ser erro.',
        image_url: null,
        likes: 31,
        dislikes: 2,
        created_at: '2026-07-08T15:50:00-03:00',
        author: { id: 15, username: 'elisa.tavares', name: 'Elisa Tavares' },
        subject: { id: 5, subject: 'Física' },
        comments_count: 15
    },
    {
        id: 6,
        title: 'Mapa mental de fotossíntese',
        content: 'Separei fase clara e fase escura em cores diferentes.',
        image_url: null,
        likes: 8,
        dislikes: 0,
        created_at: '2026-07-10T08:30:00-03:00',
        author: { id: 18, username: 'felipe.andrade', name: 'Felipe Andrade' },
        subject: { id: 6, subject: 'Biologia' },
        comments_count: 3
    },
    {
        id: 7,
        title: 'Dicas de present perfect que finalmente fizeram sentido',
        content: 'A ideia de "ponte com o presente" mudou meu jeito de estudar.',
        image_url: null,
        likes: 6,
        dislikes: 0,
        created_at: '2026-07-12T20:10:00-03:00',
        author: { id: 21, username: 'gabriela.lopes', name: 'Gabriela Lopes' },
        subject: { id: 7, subject: 'Inglês' },
        comments_count: 6
    },
    {
        id: 8,
        title: 'Diferença entre média, mediana e moda na prática',
        content: 'Usei as notas da turma como exemplo e ficou visível.',
        image_url: null,
        likes: 9,
        dislikes: 0,
        created_at: '2026-07-15T14:00:00-03:00',
        author: { id: 27, username: 'isabela.nunes', name: 'Isabela Nunes' },
        subject: { id: 1, subject: 'Matemática' },
        comments_count: 5
    },
    {
        id: 9,
        title: 'Fichamento de "Vidas Secas" capítulo por capítulo',
        content: 'Anotei o que cada personagem representa na seca.',
        image_url: null,
        likes: 14,
        dislikes: 0,
        created_at: '2026-07-16T19:25:00-03:00',
        author: { id: 30, username: 'joao.alves', name: 'João Pedro Alves' },
        subject: { id: 8, subject: 'Literatura' },
        comments_count: 8
    }
];
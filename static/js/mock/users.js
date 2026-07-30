/**
 * Usuários de exemplo, no formato da tabela `users` — sem a coluna `password`,
 * que nunca deve sair do back-end.
 *
 * Os ids batem com os autores usados em mock/posts.js e mock/comments.js:
 * é o mesmo usuário em toda a aplicação, como seria com o JOIN real.
 */

export const USUARIOS_LOCAIS = [
    { id: 4, name: 'Ana Ribeiro', username: 'ana.ribeiro', email: 'ana.ribeiro@germinare.org.br', year: { id: 2, year: '2º Tec' }, created_at: '2026-02-10T08:00:00-03:00' },
    { id: 7, name: 'Bruno Salles', username: 'bruno.salles', email: 'bruno.salles@germinare.org.br', year: { id: 2, year: '2º Tec' }, created_at: '2026-02-10T08:05:00-03:00' },
    { id: 9, name: 'Carla Menezes', username: 'carla.menezes', email: 'carla.menezes@germinare.org.br', year: { id: 3, year: '3º Tec' }, created_at: '2026-02-11T09:30:00-03:00' },
    { id: 12, name: 'Diego Farias', username: 'diego.farias', email: 'diego.farias@germinare.org.br', year: { id: 1, year: '1º Tec' }, created_at: '2026-02-12T10:15:00-03:00' },
    { id: 15, name: 'Elisa Tavares', username: 'elisa.tavares', email: 'elisa.tavares@germinare.org.br', year: { id: 3, year: '3º Tec' }, created_at: '2026-02-12T11:00:00-03:00' },
    { id: 18, name: 'Felipe Andrade', username: 'felipe.andrade', email: 'felipe.andrade@germinare.org.br', year: { id: 2, year: '2º Tec' }, created_at: '2026-02-13T14:20:00-03:00' },
    { id: 21, name: 'Gabriela Lopes', username: 'gabriela.lopes', email: 'gabriela.lopes@germinare.org.br', year: { id: 1, year: '1º Tec' }, created_at: '2026-02-14T16:45:00-03:00' },
    { id: 27, name: 'Isabela Nunes', username: 'isabela.nunes', email: 'isabela.nunes@germinare.org.br', year: { id: 3, year: '3º Tec' }, created_at: '2026-02-15T08:10:00-03:00' },
    { id: 30, name: 'João Pedro Alves', username: 'joao.alves', email: 'joao.alves@germinare.org.br', year: { id: 2, year: '2º Tec' }, created_at: '2026-02-16T13:35:00-03:00' }
];

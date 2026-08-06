/**
 * Configuração de ambiente.
 * Único lugar do projeto onde endereços e chaves de ambiente são definidos.
 */

/** Front e API usam a mesma origem, localmente e na Discloud. */
export const API_BASE_URL = '';

/** Os mocks ficam disponíveis apenas para desenvolvimento isolado do front. */
export const USAR_DADOS_LOCAIS = false;

/** Tempo máximo de espera por resposta da API, em milissegundos. */
export const TIMEOUT_MS = 8000;

/** Página para onde o usuário é levado quando a sessão não é válida. */
export const ROTA_LOGIN = '/login';

/**
 * Configuração de ambiente.
 * Único lugar do projeto onde endereços e chaves de ambiente são definidos.
 */

/** Apresentação em localhost: o back-end só roda no próprio computador. */
export const API_BASE_URL = 'http://localhost:8080';

/** Enquanto o back-end não expõe os endpoints, o front consome dados locais. */
export const USAR_DADOS_LOCAIS = true;

/** Tempo máximo de espera por resposta da API, em milissegundos. */
export const TIMEOUT_MS = 8000;

/** Página para onde o usuário é levado quando a sessão não é válida. */
export const ROTA_LOGIN = 'login.html';
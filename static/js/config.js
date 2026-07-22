/**
 * Configuração de ambiente.
 * Único lugar do projeto onde endereços e chaves de ambiente são definidos.
 */

const HOSTS_LOCAIS = ['localhost', '127.0.0.1'];

const API_LOCAL = 'http://localhost:8080';
const API_PRODUCAO = 'https://germinastack.onrender.com';

export const estaEmDesenvolvimento = HOSTS_LOCAIS.includes(window.location.hostname);

export const API_BASE_URL = estaEmDesenvolvimento ? API_LOCAL : API_PRODUCAO;

/** Enquanto o back-end não expõe os endpoints, o front consome dados locais. */
export const USAR_DADOS_LOCAIS = true;

/** Tempo máximo de espera por resposta da API, em milissegundos. */
export const TIMEOUT_MS = 8000;

/** Página para onde o usuário é levado quando a sessão não é válida. */
export const ROTA_LOGIN = 'login.html';
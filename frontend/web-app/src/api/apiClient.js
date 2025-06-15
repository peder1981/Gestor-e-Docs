import axios from 'axios';

// Obter a URL base da API do ambiente ou usar um caminho relativo por padrão
const API_BASE_URL = process.env.REACT_APP_API_URL || '';

const apiClient = axios.create({
  baseURL: `${API_BASE_URL}/api/v1/identity`,
  withCredentials: true,
  headers: {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
  },
  // Configurações avançadas do axios para cookies
  xsrfCookieName: false, // Desativa XSRF cookie
  maxRedirects: 0, // Evita redirecionamentos automáticos
});

let isRefreshing = false;
let failedQueue = [];

const processQueue = (error, token = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });

  failedQueue = [];
};

apiClient.interceptors.response.use(
  response => {
    return response;
  },
  async error => {
    const originalRequest = error.config;

    // Se a requisição está marcada para pular refresh ou já é uma tentativa de refresh
    if (error.config._skipRefresh || originalRequest._retry) {
      return Promise.reject(error);
    }

    if (error.response.status === 401) {

      if (isRefreshing) {
        return new Promise(function (resolve, reject) {
          failedQueue.push({ resolve, reject });
        })
          .then(() => apiClient(originalRequest))
          .catch(err => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;
      window.dispatchEvent(new CustomEvent('tokenRefreshStart'));

      try {
        
        // Marca a requisição de refresh para não entrar em loop
        const refreshRequest = apiClient.post('/refresh', {}, { _skipRefresh: true });
        
        try {
          const response = await refreshRequest;
          
          if (response.status === 200) {
            // Aguarda um momento para garantir que os cookies foram salvos
            await new Promise(resolve => setTimeout(resolve, 100));
          } else {
            throw new Error('Falha no refresh do token');
          }
        } catch (refreshError) {
          console.error('[apiClient] Erro no refresh:', refreshError.message);
          throw refreshError;
        }
        
        processQueue(null, null);
        return apiClient(originalRequest);
      } catch (refreshError) {
        console.error('[apiClient] Erro no refresh:', refreshError.response?.data || refreshError.message);
        processQueue(refreshError, null);
        
        // Dispara um evento para deslogar o usuário globalmente
        window.dispatchEvent(new CustomEvent('authChange', { detail: { isAuthenticated: false } }));
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
        window.dispatchEvent(new CustomEvent('tokenRefreshEnd'));
      }
    }

    return Promise.reject(error);
  }
);

export default apiClient;

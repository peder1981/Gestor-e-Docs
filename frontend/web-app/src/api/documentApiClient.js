import axios from 'axios';

const documentApiClient = axios.create({
  baseURL: '/api/v1/documents',
  withCredentials: true,
  headers: {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
  },
  // Configurações avançadas do axios para cookies
  xsrfCookieName: false, // Desativa XSRF cookie
  maxRedirects: 0, // Evita redirecionamentos automáticos
});

// Interceptor para ajustar URLs duplicadas
documentApiClient.interceptors.request.use(config => {
  // Remove possível duplicação de /api/v1/documents na URL
  if (config.url?.startsWith('/api/v1/documents')) {
    config.url = config.url.replace('/api/v1/documents', '');
  }
  return config;
});

// Interceptor para tratar erros de autenticação
documentApiClient.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      // Dispara evento para notificar o AuthContext
      window.dispatchEvent(new CustomEvent('authChange', { 
        detail: { isAuthenticated: false } 
      }));
    }
    return Promise.reject(error);
  }
);

export default documentApiClient;

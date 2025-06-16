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

// Interceptor para ajustar URLs duplicadas e adicionar logs de depuração
documentApiClient.interceptors.request.use(config => {
  // Remove possível duplicação de /api/v1/documents na URL
  if (config.url?.startsWith('/api/v1/documents')) {
    config.url = config.url.replace('/api/v1/documents', '');
  }
  console.log('Requisição API sendo enviada:', {
    url: config.url,
    method: config.method,
    data: config.data,
    headers: config.headers,
    baseURL: config.baseURL
  });
  return config;
});

// Interceptor para tratar respostas e erros
documentApiClient.interceptors.response.use(
  response => {
    console.log('Resposta da API recebida com sucesso:', {
      url: response.config.url,
      status: response.status,
      data: response.data,
      headers: response.headers
    });
    return response;
  },
  async error => {
    console.error('Erro na chamada de API:', {
      url: error.config?.url,
      method: error.config?.method,
      status: error.response?.status,
      message: error.message,
      response: error.response?.data
    });
    
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

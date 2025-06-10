# Integração de Autenticação no Frontend - Gestor-e-Docs

## Visão Geral

Este documento descreve como o frontend React se integra com o sistema de autenticação baseado em JWT com cookies HttpOnly, abordando componentes, eventos personalizados e fluxos de autenticação.

## Gerenciamento de Estado de Autenticação

### Contexto Global de Autenticação

O estado de autenticação é gerenciado através de um contexto React que fornece informações sobre o usuário autenticado para toda a aplicação:

```jsx
// AuthContext.js
import { createContext, useState, useEffect } from 'react';

export const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Verificar estado de autenticação inicial
    checkAuthStatus();
    
    // Adicionar listener para eventos de mudança de autenticação
    window.addEventListener('authChange', handleAuthChange);
    
    return () => {
      window.removeEventListener('authChange', handleAuthChange);
    };
  }, []);
  
  const checkAuthStatus = async () => {
    try {
      const response = await fetch('http://localhost:8085/api/v1/identity/me', {
        credentials: 'include' // Fundamental para enviar cookies
      });
      
      if (response.ok) {
        const user = await response.json();
        setCurrentUser(user);
        setIsAuthenticated(true);
      }
    } catch (error) {
      console.error('Erro ao verificar status de autenticação:', error);
    } finally {
      setLoading(false);
    }
  };
  
  const handleAuthChange = (event) => {
    const { authenticated, user } = event.detail;
    setIsAuthenticated(authenticated);
    setCurrentUser(user);
  };
  
  return (
    <AuthContext.Provider value={{ currentUser, isAuthenticated, loading }}>
      {children}
    </AuthContext.Provider>
  );
};
```

## Integração com API

### Cliente HTTP com Credenciais

Todas as requisições são configuradas para incluir cookies e lidar com erros de autenticação:

```jsx
// apiClient.js
import axios from 'axios';

const API_URL = 'http://localhost:8085/api/v1/identity';

const apiClient = axios.create({
  baseURL: API_URL,
  withCredentials: true, // Crucial para incluir cookies em todas as requisições
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
});

// Interceptor para tratar erros 401 (unauthorized)
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    // Se receber 401 e não for uma requisição de refresh ou login
    if (error.response?.status === 401 && 
        !originalRequest._retry &&
        !originalRequest.url.includes('/refresh') &&
        !originalRequest.url.includes('/login')) {
      
      originalRequest._retry = true;
      
      try {
        // Tentar renovar o token
        await apiClient.post('/refresh');
        // Repetir a requisição original
        return apiClient(originalRequest);
      } catch (refreshError) {
        // Se falhar o refresh, disparar evento de logout
        window.dispatchEvent(new CustomEvent('authChange', { 
          detail: { authenticated: false, user: null } 
        }));
        
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);

export default apiClient;
```

## Eventos Personalizados

O sistema utiliza eventos personalizados para sincronizar o estado de autenticação em toda a aplicação:

### Evento `authChange`

```javascript
// Disparado após login bem-sucedido
window.dispatchEvent(new CustomEvent('authChange', { 
  detail: { authenticated: true, user: userData } 
}));

// Disparado após logout
window.dispatchEvent(new CustomEvent('authChange', { 
  detail: { authenticated: false, user: null } 
}));
```

## Fluxos de Autenticação

### Login

```jsx
const login = async (email, password) => {
  try {
    setLoading(true);
    setError(null);
    
    const response = await apiClient.post('/login', { email, password });
    
    if (response.status === 200) {
      const userData = response.data;
      
      // Disparar evento de autenticação bem-sucedida
      window.dispatchEvent(new CustomEvent('authChange', { 
        detail: { authenticated: true, user: userData } 
      }));
      
      return true;
    }
  } catch (error) {
    setError('Credenciais inválidas. Por favor tente novamente.');
    console.error('Erro de login:', error);
    return false;
  } finally {
    setLoading(false);
  }
};
```

### Logout

```jsx
const handleLogout = async () => {
  try {
    await apiClient.post('/logout');
    
    // Remover tokens do estado global
    setToken(null);
    setUser(null);
    
    // Disparar evento para notificar toda a aplicação sobre o logout
    window.dispatchEvent(new CustomEvent('authChange', {
      detail: { authenticated: false, user: null }
    }));
    
    // Redirecionar para página de login
    navigate('/login');
  } catch (error) {
    console.error('Erro ao fazer logout:', error);
  }
};
```

## Componente de Proteção de Rotas

Para garantir que apenas usuários autenticados acessem determinadas rotas:

```jsx
// ProtectedRoute.js
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export const ProtectedRoute = () => {
  const { isAuthenticated, loading } = useAuth();
  
  // Mostrar indicador de carregamento enquanto verifica autenticação
  if (loading) {
    return <div>Carregando...</div>;
  }
  
  // Redirecionar para login se não estiver autenticado
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  // Renderizar a rota protegida se estiver autenticado
  return <Outlet />;
};
```

## Considerações para Produção

1. **Timeout de Sessão**: Implementar lógica de expiração de sessão por inatividade
2. **Integração com PWA**: Considerar comportamento offline e sincronização de estado
3. **Tratamento de Erros**: Melhorar UX para falhas de rede ou servidor
4. **Segurança**: Validação adicional de payload e sanitização de inputs
5. **Performance**: Otimizar verificações de refresh token para evitar requisições desnecessárias

## Testes Manuais

Use a página `test-auth.html` para:

1. Verificar se cookies são definidos após login
2. Testar refresh de token automaticamente
3. Confirmar que logout remove cookies
4. Validar redirecionamentos para login em rotas protegidas

## Melhores Práticas

1. Sempre use `credentials: 'include'` ou `withCredentials: true` em requisições
2. Nunca tente acessar os cookies HttpOnly via JavaScript (impossível por design)
3. Implemente o tratamento de erro 401 de forma consistente
4. Utilize eventos personalizados para sincronizar o estado de autenticação
5. Mantenha tempos de expiração compatíveis entre frontend e backend

---

# Frontend Authentication Integration - Gestor-e-Docs (English Version)

## Overview

This document describes how the React frontend integrates with the JWT-based authentication system using HttpOnly cookies, covering components, custom events, and authentication flows.

## Authentication State Management

### Global Authentication Context

The authentication state is managed through a React context that provides information about the authenticated user to the entire application:

```jsx
// AuthContext.js
import { createContext, useState, useEffect } from 'react';

export const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check initial authentication status
    checkAuthStatus();
    
    // Add listener for authentication change events
    window.addEventListener('authChange', handleAuthChange);
    
    return () => {
      window.removeEventListener('authChange', handleAuthChange);
    };
  }, []);
  
  const checkAuthStatus = async () => {
    try {
      const response = await fetch('http://localhost:8085/api/v1/identity/me', {
        credentials: 'include' // Crucial for sending cookies
      });
      
      if (response.ok) {
        const user = await response.json();
        setCurrentUser(user);
        setIsAuthenticated(true);
      }
    } catch (error) {
      console.error('Error checking authentication status:', error);
    } finally {
      setLoading(false);
    }
  };
  
  const handleAuthChange = (event) => {
    const { authenticated, user } = event.detail;
    setIsAuthenticated(authenticated);
    setCurrentUser(user);
  };
  
  return (
    <AuthContext.Provider value={{ currentUser, isAuthenticated, loading }}>
      {children}
    </AuthContext.Provider>
  );
};
```

## API Integration

### HTTP Client with Credentials

All requests are configured to include cookies and handle authentication errors:

```jsx
// apiClient.js
import axios from 'axios';

const API_URL = 'http://localhost:8085/api/v1/identity';

const apiClient = axios.create({
  baseURL: API_URL,
  withCredentials: true, // Critical to include cookies in all requests
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
});

// Interceptor to handle 401 (unauthorized) errors
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    // If received 401 and not a refresh or login request
    if (error.response?.status === 401 && 
        !originalRequest._retry &&
        !originalRequest.url.includes('/refresh') &&
        !originalRequest.url.includes('/login')) {
      
      originalRequest._retry = true;
      
      try {
        // Try to renew the token
        await apiClient.post('/refresh');
        // Retry the original request
        return apiClient(originalRequest);
      } catch (refreshError) {
        // If refresh fails, dispatch logout event
        window.dispatchEvent(new CustomEvent('authChange', { 
          detail: { authenticated: false, user: null } 
        }));
        
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);

export default apiClient;
```

## Custom Events

The system uses custom events to synchronize authentication state throughout the application:

### `authChange` Event

```javascript
// Triggered after successful login
window.dispatchEvent(new CustomEvent('authChange', { 
  detail: { authenticated: true, user: userData } 
}));

// Triggered after logout
window.dispatchEvent(new CustomEvent('authChange', { 
  detail: { authenticated: false, user: null } 
}));
```

## Authentication Flows

### Login

```jsx
const login = async (email, password) => {
  try {
    setLoading(true);
    setError(null);
    
    const response = await apiClient.post('/login', { email, password });
    
    if (response.status === 200) {
      const userData = response.data;
      
      // Dispatch successful authentication event
      window.dispatchEvent(new CustomEvent('authChange', { 
        detail: { authenticated: true, user: userData } 
      }));
      
      return true;
    }
  } catch (error) {
    setError('Invalid credentials. Please try again.');
    console.error('Login error:', error);
    return false;
  } finally {
    setLoading(false);
  }
};
```

### Logout

```jsx
const handleLogout = async () => {
  try {
    await apiClient.post('/logout');
    
    // Remove tokens from global state
    setToken(null);
    setUser(null);
    
    // Dispatch event to notify the entire application about logout
    window.dispatchEvent(new CustomEvent('authChange', {
      detail: { authenticated: false, user: null }
    }));
    
    // Redirect to login page
    navigate('/login');
  } catch (error) {
    console.error('Error during logout:', error);
  }
};
```

## Route Protection Component

To ensure only authenticated users access certain routes:

```jsx
// ProtectedRoute.js
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export const ProtectedRoute = () => {
  const { isAuthenticated, loading } = useAuth();
  
  // Show loading indicator while checking authentication
  if (loading) {
    return <div>Loading...</div>;
  }
  
  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  // Render the protected route if authenticated
  return <Outlet />;
};
```

## Production Considerations

1. **Session Timeout**: Implement session expiration logic for inactivity
2. **PWA Integration**: Consider offline behavior and state synchronization
3. **Error Handling**: Improve UX for network or server failures
4. **Security**: Additional payload validation and input sanitization
5. **Performance**: Optimize refresh token checks to avoid unnecessary requests

## Manual Testing

Use the `test-auth.html` page to:

1. Verify if cookies are set after login
2. Test automatic token refresh
3. Confirm that logout removes cookies
4. Validate redirects to login for protected routes

## Best Practices

1. Always use `credentials: 'include'` or `withCredentials: true` in requests
2. Never try to access HttpOnly cookies via JavaScript (impossible by design)
3. Implement 401 error handling consistently
4. Use custom events to synchronize authentication state
5. Keep expiration times compatible between frontend and backend

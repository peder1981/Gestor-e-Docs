// src/contexts/AuthContext.js
import React, { createContext, useState, useEffect, useCallback } from 'react';
import apiClient from '../api/apiClient';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true); // Para verificar o estado inicial de auth

  const fetchUserProfile = useCallback(async () => {
    try {
      const response = await apiClient.get('/me');
      
      if (response.data && response.data.user) {
        setUser(response.data.user);
        setIsAuthenticated(true);
        return response.data.user;
      } else {
        console.warn('[AuthContext] Resposta não contém dados do usuário');
        setUser(null);
        setIsAuthenticated(false);
      }
    } catch (error) {
      if (error.response?.status === 401) {
      } else {
        console.warn('[AuthContext] Erro ao buscar perfil:', error.response?.data?.message || error.message);
      }
      setUser(null);
      setIsAuthenticated(false);
    }
    return null;
  }, []);

  const verifyAuth = useCallback(async () => {
    setIsLoading(true);
    await fetchUserProfile();
    setIsLoading(false);
  }, [fetchUserProfile]);

  useEffect(() => {
    // Ao montar o componente, verifica se há uma sessão válida tentando buscar o perfil
    const checkExistingAuth = async () => {
      try {
        await verifyAuth();
      } catch (error) {
        setUser(null);
        setIsAuthenticated(false);
        setIsLoading(false);
      }
    };

    checkExistingAuth();

    // Listener para eventos de mudança de autenticação
    const handleAuthChange = (event) => {
      if (event.detail.isAuthenticated !== undefined) {
        verifyAuth();
      }
    };

    window.addEventListener('authChange', handleAuthChange);
    return () => window.removeEventListener('authChange', handleAuthChange);
  }, [verifyAuth]);

  const login = async (email, password) => {
    setIsLoading(true);
    try {
      await apiClient.post('/login', { email, password });
      
      // Aguarda um momento para garantir que os cookies foram processados pelo navegador
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // Tenta buscar o perfil do usuário. Se funcionar, significa que os cookies foram definidos corretamente.
      const loggedInUser = await fetchUserProfile();
      
      if (!loggedInUser) {
        console.error('[AuthContext] Falha ao obter dados do usuário após o login');
        throw new Error('Falha ao obter dados do usuário após o login.');
      }
      
      setIsLoading(false);
      return loggedInUser;
    } catch (error) {
      setIsAuthenticated(false);
      setUser(null);
      setIsLoading(false);
      throw error;
    }
  };

  const register = async (name, email, password, role = 'user') => {
    // O registro não loga automaticamente o usuário neste fluxo, redireciona para login.
    try {
      const response = await apiClient.post('/register', { name, email, password, role });
      return response;
    } catch (error) {
      throw error;
    }
  };

  const logout = async () => {
    setIsLoading(true);
    try {
      await apiClient.post('/logout');
    } catch (error) {
      console.error('Logout failed on server:', error.response?.data?.message || error.message);
      // Mesmo que falhe no servidor, limpa o estado do cliente
    } finally {
      setUser(null);
      setIsAuthenticated(false);
      // Cookies HttpOnly são removidos pelo backend.
      setIsLoading(false);
      // Disparar evento para notificar outras partes da aplicação, se necessário
      // window.dispatchEvent(new CustomEvent('authChange', { detail: { isAuthenticated: false } }));
    }
  };

  return (
    <AuthContext.Provider value={{ user, isAuthenticated, isLoading, login, register, logout, verifyAuth, fetchUserProfile }}>
      {!isLoading && children} {/* Renderiza children apenas quando não está carregando o estado inicial de auth */}
      {isLoading && ( /* Pode mostrar um loader global aqui se preferir, ou deixar as rotas lidarem */
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
          {/* Idealmente, um componente de Spinner global */}
          <p>Carregando aplicação...</p>
        </div>
      )}
    </AuthContext.Provider>
  );
};

export default AuthContext;

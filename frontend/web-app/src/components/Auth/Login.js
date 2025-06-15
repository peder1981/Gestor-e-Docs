import React, { useState } from 'react';
import apiClient from '../../api/apiClient';
import { useNavigate } from 'react-router-dom';

const Login = () => {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({
        email: 'admin@example.com', // Credenciais padrão do admin
        password: 'password123',
    });
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');

    const { email, password } = formData;

    const onChange = (e) =>
        setFormData({ ...formData, [e.target.name]: e.target.value });

    const onSubmit = async (e) => {
        e.preventDefault();
        setMessage('');
        setError('');

        // Log do estado inicial
        console.log('=== Iniciando login ===');
        console.log('Cookies antes do login:', document.cookie);

        try {
            // Log da requisição
            console.log('Enviando requisição de login com:', {
                email,
                baseURL: apiClient.defaults.baseURL,
                withCredentials: apiClient.defaults.withCredentials,
                headers: apiClient.defaults.headers
            });

            const response = await apiClient.post('/login', {
                email: email,
                password: password
            });
            
            // Log do sucesso
            console.log('=== Login bem-sucedido ===');
            console.log('Response data:', response.data);
            console.log('Response headers:', response.headers);
            console.log('Response status:', response.status);
            console.log('Set-Cookie header:', response.headers['set-cookie']);
            console.log('Cookies após login:', document.cookie);

            setMessage('Login bem-sucedido! Redirecionando...');

            // Validar resposta
            if (!response.headers['set-cookie'] && !document.cookie) {
                console.warn('Aviso: Nenhum cookie definido após login bem-sucedido');
            }

            // Limpar formulário e redirecionar
            setFormData({ email: '', password: '' });
            window.dispatchEvent(new CustomEvent('authChange', { 
                detail: { 
                    isAuthenticated: true,
                    timestamp: new Date().toISOString()
                } 
            }));

            // Testar endpoint /me antes de redirecionar
            try {
                const meResponse = await apiClient.get('/me');
                console.log('Teste /me bem-sucedido:', meResponse.data);
                navigate('/'); // Só redireciona se /me funcionar
            } catch (meError) {
                console.error('Erro ao validar autenticação com /me:', meError);
                throw new Error('Falha ao validar autenticação');
            }

        } catch (err) {
            console.error('=== Erro no login ===');
            
            // Log detalhado do erro
            if (err.response) {
                console.error('Response error:', {
                    data: err.response.data,
                    status: err.response.status,
                    headers: err.response.headers
                });
            } else if (err.request) {
                console.error('Request error:', err.request);
            } else {
                console.error('Error message:', err.message);
            }

            // Log do estado final após erro
            console.log('Config da requisição:', err.config);
            console.log('Cookies após erro:', document.cookie);

            // Definir mensagem de erro apropriada
            if (err.response?.data?.error) {
                setError(err.response.data.error);
            } else if (err.message === 'Falha ao validar autenticação') {
                setError('Login realizado, mas falha ao validar autenticação');
            } else {
                setError('Erro ao fazer login. Verifique suas credenciais e tente novamente.');
            }
        }
    };

    return (
        <div>
            <h2>Login</h2>
            {message && <p style={{ color: 'green' }}>{message}</p>}
            {error && <p style={{ color: 'red' }}>{error}</p>}
            <form onSubmit={onSubmit}>
                <div>
                    <label htmlFor="email">Email:</label>
                    <input
                        type="email"
                        name="email"
                        id="email"
                        value={email}
                        onChange={onChange}
                        required
                    />
                </div>
                <div>
                    <label htmlFor="password">Senha:</label>
                    <input
                        type="password"
                        name="password"
                        id="password"
                        value={password}
                        onChange={onChange}
                        required
                    />
                </div>
                <button type="submit">Entrar</button>
            </form>
        </div>
    );
};

export default Login;

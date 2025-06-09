import React, { useState } from 'react';
import apiClient from '../../apiClient';
import { useNavigate } from 'react-router-dom';

const Login = () => {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({
        email: '',
        password: '',
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

        try {
            const apiUrl = process.env.REACT_APP_IDENTITY_API_URL || 'http://localhost:8085/api/v1/identity';
            const response = await apiClient.post(`/login`, {
                email: email,
                password: password
            });
            
            console.log('Login successful:', response.data);

            if (response.data && response.data.token) {
                localStorage.setItem('jwtToken', response.data.token);
                setMessage('Login bem-sucedido! Redirecionando...'); 
                // Limpar formulário antes de redirecionar
                setFormData({ email: '', password: '' });
                navigate('/'); // Redirecionar para a página inicial
            } else {
                setError('Token não recebido do servidor.');
                console.error('Login error: Token not found in response', response.data);
            }

        } catch (err) {
            if (err.response && err.response.data && err.response.data.error) {
                setError(err.response.data.error);
            } else {
                setError('Erro ao fazer login. Verifique suas credenciais.');
            }
            console.error('Login error:', err.response || err.message || err);
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

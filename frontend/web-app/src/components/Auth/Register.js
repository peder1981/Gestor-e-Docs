import React, { useState } from 'react';
import apiClient from '../../apiClient';

const Register = () => {
    const [formData, setFormData] = useState({
        name: '',
        email: '',
        password: '',
    });
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');

    const { name, email, password } = formData;

    const onChange = (e) =>
        setFormData({ ...formData, [e.target.name]: e.target.value });

    const onSubmit = async (e) => {
        e.preventDefault();
        setMessage('');
        setError('');

        // Validação simples de senha (exemplo)
        if (password.length < 6) {
            setError('A senha deve ter pelo menos 6 caracteres.');
            return;
        }

        try {
            const response = await apiClient.post(`/register`, {
                name: name,
                email: email,
                password: password
            });
            setMessage(response.data.message || 'Usuário registrado com sucesso!');
            // Limpar formulário após sucesso
            setFormData({ name: '', email: '', password: '' });
        } catch (err) {
            if (err.response && err.response.data && err.response.data.error) {
                setError(err.response.data.error);
            } else {
                setError('Erro ao registrar. Tente novamente.');
            }
            console.error('Registration error:', err.response || err.message || err);
        }
    };

    return (
        <div>
            <h2>Registrar Novo Usuário</h2>
            {message && <p style={{ color: 'green' }}>{message}</p>}
            {error && <p style={{ color: 'red' }}>{error}</p>}
            <form onSubmit={onSubmit}>
                <div>
                    <label htmlFor="name">Nome:</label>
                    <input
                        type="text"
                        name="name"
                        id="name"
                        value={name}
                        onChange={onChange}
                        required
                    />
                </div>
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
                        // minLength="6" // Removido para usar validação customizada
                        required
                    />
                </div>
                <button type="submit">Registrar</button>
            </form>
        </div>
    );
};

export default Register;

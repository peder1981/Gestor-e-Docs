import React, { useEffect, useState } from 'react';
import apiClient from '../apiClient'; // Ajustado o caminho para apiClient
import { useNavigate } from 'react-router-dom';

const Dashboard = () => {
    const [userData, setUserData] = useState(null);
    const [error, setError] = useState('');
    const navigate = useNavigate();

    useEffect(() => {
        const fetchUserData = async () => {
            try {
                // apiClient já tem a baseURL configurada para /api/v1/identity
                // A rota /me está em /api/v1/identity/me
                const response = await apiClient.get('/me'); 
                setUserData(response.data);
            } catch (err) {
                console.error('Error fetching user data:', err);
                setError('Falha ao carregar dados do usuário. Você pode precisar fazer login novamente.');
                if (err.response && err.response.status === 401) {
                    // Não precisamos mais remover o token do localStorage
                    // Apenas disparar o evento de mudança de autenticação
                    window.dispatchEvent(new CustomEvent('authChange', { detail: { isAuthenticated: false } }));
                    navigate('/login');
                }
            }
        };

        fetchUserData();
    }, [navigate]);

    if (error) {
        return <p style={{ color: 'red' }}>{error}</p>;
    }

    if (!userData) {
        return <p>Carregando dados do usuário...</p>;
    }

    return (
        <div>
            <h2>Dashboard</h2>
            <p>Bem-vindo! Seus dados foram carregados.</p>
            <p><strong>ID do Usuário:</strong> {userData.userID}</p>
            <p><strong>Mensagem do Servidor:</strong> {userData.message}</p>
            {/* Você pode exibir mais dados do usuário aqui se eles forem retornados pela API /me */}
        </div>
    );
};

export default Dashboard;

// src/pages/SettingsPage.js
import React, { useState, useContext } from 'react';
import { Box, Typography, Paper, Grid, TextField, Button, CircularProgress, Alert } from '@mui/material';
import apiClient from '../api/apiClient';
import NotificationContext from '../contexts/NotificationContext';

const ChangePasswordForm = () => {
    const [oldPassword, setOldPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const { showNotification } = useContext(NotificationContext);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (newPassword !== confirmPassword) {
            setError('A nova senha e a confirmação não coincidem.');
            return;
        }
        if (newPassword.length < 6) { // Exemplo de validação simples
            setError('A nova senha deve ter pelo menos 6 caracteres.');
            return;
        }
        
        setIsLoading(true);
        setError('');

        try {
            // Endpoint para mudança de senha. Ajuste conforme sua API.
            // A baseURL do apiClient aponta para /api/v1/identity, então o path está correto.
            await apiClient.patch('/users/me/password', {
                old_password: oldPassword,
                new_password: newPassword,
            });
            showNotification('Senha alterada com sucesso!', 'success');
            setOldPassword('');
            setNewPassword('');
            setConfirmPassword('');
        } catch (err) {
            setError(err.response?.data?.message || 'Ocorreu um erro ao alterar a senha.');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <Box component="form" onSubmit={handleSubmit}>
            <Typography variant="h6" gutterBottom>
                Alterar Senha
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
                Para sua segurança, recomendamos o uso de senhas fortes e únicas.
            </Typography>
            {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
            <TextField
                type="password"
                label="Senha Atual"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                required
                fullWidth
                margin="normal"
                disabled={isLoading}
            />
            <TextField
                type="password"
                label="Nova Senha"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
                fullWidth
                margin="normal"
                disabled={isLoading}
            />
            <TextField
                type="password"
                label="Confirmar Nova Senha"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                fullWidth
                margin="normal"
                disabled={isLoading}
            />
            <Button
                type="submit"
                variant="contained"
                sx={{ mt: 2 }}
                disabled={isLoading}
            >
                {isLoading ? <CircularProgress size={24} /> : 'Salvar Nova Senha'}
            </Button>
        </Box>
    );
};

const SettingsPage = () => {
  return (
    <Paper sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Configurações
      </Typography>
      <Grid container spacing={4}>
        <Grid item xs={12} md={8} lg={6}>
          <ChangePasswordForm />
        </Grid>
        <Grid item xs={12} md={4} lg={6}>
            <Typography variant="h6" gutterBottom>
                Outras Configurações
            </Typography>
            <Typography color="text.secondary">
                (Área para futuras configurações, como tema da aplicação, preferências de notificação, etc.)
            </Typography>
        </Grid>
      </Grid>
    </Paper>
  );
};

export default SettingsPage;

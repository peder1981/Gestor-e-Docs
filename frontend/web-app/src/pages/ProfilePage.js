// src/pages/ProfilePage.js
import React, { useContext } from 'react';
import { Box, Typography, Paper, Avatar, Grid, Divider } from '@mui/material';
import AuthContext from '../contexts/AuthContext';
import AccountCircleIcon from '@mui/icons-material/AccountCircle';
import { format } from 'date-fns';
import ptBR from 'date-fns/locale/pt-BR';

const ProfilePage = () => {
  const { user } = useContext(AuthContext);

  if (!user) {
    return (
      <Paper sx={{ p: 3 }}>
        <Typography>Não foi possível carregar os dados do usuário.</Typography>
      </Paper>
    );
  }

  return (
    <Paper sx={{ p: 3, maxWidth: 700, mx: 'auto' }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Perfil do Usuário
      </Typography>
      <Grid container spacing={3} alignItems="center">
        <Grid item>
          <Avatar
            alt={user.name}
            src={user.avatarUrl} // Supondo que o usuário possa ter um avatar
            sx={{ width: 100, height: 100, fontSize: '3rem' }}
          >
            {!user.avatarUrl && <AccountCircleIcon sx={{ width: 100, height: 100 }} />}
          </Avatar>
        </Grid>
        <Grid item>
          <Typography variant="h5" component="h2">{user.name}</Typography>
          <Typography variant="body1" color="text.secondary">{user.email}</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ textTransform: 'capitalize' }}>
            Cargo: {user.role || 'Usuário'}
          </Typography>
        </Grid>
      </Grid>
      <Divider sx={{ my: 3 }} />
      <Box>
        <Typography variant="h6" gutterBottom>Detalhes da Conta</Typography>
        <Typography variant="body1" sx={{ mb: 1 }}>
          <strong>ID do Usuário:</strong> {user.id}
        </Typography>
        <Typography variant="body1">
          <strong>Conta criada em:</strong> {user.created_at ? format(new Date(user.created_at), 'PPP', { locale: ptBR }) : 'N/A'}
        </Typography>
      </Box>
    </Paper>
  );
};

export default ProfilePage;

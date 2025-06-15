// src/pages/DashboardPage.js
import React, { useContext } from 'react';
import { Box, Typography, Paper, Grid, Link as MuiLink } from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import AuthContext from '../contexts/AuthContext';
import DescriptionIcon from '@mui/icons-material/Description';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import FolderIcon from '@mui/icons-material/Folder';

const DashboardPage = () => {
  const { user } = useContext(AuthContext);

  return (
    <Box sx={{ flexGrow: 1 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Bem-vindo(a) ao Dashboard, {user?.name || 'Usuário'}!
      </Typography>
      <Typography variant="subtitle1" color="text.secondary" gutterBottom>
        Gerencie seus documentos de forma eficiente e segura.
      </Typography>

      <Grid container spacing={3} sx={{ mt: 2 }}>
        {/* Card de Acesso Rápido */}
        <Grid item xs={12} md={4}>
          <Paper 
            component={RouterLink} 
            to="/documents"
            sx={{
              p: 3, 
              display: 'flex', 
              flexDirection: 'column', 
              alignItems: 'center', 
              height: '100%',
              textDecoration: 'none',
              '&:hover': {
                boxShadow: 6, // Aumenta a sombra no hover
              }
            }}
          >
            <DescriptionIcon sx={{ fontSize: 60, color: 'primary.main', mb: 2 }} />
            <Typography variant="h6" component="h2" gutterBottom>
              Meus Documentos
            </Typography>
            <Typography variant="body2" color="text.secondary" textAlign="center">
              Acesse, visualize e gerencie todos os seus documentos Markdown.
            </Typography>
          </Paper>
        </Grid>

        <Grid item xs={12} md={4}>
          <Paper 
            component={RouterLink} 
            to="/documents/upload" // Supondo que teremos uma página dedicada ou modal para upload
            sx={{
              p: 3, 
              display: 'flex', 
              flexDirection: 'column', 
              alignItems: 'center', 
              height: '100%',
              textDecoration: 'none',
              '&:hover': {
                boxShadow: 6,
              }
            }}
          >
            <CloudUploadIcon sx={{ fontSize: 60, color: 'secondary.main', mb: 2 }} />
            <Typography variant="h6" component="h2" gutterBottom>
              Novo Upload
            </Typography>
            <Typography variant="body2" color="text.secondary" textAlign="center">
              Envie novos documentos Markdown para a plataforma de forma rápida.
            </Typography>
          </Paper>
        </Grid>
        
        <Grid item xs={12} md={4}>
          <Paper 
            sx={{
              p: 3, 
              display: 'flex', 
              flexDirection: 'column', 
              alignItems: 'center', 
              height: '100%',
              backgroundColor: 'grey.100', // Exemplo de card diferente
              cursor: 'not-allowed' // Exemplo de funcionalidade futura
            }}
          >
            <FolderIcon sx={{ fontSize: 60, color: 'text.disabled', mb: 2 }} />
            <Typography variant="h6" component="h2" gutterBottom color="text.disabled">
              Categorias (Em Breve)
            </Typography>
            <Typography variant="body2" color="text.disabled" textAlign="center">
              Organize seus documentos em categorias personalizadas.
            </Typography>
          </Paper>
        </Grid>

        {/* Outras seções do dashboard podem ser adicionadas aqui */}
        {/* Ex: Documentos Recentes, Atividades, etc. */}
      </Grid>
    </Box>
  );
};

export default DashboardPage;

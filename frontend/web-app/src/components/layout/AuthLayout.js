// src/components/layout/AuthLayout.js
import React from 'react';
import { Outlet } from 'react-router-dom';
import { Box, Container, Paper, Typography, Avatar } from '@mui/material';
// import LogoIcon from '@mui/icons-material/Article'; // Exemplo, ou use seu SVG
import LogoSvg from '../../assets/logo.svg'; // Importando o SVG

const AuthLayout = () => {
  return (
    <Container component="main" maxWidth="xs">
      <Paper 
        elevation={3} 
        sx={{
          marginTop: 8,
          padding: 4,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          borderRadius: 2, // Usando o borderRadius do tema
        }}
      >
        <Avatar sx={{ m: 1, bgcolor: 'primary.main', width: 56, height: 56 }}>
          {/* <LogoIcon /> */}
          <img src={LogoSvg} alt="Logo" width="32" height="32" />
        </Avatar>
        <Typography component="h1" variant="h5" gutterBottom>
          Gestor-e-Docs
        </Typography>
        <Outlet /> {/* Formulários de Login/Registro serão renderizados aqui */}
      </Paper>
      <Box mt={5} mb={2} textAlign="center">
        <Typography variant="body2" color="text.secondary">
          © {new Date().getFullYear()} Gestor-e-Docs. Todos os direitos reservados.
        </Typography>
      </Box>
    </Container>
  );
};

export default AuthLayout;

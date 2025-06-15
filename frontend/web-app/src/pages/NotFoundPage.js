// src/pages/NotFoundPage.js
import React from 'react';
import { Box, Button, Container, Typography } from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';

const NotFoundPage = () => (
  <Box
    component="main"
    sx={{
      alignItems: 'center',
      display: 'flex',
      flexGrow: 1,
      minHeight: '100vh',
      justifyContent: 'center'
    }}
  >
    <Container maxWidth="md" sx={{textAlign: 'center'}}>
      <Typography
        align="center"
        color="textPrimary"
        variant="h1"
      >
        404: A página que você procura não está aqui
      </Typography>
      <Typography
        align="center"
        color="textPrimary"
        variant="subtitle2"
        sx={{mt: 2}}
      >
        Você pode ter tentado uma rota incorreta ou veio aqui por engano.
        Seja qual for o caso, tente usar a navegação.
      </Typography>
      <Box sx={{ textAlign: 'center', mt: 4 }}>
        <Button
          component={RouterLink}
          to="/dashboard"
          variant="contained"
          color="primary"
        >
          Voltar para o Dashboard
        </Button>
      </Box>
    </Container>
  </Box>
);

export default NotFoundPage;

// src/theme/theme.js
import { createTheme } from '@mui/material/styles';
import palette from './palette';
import typography from './typography';

const theme = createTheme({
  palette: palette,
  typography: typography,
  shape: {
    borderRadius: 8, // Bordas levemente arredondadas para um look moderno
  },
  components: {
    // Exemplo de override global para botões, se necessário
    MuiButton: {
      styleOverrides: {
        root: {
          // minWidth: 'auto', // Ajuste se necessário
        },
        containedPrimary: {
          // color: 'white', // Já definido pelo contrastText da paleta
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          boxShadow: '0px 1px 3px rgba(0, 0, 0, 0.1)', // Sombra sutil para o AppBar
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          borderRight: 'none', // Remove a borda padrão se estiver usando elevação
        },
      },
    },
    // Adicione outros overrides de componentes globais aqui conforme necessário
  },
});

export default theme;

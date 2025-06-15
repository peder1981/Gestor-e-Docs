// src/theme/palette.js
const palette = {
  primary: {
    main: '#1976d2', // Um azul corporativo clássico
    light: '#42a5f5',
    dark: '#1565c0',
    contrastText: '#ffffff',
  },
  secondary: {
    main: '#9e9e9e', // Um cinza neutro para elementos secundários
    light: '#bdbdbd',
    dark: '#757575',
    contrastText: '#000000',
  },
  background: {
    default: '#f4f6f8', // Um fundo levemente acinzentado, comum em dashboards
    paper: '#ffffff', // Fundo para componentes como Cards, Menus
  },
  text: {
    primary: '#212121', // Cor de texto principal, escura e legível
    secondary: '#757575', // Cor de texto secundária, para informações menos importantes
  },
  error: {
    main: '#d32f2f',
  },
  warning: {
    main: '#ffa000',
  },
  info: {
    main: '#1976d2',
  },
  success: {
    main: '#388e3c',
  },
};

export default palette;

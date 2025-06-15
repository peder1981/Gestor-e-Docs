// src/theme/typography.js
const typography = {
  fontFamily: [
    'Roboto', // Fonte padrão do Material Design, boa para interfaces
    '-apple-system',
    'BlinkMacSystemFont',
    '"Segoe UI"',
    'Arial',
    'sans-serif',
  ].join(','),
  h1: {
    fontWeight: 500,
    fontSize: '2.25rem', // 36px
    lineHeight: 1.2,
  },
  h2: {
    fontWeight: 500,
    fontSize: '1.875rem', // 30px
    lineHeight: 1.3,
  },
  h3: {
    fontWeight: 500,
    fontSize: '1.5rem', // 24px
    lineHeight: 1.4,
  },
  h4: {
    fontWeight: 500,
    fontSize: '1.25rem', // 20px
    lineHeight: 1.5,
  },
  h5: {
    fontWeight: 500,
    fontSize: '1.125rem', // 18px
    lineHeight: 1.5,
  },
  h6: {
    fontWeight: 500,
    fontSize: '1rem', // 16px
    lineHeight: 1.5,
  },
  subtitle1: {
    fontSize: '1rem',
    fontWeight: 400,
  },
  subtitle2: {
    fontSize: '0.875rem',
    fontWeight: 500,
  },
  body1: {
    fontSize: '1rem', // 16px
    fontWeight: 400,
    lineHeight: 1.5,
  },
  body2: {
    fontSize: '0.875rem', // 14px
    fontWeight: 400,
    lineHeight: 1.43,
  },
  button: {
    textTransform: 'capitalize', // Botões com texto capitalizado podem parecer mais sóbrios
    fontWeight: 500,
  },
  caption: {
    fontSize: '0.75rem',
    fontWeight: 400,
  },
  overline: {
    fontSize: '0.75rem',
    fontWeight: 500,
    textTransform: 'uppercase',
  },
};

export default typography;

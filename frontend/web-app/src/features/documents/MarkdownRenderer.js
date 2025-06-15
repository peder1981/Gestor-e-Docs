// src/features/documents/MarkdownRenderer.js
import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Paper, Typography, Box } from '@mui/material';
import { styled } from '@mui/material/styles';

// Estilizando o container do Markdown renderizado
const StyledMarkdownContainer = styled(Box)(({ theme }) => ({
  padding: theme.spacing(2, 3),
  fontFamily: theme.typography.fontFamily,
  fontSize: theme.typography.body1.fontSize,
  lineHeight: 1.7,
  color: theme.palette.text.primary,
  '& h1': {
    ...theme.typography.h3,
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(2),
    borderBottom: `1px solid ${theme.palette.divider}`,
    paddingBottom: theme.spacing(1),
  },
  '& h2': {
    ...theme.typography.h4,
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(1.5),
  },
  '& h3': {
    ...theme.typography.h5,
    marginTop: theme.spacing(2.5),
    marginBottom: theme.spacing(1),
  },
  '& h4': {
    ...theme.typography.h6,
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(1),
  },
  '& p': {
    marginBottom: theme.spacing(1.5),
  },
  '& a': {
    color: theme.palette.primary.main,
    textDecoration: 'none',
    '&:hover': {
      textDecoration: 'underline',
    },
  },
  '& img': {
    maxWidth: '100%',
    height: 'auto',
    borderRadius: theme.shape.borderRadius,
    marginTop: theme.spacing(1),
    marginBottom: theme.spacing(1),
  },
  '& ul, & ol': {
    marginBottom: theme.spacing(1.5),
    paddingLeft: theme.spacing(3),
  },
  '& li': {
    marginBottom: theme.spacing(0.5),
  },
  '& blockquote': {
    margin: theme.spacing(2, 0, 2, 2),
    padding: theme.spacing(1, 2),
    borderLeft: `4px solid ${theme.palette.primary.light}`,
    backgroundColor: theme.palette.action.hover,
    color: theme.palette.text.secondary,
    '& p': {
      marginBottom: 0,
    },
  },
  '& pre': {
    backgroundColor: '#f5f5f5', // Um cinza claro para blocos de código
    padding: theme.spacing(2),
    borderRadius: theme.shape.borderRadius,
    overflowX: 'auto',
    fontSize: '0.875rem',
    fontFamily: 'monospace',
  },
  '& code': { // Código inline
    backgroundColor: '#f5f5f5',
    padding: '0.1em 0.4em',
    borderRadius: theme.shape.borderRadius,
    fontFamily: 'monospace',
    fontSize: '0.875rem',
  },
  '& table': {
    width: '100%',
    borderCollapse: 'collapse',
    marginBottom: theme.spacing(2),
    border: `1px solid ${theme.palette.divider}`,
  },
  '& th, & td': {
    border: `1px solid ${theme.palette.divider}`,
    padding: theme.spacing(1, 1.5),
    textAlign: 'left',
  },
  '& th': {
    backgroundColor: theme.palette.grey[100],
    fontWeight: theme.typography.fontWeightBold,
  },
  '& hr': {
    border: 'none',
    borderTop: `1px solid ${theme.palette.divider}`,
    margin: theme.spacing(3, 0),
  },
}));

const MarkdownRenderer = ({ content }) => {
  if (!content) {
    return (
      <Paper sx={{ p: 3, textAlign: 'center' }}>
        <Typography color="text.secondary">Conteúdo do documento não disponível.</Typography>
      </Paper>
    );
  }

  return (
    <StyledMarkdownContainer component={Paper} elevation={0} sx={{border: (theme) => `1px solid ${theme.palette.divider}`}}>
      <ReactMarkdown remarkPlugins={[remarkGfm]}>
        {content}
      </ReactMarkdown>
    </StyledMarkdownContainer>
  );
};

export default MarkdownRenderer;

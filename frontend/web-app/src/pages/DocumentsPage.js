// src/pages/DocumentsPage.js
import React from 'react';
import DocumentList from '../features/documents/DocumentList';
import { Box, Typography } from '@mui/material';

const DocumentsPage = () => {
  return (
    <Box sx={{ flexGrow: 1 }}>
      {/* Pode haver um título de página aqui se o MainLayout não o fornecer dinamicamente */}
      {/* <Typography variant="h4" component="h1" gutterBottom>
        Gerenciador de Documentos
      </Typography> */}
      <DocumentList />
    </Box>
  );
};

export default DocumentsPage;

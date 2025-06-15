// src/pages/UploadPage.js
import React from 'react';
import DocumentUploadForm from '../features/documents/DocumentUploadForm';
import { Container } from '@mui/material';

const UploadPage = () => {
  return (
    // O MainLayout já fornece padding, então o Container pode ser opcional 
    // ou usado para restringir a largura máxima se desejado.
    <Container maxWidth="md"> 
      <DocumentUploadForm />
    </Container>
  );
};

export default UploadPage;

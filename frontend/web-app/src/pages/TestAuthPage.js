// src/pages/TestAuthPage.js
import React from 'react';
import { TestPanel } from '../test-auth';

const TestAuthPage = () => {
  // Um componente simples que renderiza o TestPanel em uma página dedicada
  return (
    <div style={{
      minHeight: '100vh',
      padding: '20px',
      backgroundColor: '#f5f5f5'
    }}>
      <h1 style={{
        textAlign: 'center',
        marginBottom: '30px',
        color: '#333'
      }}>
        Diagnóstico do Sistema Gestor-e-Docs
      </h1>
      
      <div style={{
        maxWidth: '1200px',
        margin: '0 auto'
      }}>
        <TestPanel fullPage={true} />
      </div>
      
      <footer style={{
        marginTop: '50px',
        textAlign: 'center',
        color: '#666',
        fontSize: '14px'
      }}>
        <p>Ferramenta de diagnóstico para o sistema Gestor-e-Docs</p>
        <p>Use os botões abaixo para testar a autenticação e as ações de documentos</p>
      </footer>
    </div>
  );
};

export default TestAuthPage;

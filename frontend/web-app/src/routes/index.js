// src/routes/index.js
import React, { useContext } from 'react';
import { Routes, Route, Navigate, Outlet } from 'react-router-dom';
import AuthContext from '../contexts/AuthContext';

// Páginas
import LoginPage from '../pages/LoginPage';
import RegisterPage from '../pages/RegisterPage';
import DashboardPage from '../pages/DashboardPage'; 
import DocumentsPage from '../pages/DocumentsPage'; 
import UploadPage from '../pages/UploadPage'; // Nova página de Upload
import DocumentViewerPage from '../pages/DocumentViewerPage'; // Nova página
import ProfilePage from '../pages/ProfilePage'; // Nova página
import SettingsPage from '../pages/SettingsPage'; // Nova página
import NotFoundPage from '../pages/NotFoundPage'; // Usaremos um componente real agora

// Layouts
import AuthLayout from '../components/layout/AuthLayout';
import MainLayout from '../components/layout/MainLayout';

// Componentes de placeholder para páginas restantes
const PlaceholderComponent = ({ title }) => (
  <div style={{ padding: '20px', border: '1px dashed #ccc', margin: '20px' }}>
    <h2>{title}</h2>
    <p>Conteúdo da página virá aqui.</p>
  </div>
);

// const LoginPage = () => <PlaceholderComponent title="Login Page" />;
// const RegisterPage = () => <PlaceholderComponent title="Register Page" />;
// const DashboardPage = () => <PlaceholderComponent title="Dashboard Page (Protegida)" />;
// const DocumentsPage = () => <PlaceholderComponent title="Documents Page (Protegida)" />;


const ProtectedRoute = () => {
  const { isAuthenticated, isLoading } = useContext(AuthContext);

  if (isLoading) { 
    // O loader global no AuthContext já deve cobrir isso, mas como fallback:
    return <div style={{textAlign: 'center', padding: '50px'}}>Verificando acesso...</div>;
  }

  return isAuthenticated ? <Outlet /> : <Navigate to="/auth/login" replace />;
};

const AppRouter = () => {
  return (
    <Routes>
      {/* Rotas de Autenticação (públicas, mas com layout próprio) */}
      <Route path="auth" element={<AuthLayout />}>
        <Route path="login" element={<LoginPage />} />
        <Route path="register" element={<RegisterPage />} />
      </Route>

      {/* Rotas Protegidas (dentro do MainLayout) */}
      <Route element={<MainLayout />}>
        <Route element={<ProtectedRoute />}>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<DashboardPage />} />
          <Route path="documents" element={<DocumentsPage />} />
          <Route path="documents/upload" element={<UploadPage />} /> {/* Nova rota */}
          <Route path="documents/:documentId" element={<DocumentViewerPage />} /> {/* Nova rota */}
          <Route path="profile" element={<ProfilePage />} />
          <Route path="settings" element={<SettingsPage />} />
          {/* Outras rotas protegidas aqui (ex: /profile, /settings) */}
        </Route>
      </Route>

      {/* Rota para página não encontrada */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
};

export default AppRouter;

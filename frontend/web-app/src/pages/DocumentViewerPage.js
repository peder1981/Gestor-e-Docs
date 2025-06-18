// src/pages/DocumentViewerPage.js
import React, { useState, useEffect, useContext } from 'react';
import { useParams, useNavigate, Link as RouterLink } from 'react-router-dom';
import {
  Box,
  Typography,
  CircularProgress,
  Alert,
  Paper,
  Button,
  Breadcrumbs,
  Link as MuiLink,
  Tooltip,
  IconButton
} from '@mui/material';
import documentApiClient from '../api/documentApiClient';
import AuthContext from '../contexts/AuthContext';
import MarkdownRenderer from '../features/documents/MarkdownRenderer';
import HomeIcon from '@mui/icons-material/Home';
import DescriptionIcon from '@mui/icons-material/Description';
import GetAppIcon from '@mui/icons-material/GetApp';
import EditIcon from '@mui/icons-material/Edit'; // Para futura edição
import { format } from 'date-fns';
import ptBR from 'date-fns/locale/pt-BR';


const DocumentViewerPage = () => {
  const { documentId } = useParams();
  const [document, setDocument] = useState(null);
  const [content, setContent] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const { isAuthenticated } = useContext(AuthContext);
  const navigate = useNavigate();

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/auth/login');
      return;
    }
    if (!documentId) return;

    const fetchDocumentDetails = async () => {
      setIsLoading(true);
      setError('');
      try {
        // Busca os metadados e conteúdo do documento
        const response = await documentApiClient.get(`/${documentId}`);
        setDocument(response.data);
        
        // O conteúdo já vem na resposta da API no campo 'content'
        setContent(response.data.content || '');

      } catch (err) {
        setError(err.response?.data?.message || err.message || 'Erro ao buscar detalhes do documento.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchDocumentDetails();
  }, [documentId, isAuthenticated, navigate]);

  const handleDownload = async () => {
  if (!document) return;
  try {
      // Usar o endpoint correto que retorna o arquivo diretamente
      const downloadUrl = `${documentApiClient.defaults.baseURL}/${documentId}/download/file`;
      console.log('Iniciando download via fetch:', downloadUrl);
      
      // Usar fetch com opções para incluir credenciais
      const response = await fetch(downloadUrl, {
          method: 'GET',
          credentials: 'include', // Importante: inclui cookies para autenticação
      });
      
      if (!response.ok) {
          throw new Error(`Erro no download: ${response.status} ${response.statusText}`);
      }
      
      // Obter o blob do documento
      const blob = await response.blob();
      
      // Criar URL para o blob
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      
      // Tenta pegar o nome do arquivo do header Content-Disposition, ou usa o nome do documento
      const contentDisposition = response.headers.get('content-disposition');
      let fileName = document.name || 'documento.md';
      if (contentDisposition) {
          const fileNameMatch = contentDisposition.match(/filename\*?=['"]?(?:UTF-\d['"]*)?([^;"\n]*)/i);
          if (fileNameMatch && fileNameMatch[1]) {
              fileName = decodeURIComponent(fileNameMatch[1]);
          }
      }
      
      link.setAttribute('download', fileName);
      document.body.appendChild(link);
      link.click();
      link.parentNode.removeChild(link);
      window.URL.revokeObjectURL(url);
  } catch (err) {
      setError('Erro ao tentar baixar o arquivo: ' + err.message);
  }
};


  if (isLoading) {
    return <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 'calc(100vh - 128px)' }}><CircularProgress size={60} /></Box>;
  }

  if (error) {
    return <Alert severity="error" sx={{ m: 3 }}>{error}</Alert>;
  }

  if (!document) {
    return <Alert severity="info" sx={{ m: 3 }}>Documento não encontrado.</Alert>;
  }

  return (
    <Box sx={{ flexGrow: 1 }}>
      <Breadcrumbs aria-label="breadcrumb" sx={{ mb: 2 }}>
        <MuiLink component={RouterLink} underline="hover" color="inherit" to="/">
          <HomeIcon sx={{ mr: 0.5 }} fontSize="inherit" />
          Início
        </MuiLink>
        <MuiLink component={RouterLink} underline="hover" color="inherit" to="/documents">
          <DescriptionIcon sx={{ mr: 0.5 }} fontSize="inherit" />
          Documentos
        </MuiLink>
        <Typography color="text.primary">{document.name || 'Detalhes'}</Typography>
      </Breadcrumbs>

      <Paper sx={{ p: 2, mb: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb:2 }}>
            <Box>
                <Typography variant="h4" component="h1" gutterBottom>
                    {document.name}
                </Typography>
                <Typography variant="caption" color="text.secondary" display="block">
                    ID: {document.id}
                </Typography>
                <Typography variant="caption" color="text.secondary" display="block">
                    Tipo: {document.content_type} | Tamanho: {document.size ? `${(document.size / 1024).toFixed(2)} KB` : 'N/A'}
                </Typography>
                <Typography variant="caption" color="text.secondary" display="block">
                    Criado em: {document.created_at ? format(new Date(document.created_at), 'Pp', { locale: ptBR }) : 'N/A'}
                </Typography>
                <Typography variant="caption" color="text.secondary" display="block">
                    Última Modificação: {document.updated_at ? format(new Date(document.updated_at), 'Pp', { locale: ptBR }) : 'N/A'}
                </Typography>
            </Box>
            <Box sx={{display: 'flex', flexDirection: {xs: 'column', sm: 'row'}, gap: 1, mt: {xs: 1, sm: 0}}}>
                <Button variant="outlined" startIcon={<GetAppIcon />} onClick={handleDownload}>
                    Baixar
                </Button>
                {/* Botão de Editar (funcionalidade futura) */}
                {/* <Tooltip title="Editar Metadados (Em breve)">
                    <IconButton disabled>
                        <EditIcon />
                    </IconButton>
                </Tooltip> */}
            </Box>
        </Box>
      </Paper>

      <MarkdownRenderer content={content} />
    </Box>
  );
};

export default DocumentViewerPage;

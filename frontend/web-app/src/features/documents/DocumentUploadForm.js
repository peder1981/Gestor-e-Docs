// src/features/documents/DocumentUploadForm.js
import React, { useState, useCallback, useContext, useEffect } from 'react';
import {
  Box,
  Button,
  Typography,
  CircularProgress,
  Alert,
  LinearProgress,
  Paper,
  TextField
} from '@mui/material';
import { useDropzone } from 'react-dropzone';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import documentApiClient from '../../api/documentApiClient';
import AuthContext from '../../contexts/AuthContext';
import { useNavigate } from 'react-router-dom';


const DocumentUploadForm = () => {
  const [files, setFiles] = useState([]);
  const [description, setDescription] = useState(''); // Campo opcional
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [uploadProgress, setUploadProgress] = useState(0);
  const { isAuthenticated } = useContext(AuthContext);
  const navigate = useNavigate();

  const onDrop = useCallback((acceptedFiles) => {
    setFiles(prevFiles => [...prevFiles, ...acceptedFiles.map(file => Object.assign(file, {
      preview: URL.createObjectURL(file)
    }))]);
    setError('');
    setSuccess('');
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'text/markdown': ['.md', '.markdown'],
      // Adicione outros tipos se necessário, ex: 'application/pdf': ['.pdf']
    },
    multiple: true // Permite múltiplos arquivos, ajuste se necessário
  });

  const handleRemoveFile = (fileName) => {
    setFiles(prevFiles => prevFiles.filter(file => file.name !== fileName));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!isAuthenticated) {
      setError('Você precisa estar logado para enviar documentos.');
      return;
    }
    if (files.length === 0) {
      setError('Por favor, selecione ao menos um arquivo para enviar.');
      return;
    }

    setIsLoading(true);
    setError('');
    setSuccess('');
    setUploadProgress(0);

    const formData = new FormData();
    files.forEach(file => {
      formData.append('files', file); // O backend deve esperar um campo 'files'
    });
    if (description) {
      formData.append('description', description); // Campo opcional
    }

    try {
      const response = await documentApiClient.post('/', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent) => {
          const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          setUploadProgress(percentCompleted);
        },
      });
      
      setSuccess(`Upload bem-sucedido! ${files.length} arquivo(s) enviado(s). Você será redirecionado.`);
      setFiles([]);
      setDescription('');
      setUploadProgress(100);
      setTimeout(() => {
        navigate('/documents'); // Redireciona para a lista de documentos
      }, 2000);

    } catch (err) {
      setError(err.response?.data?.message || err.message || 'Erro ao enviar o(s) arquivo(s).');
      setUploadProgress(0);
    } finally {
      setIsLoading(false);
    }
  };

  const thumbs = files.map(file => (
    <Box key={file.name} sx={{ display: 'inline-flex', borderRadius: 2, border: '1px solid #eaeaea', mb: 1, mr: 1, width: 100, height: 100, p: 1, boxSizing: 'border-box' }}>
      <Box sx={{ display: 'flex', minWidth: 0, overflow: 'hidden', flexDirection: 'column', alignItems: 'center' }}>
        {/* <img src={file.preview} style={{ display: 'block', width: 'auto', height: '100%' }} onLoad={() => { URL.revokeObjectURL(file.preview) }} /> */}
        <Typography variant="caption" sx={{overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', width: '100%', textAlign: 'center'}}>{file.name}</Typography>
        <Button size="small" color="error" onClick={() => handleRemoveFile(file.name)}>Remover</Button>
      </Box>
    </Box>
  ));

  useEffect(() => {
    // Limpar previews para evitar memory leaks
    return () => files.forEach(file => URL.revokeObjectURL(file.preview));
  }, [files]);

  return (
    <Paper sx={{ p: 3 }}>
      <Typography variant="h5" component="h2" gutterBottom>
        Upload de Novos Documentos
      </Typography>
      <Box component="form" onSubmit={handleSubmit} sx={{ mt: 1 }}>
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}

        <Box
          {...getRootProps()}
          sx={{
            p: 3,
            border: `2px dashed ${isDragActive ? 'primary.main' : 'grey.500'}`, 
            borderRadius: 1,
            textAlign: 'center',
            cursor: 'pointer',
            mb: 2,
            backgroundColor: isDragActive ? 'action.hover' : 'transparent'
          }}
        >
          <input {...getInputProps()} />
          <CloudUploadIcon sx={{ fontSize: 48, color: 'grey.600', mb:1 }} />
          {isDragActive ? (
            <Typography>Solte os arquivos aqui ...</Typography>
          ) : (
            <Typography>Arraste e solte arquivos aqui, ou clique para selecionar (somente .md)</Typography>
          )}
        </Box>

        {files.length > 0 && (
          <Box sx={{mb: 2}}>
            <Typography variant="subtitle1">Arquivos selecionados:</Typography>
            <aside style={{ display: 'flex', flexDirection: 'row', flexWrap: 'wrap', marginTop: 1 }}>
              {thumbs}
            </aside>
          </Box>
        )}

        <TextField
          label="Descrição (Opcional)"
          fullWidth
          multiline
          rows={3}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          sx={{ mb: 2 }}
          disabled={isLoading}
        />

        {isLoading && (
          <Box sx={{ width: '100%', mb: 2 }}>
            <LinearProgress variant="determinate" value={uploadProgress} />
            <Typography variant="caption" display="block" textAlign="center">{`${uploadProgress}%`}</Typography>
          </Box>
        )}

        <Button
          type="submit"
          fullWidth
          variant="contained"
          disabled={isLoading || files.length === 0 || !!success}
          startIcon={isLoading ? <CircularProgress size={20} color="inherit" /> : <CloudUploadIcon />}
        >
          {isLoading ? 'Enviando...' : `Enviar ${files.length} Arquivo(s)`}
        </Button>
      </Box>
    </Paper>
  );
};

export default DocumentUploadForm;

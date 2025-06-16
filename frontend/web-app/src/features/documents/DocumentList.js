// src/features/documents/DocumentList.js
import React, { useState, useEffect, useCallback, useContext } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  TablePagination,
  CircularProgress,
  Typography,
  Alert,
  IconButton,
  Tooltip,
  Box,
  TextField,
  InputAdornment,
  Button,
  Menu,
  MenuItem,
  ListItemIcon, 
  ListItemText,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions
} from '@mui/material';
import { Link as RouterLink, useNavigate } from 'react-router-dom';
import documentApiClient from '../../api/documentApiClient'; 
import AuthContext from '../../contexts/AuthContext';
import NotificationContext from '../../contexts/NotificationContext'; 
import ConfirmationDialog from '../../components/common/ConfirmationDialog'; 
import { format } from 'date-fns'; 
import { ptBR } from 'date-fns/locale'; 
import EmptyState from '../../components/common/EmptyState'; 

// Ícones
import VisibilityIcon from '@mui/icons-material/Visibility';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import GetAppIcon from '@mui/icons-material/GetApp';
import SearchIcon from '@mui/icons-material/Search';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import AddIcon from '@mui/icons-material/Add';
import InboxIcon from '@mui/icons-material/Inbox'; 


const DocumentList = () => {
  const [documents, setDocuments] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [totalDocuments, setTotalDocuments] = useState(0);
  const [searchTerm, setSearchTerm] = useState('');
  const [anchorEl, setAnchorEl] = useState(null);
  const [selectedDoc, setSelectedDoc] = useState(null);
  const [docToDelete, setDocToDelete] = useState(null); 
  const [openConfirmDialog, setOpenConfirmDialog] = useState(false);
  const [openEditModal, setOpenEditModal] = useState(false);
  const [editFormData, setEditFormData] = useState({
    title: '',
    description: '',
    tags: [],
    category: ''
  });

  const { isAuthenticated } = useContext(AuthContext);
  const { showNotification } = useContext(NotificationContext); 
  const navigate = useNavigate();

  const fetchDocuments = useCallback(async () => {
    if (!isAuthenticated) return;
    setIsLoading(true);
    setError('');
    try {
      const response = await documentApiClient.get('/list', {
        params: {
          page: page + 1, 
          limit: rowsPerPage,
          search: searchTerm || undefined, 
        },
      });
      setDocuments(response.data.documents || []);
      setTotalDocuments(response.data.total_documents || 0);
    } catch (err) {
      setError(err.response?.data?.message || err.message || 'Erro ao buscar documentos.');
      setDocuments([]);
      setTotalDocuments(0);
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated, page, rowsPerPage, searchTerm]);

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleSearchChange = (event) => {
    setSearchTerm(event.target.value);
    setPage(0); 
  };

  const handleMenuOpen = (event, doc) => {
    setAnchorEl(event.currentTarget);
    setSelectedDoc(doc);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setSelectedDoc(null);
  };

  // Função para download do documento
  const handleDownload = async (doc) => {
    handleMenuClose();
    try {
      setIsLoading(true);
      
      // Fazer download usando fetch para evitar problemas de redirects e DNS
      const downloadUrl = `${documentApiClient.defaults.baseURL}/${doc.id}/download/file`;
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
      
      // Tentar extrair o nome do arquivo do cabeçalho Content-Disposition
      let filename = doc.title;
      const disposition = response.headers.get('content-disposition');
      if (disposition && disposition.includes('attachment')) {
        const filenameMatch = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(disposition);
        if (filenameMatch && filenameMatch[1]) {
          filename = filenameMatch[1].replace(/['"]]/g, '');
        }
      }
      
      // Adicionar extensão se não tiver
      if (!filename.includes('.')) {
        const contentType = response.headers.get('content-type');
        if (contentType) {
          if (contentType.includes('pdf')) {
            filename += '.pdf';
          } else if (contentType.includes('jpeg') || contentType.includes('jpg')) {
            filename += '.jpg';
          } else if (contentType.includes('png')) {
            filename += '.png';
          } else if (contentType.includes('msword') || contentType.includes('doc')) {
            filename += '.doc';
          } else if (contentType.includes('officedocument.wordprocessingml')) {
            filename += '.docx';
          }
        }
      }
      
      // Criar URL para o blob
      const url = window.URL.createObjectURL(blob);
      
      // Criar elemento para download
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      
      // Simular clique e depois remover o link
      link.click();
      document.body.removeChild(link);
      
      // Liberar a URL do objeto quando não for mais necessária
      setTimeout(() => URL.revokeObjectURL(url), 100);
      
      showNotification('Download concluído com sucesso!', 'success');
    } catch (err) {
      console.error('Erro no download:', err);
      showNotification(`Erro ao tentar baixar o arquivo: ${err.message}`, 'error');
    } finally {
      setIsLoading(false);
    }
  };

  // Função para abrir modal de edição
  const handleOpenEditModal = (doc) => {
    handleMenuClose();
    setSelectedDoc(doc);
    // Preencher formulário com dados do documento
    setEditFormData({
      title: doc.title || '',
      description: doc.description || '',
      tags: doc.tags || [],
      category: doc.category || ''
    });
    setOpenEditModal(true);
  };

  // Função para salvar metadados editados
  const handleSaveMetadata = async () => {
    if (!selectedDoc) return;
    try {
      setIsLoading(true);
      const response = await documentApiClient.put(`/${selectedDoc.id}`, editFormData);
      showNotification('Documento atualizado com sucesso!', 'success');
      setOpenEditModal(false);
      fetchDocuments(); // Recarregar lista
    } catch (err) {
      showNotification(`Erro ao atualizar documento: ${err.response?.data?.message || err.message}`, 'error');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteRequest = (doc) => {
    setDocToDelete(doc); // Usar o novo estado para documento a ser excluído
    console.log('Documento definido para exclusão:', doc.id);
    setOpenConfirmDialog(true);
    handleMenuClose();
  };

  const handleConfirmDelete = async () => {
    if (!docToDelete) {
      console.log('Nenhum documento selecionado para exclusão');
      return;
    }
    
    console.log('Iniciando exclusão do documento:', docToDelete.id);
    setOpenConfirmDialog(false);
    setIsLoading(true);
    
    try {
      console.log('Enviando solicitação de exclusão para:', `/${docToDelete.id}`);
      const response = await documentApiClient.delete(`/${docToDelete.id}`);
      console.log('Resposta da exclusão:', response.data);
      
      showNotification('Documento excluído com sucesso!', 'success');
      fetchDocuments(); 
    } catch (err) {
      console.error('Erro detalhado na exclusão:', err);
      showNotification(`Erro ao excluir documento: ${err.response?.data?.message || err.message}`, 'error');
    } finally {
      setIsLoading(false);
      setDocToDelete(null); // Limpar o documento para exclusão
    }
  };

  if (!isAuthenticated && !isLoading) {
    return <Alert severity="warning">Você precisa estar logado para ver seus documentos.</Alert>;
  }

  return (
    <Paper sx={{ width: '100%', overflow: 'hidden', p: 2 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5" component="h2">
          Meus Documentos
        </Typography>
        <Button 
          variant="contained" 
          startIcon={<AddIcon />} 
          component={RouterLink} 
          to="/documents/upload"
        >
          Novo Documento
        </Button>
      </Box>
      <TextField
        fullWidth
        variant="outlined"
        placeholder="Buscar documentos por nome..."
        value={searchTerm}
        onChange={handleSearchChange}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon />
            </InputAdornment>
          ),
        }}
        sx={{ mb: 2 }}
      />

      {isLoading && <Box sx={{ display: 'flex', justifyContent: 'center', my: 3 }}><CircularProgress /></Box>}
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      
      {!isLoading && !error && documents.length === 0 && (
        <EmptyState 
          icon={<InboxIcon fontSize="inherit" />}
          title="Nenhum documento encontrado"
          description="Parece que você ainda não enviou nenhum documento. Comece agora mesmo!"
          actionText="Enviar primeiro documento"
          actionTo="/documents/upload"
        />
      )}

      {!isLoading && !error && documents.length > 0 && (
        <TableContainer>
          <Table stickyHeader aria-label="sticky table">
            <TableHead>
              <TableRow>
                <TableCell>Nome do Arquivo</TableCell>
                <TableCell>Tipo</TableCell>
                <TableCell>Tamanho</TableCell>
                <TableCell>Última Modificação</TableCell>
                <TableCell align="right">Ações</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {documents.map((doc) => (
                <TableRow hover role="checkbox" tabIndex={-1} key={doc.id || doc.title}>
                  <TableCell component="th" scope="row">
                    {doc.title}
                  </TableCell>
                  <TableCell>{doc.content_type || 'N/A'}</TableCell>
                  <TableCell>{doc.size ? `${(doc.size / 1024).toFixed(2)} KB` : 'N/A'}</TableCell>
                  <TableCell>
                    {doc.updated_at ? format(new Date(doc.updated_at), 'Pp', { locale: ptBR }) : 'N/A'}
                  </TableCell>
                  <TableCell align="right">
                    <Tooltip title="Ver Detalhes">
                      <IconButton onClick={() => navigate(`/documents/${doc.id}`)}>
                        <VisibilityIcon />
                      </IconButton>
                    </Tooltip>
                    <IconButton aria-controls={`actions-menu-${doc.id}`} aria-haspopup="true" onClick={(e) => handleMenuOpen(e, doc)}>
                        <MoreVertIcon />
                    </IconButton>
                    <Menu
                        id={`actions-menu-${doc.id}`}
                        anchorEl={anchorEl}
                        open={Boolean(anchorEl) && selectedDoc?.id === doc.id}
                        onClose={handleMenuClose}
                    >
                        <MenuItem onClick={() => handleDownload(doc)}>
                            <ListItemIcon><GetAppIcon fontSize="small" /></ListItemIcon>
                            <ListItemText>Baixar</ListItemText>
                        </MenuItem>
                        <MenuItem onClick={() => handleOpenEditModal(doc)}>
                            <ListItemIcon><EditIcon fontSize="small" /></ListItemIcon>
                            <ListItemText>Editar Metadados</ListItemText>
                        </MenuItem>
                        <MenuItem onClick={() => handleDeleteRequest(doc)} sx={{ color: 'error.main' }}>
                            <ListItemIcon><DeleteIcon fontSize="small" sx={{ color: 'error.main' }} /></ListItemIcon>
                            <ListItemText>Excluir</ListItemText>
                        </MenuItem>
                    </Menu>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
      <TablePagination
        rowsPerPageOptions={[5, 10, 25]}
        component="div"
        count={totalDocuments}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        labelRowsPerPage="Itens por página:"
        labelDisplayedRows={({ from, to, count }) => `${from}-${to} de ${count}`}
      />
      <ConfirmationDialog 
        open={openConfirmDialog}
        onClose={() => setOpenConfirmDialog(false)}
        onConfirm={handleConfirmDelete}
        title="Confirmar Exclusão"
        contentText={`Tem certeza que deseja excluir o documento "${selectedDoc?.title}"? Esta ação não poderá ser desfeita.`}
        confirmText="Excluir"
      />

      {/* Modal de Edição de Metadados */}
      <Dialog open={openEditModal} onClose={() => setOpenEditModal(false)} maxWidth="md" fullWidth>
        <DialogTitle>Editar Metadados do Documento</DialogTitle>
        <DialogContent>
          <TextField
            margin="dense"
            label="Título"
            type="text"
            fullWidth
            value={editFormData.title}
            onChange={(e) => setEditFormData({...editFormData, title: e.target.value})}
          />
          <TextField
            margin="dense"
            label="Descrição"
            type="text"
            fullWidth
            multiline
            rows={4}
            value={editFormData.description}
            onChange={(e) => setEditFormData({...editFormData, description: e.target.value})}
          />
          <TextField
            margin="dense"
            label="Categoria"
            type="text"
            fullWidth
            value={editFormData.category}
            onChange={(e) => setEditFormData({...editFormData, category: e.target.value})}
          />
          <TextField
            margin="dense"
            label="Tags (separadas por vírgula)"
            type="text"
            fullWidth
            value={Array.isArray(editFormData.tags) ? editFormData.tags.join(', ') : ''}
            onChange={(e) => setEditFormData({...editFormData, tags: e.target.value.split(',').map(tag => tag.trim()).filter(tag => tag)})}
            helperText="Exemplo: financeiro, contrato, 2025"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenEditModal(false)}>Cancelar</Button>
          <Button onClick={handleSaveMetadata} color="primary" variant="contained">
            Salvar
          </Button>
        </DialogActions>
      </Dialog>
    </Paper>
  );
};

export default DocumentList;

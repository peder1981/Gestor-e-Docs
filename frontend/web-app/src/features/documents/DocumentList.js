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
  ListItemText 
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
  const [openConfirmDialog, setOpenConfirmDialog] = useState(false);

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

  const handleDeleteRequest = (doc) => {
    setSelectedDoc(doc); 
    setOpenConfirmDialog(true);
    handleMenuClose();
  };

  const handleConfirmDelete = async () => {
    if (!selectedDoc) return;
    setOpenConfirmDialog(false);
    setIsLoading(true); 
    try {
      await documentApiClient.delete(`/${selectedDoc.id}`);
      showNotification('Documento excluído com sucesso!', 'success');
      fetchDocuments(); 
    } catch (err) {
      showNotification(`Erro ao excluir documento: ${err.response?.data?.message || err.message}`, 'error');
    } finally {
      setIsLoading(false);
      setSelectedDoc(null);
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
                        <MenuItem onClick={() => { /* Implementar download */ handleMenuClose(); showNotification('Download não implementado.', 'info'); }}>
                            <ListItemIcon><GetAppIcon fontSize="small" /></ListItemIcon>
                            <ListItemText>Baixar</ListItemText>
                        </MenuItem>
                        <MenuItem onClick={() => { /* Implementar edição */ handleMenuClose(); showNotification('Edição não implementada.', 'info'); }}>
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
    </Paper>
  );
};

export default DocumentList;

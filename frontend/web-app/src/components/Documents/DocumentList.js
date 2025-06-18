import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import './Documents.css';
import documentApiClient from '../../api/documentApiClient';

const DocumentList = () => {
  const [documents, setDocuments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState({
    status: '',
    category: '',
  });
  
  // Estados para modais de ação
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [currentDocument, setCurrentDocument] = useState(null);
  const [editData, setEditData] = useState({
    title: '',
    tags: [],
    categories: [],
    status: ''
  });
  const [newTag, setNewTag] = useState('');
  const [newCategory, setNewCategory] = useState('');
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState(null);
  const [actionSuccess, setActionSuccess] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    fetchDocuments();
  }, [filters]);

  const fetchDocuments = async () => {
    try {
      setLoading(true);
      
      // Construir URL com parâmetros de busca e filtros
      let url = '/api/v1/documents/list?';
      if (searchTerm) url += `query=${encodeURIComponent(searchTerm)}&`;
      if (filters.status) url += `status=${encodeURIComponent(filters.status)}&`;
      if (filters.category) url += `categories=${encodeURIComponent(filters.category)}&`;
      
      const response = await fetch(url, {
        method: 'GET',
        credentials: 'include',  // Importante para enviar cookies
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        // Se o erro for 401 (não autorizado), redirecionar para login
        if (response.status === 401) {
          navigate('/login', { 
            state: { message: 'Sua sessão expirou. Por favor, faça login novamente.' } 
          });
          return;
        }
        throw new Error(`Erro ao buscar documentos: ${response.statusText}`);
      }

      const data = await response.json();
      setDocuments(data.documents || []);
    } catch (err) {
      console.error('Erro ao buscar documentos:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e) => {
    e.preventDefault();
    fetchDocuments();
  };

  const handleFilterChange = (name, value) => {
    setFilters(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleViewDocument = (id) => {
    navigate(`/documents/${id}`);
  };

  const handleCreateDocument = () => {
    navigate('/documents/new');
  };
  
  // Função para baixar um documento
  const handleDownload = async (e, docId) => {
    e.stopPropagation();
    try {
      setActionLoading(true);
      setActionError(null);
      
      // Chamada para API de download usando o endpoint correto que retorna o arquivo diretamente
      const downloadUrl = `${documentApiClient.defaults.baseURL}/${docId}/download/file`;
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
      
      // Obter nome do arquivo a partir dos headers ou usar um padrão
      const contentDisposition = response.headers.get('content-disposition');
      let filename = 'documento.pdf';
      if (contentDisposition) {
        const filenameRegex = /filename[^;=\n]*=((['"]).*)\2|[^;\n]*/;
        const matches = filenameRegex.exec(contentDisposition);
        if (matches != null && matches[1]) {
          filename = matches[1].replace(/['"]*/g, '');
        }
      }
      
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      
      // Limpeza após download
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      
      setActionSuccess('Download iniciado com sucesso!');
      setTimeout(() => setActionSuccess(''), 3000);
    } catch (err) {
      console.error('Erro ao baixar documento:', err);
      setActionError(err.message || 'Erro ao baixar o documento');
    } finally {
      setActionLoading(false);
    }
  };
  
  // Função para abrir o modal de edição de metadados
  const handleOpenEditModal = (e, doc) => {
    e.stopPropagation();
    setCurrentDocument(doc);
    setEditData({
      title: doc.title,
      tags: doc.tags || [],
      categories: doc.categories || [],
      status: doc.status || 'draft'
    });
    setShowEditModal(true);
  };
  
  // Função para adicionar tag no modal de edição
  const handleAddTag = () => {
    if (newTag.trim() && !editData.tags.includes(newTag)) {
      setEditData(prev => ({
        ...prev,
        tags: [...prev.tags, newTag]
      }));
      setNewTag('');
    }
  };
  
  // Função para remover tag no modal de edição
  const handleRemoveTag = (tagToRemove) => {
    setEditData(prev => ({
      ...prev,
      tags: prev.tags.filter(tag => tag !== tagToRemove)
    }));
  };
  
  // Função para adicionar categoria no modal de edição
  const handleAddCategory = () => {
    if (newCategory.trim() && !editData.categories.includes(newCategory)) {
      setEditData(prev => ({
        ...prev,
        categories: [...prev.categories, newCategory]
      }));
      setNewCategory('');
    }
  };
  
  // Função para remover categoria no modal de edição
  const handleRemoveCategory = (categoryToRemove) => {
    setEditData(prev => ({
      ...prev,
      categories: prev.categories.filter(category => category !== categoryToRemove)
    }));
  };
  
  // Função para lidar com alterações em campos do modal de edição
  const handleEditInputChange = (e) => {
    const { name, value } = e.target;
    setEditData(prev => ({
      ...prev,
      [name]: value
    }));
  };
  
  // Função para salvar metadados editados
  const handleSaveMetadata = async (e) => {
    e.preventDefault();
    if (!currentDocument) return;
    
    try {
      setActionLoading(true);
      setActionError(null);
      
      // Criar objeto com dados atualizados
      const updatedData = {
        title: editData.title.trim(),
        tags: editData.tags,
        categories: editData.categories,
        status: editData.status,
        // Preservar conteúdo e outros dados não editáveis
        content: currentDocument.content || ''
      };
      
      // Validar título
      if (!updatedData.title) {
        setActionError('O título é obrigatório');
        setActionLoading(false);
        return;
      }
      
      // Chamada para API de atualização
      await documentApiClient.put(`/${currentDocument.id}`, updatedData);
      
      // Fechar modal e atualizar lista
      setShowEditModal(false);
      setActionSuccess('Documento atualizado com sucesso!');
      setTimeout(() => setActionSuccess(''), 3000);
      
      // Atualizar lista de documentos
      fetchDocuments();
    } catch (err) {
      console.error('Erro ao atualizar documento:', err);
      setActionError(err.response?.data?.message || 'Erro ao atualizar o documento');
    } finally {
      setActionLoading(false);
    }
  };
  
  // Função para abrir modal de confirmação de exclusão
  const handleOpenDeleteModal = (e, doc) => {
    e.stopPropagation();
    setCurrentDocument(doc);
    setShowDeleteModal(true);
  };
  
  // Função para confirmar exclusão
  const handleConfirmDelete = async () => {
    if (!currentDocument) {
      console.log('Não há documento selecionado para exclusão');
      return;
    }
    
    console.log('Iniciando exclusão do documento:', currentDocument.id, currentDocument.title);
    
    try {
      setActionLoading(true);
      setActionError(null);
      
      // Chamada para API de exclusão
      console.log('Enviando requisição DELETE para:', `/api/v1/documents/${currentDocument.id}`);
      const response = await documentApiClient.delete(`/${currentDocument.id}`);
      console.log('Resposta da API de exclusão:', response);
      
      // Fechar modal e atualizar lista
      setShowDeleteModal(false);
      setActionSuccess('Documento excluído com sucesso!');
      setTimeout(() => setActionSuccess(''), 3000);
      
      // Atualizar lista de documentos
      fetchDocuments();
    } catch (err) {
      console.error('Erro ao excluir documento:', err);
      const errorDetails = {
        message: err.message,
        responseData: err.response?.data,
        responseStatus: err.response?.status,
        statusText: err.response?.statusText,
        stack: err.stack
      };
      console.error('Detalhes completos do erro de exclusão:', errorDetails);
      
      // Mostrar informações mais detalhadas do erro
      const errorMessage = err.response?.data?.message || 
                         err.response?.data?.error ||
                         (typeof err.response?.data === 'string' ? err.response.data : null) ||
                         err.message ||
                         'Erro ao excluir o documento';
                         
      setActionError(`Erro na exclusão: ${errorMessage}`);
    } finally {
      setActionLoading(false);
      console.log('Processo de exclusão finalizado');
    }
  };
  
  // Função para fechar modais
  const handleCloseModals = () => {
    setShowEditModal(false);
    setShowDeleteModal(false);
    setCurrentDocument(null);
    setActionError(null);
  };

  const formatDate = (dateString) => {
    const options = { 
      day: '2-digit', 
      month: '2-digit', 
      year: 'numeric', 
      hour: '2-digit', 
      minute: '2-digit' 
    };
    return new Date(dateString).toLocaleDateString('pt-BR', options);
  };

  const getStatusClass = (status) => {
    switch (status) {
      case 'draft': return 'status-draft';
      case 'review': return 'status-review';
      case 'published': return 'status-published';
      case 'archived': return 'status-archived';
      default: return '';
    }
  };

  const getStatusText = (status) => {
    switch (status) {
      case 'draft': return 'Rascunho';
      case 'review': return 'Em Revisão';
      case 'published': return 'Publicado';
      case 'archived': return 'Arquivado';
      default: return status;
    }
  };

  if (loading) return <div className="loading">Carregando documentos...</div>;

  if (error) return (
    <div className="error-container">
      <p className="error-message">Erro ao carregar documentos: {error}</p>
      <button onClick={fetchDocuments}>Tentar novamente</button>
    </div>
  );

  return (
    <div className="documents-container">
      <div className="documents-header">
        <h1>Meus Documentos</h1>
        <button 
          className="create-document-btn"
          onClick={handleCreateDocument}
        >
          Novo Documento
        </button>
      </div>

      <div className="search-filter-container">
        <form onSubmit={handleSearch} className="search-form">
          <input
            type="text"
            placeholder="Buscar documentos..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
          <button type="submit" className="search-button">Buscar</button>
        </form>

        <div className="filters">
          <select 
            value={filters.status} 
            onChange={(e) => handleFilterChange('status', e.target.value)}
            className="filter-select"
          >
            <option value="">Todos os Status</option>
            <option value="draft">Rascunho</option>
            <option value="review">Em Revisão</option>
            <option value="published">Publicado</option>
            <option value="archived">Arquivado</option>
          </select>

          <select 
            value={filters.category} 
            onChange={(e) => handleFilterChange('category', e.target.value)}
            className="filter-select"
          >
            <option value="">Todas as Categorias</option>
            <option value="relatorio">Relatório</option>
            <option value="memorando">Memorando</option>
            <option value="oficio">Ofício</option>
            <option value="procedimento">Procedimento</option>
          </select>
        </div>
      </div>

      {documents.length === 0 ? (
        <div className="no-documents">
          <p>Nenhum documento encontrado.</p>
        </div>
      ) : (
        <div className="documents-list">
          <table>
            <thead>
              <tr>
                <th>Título</th>
                <th>Última Atualização</th>
                <th>Status</th>
                <th>Versões</th>
                <th>Ações</th>
              </tr>
            </thead>
            <tbody>
              {documents.map((doc) => (
                <tr key={doc.id} onClick={() => handleViewDocument(doc.id)}>
                  <td>{doc.title}</td>
                  <td>{formatDate(doc.updated_at)}</td>
                  <td>
                    <span className={`status-badge ${getStatusClass(doc.status)}`}>
                      {getStatusText(doc.status)}
                    </span>
                  </td>
                  <td>{doc.version_count || 1}</td>
                  <td>
                    <div className="action-buttons">
                      <button 
                        onClick={(e) => {
                          e.stopPropagation();
                          handleViewDocument(doc.id);
                        }}
                        className="view-btn"
                      >
                        Visualizar
                      </button>
                      <button
                        onClick={(e) => handleDownload(e, doc.id)}
                        className="download-btn"
                        disabled={actionLoading}
                      >
                        Baixar
                      </button>
                      <button
                        onClick={(e) => handleOpenEditModal(e, doc)}
                        className="edit-btn"
                        disabled={actionLoading}
                      >
                        Editar Metadados
                      </button>
                      <button
                        onClick={(e) => handleOpenDeleteModal(e, doc)}
                        className="delete-btn"
                        disabled={actionLoading}
                      >
                        Excluir
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      
      {/* Modal de feedback para ações */}
      {actionSuccess && (
        <div className="success-message">
          {actionSuccess}
        </div>
      )}
      
      {actionError && (
        <div className="error-message">
          {actionError}
        </div>
      )}
      
      {/* Modal de edição de metadados */}
      {showEditModal && currentDocument && (
        <div className="modal-overlay">
          <div className="modal-content edit-modal">
            <div className="modal-header">
              <h2>Editar Metadados</h2>
              <button className="close-btn" onClick={handleCloseModals}>×</button>
            </div>
            
            <form onSubmit={handleSaveMetadata} className="document-form">
              {actionError && <div className="error-message">{actionError}</div>}
              
              <div className="form-group">
                <label htmlFor="title">Título</label>
                <input
                  type="text"
                  id="title"
                  name="title"
                  value={editData.title}
                  onChange={handleEditInputChange}
                  placeholder="Título do documento"
                  required
                  className="form-control"
                />
              </div>
              
              <div className="form-row">
                <div className="form-group">
                  <label htmlFor="status">Status</label>
                  <select
                    id="status"
                    name="status"
                    value={editData.status}
                    onChange={handleEditInputChange}
                    className="form-control"
                  >
                    <option value="draft">Rascunho</option>
                    <option value="review">Em Revisão</option>
                    <option value="published">Publicado</option>
                    <option value="archived">Arquivado</option>
                  </select>
                </div>
              </div>
              
              <div className="form-row">
                <div className="form-group tags-section">
                  <label>Tags</label>
                  <div className="tags-input-container">
                    <input
                      type="text"
                      value={newTag}
                      onChange={(e) => setNewTag(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddTag())}
                      placeholder="Adicionar tag"
                      className="tag-input"
                    />
                    <button 
                      type="button" 
                      onClick={handleAddTag}
                      className="add-tag-btn"
                    >
                      +
                    </button>
                  </div>
                  <div className="tags-list">
                    {editData.tags.map((tag, index) => (
                      <span key={index} className="tag">
                        {tag}
                        <button 
                          type="button"
                          onClick={() => handleRemoveTag(tag)}
                          className="remove-tag-btn"
                        >
                          ×
                        </button>
                      </span>
                    ))}
                  </div>
                </div>

                <div className="form-group categories-section">
                  <label>Categorias</label>
                  <div className="categories-input-container">
                    <input
                      type="text"
                      value={newCategory}
                      onChange={(e) => setNewCategory(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddCategory())}
                      placeholder="Adicionar categoria"
                      className="category-input"
                    />
                    <button 
                      type="button" 
                      onClick={handleAddCategory}
                      className="add-category-btn"
                    >
                      +
                    </button>
                  </div>
                  <div className="categories-list">
                    {editData.categories.map((category, index) => (
                      <span key={index} className="category">
                        {category}
                        <button 
                          type="button"
                          onClick={() => handleRemoveCategory(category)}
                          className="remove-category-btn"
                        >
                          ×
                        </button>
                      </span>
                    ))}
                  </div>
                </div>
              </div>
              
              <div className="modal-actions">
                <button 
                  type="button" 
                  className="cancel-btn" 
                  onClick={handleCloseModals}
                  disabled={actionLoading}
                >
                  Cancelar
                </button>
                <button 
                  type="submit" 
                  className="save-btn" 
                  disabled={actionLoading}
                >
                  {actionLoading ? 'Salvando...' : 'Salvar Alterações'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
      
      {/* Modal de confirmação de exclusão */}
      {showDeleteModal && currentDocument && (
        <div className="modal-overlay">
          <div className="modal-content delete-modal">
            <div className="modal-header">
              <h2>Confirmar Exclusão</h2>
              <button className="close-btn" onClick={handleCloseModals}>×</button>
            </div>
            
            <div className="modal-body">
              <p className="warning-text">
                Tem certeza que deseja excluir o documento <strong>"{currentDocument.title}"</strong>?
              </p>
              <p className="warning-text">Esta ação não pode ser desfeita.</p>
              
              {actionError && <div className="error-message">{actionError}</div>}
            </div>
            
            <div className="modal-actions">
              <button 
                type="button" 
                className="cancel-btn" 
                onClick={handleCloseModals}
                disabled={actionLoading}
              >
                Cancelar
              </button>
              <button 
                type="button" 
                className="delete-confirm-btn" 
                onClick={handleConfirmDelete}
                disabled={actionLoading}
              >
                {actionLoading ? 'Excluindo...' : 'Sim, Excluir'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default DocumentList;

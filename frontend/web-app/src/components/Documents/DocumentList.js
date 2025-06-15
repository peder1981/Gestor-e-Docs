import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import './Documents.css';

const DocumentList = () => {
  const [documents, setDocuments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState({
    status: '',
    category: '',
  });
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
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default DocumentList;

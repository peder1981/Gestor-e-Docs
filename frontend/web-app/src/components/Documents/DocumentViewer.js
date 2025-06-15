import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import './Documents.css';
import documentApiClient from '../../api/documentApiClient';

const DocumentViewer = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const [document, setDocument] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [downloadUrl, setDownloadUrl] = useState(null);

  useEffect(() => {
    fetchDocument();
  }, [id]);

  const fetchDocument = async () => {
    try {
      setLoading(true);
      
      // Buscar detalhes do documento
      const response = await documentApiClient.get(`/${id}`);

      if (!response.ok) {
        // Se não autenticado, redirecionar para login
        if (response.status === 401) {
          navigate('/login', { 
            state: { message: 'Sua sessão expirou. Por favor, faça login novamente.' } 
          });
          return;
        }
        
        throw new Error(`Erro ao carregar documento: ${response.statusText}`);
      }

      const data = await response.json();
      setDocument(data);

      // Buscar URL de download (se necessário)
      generateDownloadUrl();

    } catch (err) {
      console.error('Erro ao buscar documento:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const generateDownloadUrl = async () => {
    try {
      const response = await documentApiClient.get(`/${id}/download`);

      if (response.ok) {
        const data = await response.json();
        setDownloadUrl(data.download_url);
      }
    } catch (error) {
      console.error('Erro ao gerar URL de download:', error);
    }
  };

  const handleEdit = () => {
    navigate(`/documents/${id}/edit`);
  };

  const handleDelete = async () => {
    if (!window.confirm('Tem certeza que deseja excluir este documento?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/documents/${id}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Erro ao excluir documento: ${response.statusText}`);
      }

      // Redirecionar para a lista de documentos após exclusão bem-sucedida
      navigate('/documents', { state: { message: 'Documento excluído com sucesso' } });
    } catch (error) {
      console.error('Erro ao excluir documento:', error);
      setError(`Falha ao excluir o documento: ${error.message}`);
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'Data não disponível';
    
    const options = { 
      day: '2-digit', 
      month: '2-digit', 
      year: 'numeric', 
      hour: '2-digit', 
      minute: '2-digit' 
    };
    return new Date(dateString).toLocaleDateString('pt-BR', options);
  };

  const getStatusText = (status) => {
    switch (status) {
      case 'draft': return 'Rascunho';
      case 'review': return 'Em Revisão';
      case 'published': return 'Publicado';
      case 'archived': return 'Arquivado';
      default: return status || 'Desconhecido';
    }
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

  if (loading) {
    return <div className="loading">Carregando documento...</div>;
  }

  if (error) {
    return (
      <div className="document-viewer-container">
        <div className="error-message">
          {error}
          <button onClick={() => navigate('/documents')}>Voltar para a lista</button>
        </div>
      </div>
    );
  }

  if (!document) {
    return (
      <div className="document-viewer-container">
        <div className="error-message">
          Documento não encontrado
          <button onClick={() => navigate('/documents')}>Voltar para a lista</button>
        </div>
      </div>
    );
  }

  return (
    <div className="document-viewer-container">
      <div className="viewer-header">
        <h1>{document.title}</h1>
        <div className="viewer-actions">
          <button className="back-btn" onClick={() => navigate('/documents')}>
            Voltar
          </button>
          <button className="edit-btn" onClick={handleEdit}>
            Editar
          </button>
          {downloadUrl && (
            <a 
              href={downloadUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="download-btn"
            >
              Download
            </a>
          )}
          <button className="delete-btn" onClick={handleDelete}>
            Excluir
          </button>
        </div>
      </div>

      <div className="document-metadata">
        <div className="metadata-row">
          <div className="metadata-item">
            <span className="metadata-label">Status:</span>
            <span className={`status-badge ${getStatusClass(document.status)}`}>
              {getStatusText(document.status)}
            </span>
          </div>
          <div className="metadata-item">
            <span className="metadata-label">Última atualização:</span>
            <span>{formatDate(document.updated_at)}</span>
          </div>
          <div className="metadata-item">
            <span className="metadata-label">Criado em:</span>
            <span>{formatDate(document.created_at)}</span>
          </div>
        </div>

        {document.tags && document.tags.length > 0 && (
          <div className="metadata-row">
            <div className="metadata-item">
              <span className="metadata-label">Tags:</span>
              <div className="tags-list viewer-tags">
                {document.tags.map((tag, index) => (
                  <span key={index} className="tag">{tag}</span>
                ))}
              </div>
            </div>
          </div>
        )}

        {document.categories && document.categories.length > 0 && (
          <div className="metadata-row">
            <div className="metadata-item">
              <span className="metadata-label">Categorias:</span>
              <div className="categories-list viewer-categories">
                {document.categories.map((category, index) => (
                  <span key={index} className="category">{category}</span>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="document-content">
        <pre>{document.content}</pre>
      </div>

      {document.version_history && document.version_history.length > 0 && (
        <div className="version-history">
          <h3>Histórico de Versões</h3>
          <table>
            <thead>
              <tr>
                <th>Versão</th>
                <th>Data</th>
                <th>Descrição</th>
              </tr>
            </thead>
            <tbody>
              {document.version_history.map((version, index) => (
                <tr key={index}>
                  <td>{version.version_number}</td>
                  <td>{formatDate(version.timestamp)}</td>
                  <td>{version.description}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default DocumentViewer;

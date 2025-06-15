import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import './Documents.css';
import documentApiClient from '../../api/documentApiClient';

const DocumentEditor = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const isNewDocument = id === 'new';

  // Estados para dados do documento
  const [document, setDocument] = useState({
    title: '',
    content: '',
    tags: [],
    categories: [],
    status: 'draft'
  });
  
  // Estados para controle de UI
  const [loading, setLoading] = useState(!isNewDocument);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [successMessage, setSuccessMessage] = useState('');
  
  // Estados para tags e categorias
  const [newTag, setNewTag] = useState('');
  const [newCategory, setNewCategory] = useState('');

  // Buscar documento existente se não for novo
  useEffect(() => {
    if (!isNewDocument) {
      fetchDocument();
    }
  }, [id]);

  // Função para buscar documento existente
  const fetchDocument = async () => {
    try {
      setLoading(true);
      
      const response = await documentApiClient.get(`/${id}`);
      const data = response.data;
      setDocument(data);
      setLoading(false);
    } catch (error) {
      if (error.response?.status === 401) {
        navigate('/login', { 
          state: { message: 'Sua sessão expirou. Por favor, faça login novamente.' } 
        });
        return;
      }
      setError(error.response?.data?.message || error.message);
      setLoading(false);
    }
  };

  // Função para salvar um documento
  const saveDocument = async () => {
    // Validar campos obrigatórios
    if (!document.title.trim()) {
      setError('O título é obrigatório');
      return;
    }

    try {
      setSaving(true);
      setError(null);
      
      const documentData = {
        title: document.title,
        content: document.content,
        tags: document.tags,
        categories: document.categories,
        status: document.status,
        description: isNewDocument ? 'Criação inicial' : 'Atualização de documento'
      };

      let response;
      if (isNewDocument) {
        response = await documentApiClient.post('/', documentData);
      } else {
        response = await documentApiClient.put(`/${id}`, documentData);
      }

      setDocument(response.data);
      setSaving(false);
      setSuccessMessage('Documento salvo com sucesso!');
      
      // Redirecionar para a página de visualização após salvar
      navigate(`/documents/${response.data.id}`);

      // Limpar mensagem de sucesso após 3 segundos
      setTimeout(() => setSuccessMessage(''), 3000);
    } catch (error) {
      if (error.response?.status === 401) {
        navigate('/login', { 
          state: { message: 'Sua sessão expirou. Por favor, faça login novamente.' } 
        });
        return;
      }
      console.error('Erro ao salvar documento:', error);
      setError(error.response?.data?.message || error.message);
      setSaving(false);
    } finally {
      setSaving(false);
    }
  };

  // Função para adicionar uma tag
  const addTag = () => {
    if (newTag.trim() && !document.tags.includes(newTag)) {
      setDocument(prev => ({
        ...prev,
        tags: [...prev.tags, newTag]
      }));
      setNewTag('');
    }
  };

  // Função para remover uma tag
  const removeTag = (tagToRemove) => {
    setDocument(prev => ({
      ...prev,
      tags: prev.tags.filter(tag => tag !== tagToRemove)
    }));
  };

  // Função para adicionar uma categoria
  const addCategory = () => {
    if (newCategory.trim() && !document.categories.includes(newCategory)) {
      setDocument(prev => ({
        ...prev,
        categories: [...prev.categories, newCategory]
      }));
      setNewCategory('');
    }
  };

  // Função para remover uma categoria
  const removeCategory = (categoryToRemove) => {
    setDocument(prev => ({
      ...prev,
      categories: prev.categories.filter(category => category !== categoryToRemove)
    }));
  };

  // Função para lidar com alterações gerais nos campos
  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setDocument(prev => ({
      ...prev,
      [name]: value
    }));
  };

  // Renderizar tela de carregamento
  if (loading) {
    return <div className="loading">Carregando documento...</div>;
  }

  return (
    <div className="document-editor-container">
      <div className="editor-header">
        <h1>{isNewDocument ? 'Novo Documento' : 'Editar Documento'}</h1>
        <div className="editor-actions">
          <button 
            className="cancel-btn"
            onClick={() => navigate('/documents')}
          >
            Cancelar
          </button>
          <button 
            className="save-btn"
            onClick={saveDocument}
            disabled={saving}
          >
            {saving ? 'Salvando...' : 'Salvar'}
          </button>
        </div>
      </div>

      {error && (
        <div className="error-message">
          {error}
        </div>
      )}

      {successMessage && (
        <div className="success-message">
          {successMessage}
        </div>
      )}

      <div className="document-form">
        <div className="form-group">
          <label htmlFor="title">Título *</label>
          <input
            type="text"
            id="title"
            name="title"
            value={document.title}
            onChange={handleInputChange}
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
              value={document.status}
              onChange={handleInputChange}
              className="form-control"
            >
              <option value="draft">Rascunho</option>
              <option value="review">Em Revisão</option>
              <option value="published">Publicado</option>
              <option value="archived">Arquivado</option>
            </select>
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="content">Conteúdo</label>
          <textarea
            id="content"
            name="content"
            value={document.content}
            onChange={handleInputChange}
            placeholder="Conteúdo em formato Markdown"
            className="form-control content-editor"
            rows="15"
          />
        </div>

        <div className="form-row">
          <div className="form-group tags-section">
            <label>Tags</label>
            <div className="tags-input-container">
              <input
                type="text"
                value={newTag}
                onChange={(e) => setNewTag(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && addTag()}
                placeholder="Adicionar tag"
                className="tag-input"
              />
              <button 
                type="button" 
                onClick={addTag}
                className="add-tag-btn"
              >
                +
              </button>
            </div>
            <div className="tags-list">
              {document.tags.map((tag, index) => (
                <span key={index} className="tag">
                  {tag}
                  <button 
                    onClick={() => removeTag(tag)}
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
                onKeyPress={(e) => e.key === 'Enter' && addCategory()}
                placeholder="Adicionar categoria"
                className="category-input"
              />
              <button 
                type="button" 
                onClick={addCategory}
                className="add-category-btn"
              >
                +
              </button>
            </div>
            <div className="categories-list">
              {document.categories.map((category, index) => (
                <span key={index} className="category">
                  {category}
                  <button 
                    onClick={() => removeCategory(category)}
                    className="remove-category-btn"
                  >
                    ×
                  </button>
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {!isNewDocument && (
        <div className="document-info">
          <p className="document-info-text">
            <strong>Última atualização:</strong> {new Date().toLocaleDateString('pt-BR')}
          </p>
          <p className="document-info-text">
            <strong>ID do Documento:</strong> {id}
          </p>
        </div>
      )}
    </div>
  );
};

export default DocumentEditor;

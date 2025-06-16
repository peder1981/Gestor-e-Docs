import React, { useState } from 'react';

// Função para testar autenticação
export async function testAuth() {
  console.log('Iniciando teste de autenticação...');
  const results = [];
  
  try {
    // Teste 1: Verificar se há token de autenticação nos cookies
    const cookies = document.cookie.split(';').map(cookie => cookie.trim());
    console.log('Cookies disponíveis:', cookies);
    results.push(`Cookies disponíveis: ${cookies.join(', ') || 'Nenhum cookie encontrado'}`);
    
    // Teste 2: Fazer uma requisição para o endpoint /me para testar autenticação
    const response = await fetch('/api/v1/identity/me', {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      }
    });
    
    const data = await response.json();
    console.log('Resposta do /me:', {
      status: response.status,
      statusText: response.statusText,
      data: data
    });
    results.push(`/me status: ${response.status} ${response.statusText}`);
    
    // Teste 3: Verificar headers e logs detalhados da requisição
    const responseHeaders = Object.fromEntries([...response.headers.entries()]);
    console.log('Headers da resposta:', responseHeaders);
    results.push(`Headers da resposta: ${JSON.stringify(responseHeaders)}`);
    
    if (response.status === 401) {
      console.log('Usuário não autenticado. Tentando refresh token...');
      results.push('Usuário não autenticado. Tentando refresh token...');
      
      // Teste 4: Tentar refresh token manualmente
      const refreshResponse = await fetch('/api/v1/identity/refresh', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        }
      });
      
      const refreshData = await refreshResponse.json();
      console.log('Resposta do /refresh:', {
        status: refreshResponse.status,
        statusText: refreshResponse.statusText,
        data: refreshData
      });
      results.push(`/refresh status: ${refreshResponse.status} ${refreshResponse.statusText}`);
      
      if (refreshResponse.status === 200) {
        console.log('Refresh token bem-sucedido! Tentando novamente verificação de autenticação...');
        results.push('Refresh token bem-sucedido! Tentando novamente verificação de autenticação...');
        
        // Verificar novamente após refresh
        const meResponse = await fetch('/api/v1/identity/me', {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
          }
        });
        
        const meData = await meResponse.json();
        console.log('Resposta do /me após refresh:', {
          status: meResponse.status,
          statusText: meResponse.statusText,
          data: meData
        });
        results.push(`/me após refresh status: ${meResponse.status} ${meResponse.statusText}`);
      }
    } else if (response.status === 200) {
      console.log('Usuário autenticado com sucesso!');
      results.push('Usuário autenticado com sucesso!');
      
      // Teste 5: Verificar permissões/roles do usuário
      if (data.user && data.user.role) {
        console.log('Role do usuário:', data.user.role);
        results.push(`Role do usuário: ${data.user.role}`);
      }
      
      // Teste 6: Testar chamada para documento API
      console.log('Testando API de documentos...');
      results.push('Testando API de documentos...');
      const docsResponse = await fetch('/api/v1/documents', {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        }
      });
      
      if (!docsResponse.ok) {
        const errorText = await docsResponse.text();
        console.log('Resposta da API de documentos (erro):', {
          status: docsResponse.status,
          statusText: docsResponse.statusText,
          error: errorText
        });
        results.push(`API de documentos status: ${docsResponse.status} ${docsResponse.statusText}`);
        results.push(`Erro: ${errorText}`);
      } else {
        const docsData = await docsResponse.json();
        console.log('Resposta da API de documentos:', {
          status: docsResponse.status,
          statusText: docsResponse.statusText,
          data: docsData
        });
        results.push(`API de documentos status: ${docsResponse.status} ${docsResponse.statusText}`);
        results.push(`Documentos encontrados: ${Array.isArray(docsData) ? docsData.length : 'Não é um array'}`);
      }
    }
    
  } catch (error) {
    console.error('Erro durante teste de autenticação:', error);
    results.push(`Erro: ${error.message}`);
  }
  
  return results;
}

// Função para testar especificamente ações de documentos
export async function testDocumentActions() {
  console.log('Testando ações de documentos...');
  const results = [];
  
  try {
    results.push('Iniciando teste de ações de documentos...');
    
    // Obter lista de documentos para testar
    const docsResponse = await fetch('/api/v1/documents', {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Accept': 'application/json'
      }
    });
    
    if (!docsResponse.ok) {
      const errorText = await docsResponse.text();
      console.error(`Falha ao obter documentos: ${docsResponse.status} ${docsResponse.statusText}`);
      console.error('Detalhes do erro:', errorText);
      results.push(`Falha ao obter documentos: ${docsResponse.status} ${docsResponse.statusText}`);
      results.push(`Detalhes do erro: ${errorText}`);
      return results;
    }
    
    const docsData = await docsResponse.json();
    console.log('Documentos disponíveis:', docsData);
    
    if (!Array.isArray(docsData)) {
      results.push(`Resposta não é um array: ${JSON.stringify(docsData).substring(0, 100)}...`);
      return results;
    }
    
    results.push(`Documentos encontrados: ${docsData.length}`);
    
    if (docsData.length === 0 || !docsData[0]) {
      results.push('Nenhum documento disponível para testar ações');
      return results;
    }
    
    // Teste da ação de download
    const docId = docsData[0].id;
    const docTitle = docsData[0].title || 'Sem título';
    results.push(`Testando download do documento "${docTitle}" (ID: ${docId})...`);
    console.log(`Testando download do documento ${docId}...`);
    
    const downloadResponse = await fetch(`/api/v1/documents/${docId}/download`, {
      method: 'GET',
      credentials: 'include'
    });
    
    const downloadHeaders = Object.fromEntries([...downloadResponse.headers.entries()]);
    console.log('Resposta do download:', {
      status: downloadResponse.status,
      statusText: downloadResponse.statusText,
      headers: downloadHeaders
    });
    
    results.push(`Download status: ${downloadResponse.status} ${downloadResponse.statusText}`);
    
    if (downloadResponse.status === 200) {
      results.push('Download bem-sucedido!');
      results.push(`Content-Type: ${downloadHeaders['content-type']}`);
      results.push(`Content-Disposition: ${downloadHeaders['content-disposition'] || 'Não disponível'}`);
    } else {
      try {
        const errorText = await downloadResponse.text();
        results.push(`Erro no download: ${errorText}`);
      } catch (e) {
        results.push(`Não foi possível ler o corpo da resposta de erro`);
      }
    }
    
    // Teste da ação de editar metadados
    results.push('\nTestando edição de metadados...');
    
    // Simula a chamada para editar metadados
    const updateData = {
      title: docTitle + ' (Teste)',
      tags: ['teste'],
      categories: ['teste'],
      status: 'active'
    };
    
    const updateResponse = await fetch(`/api/v1/documents/${docId}`, {
      method: 'PUT',
      credentials: 'include',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(updateData)
    });
    
    results.push(`Edição status: ${updateResponse.status} ${updateResponse.statusText}`);
    
    if (updateResponse.status === 200) {
      const updateResult = await updateResponse.json();
      results.push('Edição bem-sucedida!');
      results.push(`Dados atualizados: ${JSON.stringify(updateResult).substring(0, 100)}...`);
    } else {
      try {
        const errorText = await updateResponse.text();
        results.push(`Erro na edição: ${errorText}`);
      } catch (e) {
        results.push(`Não foi possível ler o corpo da resposta de erro`);
      }
    }
    
  } catch (error) {
    console.error('Erro durante teste de ações de documentos:', error);
    results.push(`Erro: ${error.message}`);
  }
  
  return results;
}

// Componente React para testes
export function TestPanel({ fullPage = false }) {
  const [authResults, setAuthResults] = useState([]);
  const [actionResults, setActionResults] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [activeTest, setActiveTest] = useState('auth'); // Default para auth
  
  const runAuthTest = async () => {
    setIsLoading(true);
    setActiveTest('auth');
    setAuthResults(['Executando testes de autenticação...']);
    try {
      const results = await testAuth();
      setAuthResults(results);
    } catch (error) {
      setAuthResults([`Erro ao executar teste: ${error.message}`]);
    }
    setIsLoading(false);
  };
  
  const runActionsTest = async () => {
    setIsLoading(true);
    setActiveTest('actions');
    setActionResults(['Executando testes de ações de documentos...']);
    try {
      console.log('Chamando testDocumentActions()');
      const results = await testDocumentActions();
      console.log('Resultados obtidos:', results);
      setActionResults(results);
    } catch (error) {
      console.error('Erro ao executar testes de ações:', error);
      setActionResults([`Erro ao executar teste: ${error.message}`]);
    } finally {
      setIsLoading(false);
    }
  };

  React.useEffect(() => {
    // Auto-executar testes de autenticação quando o componente for montado em modo de página completa
    if (fullPage) {
      runAuthTest();
    }
  }, [fullPage]);
  
  // Estilo para o painel de testes
  const panelStyle = fullPage ? {
    width: '100%',
    backgroundColor: '#f8f9fa',
    border: '1px solid #ddd',
    borderRadius: '8px',
    boxShadow: '0 2px 10px rgba(0,0,0,0.1)',
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column'
  } : {
    position: 'fixed',
    bottom: '20px',
    right: '20px',
    width: '600px',
    maxHeight: '80vh',
    backgroundColor: '#f8f9fa',
    border: '1px solid #ddd',
    borderRadius: '4px',
    boxShadow: '0 2px 10px rgba(0,0,0,0.1)',
    zIndex: 9999,
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column'
  };
  
  const headerStyle = {
    padding: '12px 16px',
    borderBottom: '1px solid #ddd',
    backgroundColor: '#343a40',
    color: 'white',
    fontWeight: 'bold',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center'
  };
  
  const contentStyle = fullPage ? {
    padding: '16px',
    maxHeight: '500px',
    overflowY: 'auto',
    backgroundColor: '#fff'
  } : {
    padding: '16px',
    maxHeight: 'calc(80vh - 120px)',
    overflowY: 'auto',
    backgroundColor: '#fff'
  };
  
  const buttonContainerStyle = {
    display: 'flex',
    gap: '10px',
    padding: '12px 16px',
    borderTop: '1px solid #ddd',
    backgroundColor: '#f8f9fa'
  };
  
  const buttonStyle = (active) => ({
    padding: '12px 20px',
    borderRadius: '4px',
    border: 'none',
    backgroundColor: active ? '#007bff' : '#6c757d',
    color: 'white',
    cursor: 'pointer',
    fontWeight: 'bold',
    flex: 1,
    fontSize: fullPage ? '16px' : '14px'
  });
  
  const tabsStyle = {
    display: 'flex',
    borderBottom: '1px solid #ddd',
    backgroundColor: '#f1f3f5'
  };
  
  const tabStyle = (active) => ({
    padding: fullPage ? '15px 30px' : '10px 20px',
    borderBottom: active ? '2px solid #007bff' : 'none',
    backgroundColor: active ? '#fff' : 'transparent',
    cursor: 'pointer',
    fontWeight: active ? 'bold' : 'normal',
    color: active ? '#007bff' : '#495057',
    fontSize: fullPage ? '16px' : '14px'
  });
  
  const resultItemStyle = {
    padding: '10px 0',
    borderBottom: '1px solid #eee',
    fontSize: fullPage ? '15px' : '14px',
    fontFamily: 'monospace',
    whiteSpace: 'pre-wrap',
    wordBreak: 'break-word'
  };
  
  return (
    <div style={panelStyle}>
      <div style={headerStyle}>
        <span>Painel de Teste de Autenticação e Ações</span>
      </div>
      
      <div style={tabsStyle}>
        <div 
          style={tabStyle(activeTest === 'auth')} 
          onClick={() => setActiveTest('auth')}
        >
          Testes de Autenticação
        </div>
        <div 
          style={tabStyle(activeTest === 'actions')} 
          onClick={() => setActiveTest('actions')}
        >
          Testes de Ações
        </div>
      </div>
      
      <div style={contentStyle}>
        {activeTest === 'auth' && authResults.length > 0 ? (
          authResults.map((result, index) => (
            <div key={index} style={resultItemStyle}>{result}</div>
          ))
        ) : activeTest === 'auth' && !isLoading && (
          <div style={{padding: '20px', textAlign: 'center', color: '#666'}}>
            Clique no botão "Testar Autenticação" para iniciar os testes
          </div>
        )}
        
        {activeTest === 'actions' && actionResults.length > 0 ? (
          actionResults.map((result, index) => (
            <div key={index} style={resultItemStyle}>{result}</div>
          ))
        ) : activeTest === 'actions' && !isLoading && (
          <div style={{padding: '20px', textAlign: 'center', color: '#666'}}>
            Clique no botão "Testar Ações" para iniciar os testes
          </div>
        )}
        
        {isLoading && (
          <div style={{padding: '40px', textAlign: 'center'}}>
            <div style={{fontSize: '18px', marginBottom: '10px', fontWeight: 'bold'}}>
              Executando testes...
            </div>
            <div style={{color: '#666'}}>
              Isso pode levar alguns segundos
            </div>
          </div>
        )}
      </div>
      
      <div style={buttonContainerStyle}>
        <button 
          style={buttonStyle(activeTest === 'auth')} 
          onClick={runAuthTest} 
          disabled={isLoading}
        >
          Testar Autenticação
        </button>
        <button 
          style={buttonStyle(activeTest === 'actions')} 
          onClick={runActionsTest} 
          disabled={isLoading}
        >
          Testar Ações
        </button>
      </div>
    </div>
  );
}

// Exporta as funções para o console global também
window.runAuthTest = testAuth;
window.runDocumentActionsTest = testDocumentActions;

console.log('Funções de teste disponíveis no console:');
console.log('- window.runAuthTest(): Testar autenticação');
console.log('- window.runDocumentActionsTest(): Testar ações de documentos');

import React, { useState, useEffect } from 'react';
import {
  BrowserRouter as Router,
  Route,
  Routes,
  Link
} from 'react-router-dom';
import './App.css';
import Register from './components/Auth/Register'; // Importar o componente Register
import Login from './components/Auth/Login'; // Importar o componente Login

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('jwtToken'));

  useEffect(() => {
    const handleStorageChange = () => {
      setIsAuthenticated(!!localStorage.getItem('jwtToken'));
    };
    window.addEventListener('storage', handleStorageChange); // Ouve mudanças no localStorage (ex: login/logout em outra aba)
    // Verifica no mount inicial também, caso o evento 'storage' não seja disparado para a própria aba
    handleStorageChange();
    return () => {
      window.removeEventListener('storage', handleStorageChange);
    };
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('jwtToken');
    setIsAuthenticated(false);
    window.location.href = '/login'; // Redireciona para a página de login
  };

  return (
    <Router>
      <div className="App">
        <nav>
          <ul>
            <li>
              <Link to="/">Home</Link>
            </li>
            {!isAuthenticated ? (
              <>
                <li>
                  <Link to="/register">Registrar</Link>
                </li>
                <li>
                  <Link to="/login">Login</Link>
                </li>
              </>
            ) : (
              <li>
                <button onClick={handleLogout} style={{background: 'none', border: 'none', color: 'blue', textDecoration: 'underline', cursor: 'pointer', padding: 0, fontSize: 'inherit'}}>
                  Logout
                </button>
              </li>
            )}
            {/* Adicionar mais links de navegação aqui conforme necessário */}
          </ul>
        </nav>

        <header className="App-header">
          <h1>Gestor-e-Docs</h1>
        </header>

        <main>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/register" element={<Register />} />
            <Route path="/login" element={<Login />} />
            {/* Adicionar mais rotas aqui */}
          </Routes>
        </main>

        <footer>
          <p>&copy; {new Date().getFullYear()} Gestor-e-Docs</p>
        </footer>
      </div>
    </Router>
  );
}

// Componente Home simples para a rota principal
const Home = () => (
  <div>
    <h2>Bem-vindo ao Gestor-e-Docs</h2>
    <p>Esta é a página inicial. Use o menu de navegação para acessar outras seções.</p>
  </div>
);

export default AppWrapper;

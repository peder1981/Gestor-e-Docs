import React, { useState, useEffect } from 'react';

const GlobalLoader = () => {
  const [isRefreshing, setIsRefreshing] = useState(false);

  useEffect(() => {
    const handleTokenRefreshStart = () => {
      setIsRefreshing(true);
    };

    const handleTokenRefreshEnd = () => {
      setIsRefreshing(false);
    };

    window.addEventListener('tokenRefreshStart', handleTokenRefreshStart);
    window.addEventListener('tokenRefreshEnd', handleTokenRefreshEnd);

    return () => {
      window.removeEventListener('tokenRefreshStart', handleTokenRefreshStart);
      window.removeEventListener('tokenRefreshEnd', handleTokenRefreshEnd);
    };
  }, []);

  if (!isRefreshing) {
    return null;
  }

  // Estilo básico para o loader. Pode ser melhorado com CSS.
  const loaderStyle = {
    position: 'fixed',
    top: '10px',
    right: '10px',
    padding: '10px',
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    color: 'white',
    borderRadius: '5px',
    zIndex: 9999, // Para garantir que fique por cima de outros elementos
  };

  return (
    <div style={loaderStyle}>
      Renovando sessão...
    </div>
  );
};

export default GlobalLoader;

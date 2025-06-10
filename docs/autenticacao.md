# Sistema de Autenticação - Gestor-e-Docs

## Visão Geral

O sistema de autenticação do Gestor-e-Docs é baseado em JWT (JSON Web Tokens) com cookies HttpOnly para maior segurança. Este documento detalha a implementação, fluxo de autenticação e considerações de segurança.

## Componentes Principais

### Backend (Go - identity-service)

- **Autenticação de Usuários**: Validação de credenciais e emissão de tokens JWT
- **Gerenciamento de Tokens**: Criação de access_token (curta duração) e refresh_token (longa duração)
- **Middleware de Autenticação**: Validação de tokens para rotas protegidas
- **Renovação de Tokens**: Mecanismo de refresh para manter a sessão do usuário ativa
- **Segurança de Cookies**: Implementação de cookies HttpOnly para proteger contra ataques XSS

### Frontend (React)

- **Gerenciamento de Estado de Autenticação**: Mantém estado global sobre autenticação do usuário
- **Eventos Customizados**: Dispara eventos de authChange para sincronizar o estado
- **Interceptores de Requisição**: Garante que requisições incluam credenciais (cookies)
- **Tratamento de Expiração**: Lida com tokens expirados e logout automático

## Fluxo de Autenticação

### 1. Login
- Usuário fornece credenciais (e-mail e senha)
- Backend valida as credenciais no banco MongoDB
- Em caso de sucesso:
  - Gera access_token (15 min) e refresh_token (7 dias)
  - Armazena tokens em cookies HttpOnly
  - Retorna dados básicos do usuário

### 2. Acesso a Recursos Protegidos
- Requisição inclui automaticamente cookies HttpOnly
- Middleware verifica a validade do access_token
- Em caso de token válido:
  - Extrai informações do usuário (claims)
  - Permite acesso ao recurso solicitado

### 3. Renovação de Token (Refresh)
- Quando o access_token expira:
  - Frontend tenta acessar rota protegida e recebe 401
  - Chama automaticamente a rota /refresh
  - Backend valida o refresh_token e emite novo access_token
  - Sessão do usuário continua sem interrupção

### 4. Logout
- Quando o usuário faz logout:
  - Backend invalida os cookies (MaxAge = -1)
  - Frontend dispara evento de authChange
  - Estado de autenticação é atualizado em toda aplicação

## Configuração de Segurança

### Cookies
```go
accessTokenCookie := http.Cookie{
    Name:     "access_token",
    Value:    accessToken,
    MaxAge:   15 * 60,         // 15 minutos
    Path:     "/",
    Domain:   "localhost",     // Ajustar para domínio real em produção
    Secure:   false,           // Deve ser true em produção (HTTPS)
    HttpOnly: true,
    SameSite: http.SameSiteLaxMode,
}
```

### CORS
```go
config := cors.DefaultConfig()
config.AllowOrigins = []string{
    "http://localhost:3085", 
    // Outros domínios permitidos...
}
config.AllowCredentials = true  // Fundamental para permitir cookies
config.ExposeHeaders = []string{"Content-Length", "Set-Cookie"}
```

## Considerações de Segurança

1. **Cookies HttpOnly**: Tokens não podem ser acessados via JavaScript, mitigando riscos de XSS
2. **SameSite**: Configurado como Lax para ambiente de desenvolvimento, deve ser Strict em produção
3. **Secure Flag**: Desativado para HTTP local, deve ser ativado em produção (HTTPS)
4. **CORS**: Cuidadosamente configurado para permitir apenas origens confiáveis
5. **Token Expiration**: Access tokens com curta duração (15 minutos) limitam janela de ataque

## Migração para Produção

Ao migrar para ambiente de produção, as seguintes alterações devem ser feitas:

1. Ativar flag `Secure: true` nos cookies para exigir HTTPS
2. Atualizar `Domain` para o domínio real da aplicação
3. Configurar `SameSite: http.SameSiteStrictMode` para maior segurança
4. Atualizar `AllowOrigins` no CORS para os domínios de produção
5. Implementar rotação de refresh tokens para segurança adicional

## Depuração e Logs

Para auxiliar na depuração do sistema de autenticação, foram implementados logs detalhados:

- Logs no middleware de autenticação mostrando presença de cookies
- Logs no handler de refresh token mostrando detalhes do processo
- Logs no processo de validação de token

## Testes

Uma página de teste HTML (`test-auth.html`) está disponível para verificar o fluxo completo de autenticação. Recomenda-se usar esta página para validar configurações antes do deployment.

## Próximas Melhorias

1. **Rotação de Refresh Tokens**: Implementar invalidação de refresh tokens após uso
2. **Autenticação de Múltiplos Fatores**: Adicionar camada extra de segurança
3. **Monitoramento de Sessões**: Permitir que usuários vejam e encerrem suas sessões ativas
4. **Fingerprint do Dispositivo**: Associar tokens a dispositivos específicos
5. **Testes Automatizados**: Desenvolver testes para fluxos de autenticação

## Dicas para Desenvolvimento

1. **Ambiente Local**: Use `Domain: "localhost"` e `Secure: false` para desenvolvimento
2. **Testes Manuais**: A página `/test-auth.html` permite verificar cookies e tokens
3. **Depuração de CORS**: Monitore os logs para erros relacionados ao CORS
4. **Verificação de Cookies**: Use as ferramentas de desenvolvedor do navegador (Application > Cookies)
5. **Rotação de Chaves**: Implemente um sistema de rotação para JWT_SECRET_KEY

---

# Authentication System - Gestor-e-Docs (English Version)

## Overview

Gestor-e-Docs authentication system is based on JWT (JSON Web Tokens) with HttpOnly cookies for enhanced security. This document details the implementation, authentication flow, and security considerations.

## Main Components

### Backend (Go - identity-service)

- **User Authentication**: Credential validation and JWT token issuance
- **Token Management**: Creation of short-lived access_token and long-lived refresh_token
- **Authentication Middleware**: Token validation for protected routes
- **Token Renewal**: Refresh mechanism to keep user session active
- **Cookie Security**: HttpOnly cookies implementation to protect against XSS attacks

### Frontend (React)

- **Authentication State Management**: Maintains global state about user authentication
- **Custom Events**: Triggers authChange events to synchronize state
- **Request Interceptors**: Ensures requests include credentials (cookies)
- **Expiration Handling**: Handles expired tokens and automatic logout

## Authentication Flow

### 1. Login
- User provides credentials (email and password)
- Backend validates credentials in MongoDB database
- If successful:
  - Generates access_token (15 min) and refresh_token (7 days)
  - Stores tokens in HttpOnly cookies
  - Returns basic user data

### 2. Accessing Protected Resources
- Request automatically includes HttpOnly cookies
- Middleware verifies access_token validity
- If token is valid:
  - Extracts user information (claims)
  - Allows access to requested resource

### 3. Token Renewal (Refresh)
- When access_token expires:
  - Frontend tries to access protected route and receives 401
  - Automatically calls /refresh route
  - Backend validates refresh_token and issues new access_token
  - User session continues without interruption

### 4. Logout
- When user logs out:
  - Backend invalidates cookies (MaxAge = -1)
  - Frontend triggers authChange event
  - Authentication state is updated throughout application

## Security Configuration

### Cookies
```go
accessTokenCookie := http.Cookie{
    Name:     "access_token",
    Value:    accessToken,
    MaxAge:   15 * 60,         // 15 minutes
    Path:     "/",
    Domain:   "localhost",     // Adjust for real domain in production
    Secure:   false,           // Should be true in production (HTTPS)
    HttpOnly: true,
    SameSite: http.SameSiteLaxMode,
}
```

### CORS
```go
config := cors.DefaultConfig()
config.AllowOrigins = []string{
    "http://localhost:3085", 
    // Other allowed domains...
}
config.AllowCredentials = true  // Critical to allow cookies
config.ExposeHeaders = []string{"Content-Length", "Set-Cookie"}
```

## Security Considerations

1. **HttpOnly Cookies**: Tokens cannot be accessed via JavaScript, mitigating XSS risks
2. **SameSite**: Configured as Lax for development, should be Strict in production
3. **Secure Flag**: Disabled for local HTTP, must be enabled in production (HTTPS)
4. **CORS**: Carefully configured to allow only trusted origins
5. **Token Expiration**: Short-lived access tokens (15 minutes) limit attack window

## Production Migration

When migrating to production environment, the following changes should be made:

1. Enable `Secure: true` flag on cookies to require HTTPS
2. Update `Domain` to the actual application domain
3. Configure `SameSite: http.SameSiteStrictMode` for increased security
4. Update `AllowOrigins` in CORS to production domains
5. Implement refresh token rotation for additional security

## Debugging and Logs

To assist in debugging the authentication system, detailed logs have been implemented:

- Logs in authentication middleware showing cookie presence
- Logs in refresh token handler showing process details
- Logs in token validation process

## Testing

A HTML test page (`test-auth.html`) is available to verify the complete authentication flow. It is recommended to use this page to validate settings before deployment.

## Future Improvements

1. **Refresh Token Rotation**: Implement invalidation of refresh tokens after use
2. **Multi-Factor Authentication**: Add extra security layer
3. **Session Monitoring**: Allow users to view and terminate their active sessions
4. **Device Fingerprinting**: Associate tokens with specific devices
5. **Automated Testing**: Develop tests for authentication flows

## Development Tips

1. **Local Environment**: Use `Domain: "localhost"` and `Secure: false` for development
2. **Manual Testing**: The `/test-auth.html` page allows verification of cookies and tokens
3. **CORS Debugging**: Monitor logs for CORS-related errors
4. **Cookie Verification**: Use browser developer tools (Application > Cookies)
5. **Key Rotation**: Implement a rotation system for JWT_SECRET_KEY

# Implementação de Autenticação no Backend - Gestor-e-Docs

## Visão Geral

Este documento detalha a implementação do sistema de autenticação JWT baseado em cookies HttpOnly no backend Go (identity-service) do Gestor-e-Docs.

## Estrutura do Serviço de Identidade

```
/backend/services/identity-service/
├── db/
│   └── mongo.go              # Conexão e operações com MongoDB
├── handlers/
│   ├── auth_handler.go       # Handlers de autenticação (login, refresh, logout)
│   └── auth_middleware.go    # Middleware de proteção de rotas
├── models/
│   └── user.go               # Modelo de dados do usuário
├── utils/
│   └── token_utils.go        # Utilitários para geração/validação de tokens JWT
├── main.go                   # Ponto de entrada da aplicação e configuração de rotas
└── Dockerfile                # Configuração para build da imagem Docker
```

## Fluxo de Autenticação

### 1. Configuração de CORS

O serviço de identidade configura CORS para permitir credenciais em requisições cross-origin:

```go
// main.go
func main() {
    // ...
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{
        "http://localhost:3085",
        "http://localhost",
        "http://localhost:80",
        "http://127.0.0.1:3085",
        "http://127.0.0.1",
    }
    config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept"}
    config.AllowCredentials = true  // Permitir envio de cookies
    config.ExposeHeaders = []string{"Content-Length", "Set-Cookie"}
    config.MaxAge = 12 * time.Hour
    r.Use(cors.New(config))
    // ...
}
```

### 2. Login de Usuário

O handler de login verifica as credenciais do usuário e emite tokens JWT em cookies HttpOnly:

```go
// handlers/auth_handler.go
func LoginUser(c *gin.Context) {
    // Verificar credenciais no banco de dados
    // ...
    
    // Gerar tokens JWT
    accessToken, _ := utils.GenerateAccessToken(user.ID.Hex())
    refreshToken, _ := utils.GenerateRefreshToken(user.ID.Hex())
    
    // Definir cookies HttpOnly
    accessTokenCookie := http.Cookie{
        Name:     "access_token",
        Value:    accessToken,
        MaxAge:   15 * 60,         // 15 minutos
        Path:     "/",
        Domain:   "localhost",
        Secure:   false,           // Em produção: true
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    }
    refreshTokenCookie := http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        MaxAge:   7 * 24 * 60 * 60, // 7 dias
        Path:     "/",
        Domain:   "localhost",
        Secure:   false,            // Em produção: true
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    }
    
    http.SetCookie(c.Writer, &accessTokenCookie)
    http.SetCookie(c.Writer, &refreshTokenCookie)
    
    // Retornar resposta sem expor tokens
    c.JSON(http.StatusOK, gin.H{
        "message": "Login bem-sucedido",
        "user": gin.H{
            "id":    user.ID.Hex(),
            "name":  user.Name,
            "email": user.Email,
        },
    })
}
```

### 3. Middleware de Autenticação

O middleware protege rotas verificando a validade do token JWT presente no cookie:

```go
// handlers/auth_middleware.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Log para depuração
        log.Printf("[AuthMiddleware] Headers: %v", c.Request.Header)
        log.Printf("[AuthMiddleware] Cookies: %v", c.Request.Cookies())
        
        // Obter token do cookie
        accessToken, err := c.Cookie("access_token")
        if err != nil {
            log.Printf("[AuthMiddleware] Erro ao obter cookie access_token: %v", err)
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Cookie de acesso não encontrado"})
            c.Abort()
            return
        }
        
        // Validar token
        claims, err := utils.ValidateToken(accessToken)
        if err != nil {
            log.Printf("[AuthMiddleware] Token inválido: %v", err)
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido ou expirado"})
            c.Abort()
            return
        }
        
        // Adicionar claims ao contexto
        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

### 4. Refresh Token

O handler de refresh token verifica o refresh_token e emite novos tokens:

```go
// handlers/auth_handler.go
func RefreshToken(c *gin.Context) {
    // Log para depuração
    log.Printf("[RefreshToken] Received request. Headers: %+v", c.Request.Header)
    log.Printf("[RefreshToken] Received cookies: %+v", c.Request.Cookies())
    
    // Obter refresh token do cookie
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        log.Printf("[RefreshToken] Erro ao obter refresh_token: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token não encontrado"})
        return
    }
    
    // Validar refresh token
    claims, err := utils.ValidateToken(refreshToken)
    if err != nil {
        log.Printf("[RefreshToken] Token inválido: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token inválido"})
        return
    }
    
    // Gerar novo access token
    accessToken, _ := utils.GenerateAccessToken(claims.UserID)
    
    // Definir novo cookie access_token
    accessTokenCookie := http.Cookie{
        Name:     "access_token",
        Value:    accessToken,
        MaxAge:   15 * 60,
        Path:     "/",
        Domain:   "localhost",
        Secure:   false,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    }
    
    http.SetCookie(c.Writer, &accessTokenCookie)
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Token renovado com sucesso",
    })
}
```

### 5. Logout

O handler de logout invalida os cookies removendo-os:

```go
// handlers/auth_handler.go
func LogoutUser(c *gin.Context) {
    // Invalidar cookies definindo MaxAge negativo
    accessTokenCookie := http.Cookie{
        Name:     "access_token",
        Value:    "",
        MaxAge:   -1,
        Path:     "/",
        Domain:   "localhost",
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    }
    
    refreshTokenCookie := http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        MaxAge:   -1,
        Path:     "/",
        Domain:   "localhost",
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    }
    
    http.SetCookie(c.Writer, &accessTokenCookie)
    http.SetCookie(c.Writer, &refreshTokenCookie)
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Logout realizado com sucesso",
    })
}
```

### 6. Verificação do Usuário Atual

Rota protegida que retorna dados do usuário autenticado:

```go
// main.go
func main() {
    // ...
    
    // Grupo de rotas protegidas
    protected := r.Group("/api/v1/identity")
    protected.Use(handlers.AuthMiddleware())
    {
        protected.GET("/me", func(c *gin.Context) {
            userID, _ := c.Get("userID")
            
            // Buscar dados do usuário no banco de dados
            user, err := db.GetUserByID(userID.(string))
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "Erro ao buscar dados do usuário",
                })
                return
            }
            
            c.JSON(http.StatusOK, gin.H{
                "id":    user.ID.Hex(),
                "name":  user.Name,
                "email": user.Email,
            })
        })
    }
    
    // ...
}
```

## Utilitários JWT

```go
// utils/token_utils.go
func GenerateAccessToken(userID string) (string, error) {
    // Criar token com tempo de expiração curto (15 minutos)
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Minute * 15).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(getSecretKey()))
}

func GenerateRefreshToken(userID string) (string, error) {
    // Criar token com tempo de expiração longo (7 dias)
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(getSecretKey()))
}

func ValidateToken(tokenString string) (*TokenClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
        }
        return []byte(getSecretKey()), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &TokenClaims{
            UserID: claims["user_id"].(string),
        }, nil
    }
    
    return nil, errors.New("token inválido")
}

func getSecretKey() string {
    secret := os.Getenv("JWT_SECRET_KEY")
    if secret == "" {
        secret = "chave_secreta_insegura_para_desenvolvimento" // Em produção, deve ser configurada no ambiente
    }
    return secret
}
```

## Segurança e Considerações de Ambiente

### Ambiente de Desenvolvimento

- `Secure: false` nos cookies para permitir HTTP
- `SameSite: http.SameSiteLaxMode` para flexibilidade em testes
- Logs detalhados para depuração
- Chave JWT secreta com fallback para desenvolvimento

### Ambiente de Produção

- Configurar `Secure: true` em todos os cookies para exigir HTTPS
- Utilizar `SameSite: http.SameSiteStrictMode` para maior segurança
- Remover logs detalhados ou configurar nível de log adequado
- Definir `JWT_SECRET_KEY` como variável de ambiente com valor seguro

## Dependências

```go
// go.mod (resumido)
require (
    github.com/gin-contrib/cors v1.5.0
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-jwt/jwt/v5 v5.2.0
    go.mongodb.org/mongo-driver v1.13.1
    golang.org/x/crypto v0.17.0  // Para bcrypt
)
```

## Configuração de Rotas

```go
// main.go
func main() {
    r := gin.Default()
    
    // Configuração CORS (omitida para brevidade)
    
    // Rotas públicas
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })
    
    public := r.Group("/api/v1/identity")
    {
        public.GET("/", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"message": "Identity Service API"})
        })
        public.POST("/register", handlers.RegisterUser)
        public.POST("/login", handlers.LoginUser)
        public.POST("/logout", handlers.LogoutUser)
        public.POST("/refresh", handlers.RefreshToken)
    }
    
    // Rotas protegidas
    protected := r.Group("/api/v1/identity")
    protected.Use(handlers.AuthMiddleware())
    {
        protected.GET("/me", ...)  // Implementação omitida para brevidade
    }
    
    r.Run(":8085")
}
```

## Testes Manuais

Para testar a autenticação, utilize a página de teste HTML `/test-auth.html` que inclui:

1. Formulário de login com campos de email e senha
2. Botão para verificar estado de autenticação (GET /me)
3. Botão para testar refresh token
4. Botão para logout
5. Visualização dos cookies disponíveis (não HttpOnly)

## Melhorias Futuras

1. **Rotação de Refresh Tokens**: Implementar rotação a cada uso para maior segurança
2. **Revogação de Tokens**: Adicionar lista negra de tokens revogados
3. **Monitoramento de Sessões**: Permitir visualizar todas as sessões ativas
4. **Rate Limiting**: Proteger contra ataques de força bruta
5. **Auditoria**: Registrar todas as ações de autenticação para fins de segurança

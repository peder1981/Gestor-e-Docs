package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"gestor-e-docs/backend/services/identity-service/db"
	"gestor-e-docs/backend/services/identity-service/models"
	"gestor-e-docs/backend/services/identity-service/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// HeadHandler é um handler genérico para requisições HEAD
// Retorna apenas os headers necessários para CORS/preflight
func HeadHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

// RegisterUser registra um novo usuário no sistema
func RegisterUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existingUser models.User
	log.Printf("[RegisterUser] Attempting to find user with email: %s", user.Email)
	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		log.Printf("[RegisterUser] Email %s already exists. User: %+v", user.Email, existingUser)
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("[RegisterUser] Error finding user %s: %v", user.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking for existing user"})
		return
	}
	log.Printf("[RegisterUser] Email %s not found (err is mongo.ErrNoDocuments). Proceeding with registration.", user.Email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)
	user.Role = "user" // Define o papel padrão
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "userId": user.ID})
}

// LoginRequest define a estrutura esperada para o corpo da requisição de login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginUser autentica um usuário e retorna um token JWT
func LoginUser(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	log.Printf("[AUTH_DEBUG] Login attempt for email: %s", req.Email)

	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var foundUser models.User
	log.Printf("[LoginUser] Attempting to find user with email: %s", req.Email)
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&foundUser)
	if err == nil {
		log.Printf("[AUTH_DEBUG] User found. DB Password Hash: %s", foundUser.Password)
	}
	if err == mongo.ErrNoDocuments {
		log.Printf("[LoginUser] User with email %s not found.", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	} else if err != nil {
		log.Printf("[LoginUser] Error finding user %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding user"})
		return
	}
	log.Printf("[LoginUser] User %s found. Verifying password.", req.Email)

	if err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password)); err != nil {
		log.Printf("[AUTH_DEBUG] Password comparison FAILED for email %s. Error: %v", req.Email, err)

		log.Printf("[LoginUser] Password verification failed for user %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	log.Printf("[LoginUser] User %s authenticated successfully. Generating tokens...", req.Email)

	// Determinar o modo SameSite com base no ambiente
	sameSiteMode := http.SameSiteLaxMode // Padrão mais permissivo
	if os.Getenv("COOKIE_SAMESITE_STRICT") == "true" {
		sameSiteMode = http.SameSiteStrictMode
	}

	// Gera os tokens
	accessToken, err := utils.GenerateAccessToken(foundUser.ID.Hex())
	if err != nil {
		log.Printf("[LoginUser] Error generating access token for user %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(foundUser.ID.Hex())
	if err != nil {
		log.Printf("[LoginUser] Error generating refresh token for user %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Define os cookies com configurações mais permissivas para desenvolvimento
	isSecure := os.Getenv("GIN_MODE") == "release"
	log.Printf("[LoginUser] Setting cookies with Secure=%v for user %s", isSecure, req.Email)

	// Define o cookie de access token
	accessTokenCookie := http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		MaxAge:   15 * 60, // 15 minutos
		Path:     "/",
		Domain:   "", // Vazio para usar o domínio atual
		Secure:   true, // Sempre true pois usamos HTTPS via Nginx
		HttpOnly: true,
		SameSite: sameSiteMode, // Usar a mesma configuração dinâmica do refresh_token
	}

	// Log detalhado do cookie de access token
	log.Printf("[LoginUser] Setting access_token cookie: {Name: %s, Path: %s, Secure: %v, HttpOnly: %v, SameSite: %v}",
		accessTokenCookie.Name,
		accessTokenCookie.Path,
		accessTokenCookie.Secure,
		accessTokenCookie.HttpOnly,
		accessTokenCookie.SameSite)

	http.SetCookie(c.Writer, &accessTokenCookie)

	// Define o cookie de refresh token
	log.Printf("[LoginUser] Configurando cookie refresh_token com SameSite=%v", sameSiteMode)
	refreshTokenCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   7 * 24 * 60 * 60, // 7 dias
		Path:     "/",
		Domain:   "", // Vazio para usar o domínio atual
		Secure:   true, // Sempre true pois usamos HTTPS via Nginx
		HttpOnly: true,
		SameSite: sameSiteMode, // Usar o mesmo modo do access_token
	}
	log.Printf("[LoginUser] Cookie refresh_token configurado: Domain=%q, Path=%q, Secure=%v, HttpOnly=%v", refreshTokenCookie.Domain, refreshTokenCookie.Path, refreshTokenCookie.Secure, refreshTokenCookie.HttpOnly)

	// Log detalhado do cookie de refresh token
	log.Printf("[LoginUser] Setting refresh_token cookie: {Name: %s, Path: %s, Domain: %s, Secure: %v, HttpOnly: %v, SameSite: %v}",
		refreshTokenCookie.Name,
		refreshTokenCookie.Path,
		refreshTokenCookie.Domain,
		refreshTokenCookie.Secure,
		refreshTokenCookie.HttpOnly,
		refreshTokenCookie.SameSite)

	// Log dos headers da requisição
	log.Printf("[LoginUser] Método da requisição: %s", c.Request.Method)
	log.Printf("[LoginUser] URL da requisição: %s", c.Request.URL.String())
	log.Printf("[LoginUser] Protocolo: %s", c.Request.Proto)
	log.Printf("[LoginUser] RemoteAddr: %s", c.Request.RemoteAddr)
	log.Printf("[LoginUser] Todos os headers da requisição: %v", c.Request.Header)
	log.Printf("[LoginUser] Headers específicos: Origin=%s, Referer=%s, X-Forwarded-Proto=%s, X-Real-IP=%s, X-Forwarded-For=%s",
		c.Request.Header.Get("Origin"),
		c.Request.Header.Get("Referer"),
		c.Request.Header.Get("X-Forwarded-Proto"),
		c.Request.Header.Get("X-Real-IP"),
		c.Request.Header.Get("X-Forwarded-For"))

	http.SetCookie(c.Writer, &refreshTokenCookie)

	// Log dos headers da resposta para depuração
	log.Printf("[LoginUser] Response headers: Access-Control-Allow-Origin=%s, Access-Control-Allow-Credentials=%s",
		c.Writer.Header().Get("Access-Control-Allow-Origin"),
		c.Writer.Header().Get("Access-Control-Allow-Credentials"))

	// Log dos cookies definidos na resposta
	log.Printf("[LoginUser] Set-Cookie headers: %v", c.Writer.Header()["Set-Cookie"])

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":    foundUser.ID,
			"name":  foundUser.Name,
			"email": foundUser.Email,
		},
	})
}

// LogoutUser invalida os cookies de autenticação do usuário
func LogoutUser(c *gin.Context) {
	// Define os cookies com MaxAge -1 para instruir o navegador a excluí-los
	// Mantendo a mesma configuração dos cookies usada no login
	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	log.Printf("[LogoutUser] Cookies removidos com Secure=true e sem restrição de domínio")
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// RefreshToken renova o access_token usando um refresh_token válido.
// GetUserProfile busca e retorna os dados do usuário autenticado.
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// Convertendo o ID de string para ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	if err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// RefreshToken renova o access_token usando um refresh_token válido.
func RefreshToken(c *gin.Context) {
	log.Printf("[RefreshToken] Received request. Headers: %+v", c.Request.Header)
	log.Printf("[RefreshToken] Received cookies: %+v", c.Request.Cookies())
	refreshTokenString, err := c.Cookie("refresh_token")
	log.Printf("[RefreshToken] Attempting to read 'refresh_token' cookie. Error: %v, Value: %s", err, refreshTokenString)
	if err != nil {
		if err == http.ErrNoCookie {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading refresh token cookie"})
		return
	}

	claims, err := utils.ValidateToken(refreshTokenString)
	if err != nil {
		log.Printf("[RefreshToken] Refresh token validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	log.Printf("[RefreshToken] Refresh token for user %s is valid. Generating new access token.", claims.Subject)
	newAccessToken, err := utils.GenerateAccessToken(claims.Subject)
	if err != nil {
		log.Printf("[RefreshToken] Error generating new access token for user %s: %v", claims.Subject, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}

	// Determinar o modo SameSite com base no ambiente
	sameSiteMode := http.SameSiteLaxMode // Padrão mais permissivo
	if os.Getenv("COOKIE_SAMESITE_STRICT") == "true" {
		sameSiteMode = http.SameSiteStrictMode
	}

	// Como estamos usando HTTPS via Nginx, sempre definimos Secure como true
	log.Printf("[RefreshToken] Configurando cookie access_token com SameSite=%v", sameSiteMode)
	newAccessTokenCookie := http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		MaxAge:   15 * 60, // 15 minutos
		Path:     "/",
		Domain:   "", // Vazio para usar o domínio atual
		Secure:   true, // Sempre true pois usamos HTTPS via Nginx
		HttpOnly: true,
		SameSite: sameSiteMode,
	}
	log.Printf("[RefreshToken] Cookie access_token configurado: Domain=%q, Path=%q, Secure=%v, HttpOnly=%v", newAccessTokenCookie.Domain, newAccessTokenCookie.Path, newAccessTokenCookie.Secure, newAccessTokenCookie.HttpOnly)
	http.SetCookie(c.Writer, &newAccessTokenCookie)
	log.Printf("[RefreshToken] Cookie de access token atualizado: Domain=%q, Path=%q, Secure=%v, HttpOnly=%v, SameSite=%v",
		newAccessTokenCookie.Domain,
		newAccessTokenCookie.Path,
		newAccessTokenCookie.Secure,
		newAccessTokenCookie.HttpOnly,
		newAccessTokenCookie.SameSite)

	// Log dos headers da requisição
	log.Printf("[RefreshToken] Request headers: Origin=%s, Referer=%s",
		c.Request.Header.Get("Origin"),
		c.Request.Header.Get("Referer"))

	// Log dos headers da resposta
	log.Printf("[RefreshToken] Response headers: Access-Control-Allow-Origin=%s, Access-Control-Allow-Credentials=%s",
		c.Writer.Header().Get("Access-Control-Allow-Origin"),
		c.Writer.Header().Get("Access-Control-Allow-Credentials"))

	// Log dos cookies definidos na resposta
	log.Printf("[RefreshToken] Set-Cookie headers: %v", c.Writer.Header()["Set-Cookie"])

	log.Printf("[RefreshToken] New access token generated and set for user %s.", claims.Subject)
	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

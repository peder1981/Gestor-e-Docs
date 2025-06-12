package handlers

import (
	"context"
	"log"
	"net/http"
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
		Secure:   isSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
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
	refreshTokenCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   7 * 24 * 60 * 60, // 7 dias
		Path:     "/",
		Secure:   isSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	// Log detalhado do cookie de refresh token
	log.Printf("[LoginUser] Setting refresh_token cookie: {Name: %s, Path: %s, Secure: %v, HttpOnly: %v, SameSite: %v}",
		refreshTokenCookie.Name,
		refreshTokenCookie.Path,
		refreshTokenCookie.Secure,
		refreshTokenCookie.HttpOnly,
		refreshTokenCookie.SameSite)

	http.SetCookie(c.Writer, &refreshTokenCookie)

	// Log dos headers da resposta para depuração
	log.Printf("[LoginUser] Response headers: %+v", c.Writer.Header())

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
	// Removendo Domain e configurando Secure=true para consistência
	isSecure := os.Getenv("GIN_MODE") == "release"
	c.SetCookie("access_token", "", -1, "/", "", isSecure, true)
	c.SetCookie("refresh_token", "", -1, "/", "", isSecure, true)
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

	// Como estamos usando HTTPS, Secure deve ser true.
	newAccessTokenCookie := http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		MaxAge:   15 * 60, // 15 minutos
		Path:     "/",
		// Removendo Domain para aceitar qualquer domínio
		Secure:   os.Getenv("GIN_MODE") == "release", // Secure apenas em produção
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, &newAccessTokenCookie)
	log.Printf("[RefreshToken] Cookie de access token atualizado com Secure=true e sem restrição de domínio")

	log.Printf("[RefreshToken] New access token generated and set for user %s.", claims.Subject)
	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

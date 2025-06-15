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

// RegisterRequest define a estrutura esperada para o corpo da requisição de registro
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterUser registra um novo usuário no sistema
func RegisterUser(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Criar o modelo User a partir do request
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
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

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Captura o ID gerado pelo MongoDB
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inserted user ID"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "userId": insertedID.Hex()})
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



	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var foundUser models.User
	log.Printf("[LoginUser] Attempting to find user with email: %s", req.Email)
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&foundUser)
	if err == nil {

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(foundUser.ID.Hex())
	if err != nil {
		log.Printf("[LoginUser] Error generating refresh token for user %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

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

	http.SetCookie(c.Writer, &accessTokenCookie)

	// Define o cookie de refresh token
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

	http.SetCookie(c.Writer, &refreshTokenCookie)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":    foundUser.ID.Hex(),
			"name":  foundUser.Name,
			"email": foundUser.Email,
		},
	})
}

// LogoutUser invalida os cookies de autenticação do usuário
func LogoutUser(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	log.Printf("[LogoutUser] Cookies removidos com Secure=true e sem restrição de domínio")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetUserProfile busca e retorna os dados do usuário autenticado.
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID.Hex(),
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// RefreshToken renova o access_token usando um refresh_token válido.
func RefreshToken(c *gin.Context) {
	refreshTokenString, err := c.Cookie("refresh_token")
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	newAccessToken, err := utils.GenerateAccessToken(claims.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}

	sameSiteMode := http.SameSiteLaxMode 
	if os.Getenv("COOKIE_SAMESITE_STRICT") == "true" {
		sameSiteMode = http.SameSiteStrictMode
	}

	newAccessTokenCookie := http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		MaxAge:   15 * 60, 
		Path:     "/",
		Domain:   "", 
		Secure:   true, 
		HttpOnly: true,
		SameSite: sameSiteMode,
	}
	http.SetCookie(c.Writer, &newAccessTokenCookie)

	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"gestor-e-docs/identity-service/db"
	"gestor-e-docs/identity-service/models"
	"gestor-e-docs/identity-service/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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

	collection := db.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var foundUser models.User
	log.Printf("[LoginUser] Attempting to find user with email: %s", req.Email)
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&foundUser)
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

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password)); err != nil {
		log.Printf("[LoginUser] Password verification failed for user %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	log.Printf("[LoginUser] User %s authenticated successfully. Generating tokens...", req.Email)

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

	log.Printf("[LoginUser] Tokens generated successfully for user %s. Setting cookies.", req.Email)

	// Definir os cookies HTTP-only
	// Para desenvolvimento (HTTP), Secure deve ser false. Em produção (HTTPS), deve ser true.
	// Como estamos usando HTTPS, Secure pode ser sempre true.
	// SameSite=None requer que Secure seja true.
	accessTokenCookie := http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		MaxAge:   15 * 60, // 15 minutos
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, &accessTokenCookie)

	refreshTokenCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   7 * 24 * 60 * 60, // 7 dias
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, &refreshTokenCookie)

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
	c.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
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

	// Como estamos usando HTTPS, Secure pode ser sempre true.
	// SameSite=None requer que Secure seja true.
	newAccessTokenCookie := http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		MaxAge:   15 * 60, // 15 minutos
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, &newAccessTokenCookie)

	log.Printf("[RefreshToken] New access token generated and set for user %s.", claims.Subject)
	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

package handlers

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"gestor-e-docs/backend/services/identity-service/db"
	"gestor-e-docs/backend/services/identity-service/security"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var twoFactorAuth *security.TwoFactorAuth

// InitTwoFactorAuth inicializa o serviço de autenticação de dois fatores
func InitTwoFactorAuth() {
	twoFactorAuth = security.NewTwoFactorAuth(db.GetDatabase())
	log.Println("[2FA] Serviço de autenticação de dois fatores inicializado")
}

// GetTwoFactorAuth retorna a instância do serviço 2FA
func GetTwoFactorAuth() *security.TwoFactorAuth {
	return twoFactorAuth
}

// Setup2FARequest define o corpo da requisição para configurar 2FA
type Setup2FARequest struct {
	VerificationCode string `json:"verification_code,omitempty"`
}

// GenerateSetup2FA gera a configuração inicial de 2FA para um usuário
func GenerateSetup2FA(c *gin.Context) {
	// Verifica a autenticação do usuário
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Busca o usuário no banco para obter o e-mail (username)
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	collection := db.GetCollection("users")
	ctx := context.Background()
	
	var user struct {
		Email string `bson:"email"`
		Name  string `bson:"name"`
	}
	
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding user"})
		return
	}

	// Gera o segredo 2FA
	config, otpURL, qrCode, err := twoFactorAuth.GenerateSecret(
		ctx, 
		userID.(string),
		user.Email,
		"Gestor-e-Docs",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate 2FA secret: " + err.Error(),
		})
		return
	}

	// Codifica o QR code em base64 para envio ao frontend
	qrCodeBase64 := "data:image/png;base64," + encodeToBase64(qrCode)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "2FA setup generated",
		"secret": config.Secret,
		"otpauth_url": otpURL,
		"qr_code": qrCodeBase64,
		"backup_codes": config.BackupCodes,
	})
}

// Verify2FA verifica e ativa a autenticação de dois fatores
func Verify2FA(c *gin.Context) {
	// Verifica a autenticação do usuário
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req Setup2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Valida o código fornecido e habilita o 2FA
	err := twoFactorAuth.EnableTwoFactor(c.Request.Context(), userID.(string), req.VerificationCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "Invalid verification code: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Two-factor authentication enabled successfully",
	})
}

// Disable2FA desativa a autenticação de dois fatores
func Disable2FA(c *gin.Context) {
	// Verifica a autenticação do usuário
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req Setup2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Valida o código antes de desabilitar
	valid, err := twoFactorAuth.ValidateCode(c.Request.Context(), userID.(string), req.VerificationCode)
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": "Invalid verification code",
		})
		return
	}

	// Desabilita o 2FA
	err = twoFactorAuth.DisableTwoFactor(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "Failed to disable 2FA: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Two-factor authentication disabled successfully",
	})
}

// GetTwoFactorStatus verifica se o 2FA está habilitado para o usuário
func GetTwoFactorStatus(c *gin.Context) {
	// Verifica a autenticação do usuário
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	enabled, err := twoFactorAuth.IsTwoFactorEnabled(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check 2FA status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"enabled": enabled,
	})
}

// Login2FA verifica o código 2FA durante o login
func Login2FA(c *gin.Context) {
	userID, exists := c.Get("2fa_pending_user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No pending 2FA verification"})
		return
	}

	var req Setup2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Valida o código 2FA
	valid, err := twoFactorAuth.ValidateCode(c.Request.Context(), userID.(string), req.VerificationCode)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": "Invalid verification code",
		})
		return
	}

	// Cria token de verificação 2FA válido para a sessão
	token, err := twoFactorAuth.CreateVerificationToken(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "Failed to create verification token",
		})
		return
	}

	// Define o cookie com o token 2FA
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "2fa_token",
		Value:    token,
		MaxAge:   12 * 60 * 60, // 12 horas
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Two-factor authentication verified successfully",
	})

	// Limpar o estado de verificação pendente
	c.Set("2fa_pending_user_id", nil)
}

// TwoFactorMiddleware verifica se o usuário tem 2FA ativado e, se sim, exige verificação
func TwoFactorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pula verificação para endpoints públicos e endpoints de 2FA
		if isExemptPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Verifica se o usuário está autenticado
		userID, exists := c.Get("userID")
		if !exists {
			c.Next() // Deixa o middleware de autenticação lidar com isso
			return
		}

		// Verifica se o usuário tem 2FA habilitado
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		enabled, err := twoFactorAuth.IsTwoFactorEnabled(ctx, userID.(string))
		if err != nil {
			log.Printf("[2FA] Erro ao verificar status do 2FA: %v", err)
			c.Next() // Em caso de erro, continua sem exigir 2FA
			return
		}

		// Se o 2FA não estiver habilitado, continua normalmente
		if !enabled {
			c.Next()
			return
		}

		// Verifica o token 2FA no cookie
		cookie, err := c.Request.Cookie("2fa_token")
		if err != nil || cookie.Value == "" {
			log.Printf("[2FA] Token 2FA não encontrado para usuário %s", userID.(string))
			
			// Define o usuário com verificação 2FA pendente
			c.Set("2fa_pending_user_id", userID)
			
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": "Two-factor authentication required",
				"code": "2FA_REQUIRED",
			})
			return
		}

		// Verifica se o token 2FA é válido
		tokenUserID, err := twoFactorAuth.VerifyToken(ctx, cookie.Value)
		if err != nil || tokenUserID != userID.(string) {
			log.Printf("[2FA] Token 2FA inválido: %v", err)
			
			// Define o usuário com verificação 2FA pendente
			c.Set("2fa_pending_user_id", userID)
			
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": "Two-factor authentication required",
				"code": "2FA_REQUIRED",
			})
			return
		}

		// Usuário verificado com 2FA, continua
		c.Next()
	}
}

// isExemptPath verifica se um caminho está isento de verificação 2FA
func isExemptPath(path string) bool {
	exemptPaths := []string{
		"/api/v1/identity/login",
		"/api/v1/identity/register",
		"/api/v1/identity/logout",
		"/api/v1/identity/refresh",
		"/api/v1/identity/2fa/verify",
		"/api/v1/identity/health",
		"/api/v1/identity/metrics",
	}

	for _, exemptPath := range exemptPaths {
		if path == exemptPath {
			return true
		}
	}

	return false
}

// encodeToBase64 codifica um array de bytes para uma string base64
func encodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

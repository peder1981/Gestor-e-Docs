package security

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TwoFactorConfig representa a configuração de 2FA para um usuário
type TwoFactorConfig struct {
	UserID      string    `bson:"user_id" json:"user_id"`
	Secret      string    `bson:"secret" json:"secret,omitempty"`
	Enabled     bool      `bson:"enabled" json:"enabled"`
	LastUpdated time.Time `bson:"last_updated" json:"last_updated"`
	BackupCodes []string  `bson:"backup_codes" json:"backup_codes,omitempty"`
}

// TwoFactorVerification representa uma tentativa de verificação 2FA
type TwoFactorVerification struct {
	UserID     string    `bson:"user_id" json:"user_id"`
	Token      string    `bson:"token" json:"token"`
	Expiration time.Time `bson:"expiration" json:"expiration"`
	VerifiedAt time.Time `bson:"verified_at,omitempty" json:"verified_at,omitempty"`
}

// TwoFactorAuth gerencia a autenticação de dois fatores
type TwoFactorAuth struct {
	configCollection   *mongo.Collection
	verifyCollection   *mongo.Collection
	verificationWindow time.Duration
}

// NewTwoFactorAuth cria uma nova instância do gerenciador 2FA
func NewTwoFactorAuth(db *mongo.Database) *TwoFactorAuth {
	return &TwoFactorAuth{
		configCollection:   db.Collection("two_factor_configs"),
		verifyCollection:   db.Collection("two_factor_verifications"),
		verificationWindow: 5 * time.Minute, // Token de verificação válido por 5 minutos
	}
}

// GenerateSecret gera um novo segredo 2FA para um usuário
func (tfa *TwoFactorAuth) GenerateSecret(ctx context.Context, userID, username, issuer string) (*TwoFactorConfig, string, []byte, error) {
	// Verifica se o usuário já tem configuração 2FA
	var existingConfig TwoFactorConfig
	err := tfa.configCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&existingConfig)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, "", nil, err
	}

	// Se a configuração já existe e está habilitada, não permite gerar uma nova
	if err == nil && existingConfig.Enabled {
		return nil, "", nil, fmt.Errorf("2FA já está habilitado para este usuário")
	}

	// Configura o gerador TOTP
	opts := totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
	}
	key, err := totp.Generate(opts)
	if err != nil {
		return nil, "", nil, err
	}

	// Gera códigos de backup
	backupCodes, err := generateBackupCodes(8) // 8 códigos de backup
	if err != nil {
		return nil, "", nil, err
	}

	// Cria o QR code
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, "", nil, err
	}
	if err := png.Encode(&buf, img); err != nil {
		return nil, "", nil, err
	}
	qrCode := buf.Bytes()

	// Prepara a configuração
	config := TwoFactorConfig{
		UserID:      userID,
		Secret:      key.Secret(),
		Enabled:     false, // Inicialmente desativado até confirmação
		LastUpdated: time.Now(),
		BackupCodes: backupCodes,
	}

	// Salva ou atualiza a configuração
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": config}
	upsert := true
	_, err = tfa.configCollection.UpdateOne(ctx, filter, update,
		options.Update().SetUpsert(upsert))
	if err != nil {
		return nil, "", nil, err
	}

	return &config, key.URL(), qrCode, nil
}

// ValidateCode valida um código 2FA fornecido pelo usuário
func (tfa *TwoFactorAuth) ValidateCode(ctx context.Context, userID, code string) (bool, error) {
	// Busca a configuração 2FA do usuário
	var config TwoFactorConfig
	err := tfa.configCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, fmt.Errorf("2FA não configurado para este usuário")
		}
		return false, err
	}

	// Se 2FA não está habilitado, não valida
	if !config.Enabled {
		return false, nil
	}

	// Verifica se é um código de backup
	for i, backupCode := range config.BackupCodes {
		if backupCode == code {
			// Remove o código de backup usado
			config.BackupCodes = append(config.BackupCodes[:i], config.BackupCodes[i+1:]...)

			// Atualiza a configuração
			update := bson.M{"$set": bson.M{"backup_codes": config.BackupCodes}}
			_, err := tfa.configCollection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
			if err != nil {
				log.Printf("[2FA] Erro ao atualizar códigos de backup: %v", err)
			}

			return true, nil
		}
	}

	// Valida o código TOTP
	valid := totp.Validate(code, config.Secret)
	return valid, nil
}

// EnableTwoFactor habilita o 2FA após validação bem-sucedida
func (tfa *TwoFactorAuth) EnableTwoFactor(ctx context.Context, userID, code string) error {
	valid, err := tfa.ValidateCode(ctx, userID, code)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("código inválido")
	}

	// Atualiza o status para habilitado
	update := bson.M{"$set": bson.M{"enabled": true, "last_updated": time.Now()}}
	_, err = tfa.configCollection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	return err
}

// DisableTwoFactor desabilita o 2FA para um usuário
func (tfa *TwoFactorAuth) DisableTwoFactor(ctx context.Context, userID string) error {
	_, err := tfa.configCollection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}

// CreateVerificationToken cria um token de verificação para login 2FA
func (tfa *TwoFactorAuth) CreateVerificationToken(ctx context.Context, userID string) (string, error) {
	// Gera um token aleatório
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.StdEncoding.EncodeToString(tokenBytes)

	// Cria o registro de verificação
	verification := TwoFactorVerification{
		UserID:     userID,
		Token:      token,
		Expiration: time.Now().Add(tfa.verificationWindow),
	}

	// Salva no banco
	_, err := tfa.verifyCollection.InsertOne(ctx, verification)
	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyToken verifica se um token 2FA é válido e o marca como usado
func (tfa *TwoFactorAuth) VerifyToken(ctx context.Context, token string) (string, error) {
	// Busca o token
	var verification TwoFactorVerification
	err := tfa.verifyCollection.FindOne(ctx, bson.M{
		"token":       token,
		"expiration":  bson.M{"$gt": time.Now()},
		"verified_at": bson.M{"$exists": false},
	}).Decode(&verification)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("token inválido ou expirado")
		}
		return "", err
	}

	// Marca o token como verificado
	update := bson.M{"$set": bson.M{"verified_at": time.Now()}}
	_, err = tfa.verifyCollection.UpdateOne(
		ctx,
		bson.M{"token": token},
		update,
	)
	if err != nil {
		return "", err
	}

	return verification.UserID, nil
}

// generateBackupCodes gera códigos de backup aleatórios
func generateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Gera 10 bytes aleatórios
		b := make([]byte, 10)
		_, err := rand.Read(b)
		if err != nil {
			return nil, err
		}

		// Codifica para base32 e formata
		encoded := strings.ToUpper(base32.StdEncoding.EncodeToString(b))
		codes[i] = fmt.Sprintf("%s-%s", encoded[:5], encoded[5:10])
	}
	return codes, nil
}

// IsTwoFactorEnabled verifica se o 2FA está habilitado para um usuário
func (tfa *TwoFactorAuth) IsTwoFactorEnabled(ctx context.Context, userID string) (bool, error) {
	var config TwoFactorConfig
	err := tfa.configCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return config.Enabled, nil
}

// TwoFactorMiddleware verifica a autenticação de 2 fatores quando necessário
func TwoFactorRequiredMiddleware(tfa *TwoFactorAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pula verificação para endpoints que não precisam de 2FA
		if isExemptPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Obtém o ID do usuário do contexto (já autenticado com JWT)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Verifica se 2FA está habilitado para o usuário
		enabled, err := tfa.IsTwoFactorEnabled(c.Request.Context(), userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check 2FA status"})
			c.Abort()
			return
		}

		// Se 2FA não está habilitado, continua normalmente
		if !enabled {
			c.Next()
			return
		}

		// Verifica se o usuário está verificado com 2FA
		verified, exists := c.Get("2fa_verified")
		if !exists || !verified.(bool) {
			// Verifica se existe um token de verificação na requisição
			token := c.Request.Header.Get("X-2FA-Token")
			if token == "" {
				// Não tem token, exige autenticação 2FA
				c.JSON(http.StatusForbidden, gin.H{
					"error": "2FA verification required",
					"code":  "2FA_REQUIRED",
				})
				c.Abort()
				return
			}

			// Verifica o token
			verifiedUserID, err := tfa.VerifyToken(c.Request.Context(), token)
			if err != nil || verifiedUserID != userID.(string) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Invalid 2FA token",
					"code":  "INVALID_2FA_TOKEN",
				})
				c.Abort()
				return
			}

			// Token válido, marca usuário como verificado com 2FA
			c.Set("2fa_verified", true)
		}

		c.Next()
	}
}

// isExemptPath verifica se um caminho está isento da verificação 2FA
func isExemptPath(path string) bool {
	exemptPaths := []string{
		"/api/v1/identity/login",
		"/api/v1/identity/2fa/setup",
		"/api/v1/identity/2fa/verify",
		"/api/v1/identity/health",
		"/api/v1/identity/metrics",
	}

	for _, exemptPath := range exemptPaths {
		if strings.HasPrefix(path, exemptPath) {
			return true
		}
	}

	return false
}

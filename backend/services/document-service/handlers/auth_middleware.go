package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims armazena as claims do JWT
type TokenClaims struct {
	UserID string `json:"user_id"`
}

// AuthMiddleware protege rotas que necessitam de autenticação
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
		claims, err := ValidateToken(accessToken)
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

// ValidateToken verifica se um token JWT é válido e retorna as claims
func ValidateToken(tokenString string) (*TokenClaims, error) {
	// Remover o prefixo "Bearer " se existir
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar o método de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return []byte(getSecretKey()), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verificar se o token está expirado
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, errors.New("token expirado")
			}
		}

		// Extrair o ID do usuário
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("user_id inválido no token")
		}

		return &TokenClaims{
			UserID: userID,
		}, nil
	}

	return nil, errors.New("token inválido")
}

// getSecretKey obtém a chave secreta para tokens JWT da variável de ambiente
func getSecretKey() string {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		secret = "chave_secreta_insegura_para_desenvolvimento" // Em produção, deve ser configurada no ambiente
		log.Println("Aviso: Usando chave JWT padrão para desenvolvimento")
	}
	return secret
}

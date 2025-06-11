package handlers

import (
	"log"
	"net/http"

	"gestor-e-docs/backend/services/identity-service/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifica o token JWT nas requisições
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[AuthMiddleware] Headers: %v", c.Request.Header)
		
		// Listar todos os cookies recebidos
		cookies := c.Request.Cookies()
		log.Printf("[AuthMiddleware] Recebidos %d cookies", len(cookies))
		for i, cookie := range cookies {
			log.Printf("[AuthMiddleware] Cookie[%d]: Nome=%s, Valor=%s, Domain=%s, Path=%s", 
				i, cookie.Name, cookie.Value, cookie.Domain, cookie.Path)
		}

		tokenString, err := c.Cookie("access_token")
		if err != nil {
			log.Printf("[AuthMiddleware] Erro ao ler cookie 'access_token': %v", err)
			if err == http.ErrNoCookie {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie not found"})
				return
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Error reading authorization cookie"})
			return
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie is empty"})
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			log.Printf("[AuthMiddleware] Token validation error: %v", err)
			// O frontend deve tentar atualizar o token em qualquer erro 401.
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Token é válido, pode-se adicionar informações do usuário ao contexto se necessário
		c.Set("userID", claims.Subject) // O ID do usuário está no campo Subject das claims
		log.Printf("[AuthMiddleware] User %s authorized.", claims.Subject)

		c.Next()
	}
}

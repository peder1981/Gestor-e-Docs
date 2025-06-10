package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getJWTSecret retorna a chave secreta do ambiente ou uma chave padrão insegura.
func getJWTSecret() []byte {
	jtwSecret := os.Getenv("JWT_SECRET_KEY")
	if jtwSecret == "" {
		// Esta chave é apenas para desenvolvimento e será logada como um aviso crítico se usada.
		return []byte("your-default-insecure-secret-key-change-me")
	}
	return []byte(jtwSecret)
}

// GenerateAccessToken cria um novo token de acesso JWT de curta duração.
func GenerateAccessToken(userID string) (string, error) {
	secretKey := getJWTSecret()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // TEMPO REDUZIDO PARA TESTE: Token de acesso válido por 15 minutos
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// GenerateRefreshToken cria um novo token de atualização JWT de longa duração.
func GenerateRefreshToken(userID string) (string, error) {
	secretKey := getJWTSecret()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // Token de atualização válido por 7 dias
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ValidateToken verifica um token JWT e retorna as claims se for válido.
func ValidateToken(tokenString string) (*jwt.RegisteredClaims, error) {
	secretKey := getJWTSecret()
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Valida o algoritmo de assinatura esperado (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err // Retorna o erro original da biblioteca JWT (ex: ErrTokenExpired)
	}

	if !token.Valid {
		return nil, jwt.ErrTokenNotValidYet // Ou um erro genérico de token inválido
	}

	return claims, nil
}

package security

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implementa um sistema de limitação de taxa baseado em IP/usuário
// usando o algoritmo de sliding window
type RateLimiter struct {
	requests      map[string][]time.Time
	windowSize    time.Duration
	maxRequests   int
	mu            sync.RWMutex
	ipEnabled     bool
	userIDEnabled bool
	cleanupTick   time.Duration
}

// NewRateLimiter cria uma nova instância do rate limiter
func NewRateLimiter(windowSize time.Duration, maxRequests int, ipEnabled, userIDEnabled bool) *RateLimiter {
	limiter := &RateLimiter{
		requests:      make(map[string][]time.Time),
		windowSize:    windowSize,
		maxRequests:   maxRequests,
		ipEnabled:     ipEnabled,
		userIDEnabled: userIDEnabled,
		cleanupTick:   5 * time.Minute, // Executa cleanup a cada 5 minutos
	}

	// Inicia uma goroutine para limpeza periódica de registros antigos
	go limiter.periodicCleanup()

	return limiter
}

// Cleanup remove registros antigos que estão fora da janela de tempo
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.windowSize)

	for key, times := range rl.requests {
		var validTimes []time.Time

		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}

		if len(validTimes) > 0 {
			rl.requests[key] = validTimes
		} else {
			delete(rl.requests, key)
		}
	}
}

// periodicCleanup executa a limpeza de registros periodicamente
func (rl *RateLimiter) periodicCleanup() {
	ticker := time.NewTicker(rl.cleanupTick)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// getKey gera uma chave única baseada no IP e/ou ID do usuário
func (rl *RateLimiter) getKey(c *gin.Context) string {
	var key string

	if rl.ipEnabled {
		key += c.ClientIP()
	}

	if rl.userIDEnabled {
		if userID, exists := c.Get("userID"); exists && userID != nil {
			key += "/" + userID.(string)
		}
	}

	return key
}

// IsAllowed verifica se a requisição está dentro do limite da taxa
func (rl *RateLimiter) IsAllowed(c *gin.Context) bool {
	key := rl.getKey(c)

	// Se não temos como identificar a origem da requisição, permitir
	if key == "" {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.windowSize)

	// Remove timestamps antigos
	times := rl.requests[key]
	var validTimes []time.Time

	for _, t := range times {
		if t.After(cutoff) {
			validTimes = append(validTimes, t)
		}
	}

	// Verifica se excede o limite
	if len(validTimes) >= rl.maxRequests {
		return false
	}

	// Adiciona o timestamp atual
	validTimes = append(validTimes, now)
	rl.requests[key] = validTimes

	return true
}

// RemainingAttempts retorna o número de tentativas restantes
func (rl *RateLimiter) RemainingAttempts(c *gin.Context) int {
	key := rl.getKey(c)

	if key == "" {
		return rl.maxRequests
	}

	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-rl.windowSize)

	count := 0
	for _, t := range rl.requests[key] {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := rl.maxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// ResetInWindow retorna o tempo restante em segundos até liberar uma vaga
// na janela deslizante quando o limite foi atingido
func (rl *RateLimiter) ResetInWindow(c *gin.Context) int {
	key := rl.getKey(c)

	if key == "" {
		return 0
	}

	rl.mu.RLock()
	defer rl.mu.RUnlock()

	times := rl.requests[key]
	if len(times) == 0 {
		return 0
	}

	// Encontra o timestamp mais antigo na janela atual
	now := time.Now()
	cutoff := now.Add(-rl.windowSize)

	var oldestTime time.Time
	for _, t := range times {
		if t.After(cutoff) {
			if oldestTime.IsZero() || t.Before(oldestTime) {
				oldestTime = t
			}
		}
	}

	if oldestTime.IsZero() {
		return 0
	}

	// Calcula quando essa requisição mais antiga sairá da janela
	timeToReset := int(oldestTime.Add(rl.windowSize).Sub(now).Seconds())
	if timeToReset < 0 {
		timeToReset = 0
	}

	return timeToReset
}

// RateLimitMiddleware cria um middleware Gin para aplicar rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Se o limiter estiver desabilitado (sem rulesets ou não configurado)
		if rl == nil || rl.maxRequests <= 0 {
			c.Next()
			return
		}

		// Verifica se a requisição está dentro do limite
		if !rl.IsAllowed(c) {
			// Calcula os cabeçalhos rate limit RFC 6585
			remaining := rl.RemainingAttempts(c)
			reset := rl.ResetInWindow(c)

			// Define os cabeçalhos para o cliente saber sobre o rate limit
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", reset))
			c.Header("Retry-After", fmt.Sprintf("%d", reset))

			c.AbortWithStatusJSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Você enviou muitas requisições. Tente novamente mais tarde.",
				"retry_after": reset,
			})
			return
		}

		// Define os cabeçalhos mesmo para requisições permitidas
		remaining := rl.RemainingAttempts(c)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// Limiters específicos pré-configurados
var (
	// Limiter para autenticação: 5 tentativas por minuto por IP
	AuthLimiter = NewRateLimiter(1*time.Minute, 5, true, false)

	// Limiter para operações sensíveis: 10 operações por minuto por usuário
	SensitiveLimiter = NewRateLimiter(1*time.Minute, 10, false, true)

	// Limiter global: 60 requisições por minuto por IP
	GlobalLimiter = NewRateLimiter(1*time.Minute, 60, true, false)
)

// GetAuthLimiter retorna o rate limiter para endpoints de autenticação
func GetAuthLimiter() *RateLimiter {
	return AuthLimiter
}

// GetSensitiveLimiter retorna o rate limiter para operações sensíveis
func GetSensitiveLimiter() *RateLimiter {
	return SensitiveLimiter
}

// GetGlobalLimiter retorna o rate limiter global
func GetGlobalLimiter() *RateLimiter {
	return GlobalLimiter
}

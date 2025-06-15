package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"context"
	"gestor-e-docs/backend/services/identity-service/db"
	"gestor-e-docs/backend/services/identity-service/handlers"
	"gestor-e-docs/backend/services/identity-service/metrics"
	"gestor-e-docs/backend/services/identity-service/models"
	"gestor-e-docs/backend/services/identity-service/security"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// ensureAdminUserExists verifica se o usuário administrador padrão existe e, se não, o cria.
func ensureAdminUserExists() {
	collection := db.GetCollection("users")
	ctx := context.Background()

	adminEmail := "admin@example.com"
	filter := bson.M{"email": adminEmail}

	var existingUser models.User
	err := collection.FindOne(ctx, filter).Decode(&existingUser)

	if err == mongo.ErrNoDocuments {
		log.Println("[DB_INIT] Usuário admin não encontrado, criando...")

		adminPassword := "password123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("[DB_INIT] Falha ao gerar hash da senha do admin: %v", err)
		}

		adminUser := models.User{
			Name:      "Admin",
			Email:     adminEmail,
			Password:  string(hashedPassword),
			Role:      "admin",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = collection.InsertOne(ctx, adminUser)
		if err != nil {
			log.Fatalf("[DB_INIT] Falha ao inserir usuário admin: %v", err)
		}
		log.Println("[DB_INIT] Usuário admin criado com sucesso!")

	} else if err != nil {
		log.Fatalf("[DB_INIT] Erro ao verificar a existência do usuário admin: %v", err)
	} else {
		log.Println("[DB_INIT] Usuário admin já existe.")
	}
}

func main() {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8085" // Porta padrão se não definida
	}

	// Inicializar conexão com o banco de dados
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.DisconnectDB() // Garante que a conexão será fechada ao sair

	// Garante que o usuário administrador exista
	ensureAdminUserExists()

	// Inicializar métricas do Prometheus
	metrics.Init()
	
	// Inicializar o serviço de autenticação de dois fatores
	handlers.InitTwoFactorAuth()
	
	// Inicializar o log de auditoria
	auditLogger := security.NewAuditLogger(db.GetCollection("audit_logs"), true, true)
	
	// Inicializar os rate limiters
	// Usar limiters pré-configurados do pacote security
	// Limitador global já está configurado para 60 requisições por minuto por IP
	globalLimiter := security.GetGlobalLimiter()
	
	// Limitador para autenticação configurado para 5 tentativas por minuto por IP
	authLimiter := security.GetAuthLimiter()
	
	// Limitador para operações sensíveis configurado para 10 operações por minuto por usuário
	sensitiveLimiter := security.GetSensitiveLimiter()

	r := gin.Default()

	// Adicionar middleware de métricas do Prometheus
	r.Use(metrics.PrometheusMiddleware())

	// Middleware para tratar requisições HEAD
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "HEAD" {
			// Permitir que o Gin continue o processamento da rota
			// Se a rota existir, será processada normalmente
			// Se não existir, o Gin retornará 404 automaticamente
			c.Request.Method = "GET"
		}
		c.Next()
	})

	// Configuração do CORS baseada em variáveis de ambiente
	config := cors.DefaultConfig()
	
	// Obter origens permitidas do ambiente ou usar padrões de desenvolvimento
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		// Em desenvolvimento, permitir localhost em várias portas
		config.AllowOrigins = []string{
			"https://localhost",
			"http://localhost",
			"http://127.0.0.1",
			"http://localhost:3085",
		}
	} else {
		// Em produção, usar as origens configuradas
		config.AllowOrigins = strings.Split(allowedOrigins, ",")
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"Accept",
		"Cookie",
		"X-CSRF-Token",
		"X-Requested-With",
	}
	config.AllowCredentials = true // Essencial para cookies
	config.ExposeHeaders = []string{
		"Content-Length",
		"Content-Type",
		"Set-Cookie",
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
	}
	config.MaxAge = 12 * time.Hour

	r.Use(cors.New(config))
	
	// Adiciona o middleware de rate limiting global em todas as rotas
	r.Use(globalLimiter.RateLimitMiddleware())

	// Rota de health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Endpoint de métricas do Prometheus
	r.GET("/metrics", metrics.PrometheusHandler())

	// Rota de API V1 - placeholder
	apiV1 := r.Group("/api/v1/identity")
	
	// Adicionar middleware de auditoria a todas as rotas da API
	apiV1.Use(security.AuditMiddleware(auditLogger))
	{
		public := apiV1.Group("")
		{
			public.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Identity Service v1"})
			})
			// Rotas POST principais com rate limiting específico
			public.POST("/register", authLimiter.RateLimitMiddleware(), handlers.RegisterUser)
			public.POST("/login", authLimiter.RateLimitMiddleware(), handlers.LoginUser)
			public.POST("/logout", handlers.LogoutUser)
			public.POST("/refresh", handlers.RefreshToken) // Rota de refresh é pública

			// Handlers HEAD para CORS/preflight
			public.HEAD("/register", handlers.HeadHandler)
			public.HEAD("/login", handlers.HeadHandler)
			public.HEAD("/logout", handlers.HeadHandler)
			public.HEAD("/refresh", handlers.HeadHandler)
		}

		// Rota específica para verificação 2FA no login
		public.POST("/2fa/verify", handlers.Login2FA)
			
		protected := apiV1.Group("")
		protected.Use(handlers.AuthMiddleware()) // Aplicar middleware de autenticação a este grupo
		protected.Use(handlers.TwoFactorMiddleware()) // Aplicar middleware de 2FA após autenticação
		{
			protected.GET("/me", handlers.GetUserProfile)
		
			// Rotas para o gerenciamento de 2FA com limitação de taxa para operações sensíveis
			twoFactorGroup := protected.Group("/2fa")
			{
				// Aplicar sensitiveLimiter para operações sensíveis
				twoFactorGroup.GET("/setup", sensitiveLimiter.RateLimitMiddleware(), handlers.GenerateSetup2FA)
				twoFactorGroup.POST("/setup", sensitiveLimiter.RateLimitMiddleware(), handlers.Verify2FA)
				twoFactorGroup.DELETE("/disable", sensitiveLimiter.RateLimitMiddleware(), handlers.Disable2FA)
				twoFactorGroup.GET("/status", handlers.GetTwoFactorStatus)
			}
		}
	}

	log.Printf("Identity service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

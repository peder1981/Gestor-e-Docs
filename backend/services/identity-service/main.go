package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"gestor-e-docs/identity-service/db"
	"gestor-e-docs/identity-service/handlers"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors" // Adicionar import do CORS
)

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

	r := gin.Default()

	// Configuração do CORS
	config := cors.DefaultConfig()
	// Permitir origens mais flexíveis para desenvolvimento
config.AllowOrigins = []string{
	"http://localhost:3085", 
	"http://localhost",
	"http://localhost:80",
	"http://127.0.0.1:3085",
	"http://127.0.0.1",
	"http://127.0.0.1:33023", // Adicionar porta do proxy temporário para testes
}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept"}
	config.AllowCredentials = true // Permitir envio de cookies
	config.ExposeHeaders = []string{"Content-Length", "Set-Cookie"} // Expor cabeçalhos Set-Cookie
	config.MaxAge = 12 * time.Hour // Aumentar tempo de cache para preflight requests
	r.Use(cors.New(config)) // Aplicar o middleware CORS

	// Rota de health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Rota de API V1 - placeholder
	apiV1 := r.Group("/api/v1/identity")
	{
		public := apiV1.Group("")
		{
			public.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Identity Service v1"})
			})
			public.POST("/register", handlers.RegisterUser)
			public.POST("/login", handlers.LoginUser)
			public.POST("/logout", handlers.LogoutUser)
			public.POST("/refresh", handlers.RefreshToken) // Rota de refresh é pública
		}

		protected := apiV1.Group("")
		protected.Use(handlers.AuthMiddleware()) // Aplicar middleware de autenticação a este grupo
		{
			protected.GET("/me", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "User is authenticated", "userID": userID})
			})
		}
	}

	log.Printf("Identity service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"gestor-e-docs/backend/services/identity-service/db"
	"gestor-e-docs/backend/services/identity-service/handlers"
	"context"

	"gestor-e-docs/backend/services/identity-service/models"

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

	r := gin.Default()

	// Configuração do CORS
	config := cors.DefaultConfig()
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
	config.AllowCredentials = true
	config.ExposeHeaders = []string{"Content-Length", "Set-Cookie"}
	config.MaxAge = 12 * time.Hour
	r.Use(cors.New(config))

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
			protected.GET("/me", handlers.GetUserProfile)
		}
	}

	log.Printf("Identity service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

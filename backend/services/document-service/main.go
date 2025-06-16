package main

import (
	"context"
	"gestor-e-docs/document-service/db"
	"gestor-e-docs/document-service/handlers"
	"gestor-e-docs/document-service/metrics"
	"gestor-e-docs/document-service/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Inicializar métricas do Prometheus
	metrics.Init()

	// Inicializar conexão com MongoDB
	err := db.ConnectDatabase()
	if err != nil {
		log.Fatalf("Falha ao conectar com MongoDB: %v", err)
	}
	defer db.DisconnectDatabase()

	// Inicializar cliente MinIO
	_, err = storage.GetMinioClient()
	if err != nil {
		log.Fatalf("Falha ao inicializar cliente MinIO: %v", err)
	}

	// Configurar o router
	r := gin.Default()

	// Adicionar middleware de métricas do Prometheus
	r.Use(metrics.PrometheusMiddleware())

	// Configuração CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"https://localhost",
		"http://localhost:3085",
		"http://localhost",
		"http://localhost:80",
		"http://127.0.0.1:3085",
		"http://127.0.0.1",
		"https://127.0.0.1",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept"}
	config.AllowCredentials = true  // Permitir envio de cookies
	config.ExposeHeaders = []string{"Content-Length", "Set-Cookie"}
	config.MaxAge = 12 * time.Hour
	r.Use(cors.New(config))

	// Rotas de saúde/diagnóstico
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Endpoint de métricas do Prometheus
	r.GET("/metrics", metrics.PrometheusHandler())

	// Grupo de rotas da API
	api := r.Group("/api/v1/documents")
	
	// Rotas públicas
	api.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Document Service API"})
	})

	// Rotas protegidas
	protected := api.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		// Operações CRUD de documentos
		protected.POST("/", handlers.CreateDocument)
		protected.GET("/:id", handlers.GetDocument)
		protected.PUT("/:id", handlers.UpdateDocument)
		protected.DELETE("/:id", handlers.DeleteDocument)
		protected.GET("/list", handlers.ListDocuments)
		protected.GET("/:id/download", handlers.DownloadDocument)
		protected.GET("/:id/download/file", handlers.DownloadDocumentFile)
	}

	// Determinar a porta do servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8185" // Porta padrão para o document-service
		log.Printf("Porta não especificada. Usando a porta padrão: %s", port)
	}

	// Configurar servidor HTTP com graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Iniciar o servidor em uma goroutine
	go func() {
		log.Printf("Servidor iniciado na porta %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	// Configurar canal para capturar sinais de encerramento
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Bloquear até receber um sinal
	<-quit
	log.Println("Desligando o servidor...")

	// Contexto com timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Tentar shutdown graceful
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erro durante o shutdown do servidor: %v", err)
	}

	log.Println("Servidor encerrado com sucesso")
}

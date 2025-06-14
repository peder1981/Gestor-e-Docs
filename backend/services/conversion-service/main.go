package main

import (
	"context"
	"gestor-e-docs/conversion-service/handlers"
	"gestor-e-docs/conversion-service/metrics"
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

	// Configurar o router
	r := gin.Default()

	// Adicionar middleware de métricas do Prometheus
	r.Use(metrics.PrometheusMiddleware())

	// Configuração CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3085",
		"http://localhost",
		"http://localhost:80",
		"http://127.0.0.1:3085",
		"http://127.0.0.1",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept"}
	config.AllowCredentials = true
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
	api := r.Group("/api/v1/convert")
	
	// Rotas públicas
	api.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Conversion Service API"})
	})

	// Rotas públicas
	api.GET("/formats", handlers.ListSupportedFormats)

	// Rotas protegidas
	protected := api.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.POST("/markdown-to-pdf", handlers.ConvertMarkdownToPDF)
		protected.POST("/markdown-to-html", handlers.ConvertMarkdownToHTML)
	}

	// Determinar a porta do servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8285" // Porta padrão para o conversion-service
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

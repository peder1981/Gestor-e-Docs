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
	// Configurar logs
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Iniciando Conversion Service...")

	// Inicializar métricas do Prometheus
	metrics.Init()

	// Inicializar queue de conversão com 3 workers
	handlers.InitializeQueue(3)
	log.Println("Queue de conversão inicializada com 3 workers")

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

	// Rotas protegidas (síncronas)
	protected := api.Group("/")
	protected.Use(handlers.AuthMiddleware())
	protected.Use(handlers.ConversionValidationMiddleware())
	{
		protected.POST("/markdown-to-pdf", handlers.ConvertMarkdownToPDF)
		protected.POST("/markdown-to-html", handlers.ConvertMarkdownToHTML)
		protected.POST("/markdown-to-docx", handlers.ConvertMarkdownToDOCX)
		protected.POST("/markdown-to-latex", handlers.ConvertMarkdownToLaTeX)
	}

	// Rotas protegidas (assíncronas)
	async := api.Group("/async")
	async.Use(handlers.AuthMiddleware())
	async.Use(handlers.ConversionValidationMiddleware())
	{
		async.POST("/markdown-to-pdf", handlers.AsyncConvertMarkdownToPDF)
		async.POST("/markdown-to-html", handlers.AsyncConvertMarkdownToHTML)
		async.POST("/markdown-to-docx", handlers.AsyncConvertMarkdownToDOCX)
		async.POST("/markdown-to-latex", handlers.AsyncConvertMarkdownToLaTeX)
	}

	// Rotas para gerenciamento de jobs (protegidas, sem validação de conteúdo)
	jobs := api.Group("/jobs")
	jobs.Use(handlers.AuthMiddleware())
	{
		jobs.GET("/:jobId/status", handlers.GetJobStatus)
		jobs.GET("/:jobId/download", handlers.DownloadJobResult)
		jobs.GET("/stats", handlers.GetQueueStats)
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

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Servidor finalizando...")

	// Finalizar queue
	handlers.GetQueue().Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Erro ao finalizar servidor:", err)
	}
	log.Println("Servidor finalizado")
}

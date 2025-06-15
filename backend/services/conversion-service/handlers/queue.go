package handlers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// ConversionJob representa um trabalho de conversão na queue
type ConversionJob struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"` // pdf, html, docx, latex
	Content      string     `json:"content"`
	Title        string     `json:"title"`
	Status       string     `json:"status"` // pending, processing, completed, failed
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ResultData   []byte     `json:"result_data,omitempty"`
	ResultType   string     `json:"result_type,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
}

// ConversionQueue gerencia a queue de conversões
type ConversionQueue struct {
	jobs    map[string]*ConversionJob
	pending chan *ConversionJob
	workers int
	mutex   sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewConversionQueue cria uma nova queue de conversão
func NewConversionQueue(workers int) *ConversionQueue {
	ctx, cancel := context.WithCancel(context.Background())

	queue := &ConversionQueue{
		jobs:    make(map[string]*ConversionJob),
		pending: make(chan *ConversionJob, 100), // Buffer de 100 jobs
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Iniciar workers
	for i := 0; i < workers; i++ {
		go queue.worker(i)
	}

	return queue
}

// AddJob adiciona um novo job à queue
func (q *ConversionQueue) AddJob(jobType, content, title, userID string) string {
	job := &ConversionJob{
		ID:        generateJobID(),
		Type:      jobType,
		Content:   content,
		Title:     title,
		Status:    "pending",
		CreatedAt: time.Now(),
		UserID:    userID,
	}

	q.mutex.Lock()
	q.jobs[job.ID] = job
	q.mutex.Unlock()

	// Enviar para channel com timeout
	select {
	case q.pending <- job:
		log.Printf("Job %s adicionado à queue (tipo: %s)", job.ID, job.Type)
	case <-time.After(5 * time.Second):
		// Queue cheia, marcar como falha
		q.mutex.Lock()
		job.Status = "failed"
		job.ErrorMessage = "Queue de processamento cheia, tente novamente"
		now := time.Now()
		job.CompletedAt = &now
		q.mutex.Unlock()
		log.Printf("Falha ao adicionar job %s à queue: queue cheia", job.ID)
	}

	return job.ID
}

// GetJob recupera informações de um job
func (q *ConversionQueue) GetJob(jobID string) (*ConversionJob, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	job, exists := q.jobs[jobID]
	if !exists {
		return nil, false
	}

	// Criar cópia sem dados sensíveis para resposta
	result := *job
	if result.Status == "completed" && len(result.ResultData) > 0 {
		// Não incluir dados binários na resposta JSON
		result.ResultData = nil
	}

	return &result, true
}

// GetJobResult recupera o resultado de um job completo
func (q *ConversionQueue) GetJobResult(jobID string) ([]byte, string, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	job, exists := q.jobs[jobID]
	if !exists {
		return nil, "", fmt.Errorf("job não encontrado")
	}

	if job.Status != "completed" {
		return nil, "", fmt.Errorf("job ainda não foi concluído")
	}

	if job.ErrorMessage != "" {
		return nil, "", fmt.Errorf("job falhou: %s", job.ErrorMessage)
	}

	return job.ResultData, job.ResultType, nil
}

// worker processa jobs da queue
func (q *ConversionQueue) worker(id int) {
	log.Printf("Worker %d iniciado", id)

	for {
		select {
		case <-q.ctx.Done():
			log.Printf("Worker %d finalizando", id)
			return
		case job := <-q.pending:
			q.processJob(job, id)
		}
	}
}

// processJob processa um job individual
func (q *ConversionQueue) processJob(job *ConversionJob, workerID int) {
	log.Printf("Worker %d processando job %s (tipo: %s)", workerID, job.ID, job.Type)

	// Atualizar status para processando
	q.mutex.Lock()
	job.Status = "processing"
	q.mutex.Unlock()

	var result []byte
	var resultType string
	var err error

	// Processar baseado no tipo
	switch job.Type {
	case "pdf":
		result, err = gotenberg.ConvertMarkdownToPDF(job.Content, job.Title)
		resultType = "application/pdf"
	case "html":
		result, err = gotenberg.ConvertMarkdownToHTML(job.Content)
		resultType = "text/html"
	case "docx":
		result, err = gotenberg.ConvertMarkdownToDOCX(job.Content, job.Title)
		resultType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "latex":
		result, err = gotenberg.ConvertMarkdownToLaTeX(job.Content, job.Title)
		resultType = "application/x-tex"
	default:
		err = fmt.Errorf("tipo de conversão não suportado: %s", job.Type)
	}

	// Atualizar resultado
	q.mutex.Lock()
	now := time.Now()
	job.CompletedAt = &now

	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		log.Printf("Worker %d: job %s falhou: %v", workerID, job.ID, err)
	} else {
		job.Status = "completed"
		job.ResultData = result
		job.ResultType = resultType
		log.Printf("Worker %d: job %s concluído com sucesso", workerID, job.ID)
	}
	q.mutex.Unlock()
}

// CleanupOldJobs remove jobs antigos da memória
func (q *ConversionQueue) CleanupOldJobs() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // Remover jobs com mais de 24h

	for id, job := range q.jobs {
		if job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
			delete(q.jobs, id)
			log.Printf("Job antigo %s removido da memória", id)
		}
	}
}

// Shutdown finaliza a queue gracefully
func (q *ConversionQueue) Shutdown() {
	log.Println("Finalizando queue de conversão...")
	q.cancel()
	close(q.pending)
}

// QueueStats retorna estatísticas da queue
func (q *ConversionQueue) QueueStats() map[string]interface{} {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_jobs": len(q.jobs),
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
		"workers":    q.workers,
	}

	for _, job := range q.jobs {
		switch job.Status {
		case "pending":
			stats["pending"] = stats["pending"].(int) + 1
		case "processing":
			stats["processing"] = stats["processing"].(int) + 1
		case "completed":
			stats["completed"] = stats["completed"].(int) + 1
		case "failed":
			stats["failed"] = stats["failed"].(int) + 1
		}
	}

	return stats
}

// Instância global da queue
var conversionQueue *ConversionQueue

// InitializeQueue inicializa a queue global
func InitializeQueue(workers int) {
	conversionQueue = NewConversionQueue(workers)

	// Iniciar limpeza periódica
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			conversionQueue.CleanupOldJobs()
		}
	}()
}

// GetQueue retorna a queue global
func GetQueue() *ConversionQueue {
	return conversionQueue
}

// generateJobID gera um ID único para o job
func generateJobID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 16

	// O global rand já é automaticamente seedado em Go 1.20+, não é necessário chamar Seed
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

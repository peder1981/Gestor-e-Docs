package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AsyncConversionResponse resposta para conversões assíncronas
type AsyncConversionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	JobID   string `json:"job_id"`
}

// JobStatusResponse resposta para status de job
type JobStatusResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Job     *ConversionJob `json:"job,omitempty"`
}

// QueueStatsResponse resposta para estatísticas da queue
type QueueStatsResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Stats   map[string]interface{} `json:"stats,omitempty"`
}

// AsyncConvertMarkdownToPDF inicia conversão assíncrona para PDF
func AsyncConvertMarkdownToPDF(c *gin.Context) {
	req := GetValidatedRequest(c)
	if req == nil {
		c.JSON(http.StatusInternalServerError, AsyncConversionResponse{
			Success: false,
			Message: "Erro interno: requisição não validada",
		})
		return
	}

	// Obter ID do usuário do contexto de autenticação
	userID := getUserIDFromContext(c)
	
	// Adicionar job à queue
	jobID := GetQueue().AddJob("pdf", req.Content, req.Title, userID)
	
	c.JSON(http.StatusAccepted, AsyncConversionResponse{
		Success: true,
		Message: "Conversão para PDF iniciada",
		JobID:   jobID,
	})
}

// AsyncConvertMarkdownToHTML inicia conversão assíncrona para HTML
func AsyncConvertMarkdownToHTML(c *gin.Context) {
	req := GetValidatedRequest(c)
	if req == nil {
		c.JSON(http.StatusInternalServerError, AsyncConversionResponse{
			Success: false,
			Message: "Erro interno: requisição não validada",
		})
		return
	}

	userID := getUserIDFromContext(c)
	jobID := GetQueue().AddJob("html", req.Content, req.Title, userID)
	
	c.JSON(http.StatusAccepted, AsyncConversionResponse{
		Success: true,
		Message: "Conversão para HTML iniciada",
		JobID:   jobID,
	})
}

// AsyncConvertMarkdownToDOCX inicia conversão assíncrona para DOCX
func AsyncConvertMarkdownToDOCX(c *gin.Context) {
	req := GetValidatedRequest(c)
	if req == nil {
		c.JSON(http.StatusInternalServerError, AsyncConversionResponse{
			Success: false,
			Message: "Erro interno: requisição não validada",
		})
		return
	}

	userID := getUserIDFromContext(c)
	jobID := GetQueue().AddJob("docx", req.Content, req.Title, userID)
	
	c.JSON(http.StatusAccepted, AsyncConversionResponse{
		Success: true,
		Message: "Conversão para DOCX iniciada",
		JobID:   jobID,
	})
}

// AsyncConvertMarkdownToLaTeX inicia conversão assíncrona para LaTeX
func AsyncConvertMarkdownToLaTeX(c *gin.Context) {
	req := GetValidatedRequest(c)
	if req == nil {
		c.JSON(http.StatusInternalServerError, AsyncConversionResponse{
			Success: false,
			Message: "Erro interno: requisição não validada",
		})
		return
	}

	userID := getUserIDFromContext(c)
	jobID := GetQueue().AddJob("latex", req.Content, req.Title, userID)
	
	c.JSON(http.StatusAccepted, AsyncConversionResponse{
		Success: true,
		Message: "Conversão para LaTeX iniciada",
		JobID:   jobID,
	})
}

// GetJobStatus retorna o status de um job
func GetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, JobStatusResponse{
			Success: false,
			Message: "ID do job é obrigatório",
		})
		return
	}

	job, exists := GetQueue().GetJob(jobID)
	if !exists {
		c.JSON(http.StatusNotFound, JobStatusResponse{
			Success: false,
			Message: "Job não encontrado",
		})
		return
	}

	c.JSON(http.StatusOK, JobStatusResponse{
		Success: true,
		Message: "Status do job recuperado com sucesso",
		Job:     job,
	})
}

// DownloadJobResult faz download do resultado de um job completo
func DownloadJobResult(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, ConversionResponse{
			Success: false,
			Message: "ID do job é obrigatório",
		})
		return
	}

	// Verificar se o job existe e está completo
	job, exists := GetQueue().GetJob(jobID)
	if !exists {
		c.JSON(http.StatusNotFound, ConversionResponse{
			Success: false,
			Message: "Job não encontrado",
		})
		return
	}

	if job.Status != "completed" {
		c.JSON(http.StatusBadRequest, ConversionResponse{
			Success: false,
			Message: "Job ainda não foi concluído. Status: " + job.Status,
		})
		return
	}

	// Obter resultado
	resultData, resultType, err := GetQueue().GetJobResult(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro ao recuperar resultado: " + err.Error(),
		})
		return
	}

	// Configurar headers para download
	filename := getFilenameForType(job.Type, job.Title)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", resultType)

	// Retornar dados baseado no tipo
	if job.Type == "html" || job.Type == "latex" {
		c.String(http.StatusOK, string(resultData))
	} else {
		c.Data(http.StatusOK, resultType, resultData)
	}
}

// GetQueueStats retorna estatísticas da queue
func GetQueueStats(c *gin.Context) {
	stats := GetQueue().QueueStats()
	
	c.JSON(http.StatusOK, QueueStatsResponse{
		Success: true,
		Message: "Estatísticas da queue recuperadas com sucesso",
		Stats:   stats,
	})
}

// getUserIDFromContext extrai o ID do usuário do contexto de autenticação
func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return "anonymous"
}

// getFilenameForType retorna o nome de arquivo apropriado para o tipo
func getFilenameForType(jobType, title string) string {
	var extension string
	switch jobType {
	case "pdf":
		extension = ".pdf"
	case "html":
		extension = ".html"
	case "docx":
		extension = ".docx"
	case "latex":
		extension = ".tex"
	default:
		extension = ".txt"
	}

	if title != "" {
		return title + extension
	}
	return "document" + extension
}

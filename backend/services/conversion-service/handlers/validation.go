package handlers

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResponse representa uma resposta de erro de validação
type ValidationResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// ValidateConversionRequest valida uma requisição de conversão
func ValidateConversionRequest(req *ConversionRequest) []ValidationError {
	var errors []ValidationError

	// Validar conteúdo obrigatório
	if strings.TrimSpace(req.Content) == "" {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: "Conteúdo é obrigatório e não pode estar vazio",
		})
	}

	// Validar tamanho do conteúdo (máximo 10MB)
	const maxContentSize = 10 * 1024 * 1024 // 10MB
	if len(req.Content) > maxContentSize {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: "Conteúdo excede o tamanho máximo permitido (10MB)",
		})
	}

	// Validar se o conteúdo é UTF-8 válido
	if !utf8.ValidString(req.Content) {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: "Conteúdo deve estar em formato UTF-8 válido",
		})
	}

	// Validar título (se fornecido)
	if req.Title != "" {
		// Remover espaços em branco
		req.Title = strings.TrimSpace(req.Title)
		
		// Validar tamanho do título
		if len(req.Title) > 255 {
			errors = append(errors, ValidationError{
				Field:   "title",
				Message: "Título não pode exceder 255 caracteres",
			})
		}

		// Validar caracteres inválidos para nome de arquivo
		invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
		for _, char := range invalidChars {
			if strings.Contains(req.Title, char) {
				errors = append(errors, ValidationError{
					Field:   "title",
					Message: "Título contém caracteres inválidos para nome de arquivo",
				})
				break
			}
		}
	}

	return errors
}

// ValidateMarkdownContent valida se o conteúdo parece ser Markdown válido
func ValidateMarkdownContent(content string) []ValidationError {
	var errors []ValidationError

	// Verificar se há indicadores básicos de Markdown
	hasMarkdownIndicators := false
	
	// Verificar cabeçalhos
	if strings.Contains(content, "#") {
		hasMarkdownIndicators = true
	}
	
	// Verificar listas
	if strings.Contains(content, "- ") || strings.Contains(content, "* ") || strings.Contains(content, "+ ") {
		hasMarkdownIndicators = true
	}
	
	// Verificar texto em negrito/itálico
	if strings.Contains(content, "**") || strings.Contains(content, "*") {
		hasMarkdownIndicators = true
	}
	
	// Verificar código
	if strings.Contains(content, "`") || strings.Contains(content, "```") {
		hasMarkdownIndicators = true
	}
	
	// Verificar links
	if strings.Contains(content, "[") && strings.Contains(content, "]") {
		hasMarkdownIndicators = true
	}

	// Se o conteúdo é muito grande mas não tem indicadores de Markdown, avisar
	if len(content) > 1000 && !hasMarkdownIndicators {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: "Conteúdo não parece conter formatação Markdown válida",
		})
	}

	return errors
}

// ConversionValidationMiddleware é um middleware para validar requisições de conversão
func ConversionValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ConversionRequest
		
		// Tentar fazer bind da requisição
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Success: false,
				Message: "Erro ao processar requisição JSON",
				Errors: []ValidationError{
					{
						Field:   "json",
						Message: err.Error(),
					},
				},
			})
			c.Abort()
			return
		}

		// Validar requisição
		validationErrors := ValidateConversionRequest(&req)
		
		// Validar conteúdo Markdown
		markdownErrors := ValidateMarkdownContent(req.Content)
		validationErrors = append(validationErrors, markdownErrors...)

		// Se há erros de validação, retornar erro
		if len(validationErrors) > 0 {
			c.JSON(http.StatusBadRequest, ValidationResponse{
				Success: false,
				Message: "Erro de validação nos dados de entrada",
				Errors:  validationErrors,
			})
			c.Abort()
			return
		}

		// Armazenar a requisição validada no contexto
		c.Set("validated_request", req)
		c.Next()
	}
}

// GetValidatedRequest recupera a requisição validada do contexto
func GetValidatedRequest(c *gin.Context) *ConversionRequest {
	if req, exists := c.Get("validated_request"); exists {
		if validatedReq, ok := req.(ConversionRequest); ok {
			return &validatedReq
		}
	}
	return nil
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	gotenberg *GotenbergClient
)

func init() {
	gotenberg = NewGotenbergClient()
}

// ConvertMarkdownToPDF converte um documento Markdown para PDF
func ConvertMarkdownToPDF(c *gin.Context) {
	// Recuperar requisição já validada do middleware
	req := GetValidatedRequest(c)
	if req == nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro interno: requisição não validada",
		})
		return
	}

	// Converter para PDF usando Gotenberg
	pdfData, err := gotenberg.ConvertMarkdownToPDF(req.Content, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro na conversão para PDF: " + err.Error(),
		})
		return
	}

	// Configurar headers para download do PDF
	filename := "document.pdf"
	if req.Title != "" {
		filename = req.Title + ".pdf"
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// ConvertMarkdownToHTML converte um documento Markdown para HTML
func ConvertMarkdownToHTML(c *gin.Context) {
	// Recuperar requisição já validada do middleware
	req := GetValidatedRequest(c)
	if req == nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro interno: requisição não validada",
		})
		return
	}

	// Converter para HTML usando Gotenberg
	htmlData, err := gotenberg.ConvertMarkdownToHTML(req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro na conversão para HTML: " + err.Error(),
		})
		return
	}

	// Retornar o HTML
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(htmlData))
}

// ConvertMarkdownToDOCX converte um documento Markdown para DOCX
func ConvertMarkdownToDOCX(c *gin.Context) {
	// Recuperar requisição já validada do middleware
	req := GetValidatedRequest(c)
	if req == nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro interno: requisição não validada",
		})
		return
	}

	// Converter para DOCX usando Gotenberg
	docxData, err := gotenberg.ConvertMarkdownToDOCX(req.Content, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro na conversão para DOCX: " + err.Error(),
		})
		return
	}

	// Configurar headers para download do DOCX
	filename := "document.docx"
	if req.Title != "" {
		filename = req.Title + ".docx"
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxData)
}

// ConvertMarkdownToLaTeX converte um documento Markdown para LaTeX
func ConvertMarkdownToLaTeX(c *gin.Context) {
	// Recuperar requisição já validada do middleware
	req := GetValidatedRequest(c)
	if req == nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro interno: requisição não validada",
		})
		return
	}

	// Converter para LaTeX usando conversor interno
	latexData, err := gotenberg.ConvertMarkdownToLaTeX(req.Content, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConversionResponse{
			Success: false,
			Message: "Erro na conversão para LaTeX: " + err.Error(),
		})
		return
	}

	// Retornar o LaTeX
	filename := "document.tex"
	if req.Title != "" {
		filename = req.Title + ".tex"
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/x-tex")
	c.String(http.StatusOK, string(latexData))
}

// ListSupportedFormats retorna os formatos suportados para conversão
func ListSupportedFormats(c *gin.Context) {
	formats := []SupportedFormat{
		{
			ID:          "markdown-to-pdf",
			Name:        "Markdown para PDF",
			Description: "Converte documentos Markdown para formato PDF",
			InputType:   "text/markdown",
			OutputType:  "application/pdf",
		},
		{
			ID:          "markdown-to-html",
			Name:        "Markdown para HTML",
			Description: "Converte documentos Markdown para HTML",
			InputType:   "text/markdown",
			OutputType:  "text/html",
		},
		{
			ID:          "markdown-to-docx",
			Name:        "Markdown para DOCX",
			Description: "Converte documentos Markdown para Microsoft Word (DOCX)",
			InputType:   "text/markdown",
			OutputType:  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		{
			ID:          "markdown-to-latex",
			Name:        "Markdown para LaTeX",
			Description: "Converte documentos Markdown para LaTeX",
			InputType:   "text/markdown",
			OutputType:  "application/x-tex",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"supported_formats": formats,
		"total_formats":     len(formats),
	})
}

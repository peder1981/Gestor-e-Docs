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
	var req ConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ConversionResponse{
			Success: false,
			Message: "Erro ao ler requisição: " + err.Error(),
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
	c.Header("Content-Disposition", "attachment; filename=document.pdf")
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// ConvertMarkdownToHTML converte um documento Markdown para HTML
func ConvertMarkdownToHTML(c *gin.Context) {
	var req ConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ConversionResponse{
			Success: false,
			Message: "Erro ao ler requisição: " + err.Error(),
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

// ListSupportedFormats retorna os formatos suportados para conversão
func ListSupportedFormats(c *gin.Context) {
	formats := []string{
		"markdown-to-pdf",
		"markdown-to-html",
	}
	
	c.JSON(http.StatusOK, gin.H{
		"supported_formats": formats,
	})
}

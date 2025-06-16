package handlers

import (
	"bytes"
	"gestor-e-docs/document-service/db"
	"gestor-e-docs/document-service/storage"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// DownloadDocumentFile faz o download direto do arquivo do documento
func DownloadDocumentFile(c *gin.Context) {
	docID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Buscar documento
	doc, err := db.DbCollections.Documents.GetDocumentByID(docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Documento não encontrado"})
		return
	}

	// Verificar permissões
	if !hasReadAccess(doc, userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para acessar este documento"})
		return
	}

	objectPath := doc.StoragePath

	// Obter o arquivo do MinIO
	minioClient, err := storage.GetMinioClient()
	if err != nil {
		log.Printf("Erro ao obter cliente MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no sistema de armazenamento"})
		return
	}

	// Obter o arquivo do MinIO
	contentBytes, err := minioClient.GetDocument(objectPath)
	if err != nil {
		log.Printf("Erro ao obter objeto do MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao acessar arquivo"})
		return
	}
	
	// Criar um reader a partir dos bytes
	object := bytes.NewReader(contentBytes)

	// Atualizar contador de visualizações de forma assíncrona
	updateViewCountAsync(docID)

	// Determinar o tipo de conteúdo
	contentType := http.DetectContentType([]byte{}) // Placeholder
	if doc.Metadata.OriginalExtension != "" {
		switch doc.Metadata.OriginalExtension {
		case "md":
			contentType = "text/markdown"
		case "pdf":
			contentType = "application/pdf"
		case "doc", "docx":
			contentType = "application/msword"
		case "xls", "xlsx":
			contentType = "application/vnd.ms-excel"
		default:
			contentType = "application/octet-stream"
		}
	}

	// Nome do arquivo para download
	filename := doc.Title
	if filepath.Ext(filename) == "" {
		// Adicionar extensão se não tiver
		filename = filename + "." + doc.Metadata.OriginalExtension
	}

	// Configurar cabeçalhos para forçar download
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", contentType)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	// Stream do arquivo direto para o cliente
	c.DataFromReader(http.StatusOK, doc.Metadata.FileSize, contentType, object, nil)
}

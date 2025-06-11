package handlers

import (
	"context"
	"gestor-e-docs/document-service/db"
	"gestor-e-docs/document-service/models"
	"gestor-e-docs/document-service/storage"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateDocument cria um novo documento
func CreateDocument(c *gin.Context) {
	// Extrair userID do token JWT (adicionado pelo middleware de autenticação)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Parse do corpo da requisição
	var docRequest models.DocumentCreate
	if err := c.ShouldBindJSON(&docRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Garantir que o AuthorID seja o userID da sessão
	docRequest.AuthorID = userID.(string)

	// Criar um novo documento com base nos dados recebidos
	now := time.Now()
	newDoc := models.Document{
		Title:      docRequest.Title,
		Content:    docRequest.Content,
		AuthorID:   docRequest.AuthorID,
		CreatedAt:  now,
		UpdatedAt:  now,
		Tags:       docRequest.Tags,
		Categories: docRequest.Categories,
		Status:     models.StatusDraft,
		Permissions: models.DocumentPermissions{
			OwnerID:  docRequest.AuthorID,
			IsPublic: docRequest.IsPublic,
			ReadAccess: []string{},
			WriteAccess: []string{},
			AdminAccess: []string{},
		},
		Metadata: models.DocumentMetadata{
			FileSize:          int64(len(docRequest.Content)),
			OriginalExtension: "md",
			LastViewedAt:      now,
			ViewCount:         0,
			IsTemplate:        false,
			Keywords:          []string{},
			CustomFields:      map[string]interface{}{},
		},
	}

	// Salvar conteúdo no MinIO
	minioClient, err := storage.GetMinioClient()
	if err != nil {
		log.Printf("Erro ao obter cliente MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no sistema de armazenamento"})
		return
	}

	// Criar um ID temporário para o documento enquanto não temos o ID do MongoDB
	tempID := primitive.NewObjectID().Hex()
	objectPath, err := minioClient.UploadDocument(
		[]byte(docRequest.Content),
		docRequest.AuthorID,
		tempID,
		"text/markdown",
	)
	if err != nil {
		log.Printf("Erro ao fazer upload do documento: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao armazenar o documento"})
		return
	}

	// Definir o caminho de armazenamento no documento
	newDoc.StoragePath = objectPath

	// Salvar o documento no MongoDB
	err = db.DbCollections.Documents.InsertDocument(&newDoc)
	if err != nil {
		log.Printf("Erro ao inserir documento no MongoDB: %v", err)
		
		// Tentar limpar o arquivo do MinIO em caso de falha
		deleteErr := minioClient.DeleteDocument(objectPath)
		if deleteErr != nil {
			log.Printf("Erro ao limpar arquivo do MinIO após falha: %v", deleteErr)
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao salvar o documento"})
		return
	}

	// Retornar o documento criado
	c.JSON(http.StatusCreated, gin.H{
		"message": "Documento criado com sucesso",
		"id": newDoc.ID.Hex(),
		"title": newDoc.Title,
	})
}

// GetDocument busca um documento pelo ID
func GetDocument(c *gin.Context) {
	docID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Buscar documento no MongoDB
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

	// Buscar conteúdo do documento no MinIO
	minioClient, err := storage.GetMinioClient()
	if err != nil {
		log.Printf("Erro ao obter cliente MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no sistema de armazenamento"})
		return
	}

	content, err := minioClient.GetDocument(doc.StoragePath)
	if err != nil {
		log.Printf("Erro ao buscar documento no MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao recuperar conteúdo do documento"})
		return
	}

	// Atualizar contadores de visualização
	updateViewCountAsync(doc.ID.Hex())

	// Preparar resposta
	doc.Content = string(content)

	c.JSON(http.StatusOK, doc)
}

// UpdateDocument atualiza um documento existente
func UpdateDocument(c *gin.Context) {
	docID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Buscar documento atual no MongoDB
	doc, err := db.DbCollections.Documents.GetDocumentByID(docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Documento não encontrado"})
		return
	}

	// Verificar permissões de escrita
	if !hasWriteAccess(doc, userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para editar este documento"})
		return
	}

	// Parse do corpo da requisição
	var docUpdate models.DocumentUpdate
	if err := c.ShouldBindJSON(&docUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Se o conteúdo foi atualizado, salvar nova versão no MinIO
	if docUpdate.Content != "" {
		minioClient, err := storage.GetMinioClient()
		if err != nil {
			log.Printf("Erro ao obter cliente MinIO: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no sistema de armazenamento"})
			return
		}

		newObjectPath, err := minioClient.UploadDocument(
			[]byte(docUpdate.Content),
			userID.(string),
			docID,
			"text/markdown",
		)
		if err != nil {
			log.Printf("Erro ao fazer upload da nova versão: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao armazenar a nova versão"})
			return
		}

		// Atualizar o caminho no documento
		docUpdate.Content = newObjectPath
	}

	// Atualizar no MongoDB
	err = db.DbCollections.Documents.UpdateDocument(docID, &docUpdate, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar o documento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Documento atualizado com sucesso",
		"id": docID,
	})
}

// ListDocuments lista documentos do usuário com paginação e filtros
func ListDocuments(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Extrair parâmetros de consulta
	var query models.DocumentSearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Se não for especificado, listar apenas documentos do usuário
	if query.AuthorID == "" {
		query.AuthorID = userID.(string)
	}

	// Buscar documentos que o usuário tem acesso
	docs, err := db.DbCollections.Documents.SearchDocuments(&query)
	if err != nil {
		log.Printf("Erro ao buscar documentos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar documentos"})
		return
	}

	// Contar total para paginação
	filter := bson.M{"author_id": query.AuthorID}
	if query.Query != "" {
		filter["$text"] = bson.M{"$search": query.Query}
	}
	total, err := db.DbCollections.Documents.CountDocuments(filter)
	if err != nil {
		log.Printf("Erro ao contar documentos: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": docs,
		"total": total,
		"offset": query.Offset,
		"limit": query.Limit,
	})
}

// DeleteDocument exclui um documento
func DeleteDocument(c *gin.Context) {
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

	// Verificar se o usuário é o dono ou tem acesso administrativo
	if !hasAdminAccess(doc, userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para excluir este documento"})
		return
	}

	// Excluir documento do MinIO
	minioClient, err := storage.GetMinioClient()
	if err != nil {
		log.Printf("Erro ao obter cliente MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no sistema de armazenamento"})
		return
	}

	// Excluir o arquivo principal e todas as versões
	err = minioClient.DeleteDocument(doc.StoragePath)
	if err != nil {
		log.Printf("Aviso: Erro ao excluir arquivo do MinIO: %v", err)
		// Continuar mesmo com erro, para pelo menos excluir do MongoDB
	}

	// Tentar excluir todas as versões
	for _, version := range doc.VersionHistory {
		if version.StoragePath != doc.StoragePath {
			err = minioClient.DeleteDocument(version.StoragePath)
			if err != nil {
				log.Printf("Aviso: Erro ao excluir versão %d do MinIO: %v", version.VersionNumber, err)
			}
		}
	}

	// Excluir do MongoDB
	err = db.DbCollections.Documents.DeleteDocument(docID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao excluir o documento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Documento excluído com sucesso",
	})
}

// DownloadDocument gera URL temporária para download do documento
func DownloadDocument(c *gin.Context) {
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

	// Buscar versão específica (opcional)
	versionStr := c.Query("version")
	objectPath := doc.StoragePath

	if versionStr != "" {
		// Se foi solicitada uma versão específica
		// Aqui deveria ter lógica para determinar o caminho da versão específica
		// Por simplicidade, estou ignorando este caso
	}

	// Gerar URL para download
	minioClient, err := storage.GetMinioClient()
	if err != nil {
		log.Printf("Erro ao obter cliente MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no sistema de armazenamento"})
		return
	}

	// URL válida por 1 hora
	url, err := minioClient.GetDocumentURL(objectPath, time.Hour)
	if err != nil {
		log.Printf("Erro ao gerar URL de download: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar link de download"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": url,
		"expires_in": "1 hora",
		"filename": doc.Title + ".md",
	})
}

// Funções auxiliares para verificação de permissões

// hasReadAccess verifica se um usuário tem permissão de leitura
func hasReadAccess(doc *models.Document, userID string) bool {
	// O dono sempre tem acesso
	if doc.Permissions.OwnerID == userID {
		return true
	}

	// Documento público
	if doc.Permissions.IsPublic {
		return true
	}

	// Usuário com acesso explícito de leitura, escrita ou admin
	for _, id := range doc.Permissions.ReadAccess {
		if id == userID {
			return true
		}
	}
	for _, id := range doc.Permissions.WriteAccess {
		if id == userID {
			return true
		}
	}
	for _, id := range doc.Permissions.AdminAccess {
		if id == userID {
			return true
		}
	}

	return false
}

// hasWriteAccess verifica se um usuário tem permissão de escrita
func hasWriteAccess(doc *models.Document, userID string) bool {
	// O dono sempre tem acesso
	if doc.Permissions.OwnerID == userID {
		return true
	}

	// Usuário com acesso de escrita ou admin
	for _, id := range doc.Permissions.WriteAccess {
		if id == userID {
			return true
		}
	}
	for _, id := range doc.Permissions.AdminAccess {
		if id == userID {
			return true
		}
	}

	return false
}

// hasAdminAccess verifica se um usuário tem permissões administrativas
func hasAdminAccess(doc *models.Document, userID string) bool {
	// O dono sempre tem acesso admin
	if doc.Permissions.OwnerID == userID {
		return true
	}

	// Usuário com acesso admin
	for _, id := range doc.Permissions.AdminAccess {
		if id == userID {
			return true
		}
	}

	return false
}

// updateViewCountAsync atualiza o contador de visualizações de forma assíncrona
func updateViewCountAsync(docID string) {
	go func() {
		// Convertendo o docID para ObjectID
		docObjID, err := primitive.ObjectIDFromHex(docID)
		if err != nil {
			log.Printf("Erro ao converter docID para ObjectID: %v", err)
			return
		}

		// Atualização do contador e timestamp
		update := bson.M{
			"$inc": bson.M{"metadata.view_count": 1},
			"$set": bson.M{"metadata.last_viewed_at": time.Now()},
		}

		// Usando a coleção diretamente para evitar bloquear o método principal
		collection := db.DbCollections.Documents.Collection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = collection.UpdateOne(
			ctx,
			bson.M{"_id": docObjID},
			update,
		)
		if err != nil {
			log.Printf("Erro ao atualizar contador de visualizações: %v", err)
		}
	}()
}

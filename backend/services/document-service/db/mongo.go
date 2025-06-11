package db

import (
	"context"
	"log"
	"os"
	"time"

	"gestor-e-docs/document-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var database *mongo.Database

// DocCollection é uma struct para encapsular operações em uma coleção específica
type DocCollection struct {
	Collection *mongo.Collection
}

// Collections contém referências a todas as coleções do banco de dados
type Collections struct {
	Documents *DocCollection
}

// DbCollections contém todas as coleções do banco de dados
var DbCollections Collections

// ConnectDatabase estabelece conexão com o MongoDB
func ConnectDatabase() error {
	// Obter a URI do MongoDB da variável de ambiente ou usar valor padrão
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
		log.Println("Aviso: Usando URI padrão do MongoDB. Considere definir a variável de ambiente MONGO_URI.")
	}

	// Obter o nome do banco de dados da variável de ambiente ou usar valor padrão
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "gestor_docs"
		log.Println("Aviso: Usando nome padrão do banco de dados. Considere definir a variável de ambiente MONGO_DB_NAME.")
	}

	// Configurar o contexto com timeout para a conexão
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Conectar ao MongoDB
	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Verificar a conexão
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	log.Println("Conexão com MongoDB estabelecida com sucesso")
	database = client.Database(dbName)

	// Inicializar coleções
	initCollections()

	// Criar índices para melhor performance nas consultas
	createIndices()

	return nil
}

// DisconnectDatabase fecha a conexão com o MongoDB
func DisconnectDatabase() {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := client.Disconnect(ctx)
		if err != nil {
			log.Printf("Erro ao desconectar do MongoDB: %v", err)
			return
		}

		log.Println("Conexão com MongoDB fechada com sucesso")
	}
}

// initCollections inicializa as coleções do banco de dados
func initCollections() {
	DbCollections = Collections{
		Documents: &DocCollection{
			Collection: database.Collection("documents"),
		},
	}
}

// createIndices cria índices para otimizar consultas frequentes
func createIndices() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Índices para a coleção de documentos
	documentIndices := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "title", Value: "text"}, {Key: "content", Value: "text"}},
			Options: options.Index().
				SetName("text_search").
				SetWeights(bson.D{
					{Key: "title", Value: 10},
					{Key: "content", Value: 5},
				}),
		},
		{
			Keys:    bson.D{{Key: "author_id", Value: 1}},
			Options: options.Index().SetName("author_id_idx"),
		},
		{
			Keys:    bson.D{{Key: "tags", Value: 1}},
			Options: options.Index().SetName("tags_idx"),
		},
		{
			Keys:    bson.D{{Key: "categories", Value: 1}},
			Options: options.Index().SetName("categories_idx"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("status_idx"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_idx"),
		},
		{
			Keys:    bson.D{{Key: "updated_at", Value: -1}},
			Options: options.Index().SetName("updated_at_idx"),
		},
	}

	_, err := DbCollections.Documents.Collection.Indexes().CreateMany(ctx, documentIndices)
	if err != nil {
		log.Printf("Erro ao criar índices para a coleção de documentos: %v", err)
	} else {
		log.Println("Índices criados com sucesso para a coleção de documentos")
	}
}

// Métodos do DocCollection para operações CRUD

// InsertDocument insere um novo documento no banco de dados
func (c *DocCollection) InsertDocument(doc *models.Document) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Configurar campos de criação
	doc.ID = primitive.NewObjectID()
	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	// Inicializar histórico de versões
	doc.VersionHistory = []models.Version{
		{
			VersionNumber: 1,
			CreatedAt:     now,
			AuthorID:      doc.AuthorID,
			Description:   "Criação inicial do documento",
			StoragePath:   doc.StoragePath,
		},
	}

	_, err := c.Collection.InsertOne(ctx, doc)
	return err
}

// UpdateDocument atualiza um documento existente
func (c *DocCollection) UpdateDocument(id string, update *models.DocumentUpdate, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Obter o documento atual para calcular a nova versão
	var currentDoc models.Document
	err = c.Collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&currentDoc)
	if err != nil {
		return err
	}

	// Construir o documento de atualização
	now := time.Now()
	updateFields := bson.M{
		"updated_at": now,
	}

	if update.Title != "" {
		updateFields["title"] = update.Title
	}
	if update.Content != "" {
		updateFields["content"] = update.Content
	}
	if update.Tags != nil {
		updateFields["tags"] = update.Tags
	}
	if update.Categories != nil {
		updateFields["categories"] = update.Categories
	}
	if update.Status != "" {
		updateFields["status"] = update.Status
	}

	// Criar uma nova versão se o conteúdo foi alterado
	if update.Content != "" {
		// Gerar um novo caminho de armazenamento para esta versão
		versionPath := currentDoc.StoragePath + ".v" + 
			primitive.NewObjectID().Hex()

		newVersion := models.Version{
			VersionNumber: len(currentDoc.VersionHistory) + 1,
			CreatedAt:     now,
			AuthorID:      userID,
			Description:   update.Description,
			StoragePath:   versionPath,
		}

		updateFields["$push"] = bson.M{
			"version_history": newVersion,
		}
	}

	_, err = c.Collection.UpdateOne(
		ctx,
		bson.M{"_id": docID},
		bson.M{"$set": updateFields},
	)
	return err
}

// GetDocumentByID busca um documento pelo ID
func (c *DocCollection) GetDocumentByID(id string) (*models.Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var document models.Document
	err = c.Collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&document)
	if err != nil {
		return nil, err
	}

	return &document, nil
}

// SearchDocuments busca documentos com base em critérios de pesquisa
func (c *DocCollection) SearchDocuments(query *models.DocumentSearchQuery) ([]models.DocumentListItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}

	// Aplicar filtros de busca
	if query.Query != "" {
		filter["$text"] = bson.M{"$search": query.Query}
	}
	if len(query.Tags) > 0 {
		filter["tags"] = bson.M{"$in": query.Tags}
	}
	if len(query.Categories) > 0 {
		filter["categories"] = bson.M{"$in": query.Categories}
	}
	if query.AuthorID != "" {
		filter["author_id"] = query.AuthorID
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}

	// Filtros de data
	dateFilter := bson.M{}
	if query.DateFrom != "" {
		dateFrom, err := time.Parse(time.RFC3339, query.DateFrom)
		if err == nil {
			dateFilter["$gte"] = dateFrom
		}
	}
	if query.DateTo != "" {
		dateTo, err := time.Parse(time.RFC3339, query.DateTo)
		if err == nil {
			dateFilter["$lte"] = dateTo
		}
	}
	if len(dateFilter) > 0 {
		filter["created_at"] = dateFilter
	}

	// Configurar ordenação
	opts := options.Find()
	if query.SortBy != "" {
		sortOrder := 1 // Crescente por padrão
		if query.SortOrder == "desc" {
			sortOrder = -1
		}
		opts.SetSort(bson.D{{Key: query.SortBy, Value: sortOrder}})
	} else {
		// Ordenação padrão por data de atualização decrescente
		opts.SetSort(bson.D{{Key: "updated_at", Value: -1}})
	}

	// Configurar paginação
	if query.Limit <= 0 {
		query.Limit = 10 // Limite padrão
	}
	if query.Limit > 100 {
		query.Limit = 100 // Limite máximo
	}
	opts.SetSkip(int64(query.Offset))
	opts.SetLimit(int64(query.Limit))

	// Executar consulta
	cursor, err := c.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Converter resultados para a lista de documentos
	var results []models.DocumentListItem
	for cursor.Next(ctx) {
		var doc models.Document
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}

		// Converter para DocumentListItem
		item := models.DocumentListItem{
			ID:           doc.ID,
			Title:        doc.Title,
			AuthorID:     doc.AuthorID,
			CreatedAt:    doc.CreatedAt,
			UpdatedAt:    doc.UpdatedAt,
			Status:       doc.Status,
			Tags:         doc.Tags,
			Categories:   doc.Categories,
			VersionCount: len(doc.VersionHistory),
		}
		results = append(results, item)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteDocument remove um documento pelo ID
func (c *DocCollection) DeleteDocument(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = c.Collection.DeleteOne(ctx, bson.M{"_id": docID})
	return err
}

// CountDocuments conta documentos com base em critérios
func (c *DocCollection) CountDocuments(filter bson.M) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.Collection.CountDocuments(ctx, filter)
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Document representa a estrutura principal de um documento no sistema
type Document struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title           string               `bson:"title" json:"title"`
	Content         string               `bson:"content" json:"content"`
	AuthorID        string               `bson:"author_id" json:"author_id"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `bson:"updated_at" json:"updated_at"`
	Tags            []string             `bson:"tags" json:"tags"`
	Categories      []string             `bson:"categories" json:"categories"`
	Status          DocumentStatus       `bson:"status" json:"status"`
	VersionHistory  []Version            `bson:"version_history" json:"version_history"`
	StoragePath     string               `bson:"storage_path" json:"storage_path"`
	Permissions     DocumentPermissions  `bson:"permissions" json:"permissions"`
	Metadata        DocumentMetadata     `bson:"metadata" json:"metadata"`
}

// Version representa uma versão específica do documento
type Version struct {
	VersionNumber int       `bson:"version_number" json:"version_number"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	AuthorID      string    `bson:"author_id" json:"author_id"`
	Description   string    `bson:"description" json:"description"`
	StoragePath   string    `bson:"storage_path" json:"storage_path"`
}

// DocumentStatus representa o estado atual do documento
type DocumentStatus string

const (
	StatusDraft     DocumentStatus = "draft"
	StatusReview    DocumentStatus = "review"
	StatusPublished DocumentStatus = "published"
	StatusArchived  DocumentStatus = "archived"
)

// DocumentPermissions define quem pode acessar e modificar o documento
type DocumentPermissions struct {
	OwnerID       string   `bson:"owner_id" json:"owner_id"`
	IsPublic      bool     `bson:"is_public" json:"is_public"`
	ReadAccess    []string `bson:"read_access" json:"read_access"` // IDs de usuários com acesso de leitura
	WriteAccess   []string `bson:"write_access" json:"write_access"` // IDs de usuários com acesso de escrita
	AdminAccess   []string `bson:"admin_access" json:"admin_access"` // IDs de usuários com acesso administrativo
}

// DocumentMetadata contém informações adicionais sobre o documento
type DocumentMetadata struct {
	FileSize          int64     `bson:"file_size" json:"file_size"`
	OriginalExtension string    `bson:"original_extension" json:"original_extension"`
	LastViewedAt      time.Time `bson:"last_viewed_at" json:"last_viewed_at"`
	ViewCount         int       `bson:"view_count" json:"view_count"`
	IsTemplate        bool      `bson:"is_template" json:"is_template"`
	Keywords          []string  `bson:"keywords" json:"keywords"`
	CustomFields      map[string]interface{} `bson:"custom_fields" json:"custom_fields"`
}

// DocumentCreate representa os dados necessários para criar um novo documento
type DocumentCreate struct {
	Title      string   `json:"title" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	AuthorID   string   `json:"author_id" binding:"required"`
	Tags       []string `json:"tags"`
	Categories []string `json:"categories"`
	IsPublic   bool     `json:"is_public"`
}

// DocumentUpdate representa os dados para atualização de um documento existente
type DocumentUpdate struct {
	Title       string         `json:"title"`
	Content     string         `json:"content"`
	Tags        []string       `json:"tags"`
	Categories  []string       `json:"categories"`
	Status      DocumentStatus `json:"status"`
	Description string         `json:"description"` // Descrição da alteração para histórico de versões
}

// DocumentListItem representa um item resumido na listagem de documentos
type DocumentListItem struct {
	ID           primitive.ObjectID `json:"id"`
	Title        string             `json:"title"`
	AuthorID     string             `json:"author_id"`
	AuthorName   string             `json:"author_name,omitempty"` // Será preenchido em tempo de execução
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Status       DocumentStatus     `json:"status"`
	Tags         []string           `json:"tags"`
	Categories   []string           `json:"categories"`
	VersionCount int                `json:"version_count"`
}

// DocumentSearchQuery representa os parâmetros para busca de documentos
type DocumentSearchQuery struct {
	Query       string   `form:"query"`
	Tags        []string `form:"tags"`
	Categories  []string `form:"categories"`
	AuthorID    string   `form:"author_id"`
	Status      string   `form:"status"`
	SortBy      string   `form:"sort_by"`
	SortOrder   string   `form:"sort_order"`
	DateFrom    string   `form:"date_from"`
	DateTo      string   `form:"date_to"`
	Offset      int      `form:"offset"`
	Limit       int      `form:"limit"`
}

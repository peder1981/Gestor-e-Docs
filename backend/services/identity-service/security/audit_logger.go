package security

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AuditEvent representa um evento de auditoria
type AuditEvent struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Action      string             `bson:"action" json:"action"`
	Resource    string             `bson:"resource" json:"resource"`
	ResourceID  string             `bson:"resource_id" json:"resource_id"`
	IPAddress   string             `bson:"ip_address" json:"ip_address"`
	UserAgent   string             `bson:"user_agent" json:"user_agent"`
	RequestData json.RawMessage    `bson:"request_data,omitempty" json:"request_data,omitempty"`
	Status      int                `bson:"status" json:"status"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
}

// AuditLogger gerencia o registro de eventos de auditoria
type AuditLogger struct {
	collection  *mongo.Collection
	enabled     bool
	logRequests bool
}

// NewAuditLogger cria uma nova instância do logger de auditoria
func NewAuditLogger(collection *mongo.Collection, enabled, logRequests bool) *AuditLogger {
	return &AuditLogger{
		collection:  collection,
		enabled:     enabled,
		logRequests: logRequests,
	}
}

// LogEvent registra um evento de auditoria
func (al *AuditLogger) LogEvent(ctx context.Context, event AuditEvent) {
	if !al.enabled {
		return
	}

	// Certifica-se de que o timestamp está definido
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Insere o evento no MongoDB
	_, err := al.collection.InsertOne(ctx, event)
	if err != nil {
		log.Printf("[AuditLogger] Falha ao registrar evento: %v", err)
	}
}

// AuditMiddleware retorna um middleware Gin para registrar eventos de auditoria
func AuditMiddleware(logger *AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip para endpoints de saúde ou métricas
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Prepara os dados para auditoria
		userID := "anonymous"
		if id, exists := c.Get("userID"); exists && id != nil {
			userID = id.(string)
		}

		// Extrai o path e o resourceID (se houver)
		path := c.Request.URL.Path
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param("userId")
		}

		// Prepara os dados de requisição se habilitado
		var requestData json.RawMessage
		if logger.logRequests && c.Request.ContentLength > 0 {
			buf := make([]byte, 1024) // limite de 1kb para dados de requisição
			n, err := c.Request.Body.Read(buf)
			if err != nil {
				log.Printf("[AuditLogger] Erro ao ler corpo da requisição: %v", err)
			} else if n > 0 {
				requestData = buf[:n]
			}
			// Redefine o body para outros handlers poderem ler
			c.Request.Body = &bodyReader{buf: buf[:n], Reader: c.Request.Body}
		}

		// Executa o handler e captura o status
		c.Next()
		status := c.Writer.Status()

		// Registra o evento
		event := AuditEvent{
			UserID:      userID,
			Action:      c.Request.Method,
			Resource:    path,
			ResourceID:  resourceID,
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			RequestData: requestData,
			Status:      status,
			Timestamp:   time.Now(),
		}

		go logger.LogEvent(context.Background(), event)
	}
}

// bodyReader é um helper para permitir a leitura do body mais de uma vez
type bodyReader struct {
	buf    []byte
	Reader io.ReadCloser
}

// Read reimplementa a interface io.Reader
func (b *bodyReader) Read(p []byte) (n int, err error) {
	if len(b.buf) > 0 {
		n = copy(p, b.buf)
		b.buf = b.buf[n:]
		return n, nil
	}
	return b.Reader.Read(p)
}

// Close implementa a interface io.ReadCloser
func (b *bodyReader) Close() error {
	return b.Reader.Close()
}

// GetAuditEvents retorna eventos de auditoria com filtragem
func (al *AuditLogger) GetAuditEvents(ctx context.Context, filter map[string]interface{}, limit, skip int) ([]AuditEvent, error) {
	if !al.enabled {
		return nil, nil
	}

	if limit <= 0 {
		limit = 100 // limite padrão
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}}) // mais recentes primeiro

	cursor, err := al.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []AuditEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

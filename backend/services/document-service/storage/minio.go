package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient encapsula as operações de armazenamento com MinIO
type MinioClient struct {
	Client     *minio.Client
	BucketName string
}

var minioInstance *MinioClient

// GetMinioClient retorna uma instância singleton do cliente MinIO
func GetMinioClient() (*MinioClient, error) {
	if minioInstance != nil {
		return minioInstance, nil
	}

	// Parâmetros de conexão do MinIO
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	useSSL := false

	// Valores padrão para desenvolvimento local
	if endpoint == "" {
		endpoint = "minio_server:9000" // Valor padrão
		log.Println("Usando endpoint padrão do MinIO:", endpoint)
	}
	
	// Implementação de service discovery robusto
	// Em vez de usar IP hardcoded, confiaremos no DNS interno do Docker que resolve nomes de serviço
	// e implementaremos uma lógica de retry robusta com backoff exponencial
	log.Printf("Usando service discovery para conectar ao MinIO em: %s", endpoint)
	
	// Implementação de mecanismo robusto de retry com backoff exponencial para service discovery
	maxRetries := 15       // Aumentar número máximo de tentativas para descoberta de serviço
	baseInterval := 250 * time.Millisecond // Intervalo base para backoff exponencial
	maxInterval := 30 * time.Second        // Intervalo máximo entre tentativas
	
	log.Printf("Serviço de discovery iniciado para conexão com MinIO em %s", endpoint)
	log.Printf("Aguardando 2 segundos para garantir que todos os serviços estejam inicializados...")
	time.Sleep(time.Second * 2)  // Pequena pausa inicial para garantir que o DNS está funcional
	if accessKeyID == "" {
		accessKeyID = "minioadmin"  // Alterado para corresponder ao docker-compose
		log.Println("Usando accessKeyID padrão do MinIO")
	}
	if secretAccessKey == "" {
		secretAccessKey = "minioadmin"  // Alterado para corresponder ao docker-compose
		log.Println("Usando secretAccessKey padrão do MinIO")
	}
	if bucketName == "" {
		bucketName = "documents"
		log.Println("Usando bucket padrão do MinIO:", bucketName)
	}

	// Inicializar o cliente MinIO com mecanismo de retry simplificado
	log.Printf("Iniciando conexão com MinIO em %s", endpoint)
	
	// Configuração melhorada para resolver problemas de hostname em ambientes Docker
	minioOptions := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: "", // Definir região vazia para evitar validação de hostname
		BucketLookup: minio.BucketLookupAuto, // Permitir busca automática do bucket
		// Desativando validação de nome de host para permitir IPs e hostnames Docker internos
		Transport: &http.Transport{
			DisableCompression: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Em produção deve ser false e usar certificados válidos
			},
			// Timeout mais generoso para lidar com DNS lento em ambientes containerizados
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
	
	// Criar cliente com as opções robustas
	log.Printf("Conectando ao MinIO em: %s (secure=%v)", endpoint, useSSL)
	client, err := minio.New(endpoint, minioOptions)
	
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cliente MinIO: %v", err)
	}
	
	log.Printf("Cliente MinIO inicializado, testando acesso ao bucket '%s'", bucketName)
	
	// Verificar se o bucket existe usando backoff exponencial
	ctx := context.Background()
	
	// Tentativas de verificar o bucket com backoff exponencial
	var exists bool
	var bucketErr error
	
	for i := 0; i < maxRetries; i++ {
		log.Printf("Tentativa %d de verificar bucket '%s'", i+1, bucketName)
		exists, bucketErr = client.BucketExists(ctx, bucketName)
		
		if bucketErr == nil {
			log.Printf("Bucket '%s' verificado com sucesso", bucketName)
			break
		}
		
		// Calcular próximo intervalo com backoff exponencial
		// waitTime = min(baseInterval * 2^attempt, maxInterval)
		waitTime := baseInterval * time.Duration(1<<uint(i))
		if waitTime > maxInterval {
			waitTime = maxInterval
		}
		
		log.Printf("Falha na tentativa %d: %v. Backoff exponencial: aguardando %v antes da próxima tentativa...", 
			i+1, bucketErr, waitTime)
		time.Sleep(waitTime)
	}
	
	if bucketErr != nil {
		return nil, fmt.Errorf("falha ao verificar bucket após %d tentativas: %v", maxRetries, bucketErr)
	}

	// O bucket já foi verificado, mas pode não existir ainda
	if !exists {
		makeBucketErr := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if makeBucketErr != nil {
			return nil, fmt.Errorf("falha ao criar bucket %s: %v", bucketName, makeBucketErr)
		}
		log.Printf("Bucket %s criado com sucesso", bucketName)
	} else {
		log.Printf("Bucket %s já existe", bucketName)
	}

	// Definir política de acesso ao bucket (varia conforme o ambiente)
	if os.Getenv("APP_ENV") != "production" {
		// Em dev/test, permitir acesso de leitura para download facilitado
		policy := `{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::` + bucketName + `/*"]
				}
			]
		}`
		policyErr := client.SetBucketPolicy(ctx, bucketName, policy)
		if policyErr != nil {
			log.Printf("Aviso: falha ao definir política do bucket: %v", policyErr)
		}
	}

	minioInstance = &MinioClient{
		Client:     client,
		BucketName: bucketName,
	}

	return minioInstance, nil
}

// UploadDocument faz upload de um documento para o MinIO
func (m *MinioClient) UploadDocument(content []byte, userID string, documentID string, contentType string) (string, error) {
	// Criar um path único para o arquivo
	// Formato: userID/documentID/<uuid>.md
	fileUUID := uuid.New().String()
	objectName := filepath.Join(userID, documentID, fileUUID+".md")

	ctx := context.Background()
	reader := bytes.NewReader(content)
	contentSize := int64(len(content))

	// Configurar opções de upload
	opts := minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"user-id":     userID,
			"document-id": documentID,
			"uploaded-at": time.Now().Format(time.RFC3339),
		},
	}

	// Upload do arquivo
	info, err := m.Client.PutObject(ctx, m.BucketName, objectName, reader, contentSize, opts)
	if err != nil {
		return "", fmt.Errorf("falha ao fazer upload do documento: %v", err)
	}

	log.Printf("Arquivo '%s' de tamanho %d bytes enviado com sucesso", objectName, info.Size)
	return objectName, nil
}

// GetDocument recupera um documento do MinIO
func (m *MinioClient) GetDocument(objectName string) ([]byte, error) {
	ctx := context.Background()
	obj, err := m.Client.GetObject(ctx, m.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("falha ao obter objeto do MinIO: %v", err)
	}
	defer obj.Close()

	// Lê o conteúdo completo
	content, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler conteúdo do objeto: %v", err)
	}

	return content, nil
}

// DeleteDocument remove um documento do MinIO
func (m *MinioClient) DeleteDocument(objectName string) error {
	ctx := context.Background()
	err := m.Client.RemoveObject(ctx, m.BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("falha ao remover objeto do MinIO: %v", err)
	}
	return nil
}

// ListDocuments lista todos os documentos de um usuário
func (m *MinioClient) ListDocuments(userID string) ([]minio.ObjectInfo, error) {
	ctx := context.Background()
	prefix := userID + "/"
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	objectCh := m.Client.ListObjects(ctx, m.BucketName, opts)
	
	var objects []minio.ObjectInfo
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("erro ao listar objetos: %v", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// GetDocumentURL gera uma URL pré-assinada para acesso temporário a um documento
func (m *MinioClient) GetDocumentURL(objectName string, expiry time.Duration) (string, error) {
	ctx := context.Background()
	presignedURL, err := m.Client.PresignedGetObject(ctx, m.BucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("falha ao gerar URL pré-assinada: %v", err)
	}
	return presignedURL.String(), nil
}

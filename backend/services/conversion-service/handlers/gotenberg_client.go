package handlers

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

// GotenbergClient é um cliente para comunicação com a API do Gotenberg
type GotenbergClient struct {
	baseURL string
}

// NewGotenbergClient cria um novo cliente Gotenberg
func NewGotenbergClient() *GotenbergClient {
	baseURL := os.Getenv("GOTENBERG_API_URL")
	if baseURL == "" {
		baseURL = "http://gotenberg:3000"
	}
	return &GotenbergClient{baseURL: baseURL}
}

// ConvertMarkdownToPDF converte conteúdo Markdown para PDF usando Gotenberg
func (c *GotenbergClient) ConvertMarkdownToPDF(content string, title string) ([]byte, error) {
	// Preparar o buffer e o writer multipart
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Criar arquivo markdown
	markdownFile, err := writer.CreateFormFile("files", "input.md")
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo markdown: %v", err)
	}
	_, err = io.Copy(markdownFile, strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("erro ao escrever conteúdo markdown: %v", err)
	}

	// Adicionar título ao documento
	if title != "" {
		err = writer.WriteField("title", title)
		if err != nil {
			return nil, fmt.Errorf("erro ao adicionar título: %v", err)
		}
	}

	// Fechar o writer multipart
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("erro ao fechar writer multipart: %v", err)
	}

	// Criar requisição para o Gotenberg
	req, err := http.NewRequest("POST", c.baseURL+"/forms/chromium/convert/markdown", body)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Enviar requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %v", err)
	}
	defer resp.Body.Close()

	// Verificar status da resposta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na conversão: status %d", resp.StatusCode)
	}

	// Ler resposta
	return io.ReadAll(resp.Body)
}

// ConvertMarkdownToHTML converte conteúdo Markdown para HTML usando Gotenberg
func (c *GotenbergClient) ConvertMarkdownToHTML(content string) ([]byte, error) {
	// Preparar o buffer e o writer multipart
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Criar arquivo markdown
	markdownFile, err := writer.CreateFormFile("files", "input.md")
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo markdown: %v", err)
	}
	_, err = io.Copy(markdownFile, strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("erro ao escrever conteúdo markdown: %v", err)
	}

	// Fechar o writer multipart
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("erro ao fechar writer multipart: %v", err)
	}

	// Criar requisição para o Gotenberg
	req, err := http.NewRequest("POST", c.baseURL+"/forms/chromium/convert/markdown", body)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "text/html") // Solicitar HTML em vez de PDF

	// Enviar requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %v", err)
	}
	defer resp.Body.Close()

	// Verificar status da resposta
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na conversão: status %d", resp.StatusCode)
	}

	// Ler resposta
	return io.ReadAll(resp.Body)
}

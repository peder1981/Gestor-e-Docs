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

	// Criar arquivo HTML que incorpora o markdown
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        h1, h2, h3 { color: #333; }
        code { background-color: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
        pre { background-color: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <div id="content"></div>
    <script>
        const markdownContent = %s;
        document.getElementById('content').innerHTML = marked.parse(markdownContent);
    </script>
</body>
</html>`, title, fmt.Sprintf("`%s`", strings.ReplaceAll(content, "`", "\\`")))

	// Criar arquivo index.html (obrigatório pelo Gotenberg)
	htmlFile, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo HTML: %v", err)
	}
	_, err = io.Copy(htmlFile, strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("erro ao escrever conteúdo HTML: %v", err)
	}

	// Fechar o writer multipart
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("erro ao fechar writer multipart: %v", err)
	}

	// Criar requisição para o Gotenberg
	req, err := http.NewRequest("POST", c.baseURL+"/forms/chromium/convert/html", body)
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

	// Criar arquivo HTML que incorpora o markdown
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Converted Document</title>
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        h1, h2, h3 { color: #333; }
        code { background-color: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
        pre { background-color: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <div id="content"></div>
    <script>
        const markdownContent = %s;
        document.getElementById('content').innerHTML = marked.parse(markdownContent);
    </script>
</body>
</html>`, fmt.Sprintf("`%s`", strings.ReplaceAll(content, "`", "\\`")))

	// Criar arquivo index.html (obrigatório pelo Gotenberg)
	htmlFile, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("erro ao criar arquivo HTML: %v", err)
	}
	_, err = io.Copy(htmlFile, strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("erro ao escrever conteúdo HTML: %v", err)
	}

	// Fechar o writer multipart
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("erro ao fechar writer multipart: %v", err)
	}

	// Criar requisição para o Gotenberg (usar endpoint HTML em vez de markdown)
	req, err := http.NewRequest("POST", c.baseURL+"/forms/chromium/convert/html", body)
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

	// Ler resposta (que será um PDF, mas vamos extrair o HTML)
	return io.ReadAll(resp.Body)
}

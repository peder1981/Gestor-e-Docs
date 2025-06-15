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

// ConvertMarkdownToDOCX converte conteúdo Markdown para DOCX usando pandoc via LibreOffice
func (c *GotenbergClient) ConvertMarkdownToDOCX(content string, title string) ([]byte, error) {
	// Preparar o buffer e o writer multipart
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Criar arquivo HTML bem formatado que será convertido para DOCX via LibreOffice
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <style>
        @page {
            margin: 2cm;
        }
        body { 
            font-family: 'Times New Roman', serif; 
            font-size: 12pt;
            line-height: 1.5; 
            color: #000;
            max-width: none;
            margin: 0;
            padding: 0;
        }
        h1 { font-size: 18pt; font-weight: bold; margin: 24pt 0 12pt 0; }
        h2 { font-size: 16pt; font-weight: bold; margin: 18pt 0 6pt 0; }
        h3 { font-size: 14pt; font-weight: bold; margin: 12pt 0 6pt 0; }
        p { margin: 6pt 0; text-align: justify; }
        code { 
            font-family: 'Courier New', monospace; 
            background-color: #f0f0f0; 
            padding: 2pt 4pt; 
            border: 1pt solid #ccc;
        }
        pre { 
            font-family: 'Courier New', monospace;
            background-color: #f0f0f0; 
            padding: 12pt; 
            border: 1pt solid #ccc;
            margin: 12pt 0;
            white-space: pre-wrap;
        }
        ul, ol { margin: 6pt 0; padding-left: 24pt; }
        li { margin: 3pt 0; }
        blockquote {
            margin: 12pt 24pt;
            padding: 6pt 12pt;
            border-left: 4pt solid #ccc;
            font-style: italic;
        }
        table {
            border-collapse: collapse;
            width: 100%%;
            margin: 12pt 0;
        }
        th, td {
            border: 1pt solid #000;
            padding: 6pt;
            text-align: left;
        }
        th {
            background-color: #f0f0f0;
            font-weight: bold;
        }
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

	// Usar endpoint do LibreOffice para gerar DOCX
	req, err := http.NewRequest("POST", c.baseURL+"/forms/libreoffice/convert", body)
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

// ConvertMarkdownToLaTeX converte conteúdo Markdown para LaTeX
func (c *GotenbergClient) ConvertMarkdownToLaTeX(content string, title string) ([]byte, error) {
	// Converter Markdown para LaTeX usando conversor interno
	latexContent := c.markdownToLatex(content, title)
	return []byte(latexContent), nil
}

// markdownToLatex é um conversor simples de Markdown para LaTeX
func (c *GotenbergClient) markdownToLatex(content string, title string) string {
	// Cabeçalho LaTeX básico
	latex := `\documentclass[12pt,a4paper]{article}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage[brazil]{babel}
\usepackage{amsmath}
\usepackage{amsfonts}
\usepackage{amssymb}
\usepackage{graphicx}
\usepackage{url}
\usepackage{listings}
\usepackage{xcolor}
\usepackage{geometry}
\geometry{margin=2.5cm}

\lstset{
    basicstyle=\ttfamily\small,
    backgroundcolor=\color{gray!10},
    frame=single,
    breaklines=true,
    captionpos=b,
    numbers=left,
    numberstyle=\tiny\color{gray},
    keywordstyle=\color{blue},
    commentstyle=\color{green!60!black},
    stringstyle=\color{red}
}

`
	
	if title != "" {
		latex += fmt.Sprintf("\\title{%s}\n", c.escapeLatex(title))
		latex += "\\author{}\n"
		latex += "\\date{\\today}\n"
	}
	
	latex += "\\begin{document}\n"
	
	if title != "" {
		latex += "\\maketitle\n\n"
	}

	// Converter conteúdo Markdown básico para LaTeX
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	codeLanguage := ""
	
	for _, line := range lines {
		line = strings.TrimRight(line, " ")
		
		// Detectar blocos de código
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				// Início do bloco de código
				inCodeBlock = true
				codeLanguage = strings.TrimPrefix(line, "```")
				if codeLanguage == "" {
					codeLanguage = "text"
				}
				latex += fmt.Sprintf("\\begin{lstlisting}[language=%s]\n", codeLanguage)
			} else {
				// Fim do bloco de código
				inCodeBlock = false
				latex += "\\end{lstlisting}\n\n"
			}
			continue
		}
		
		if inCodeBlock {
			latex += line + "\n"
			continue
		}
		
		// Títulos
		if strings.HasPrefix(line, "### ") {
			latex += fmt.Sprintf("\\subsubsection{%s}\n", c.escapeLatex(strings.TrimPrefix(line, "### ")))
		} else if strings.HasPrefix(line, "## ") {
			latex += fmt.Sprintf("\\subsection{%s}\n", c.escapeLatex(strings.TrimPrefix(line, "## ")))
		} else if strings.HasPrefix(line, "# ") {
			latex += fmt.Sprintf("\\section{%s}\n", c.escapeLatex(strings.TrimPrefix(line, "# ")))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// Lista não ordenada (simplificada)
			latex += fmt.Sprintf("\\item %s\n", c.escapeLatex(strings.TrimPrefix(line, "- ")))
		} else if line == "" {
			latex += "\n"
		} else {
			// Parágrafo normal
			if line != "" {
				// Tratar texto em negrito e itálico básico
				processedLine := c.processInlineFormatting(line)
				latex += processedLine + "\n\n"
			}
		}
	}
	
	latex += "\\end{document}\n"
	
	return latex
}

// escapeLatex escapa caracteres especiais do LaTeX
func (c *GotenbergClient) escapeLatex(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\textbackslash{}")
	text = strings.ReplaceAll(text, "{", "\\{")
	text = strings.ReplaceAll(text, "}", "\\}")
	text = strings.ReplaceAll(text, "$", "\\$")
	text = strings.ReplaceAll(text, "&", "\\&")
	text = strings.ReplaceAll(text, "%", "\\%")
	text = strings.ReplaceAll(text, "#", "\\#")
	text = strings.ReplaceAll(text, "^", "\\textasciicircum{}")
	text = strings.ReplaceAll(text, "_", "\\_")
	text = strings.ReplaceAll(text, "~", "\\textasciitilde{}")
	return text
}

// processInlineFormatting processa formatação inline básica (negrito, itálico, código)
func (c *GotenbergClient) processInlineFormatting(text string) string {
	// Código inline
	text = strings.ReplaceAll(text, "`", "\\texttt{")
	
	// Negrito (**texto**)
	boldRegex := `\*\*(.*?)\*\*`
	text = strings.ReplaceAll(text, boldRegex, "\\textbf{$1}")
	
	// Itálico (*texto*)
	italicRegex := `\*(.*?)\*`
	text = strings.ReplaceAll(text, italicRegex, "\\textit{$1}")
	
	return c.escapeLatex(text)
}

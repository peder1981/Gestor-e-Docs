package handlers

// ConversionRequest representa uma requisição de conversão de documento
type ConversionRequest struct {
	Content string `json:"content" binding:"required"`
	Title   string `json:"title"`
}

// ConversionResponse representa uma resposta de conversão de documento
type ConversionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    []byte `json:"data,omitempty"`
}

// SupportedFormat representa um formato de conversão suportado
type SupportedFormat struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	InputType   string `json:"input_type"`
	OutputType  string `json:"output_type"`
}

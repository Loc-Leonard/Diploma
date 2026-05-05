package cv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileProcessor interface {
	ProcessFile(ctx context.Context, filePath string, originalFilename string) (*FileProcessResult, error)
}

type FileProcessResult struct {
	SourcePath    string          `json:"source_path"`
	PredictedType string          `json:"predicted_type"`
	OCR           OCRPayload      `json:"ocr"`
	Extraction    ExtractionEntry `json:"extraction"`
	RawJSON       json.RawMessage `json:"-"`
}

type OCRPayload struct {
	Method            string   `json:"method"`
	AverageConfidence *float64 `json:"average_confidence"`
}

type ExtractionEntry struct {
	DocumentPath     string  `json:"document_path"`
	DocumentType     string  `json:"document_type"`
	MaterialName     *string `json:"material_name"`
	Quantity         *int    `json:"quantity"`
	Unit             *string `json:"unit"`
	DocumentNumber   *string `json:"document_number"`
	Confidence       float64 `json:"confidence"`
	ExtractionMethod string  `json:"extraction_method"`
}

type HTTPProcessor struct {
	BaseURL string
	Client  *http.Client
}

func (p HTTPProcessor) ProcessFile(ctx context.Context, filePath string, originalFilename string) (*FileProcessResult, error) {
	fileBytes, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("read file for cv: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", originalFilename)
	if err != nil {
		return nil, fmt.Errorf("build multipart request: %w", err)
	}
	if _, err := part.Write(fileBytes); err != nil {
		return nil, fmt.Errorf("write multipart request: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(p.BaseURL, "/")+"/process-file",
		body,
	)
	if err != nil {
		return nil, fmt.Errorf("create cv request: %w", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call cv service: %w", err)
	}
	defer response.Body.Close()

	var result FileProcessResult
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read cv response: %w", err)
	}
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("cv service returned %d: %s", response.StatusCode, string(responseBytes))
	}
	if err := json.Unmarshal(responseBytes, &result); err != nil {
		return nil, fmt.Errorf("decode cv response: %w", err)
	}
	result.RawJSON = append(result.RawJSON[:0], responseBytes...)
	return &result, nil
}

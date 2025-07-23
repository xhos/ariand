package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// Receipt represents the structured data from a parsed receipt.
type Receipt struct {
	Merchant string  `json:"merchant"`
	Date     string  `json:"date"`
	Total    float64 `json:"total"`
	Items    []struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
		Qty   int     `json:"qty"`
	} `json:"items"`
}

// ReceiptService defines the client for the receipts microservice.
type ReceiptService interface {
	Parse(ctx context.Context, file io.Reader, fileName string) (*Receipt, error)
}

type receiptSvc struct {
	client  *http.Client
	baseURL string
}

// NewReceiptService creates a new client for the receipts microservice.
func newReceiptSvc() ReceiptService {
	receiptsURL := os.Getenv("RECEIPTS_API_URL")
	if receiptsURL == "" {
		receiptsURL = "http://localhost:8081" // Default URL
	}

	return &receiptSvc{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: receiptsURL,
	}
}

// Parse sends an image file to the /parse endpoint.
func (s *receiptSvc) Parse(ctx context.Context, file io.Reader, fileName string) (*Receipt, error) {
	// Create a buffer to write our multipart form data.
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field.
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file to buffer: %w", err)
	}
	writer.Close()

	// Create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/parse", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request.
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to receipts service failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("receipts service returned non-200 status: %s", resp.Status)
	}

	// Decode the JSON response.
	var receipt Receipt
	if err := json.NewDecoder(resp.Body).Decode(&receipt); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &receipt, nil
}

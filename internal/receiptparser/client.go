package receiptparser

import (
	ariandv1 "ariand/gen/go/ariand/v1"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"time"
)

// ParsedReceipt represents the  response from the parser microservice
type ParsedReceipt struct {
	Merchant string  `json:"merchant"`
	Date     string  `json:"date"` // YYYY-MM-DD
	Total    float64 `json:"total"`
	Items    []struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
		Qty   float64 `json:"qty"`
	} `json:"items"`
}

// Client is the interface for communicating with the parser
type Client interface {
	Parse(
		ctx context.Context,
		file io.Reader,
		filename string,
		provider ariandv1.ReceiptEngine,
	) (
		*ParsedReceipt,
		[]byte, error,
	)
}

// parserClient is the implementation of Client
type parserClient struct {
	httpClient *http.Client
	baseURL    string
}

func New(baseURL string, timeout time.Duration) Client {
	return &parserClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

// Parse sends an image file to the /parse endpoint and returns ParsedReceipt
func (c *parserClient) Parse(
	ctx context.Context,
	file io.Reader,
	filename string,
	provider ariandv1.ReceiptEngine,
) (*ParsedReceipt, []byte, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// add the file to the request
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, nil, fmt.Errorf("failed to copy file to multipart buffer: %w", err)
	}

	writer.Close()

	// TODO: add support for that in the python code
	if err := writer.WriteField("provider", string(provider)); err != nil {
		return nil, nil, fmt.Errorf("failed to write provider field: %w", err)
	}

	// create and send the request
	url := fmt.Sprintf("%s/parse", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create parser request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request to parser service at %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	// read the raw response
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read parser response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, rawBody, fmt.Errorf("parser service returned non-OK status: %d. Body: %s", resp.StatusCode, string(rawBody))
	}

	// decode response itno JSON
	var parsedData ParsedReceipt
	if err := json.Unmarshal(rawBody, &parsedData); err != nil {
		return nil, rawBody, fmt.Errorf("failed to decode parser JSON response: %w", err)
	}

	return &parsedData, rawBody, nil
}

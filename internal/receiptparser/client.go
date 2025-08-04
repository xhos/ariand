package receiptparser

import (
	ariandv1 "ariand/gen/go/arian/v1"
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is the interface for communicating with the parser
type Client interface {
	Parse(
		ctx context.Context,
		file io.Reader,
		filename string,
		contentType string,
		engine *ariandv1.ReceiptEngine,
	) (*ariandv1.Receipt, error)

	GetStatus(ctx context.Context) (*ariandv1.GetStatusResponse, error)
}

// grpcClient is the implementation of Client using gRPC
type grpcClient struct {
	client ariandv1.ReceiptParsingServiceClient
	conn   *grpc.ClientConn
}

func New(address string, timeout time.Duration) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to receipt parsing service at %s: %w", address, err)
	}

	client := ariandv1.NewReceiptParsingServiceClient(conn)

	return &grpcClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}

// Parse sends an image to the gRPC service and returns the parsed receipt
func (c *grpcClient) Parse(
	ctx context.Context,
	file io.Reader,
	filename string,
	contentType string,
	engine *ariandv1.ReceiptEngine,
) (*ariandv1.Receipt, error) {
	// Read the file data
	imageData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Create the gRPC request
	req := &ariandv1.ParseImageRequest{
		ImageData:   imageData,
		ContentType: contentType,
		Engine:      engine,
	}

	// Call the gRPC service
	resp, err := c.client.ParseImage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image via gRPC: %w", err)
	}

	return resp.Receipt, nil
}

// GetStatus returns the status of available parsing providers
func (c *grpcClient) GetStatus(ctx context.Context) (*ariandv1.GetStatusResponse, error) {
	req := &ariandv1.GetStatusRequest{}

	resp, err := c.client.GetStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get status via gRPC: %w", err)
	}

	return resp, nil
}

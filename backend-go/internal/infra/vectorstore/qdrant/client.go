package qdrant

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	sdk "github.com/qdrant/go-client/qdrant"
	"github.com/ran/demo/backend-go/internal/domain/models"
)

// QdrantClient wraps the Qdrant SDK GrpcClient for our specific needs
// and exposes convenience methods
// It embeds the GrpcClient for direct access
// (if you want to hide the SDK, use composition instead)
type QdrantClient struct {
	grpcClient *sdk.GrpcClient
}

// NewQdrantClient creates a new Qdrant client using the official SDK GrpcClient
func NewQdrantClient(endpoint, apiKey string) (*QdrantClient, error) {
	if endpoint == "" {
		return nil, errors.New("qdrant endpoint is required")
	}
	// Parse host and optional port
	config := &sdk.Config{}
	hostOnly, portStr, splitErr := net.SplitHostPort(endpoint)
	if splitErr == nil {
		// endpoint contains port
		config.Host = hostOnly
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port in endpoint %q: %w", endpoint, err)
		}
		config.Port = port
	} else {
		config.Host = endpoint
	}
	// Always use TLS for secure connection
	config.UseTLS = true
	// Set API key if provided
	if apiKey != "" {
		config.APIKey = apiKey
	}
	grpcClient, err := sdk.NewGrpcClient(config)
	if err != nil {
		return nil, err
	}
	return &QdrantClient{grpcClient: grpcClient}, nil
}

// Close closes the gRPC connection
func (q *QdrantClient) Close() error {
	if q.grpcClient != nil {
		return q.grpcClient.Close()
	}
	return nil
}

// EnsureCollection ensures a collection exists with the specified parameters
func (q *QdrantClient) EnsureCollection(ctx context.Context, collectionName string, vectorSize uint64, distance sdk.Distance) error {
	collections := q.grpcClient.Collections()
	existsResp, err := collections.CollectionExists(ctx, &sdk.CollectionExistsRequest{
		CollectionName: collectionName,
	})
	if err != nil {
		return err
	}
	if existsResp.GetResult().GetExists() {
		return nil
	}
	_, err = collections.Create(ctx, &sdk.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: sdk.NewVectorsConfig(&sdk.VectorParams{
			Size:     vectorSize,
			Distance: distance,
		}),
	})
	return err
}

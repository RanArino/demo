package storage

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// QdrantClient wraps the Qdrant client with our specific needs
type QdrantClient struct {
	client      qdrant.QdrantClient
	collections qdrant.CollectionsClient
	conn        *grpc.ClientConn
}

// NewQdrantClient creates a new Qdrant client
func NewQdrantClient(endpoint string) (*QdrantClient, error) {
	// Create gRPC connection
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// Create Qdrant clients
	client := qdrant.NewQdrantClient(conn)
	collections := qdrant.NewCollectionsClient(conn)

	return &QdrantClient{
		client:      client,
		collections: collections,
		conn:        conn,
	}, nil
}

// Close closes the gRPC connection
func (q *QdrantClient) Close() error {
	return q.conn.Close()
}

// EnsureCollection ensures a collection exists with the specified parameters
func (q *QdrantClient) EnsureCollection(ctx context.Context, name string, vectorSize uint64) error {
	// Check if collection exists
	_, err := q.collections.Get(ctx, &qdrant.GetCollectionInfoRequest{
		CollectionName: name,
	})

	if err != nil {
		// Create collection if it doesn't exist
		_, err = q.collections.Create(ctx, &qdrant.CreateCollection{
			CollectionName: name,
			VectorsConfig: &qdrant.VectorsConfig{
				Config: &qdrant.VectorsConfig_Params{
					Params: &qdrant.VectorParams{
						Size:     vectorSize,
						Distance: qdrant.Distance_Cosine,
					},
				},
			},
		})
	}

	return err
}

package storage

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
	"github.com/ran/demo/backend-go/internal/core/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// QdrantClient wraps the Qdrant client with our specific needs
type QdrantClient struct {
	client      qdrant.QdrantClient
	collections qdrant.CollectionsClient
	points      qdrant.PointsClient
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
	points := qdrant.NewPointsClient(conn)

	return &QdrantClient{
		client:      client,
		collections: collections,
		points:      points,
		conn:        conn,
	}, nil
}

// Close closes the gRPC connection
func (q *QdrantClient) Close() error {
	return q.conn.Close()
}

// EnsureCollection ensures a collection exists with the specified parameters
func (q *QdrantClient) EnsureCollection(ctx context.Context, name string, vectorSize uint64, distance qdrant.Distance) error {
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
						Distance: distance,
					},
				},
			},
		})
	}

	return err
}

// UpsertPoints adds or updates points in a collection
func (q *QdrantClient) UpsertPoints(ctx context.Context, collectionName string, points []*qdrant.PointStruct) error {
	_, err := q.points.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})

	return err
}

// Search performs a similarity search in the collection
func (q *QdrantClient) Search(ctx context.Context, collectionName string, vector []float32, limit uint64, filter *qdrant.Filter) ([]*qdrant.ScoredPoint, error) {
	response, err := q.points.Search(ctx, &qdrant.SearchPoints{
		CollectionName: collectionName,
		Vector:         vector,
		Filter:         filter,
		Limit:          limit,
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

// DeletePoints removes points from a collection
func (q *QdrantClient) DeletePoints(ctx context.Context, collectionName string, pointIDs []string) error {
	pointIDsList := make([]*qdrant.PointId, len(pointIDs))
	for i, id := range pointIDs {
		pointIDsList[i] = &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: id,
			},
		}
	}

	_, err := q.points.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: pointIDsList,
				},
			},
		},
	})

	return err
}

// GetPoints retrieves points by their IDs
func (q *QdrantClient) GetPoints(ctx context.Context, collectionName string, pointIDs []string) ([]*qdrant.RetrievedPoint, error) {
	pointIDsList := make([]*qdrant.PointId, len(pointIDs))
	for i, id := range pointIDs {
		pointIDsList[i] = &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: id,
			},
		}
	}

	response, err := q.points.Get(ctx, &qdrant.GetPoints{
		CollectionName: collectionName,
		Ids:            pointIDsList,
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
		WithVectors: &qdrant.WithVectorsSelector{
			SelectorOptions: &qdrant.WithVectorsSelector_Enable{
				Enable: true,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

// CountPoints returns the total number of points in a collection
func (q *QdrantClient) CountPoints(ctx context.Context, collectionName string) (uint64, error) {
	response, err := q.points.Count(ctx, &qdrant.CountPoints{
		CollectionName: collectionName,
	})

	if err != nil {
		return 0, err
	}

	return response.Result.Count, nil
}

// Helper function to create a point from a document chunk
func CreatePointFromChunk(chunk *models.Chunk) (*qdrant.PointStruct, error) {
	// Create payload map
	payload := make(map[string]*qdrant.Value)

	// Add document ID (string)
	payload["document_id"] = qdrant.NewValueString(chunk.DocumentID)

	// Add chunk ID (string)
	payload["chunk_id"] = qdrant.NewValueString(chunk.ID)

	// Add text (string)
	payload["text"] = qdrant.NewValueString(chunk.Text)

	// Add index (integer)
	payload["index"] = qdrant.NewValueInt(int64(chunk.Index))

	// Add token count (integer)
	payload["token_count"] = qdrant.NewValueInt(int64(chunk.TokenCount))

	// Create point
	return &qdrant.PointStruct{
		Id: &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: chunk.ID,
			},
		},
		Vectors: &qdrant.Vectors{
			VectorsOptions: &qdrant.Vectors_Vector{
				Vector: &qdrant.Vector{
					Data: chunk.Embedding,
				},
			},
		},
		Payload: payload,
	}, nil
}

// Helper function to create a point from a summary
func CreatePointFromSummary(summary *models.Summary) (*qdrant.PointStruct, error) {
	// Create payload map
	payload := make(map[string]*qdrant.Value)

	// Add document ID (string)
	payload["document_id"] = qdrant.NewValueString(summary.DocumentID)

	// Add summary ID (string)
	payload["summary_id"] = qdrant.NewValueString(summary.ID)

	// Add text (string)
	payload["text"] = qdrant.NewValueString(summary.Text)

	// Add is_summary flag (boolean)
	payload["is_summary"] = qdrant.NewValueBool(true)

	// Create point
	return &qdrant.PointStruct{
		Id: &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: summary.ID,
			},
		},
		Vectors: &qdrant.Vectors{
			VectorsOptions: &qdrant.Vectors_Vector{
				Vector: &qdrant.Vector{
					Data: summary.Embedding,
				},
			},
		},
		Payload: payload,
	}, nil
}

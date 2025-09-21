package python

import (
	"context"
	"fmt"
	"sync"
	"time"

	privpb "demo/ms_canvas/go_app/api/proto/private/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"demo/ms_canvas/go_app/internal/config"
)

// GatewayFactory manages creation and lifecycle of Python service gateways
type GatewayFactory struct {
	config   config.Config
	gateways map[string]*Gateway
	mu       sync.RWMutex
}

// ServiceType represents different Python service types
type ServiceType string

const (
	// ChunkingEmbeddingService handles text chunking and embedding
	ChunkingEmbeddingService ServiceType = "chunking_embedding"
	// Future services can be added here:
	// SummarizationService ServiceType = "summarization"
	// ClusteringService    ServiceType = "clustering"
	// GraphBuilderService  ServiceType = "graph_builder"
)

// Gateway handles communication with Python services
type Gateway struct {
	client privpb.CanvasClient
	conn   *grpc.ClientConn
}

// Config holds configuration for the Python gateway
type Config struct {
	Host        string
	Port        int
	MaxMsgBytes int
}

// ServiceConfig holds configuration for specific service types
type ServiceConfig struct {
	Host        string
	Port        int
	MaxMsgBytes int
	Timeout     time.Duration
}

// NewGatewayFactory creates a new factory for managing Python service gateways
func NewGatewayFactory(config config.Config) *GatewayFactory {
	return &GatewayFactory{
		config:   config,
		gateways: make(map[string]*Gateway),
	}
}

// GetGateway returns a gateway for the specified service type, creating it if necessary
func (f *GatewayFactory) GetGateway(serviceType ServiceType) (*Gateway, error) {
	key := string(serviceType)

	f.mu.RLock()
	gateway, exists := f.gateways[key]
	f.mu.RUnlock()

	if exists {
		return gateway, nil
	}

	// Create new gateway with write lock
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check pattern
	if gateway, exists := f.gateways[key]; exists {
		return gateway, nil
	}

	// Create gateway based on service type
	gatewayConfig, err := f.getServiceConfig(serviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get config for service %s: %w", serviceType, err)
	}

	gateway, err = f.newGateway(gatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway for service %s: %w", serviceType, err)
	}

	f.gateways[key] = gateway
	return gateway, nil
}

// GetChunkingEmbeddingGateway is a convenience method for the chunking/embedding service
func (f *GatewayFactory) GetChunkingEmbeddingGateway() (*Gateway, error) {
	return f.GetGateway(ChunkingEmbeddingService)
}

// getServiceConfig returns configuration for a specific service type
func (f *GatewayFactory) getServiceConfig(serviceType ServiceType) (Config, error) {
	switch serviceType {
	case ChunkingEmbeddingService:
		return Config{
			Host:        f.config.PythonService.Host,
			Port:        f.config.PythonService.Port,
			MaxMsgBytes: f.config.PythonService.MaxMsgBytes,
		}, nil

	// Future service configurations:
	// case SummarizationService:
	//     return Config{
	//         Host:        f.config.PythonService.Host,
	//         Port:        f.config.PythonSummarizationService.Port, // Different port
	//         MaxMsgBytes: f.config.PythonService.MaxMsgBytes,
	//     }, nil

	default:
		return Config{}, fmt.Errorf("unknown service type: %s", serviceType)
	}
}

// HealthCheck performs health checks on all active gateways
func (f *GatewayFactory) HealthCheck(ctx context.Context) map[ServiceType]error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	results := make(map[ServiceType]error)

	for key, gateway := range f.gateways {
		serviceType := ServiceType(key)
		_, err := gateway.Healthz(ctx)
		results[serviceType] = err
	}

	return results
}

// Close closes all active gateways
func (f *GatewayFactory) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var lastErr error
	for key, gateway := range f.gateways {
		if err := gateway.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close gateway %s: %w", key, err)
		}
	}

	// Clear the map
	f.gateways = make(map[string]*Gateway)

	return lastErr
}

// GetActiveServices returns list of currently active service types
func (f *GatewayFactory) GetActiveServices() []ServiceType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	services := make([]ServiceType, 0, len(f.gateways))
	for key := range f.gateways {
		services = append(services, ServiceType(key))
	}

	return services
}

// Reconnect attempts to reconnect a specific service gateway
func (f *GatewayFactory) Reconnect(serviceType ServiceType) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := string(serviceType)

	// Close existing gateway if it exists
	if gateway, exists := f.gateways[key]; exists {
		_ = gateway.Close()
		delete(f.gateways, key)
	}

	// Create new gateway
	gatewayConfig, err := f.getServiceConfig(serviceType)
	if err != nil {
		return fmt.Errorf("failed to get config for service %s: %w", serviceType, err)
	}

	gateway, err := f.newGateway(gatewayConfig)
	if err != nil {
		return fmt.Errorf("failed to reconnect gateway for service %s: %w", serviceType, err)
	}

	f.gateways[key] = gateway
	return nil
}

// Example usage patterns for future services:

// Future: GetSummarizationGateway for LLM summarization service
// func (f *GatewayFactory) GetSummarizationGateway() (*SummarizationGateway, error) {
//     gateway, err := f.GetGateway(SummarizationService)
//     if err != nil {
//         return nil, err
//     }
//     return &SummarizationGateway{Gateway: gateway}, nil
// }

// Future: GetClusteringGateway for clustering service
// func (f *GatewayFactory) GetClusteringGateway() (*ClusteringGateway, error) {
//     gateway, err := f.GetGateway(ClusteringService)
//     if err != nil {
//         return nil, err
//     }
//     return &ClusteringGateway{Gateway: gateway}, nil
// }

// ===== CENTRALIZED GATEWAY CREATION AND MANAGEMENT =====

// newGateway creates a new gateway to the Python service - CENTRALIZED CREATION
func (f *GatewayFactory) newGateway(config Config) (*Gateway, error) {
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(config.MaxMsgBytes),
			grpc.MaxCallSendMsgSize(config.MaxMsgBytes),
		),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Python service at %s: %w", address, err)
	}

	client := privpb.NewCanvasClient(conn)

	return &Gateway{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes a gateway connection
func (g *Gateway) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

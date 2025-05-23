# Architecture Decision Report: Hierarchical Text Visualization System

**Document Version:** 1.0  
**Last Updated:** 2025-05-23  

## Executive Summary

This document outlines the architectural decision process and final design for the Hierarchical Text Visualization production system. After evaluating multiple architectural patterns, we have adopted a **Hybrid Shared-Database Distributed Services Architecture** with selective microservice extraction, supported by a **multi-protocol communication strategy**.

## 1. Initial Architecture Considerations

### 1.1 Microservices Architecture (Initial Consideration)

**Initial Appeal:**
- Service independence and scalability
- Technology diversity support (Go, Python, Next.js)
- Team autonomy and parallel development
- Modern industry practices

**Identified Risks and Challenges:**
```
❌ Network Latency Issues:
- Real-time 3D visualization requires <50ms response times
- Chat streaming would be interrupted by service boundaries
- Interactive search highlighting needs immediate data access

❌ Database Complexity:
- Document ↔ Space ↔ User relationships require ACID transactions
- Complex permission checks across multiple entities
- Vector metadata synchronization between services

❌ Operational Overhead:
- Multiple databases to manage and backup
- Service discovery and coordination complexity
- Distributed debugging and monitoring challenges

❌ Development Constraints:
- Limited microservices experience in team
- Small team size vs. operational complexity
- Need for rapid iteration and feature development
```

### 1.2 Data Relationship Analysis

Our system exhibits tightly coupled data relationships that challenged traditional microservice boundaries:

```sql
-- Complex multi-entity relationships requiring transactions
documents ←→ document_space_assignments ←→ spaces
    ↓                                         ↓
document_content                         space_members
    ↓                                         ↓
document_permissions ←← users →→ user_permissions
    ↓
vector_metadata ←→ qdrant_vectors
    ↓
chat_messages ←→ chat_context_links
```

**Key Insight:** Traditional microservice data separation would introduce significant network overhead and consistency challenges for our use cases.

## 2. Alternative Architecture Evaluation

### 2.1 Modular Monolith (Proposed Solution)

**Architecture Characteristics:**
- Single application with well-defined internal modules
- Shared database for all business entities
- Clear service boundaries within the monolith
- External services only for specialized technologies

**Benefits:**
- ✅ ACID transactions across all entities
- ✅ Low latency for real-time features
- ✅ Simple development and debugging
- ✅ Single deployment unit

**Limitations Identified:**
- ❌ Limited service isolation
- ❌ Scaling constraints (scale entire application)
- ❌ Technology stack limitations
- ❌ Team boundary enforcement challenges

### 2.2 Requirements for Service Separation

**Scalability Requirements:**
- ML processing needs GPU resources
- Chat service requires different scaling patterns
- Document processing has burst workloads
- Visualization generation is compute-intensive

**Management Requirements:**
- Independent deployment cycles
- Service-specific monitoring and alerting
- Technology stack flexibility
- Team ownership boundaries

## 3. Final Architecture Decision

### 3.1 Hybrid Shared-Database Distributed Services

**Core Principle:** Separate services for operational flexibility while maintaining shared data access for consistency.

```
Architecture Components:

🔄 Shared-Database Services (Uncertain Boundaries):
├── Document Service       # Document processing and metadata
├── Space Service         # Space management and organization  
├── Chat Service          # RAG chat and conversation management
├── Search Service        # Vector search and indexing
├── Visualization Service # Graph layout and coordinate generation
└── User Profile Service  # User preferences and settings

🔒 Dedicated Microservices (Clear Boundaries):
├── Authentication Service # User auth, sessions, tokens
├── Notification Service   # Email, push, in-app notifications
├── File Storage Service   # S3 integration and file management
└── Audit Service         # Logging and compliance tracking

🌐 External Services:
├── ML Service (Python)   # Embeddings, clustering, dimensionality reduction
├── Qdrant Vector Store   # Vector similarity search
├── Gemini API           # LLM text generation
└── AWS Infrastructure   # S3, SQS, monitoring
```

### 3.2 Service Boundary Decision Matrix

```go
type ServiceClassification struct {
    // Immediate Microservice Criteria
    WellUnderstoodDomain     bool  // Auth, storage, notifications
    RegulatoryRequirements   bool  // Security isolation needs
    IndependentBusinessValue bool  // Can be monetized separately
    ClearDataOwnership      bool  // No foreign key dependencies
    
    // Shared-Database Service Criteria  
    UncertainBoundaries     bool  // Document vs Space ownership unclear
    TightDataCoupling       bool  // Complex cross-entity transactions
    FrequentDataAccess      bool  // Real-time features requiring low latency
    SharedBusinessLogic     bool  // Overlapping business rules
}

// Decision Logic:
func ClassifyService(name string, criteria ServiceClassification) ServiceType {
    if criteria.WellUnderstoodDomain && 
       criteria.ClearDataOwnership && 
       !criteria.TightDataCoupling {
        return MICROSERVICE
    }
    
    if criteria.UncertainBoundaries || 
       criteria.TightDataCoupling || 
       criteria.FrequentDataAccess {
        return SHARED_DATABASE_SERVICE
    }
    
    return EVALUATE_LATER
}
```

### 3.3 Architecture Evolution Strategy

**Progressive Service Extraction:**
```
Phase 1: Hybrid Launch
├── Start with immediate microservices (auth, notifications)
├── Deploy shared-database services for core business logic
└── Monitor service boundaries and data access patterns

Phase 2: Boundary Emergence
├── Analyze actual data access patterns
├── Identify clear service ownership
├── Monitor scaling and team boundary formation
└── Measure extraction criteria satisfaction

Phase 3: Selective Extraction
├── Extract services meeting microservice criteria
├── Migrate dedicated data models
├── Implement service-to-service communication
└── Maintain shared database for remaining services
```

**Extraction Triggers:**
```go
func ShouldExtractService(serviceName string) bool {
    return (HasClearDataOwnership(serviceName) && 
            HasDedicatedTeam(serviceName)) ||
           (RequiresDifferentScaling(serviceName) && 
            HasStableInterfaces(serviceName)) ||
           (RegulatoryIsolationRequired(serviceName))
}
```

## 4. Communication Strategy

### 4.1 Multi-Protocol Approach

Our hybrid architecture employs different communication protocols optimized for specific interaction patterns:

```
Frontend ↔ Backend:     REST + WebSockets
Service ↔ Service:      gRPC  
Backend ↔ External:     REST
```

### 4.2 Protocol Selection Rationale

#### 4.2.1 Frontend Communication: REST + WebSockets

**REST for CRUD Operations:**
```typescript
// Standard HTTP operations
GET    /api/v1/spaces/{id}
POST   /api/v1/documents/upload  
PUT    /api/v1/spaces/{id}/settings
DELETE /api/v1/documents/{id}
```

**WebSockets for Real-time Features:**
```typescript
// Real-time streaming connections
ws://api/ws/chat/{space_id}           // Streaming chat responses
ws://api/ws/visualization/{space_id}  // Live graph updates  
ws://api/ws/processing/{doc_id}       // Document processing status
```

**Benefits:**
- ✅ Native browser support, no additional libraries
- ✅ Excellent debugging tools (browser DevTools, Postman)
- ✅ HTTP caching for static data
- ✅ Real-time capabilities via WebSockets
- ✅ Simple error handling and retry logic

#### 4.2.2 Service-to-Service Communication: gRPC

**Protocol Buffer Definitions:**
```protobuf
service DocumentService {
  rpc ProcessDocument(ProcessDocumentRequest) returns (ProcessDocumentResponse);
  rpc GetRelevantDocuments(SearchRequest) returns (DocumentList);
  rpc StreamProcessingStatus(StatusRequest) returns (stream ProcessingUpdate);
}

service ChatService {
  rpc GenerateResponse(ChatRequest) returns (stream ChatChunk);
  rpc GetChatHistory(HistoryRequest) returns (ChatHistory);
}
```

**Benefits:**
- ✅ High performance: Binary protocol with HTTP/2 multiplexing
- ✅ Type safety: Protocol Buffers prevent API contract violations
- ✅ Streaming support: Perfect for chat and processing updates
- ✅ Code generation: Automatic client/server stub generation
- ✅ Observability: Built-in metrics and tracing support

#### 4.2.3 External Service Communication: REST

**External Service Integration:**
```go
type ExternalClients struct {
    MLService    *http.Client      // Python ML service REST API
    QdrantClient *qdrant.Client    // Qdrant Go client (REST wrapper)
    GeminiClient *genai.Client     // Google Gemini API client
    S3Client     *s3.Client        // AWS S3 SDK
}
```

**Benefits:**
- ✅ Standard protocol support across all external services
- ✅ Extensive tooling and debugging support
- ✅ Wide library availability in all languages
- ✅ Simple integration with cloud services

### 4.3 Communication Architecture Implementation

```go
// API Gateway orchestrating multiple protocols
type APIGateway struct {
    // HTTP/REST server
    restServer *gin.Engine
    
    // WebSocket hubs for real-time features
    chatHub          *websocket.Hub
    visualizationHub *websocket.Hub
    processingHub    *websocket.Hub
    
    // gRPC clients for internal services
    documentClient    pb.DocumentServiceClient
    chatClient        pb.ChatServiceClient  
    visualClient      pb.VisualizationServiceClient
    searchClient      pb.SearchServiceClient
    
    // REST clients for external services
    mlClient       *MLServiceClient
    storageClient  *StorageServiceClient
}

// Example: Chat request flow using multiple protocols
func (gw *APIGateway) HandleChatMessage(ws *websocket.Conn, msg ChatMessage) error {
    // 1. Get relevant documents via gRPC
    docs, err := gw.documentClient.GetRelevantDocuments(context.Background(), 
        &pb.SearchRequest{
            SpaceId: msg.SpaceID,
            Query:   msg.Content,
        })
    
    // 2. Get vector embeddings via gRPC  
    vectors, err := gw.searchClient.VectorSearch(context.Background(),
        &pb.VectorSearchRequest{
            Query: msg.Content,
            TopK:  10,
        })
    
    // 3. Generate response via external REST API
    response, err := gw.mlClient.GenerateResponse(context.Background(), MLRequest{
        Query:     msg.Content,
        Context:   docs,
        Vectors:   vectors,
    })
    
    // 4. Stream response back via WebSocket
    return gw.streamChatResponse(ws, response)
}
```

## 5. Implementation Architecture

### 5.1 Repository Structure

```
demo-production/
├── cmd/                          # Service entry points
│   ├── orchestrator/             # Main API gateway
│   ├── auth-service/             # Authentication microservice
│   ├── notification-service/     # Notification microservice
│   ├── document-service/         # Document processing service
│   ├── chat-service/             # Chat/RAG service
│   └── visualization-service/    # Graph generation service
│
├── internal/                     # Shared internal packages
│   ├── orchestrator/             # API gateway logic
│   │   ├── handlers/             # HTTP/WebSocket handlers
│   │   ├── grpc_clients/         # gRPC client implementations
│   │   └── websocket/            # WebSocket hub management
│   │
│   ├── services/                 # Business logic implementations
│   │   ├── document/             # Document service logic
│   │   ├── chat/                 # Chat service logic
│   │   ├── visualization/        # Visualization logic
│   │   └── shared/               # Common business utilities
│   │
│   ├── repository/               # Data access layer
│   │   ├── document_repo.go      # Document database operations
│   │   ├── space_repo.go         # Space database operations
│   │   ├── chat_repo.go          # Chat database operations
│   │   └── vector_repo.go        # Vector metadata operations
│   │
│   └── clients/                  # External service clients
│       ├── ml_client.go          # ML service REST client
│       ├── storage_client.go     # S3 storage client
│       └── gemini_client.go      # Gemini API client
│
├── pkg/                          # Public packages
│   ├── api/                      # API definitions
│   │   ├── proto/                # gRPC Protocol Buffers
│   │   ├── rest/                 # REST API schemas
│   │   └── websocket/            # WebSocket message types
│   │
│   ├── models/                   # Shared data models
│   └── database/                 # Database utilities
│
├── deployments/                  # Deployment configurations
│   ├── docker-compose.full.yml   # All services distributed
│   ├── docker-compose.minimal.yml# Orchestrator only
│   └── kubernetes/               # K8s manifests
│
├── migrations/                   # Database schema migrations
│   ├── shared/                   # Shared database tables
│   └── auth/                     # Auth service dedicated tables
│
└── scripts/                      # Operational scripts
    ├── deploy-service.sh         # Deploy individual service
    ├── migrate-to-microservice.sh# Extract service to microservice
    └── monitor-boundaries.sh     # Monitor service extraction signals
```

### 5.2 Database Architecture

```sql
-- Shared Core Database (PostgreSQL)
CREATE DATABASE demo_core;

-- Core business entities (shared across services)
CREATE TABLE spaces (...);
CREATE TABLE documents (...);
CREATE TABLE document_space_assignments (...);
CREATE TABLE document_content (...);
CREATE TABLE chat_messages (...);
CREATE TABLE chat_sessions (...);
CREATE TABLE vector_metadata (...);
CREATE TABLE search_indexes (...);

-- Dedicated Microservice Databases
CREATE DATABASE auth_service;
CREATE DATABASE notification_service;

-- External Databases
-- Qdrant: Vector embeddings and similarity search
-- Redis: Caching and session storage
```

### 5.3 Deployment Strategy

```yaml
# docker-compose.full.yml - Production deployment
version: '3.8'
services:
  # API Gateway
  orchestrator:
    build: ./cmd/orchestrator
    ports: ["8080:8080"]
    environment:
      - SHARED_DB_URL=postgresql://user:pass@postgres:5432/demo_core
      - AUTH_SERVICE_URL=http://auth-service:8081
      - DOCUMENT_SERVICE_URL=http://document-service:8082
      - CHAT_SERVICE_URL=http://chat-service:8083
    depends_on: [postgres, auth-service]
    
  # Microservices (dedicated databases)  
  auth-service:
    build: ./cmd/auth-service
    environment:
      - AUTH_DB_URL=postgresql://user:pass@postgres:5432/auth_service
    depends_on: [postgres]
    
  notification-service:
    build: ./cmd/notification-service  
    environment:
      - NOTIFICATION_DB_URL=postgresql://user:pass@postgres:5432/notification_service
    depends_on: [postgres]
    
  # Shared-database services
  document-service:
    build: ./cmd/document-service
    environment:
      - SHARED_DB_URL=postgresql://user:pass@postgres:5432/demo_core
      - QDRANT_URL=http://qdrant:6333
    depends_on: [postgres, qdrant]
    
  chat-service:
    build: ./cmd/chat-service
    environment:
      - SHARED_DB_URL=postgresql://user:pass@postgres:5432/demo_core
      - QDRANT_URL=http://qdrant:6333
      - ML_SERVICE_URL=http://ml-service:8084
    depends_on: [postgres, qdrant, ml-service]
    
  visualization-service:
    build: ./cmd/visualization-service
    environment:
      - SHARED_DB_URL=postgresql://user:pass@postgres:5432/demo_core
      - QDRANT_URL=http://qdrant:6333
    depends_on: [postgres, qdrant]
    
  # External services
  ml-service:
    build: ./ml-service
    environment:
      - QDRANT_URL=http://qdrant:6333
    depends_on: [qdrant]
    
  # Infrastructure
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: demo_core
      POSTGRES_USER: user  
      POSTGRES_PASSWORD: pass
    volumes: [postgres_data:/var/lib/postgresql/data]
    
  qdrant:
    image: qdrant/qdrant
    volumes: [qdrant_data:/qdrant/storage]
    
  redis:
    image: redis:7
    volumes: [redis_data:/data]
```

## 6. Benefits and Trade-offs

### 6.1 Architecture Benefits

**Performance Optimization:**
```
✅ Low Latency: Shared database eliminates network calls for data access
✅ Real-time Features: WebSocket streaming without service boundaries
✅ Efficient Communication: gRPC for high-frequency service calls
✅ Optimized Protocols: Right protocol for each interaction pattern
```

**Development Experience:**
```
✅ Rapid Development: Shared types and database schemas
✅ Easy Debugging: Single database for transaction tracing
✅ Simple Testing: Can run entire system locally
✅ Flexible Deployment: Services can be combined or separated
```

**Operational Simplicity:**
```
✅ Single Database: Simplified backup, monitoring, and maintenance
✅ Gradual Evolution: Can extract microservices when boundaries clarify
✅ Cost Efficiency: No premature infrastructure complexity
✅ Risk Mitigation: Proven patterns with escape hatches
```

### 6.2 Trade-offs and Limitations

**Potential Constraints:**
```
⚠️ Database Bottleneck: Shared database may become scaling constraint
⚠️ Schema Coordination: Changes require coordination across services
⚠️ Technology Lock-in: Services must use compatible database drivers
⚠️ Deployment Coupling: Some schema changes affect multiple services
```

**Mitigation Strategies:**
```
✅ Database Optimization: Connection pooling, read replicas, caching
✅ Schema Versioning: Backward-compatible migration strategies
✅ Service Extraction: Clear criteria for microservice promotion
✅ Monitoring: Track extraction signals and performance metrics
```

## 7. Success Metrics and Evolution Criteria

### 7.1 Performance Targets

```
Real-time Features:
├── 3D Visualization Response: < 50ms
├── Chat Message Streaming: < 100ms first token
├── Search Result Highlighting: < 200ms
└── Document Processing Updates: < 500ms

System Reliability:
├── API Gateway Uptime: 99.9%
├── Database Availability: 99.95%
├── Service Communication: < 1% error rate
└── Data Consistency: Zero data loss tolerance
```

### 7.2 Microservice Extraction Criteria

```go
// Automated monitoring for extraction signals
type ExtractionMetrics struct {
    DataOwnershipClarity    float64  // % of tables owned by single service
    ScalingDivergence      float64  // Coefficient of variation in resource usage
    TeamBoundaryFormation  bool     // Dedicated team assignment
    CrossServiceQueries    int      // Number of cross-service data queries
    DatabaseContention     float64  // Lock wait time percentage
}

func MonitorExtractionSignals() {
    metrics := collectMetrics()
    
    if metrics.DataOwnershipClarity > 0.8 &&
       metrics.ScalingDivergence > 0.5 &&
       metrics.TeamBoundaryFormation &&
       metrics.CrossServiceQueries < 10 {
        
        recommendMicroserviceExtraction()
    }
}
```

## 8. Conclusion

The **Hybrid Shared-Database Distributed Services Architecture** provides an optimal balance for our hierarchical text visualization system by:

1. **Addressing Initial Concerns:** Eliminates network latency and database consistency issues while maintaining service separation capabilities

2. **Supporting Evolution:** Provides clear criteria and mechanisms for transitioning to microservices when boundaries clarify

3. **Optimizing Communication:** Uses appropriate protocols (REST, WebSocket, gRPC) for different interaction patterns

4. **Enabling Scalability:** Allows independent scaling of services while maintaining data consistency

5. **Facilitating Development:** Supports rapid iteration with clear service boundaries and shared infrastructure

This architecture acknowledges the uncertainty inherent in early-stage system design while providing concrete mechanisms for evolution based on real-world usage patterns and organizational growth.

The decision represents a pragmatic approach to modern software architecture that prioritizes **business value delivery** over **architectural purity**, while maintaining **clear evolution paths** for future growth.


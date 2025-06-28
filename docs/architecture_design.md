# Architecture Decision Report: Hierarchical Text Visualization System

**Document Version:** 3.0  
**Last Updated:** 2025-06-23    

## Executive Summary

This document outlines the architectural decision process and final design for the Hierarchical Text Visualization production system. After evaluating multiple architectural patterns, we have adopted a **Full Microservices Architecture** with service-specific databases, supported by a **multi-repository distribution strategy** and **event-driven communication** patterns.

## 1. Initial Architecture Considerations

### 1.1 Microservices Architecture (Initial Consideration)

**Initial Appeal:**
- Service independence and scalability
- Technology diversity support (Go, Python, Next.js)
- Team autonomy and parallel development
- Modern industry practices

**Identified Risks and Challenges:**
```
✅ Network Latency Solutions:
- RESTful API design with efficient caching strategies
- Event streaming for real-time updates (Apache Kafka)
- Intelligent caching layers and CDN integration
- gRPC for high-performance inter-service communication

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

### 2.2 Shared-Database Distributed Services (Selected Solution)

**Scalability Requirements:**
- ML processing needs GPU resources
- Chat service requires different scaling patterns
- Document processing has burst workloads
- Visualization generation is compute-intensive

**Strategic Benefits:**
- ACID transactions across all entities maintained
- Low latency for real-time features preserved
- Service boundaries established for future evolution
- Reduced operational overhead compared to full microservices
- Technology stack flexibility for specific services

**Managed Trade-offs:**
- Service coordination through shared database schema
- Careful API design to maintain service boundaries
- Gradual evolution path to full microservices when needed
- Shared infrastructure reduces operational complexity

## 3. Initial Architecture Decision: Shared-Database Distributed Services (Phase 1)

**Decision Rationale:** Based on the evaluation in section 2, we initially selected the Shared-Database Distributed Services approach to balance operational simplicity with service autonomy.

### 3.1 Multi-Repository Strategy with Shared Database

**Core Principle:** Distributed repository management with centralized orchestration and shared-database services for operational flexibility.

**Repository Distribution (Phase 1 Plan):**

**scaler (Central Hub Repository):**
- Orchestration & Deployment infrastructure
- API Gateway (Kong) configuration for public APIs
- Architecture documentation and service specifications
- Repository governance and feature implementation guidelines

**scaler-hub (Shared-Database Services):**
- Document Service for document processing and metadata
- Chat Service for RAG chat and conversation management  
- Vector Service for vector similarity search and indexing
- ML Service for document clustering, 2D/3D positioning, and model management
- Shared PostgreSQL database with coordinated schema management
- Common business logic and external service wrappers
- Internal gRPC APIs for service communication

**scaler-front (Frontend Repository):**
- Next.js application with RESTful API integration
- 3D visualization components and real-time chat interface
- WebSocket management for real-time features

**Individual Microservices (Selective Extraction):**
- scaler-auth for user authentication and sessions
- scaler-audit for logging and compliance tracking
- scaler-notifications for messaging services

### 3.2 Database Strategy (Phase 1)

**Shared Database Approach:**
- Central PostgreSQL database for core business entities
- Coordinated schema management across services
- ACID transactions maintained across service boundaries
- Selective polyglot persistence for specialized needs (DynamoDB for Canvas, ClickHouse for Activity)

### 3.3 Communication Strategy (Phase 1)

**Hybrid Communication Patterns:**
- RESTful APIs for public endpoints via Kong gateway
- gRPC for internal service-to-service communication
- WebSockets for real-time features
- Shared database access for complex queries requiring joins

## 4. Revised Final Architecture: Full Microservices (Phase 2)

**Architecture Evolution Rationale:**

During the database table development phase, detailed analysis revealed that the service boundaries were clearer than initially anticipated, and the data relationships could be effectively managed through event-driven patterns. This discovery, combined with the team's growing confidence in microservices patterns and the need for independent scaling, led to the decision to transition to a full microservices architecture.

**Key Factors Leading to Revision:**
- **Clear Service Boundaries**: Database design revealed natural domain boundaries with minimal cross-service dependencies
- **Independent Scaling Requirements**: Different services showed distinct scaling patterns (GPU for ML, high concurrency for chat, storage optimization for knowledge)
- **Technology Optimization**: Opportunity to use optimal technology stacks per service (Go for performance, Python for ML)
- **Team Growth**: Increased confidence in distributed systems patterns and operational capabilities
- **Event-Driven Maturity**: Better understanding of event streaming and eventual consistency patterns

### 4.1 Complete Microservices Architecture

**Core Business Microservices:**

```
🔍 Knowledge Service (scaler-knowledge):
- Document upload, processing, metadata management
- Content source management and blob storage
- Space creation, settings, and permissions
- Content extraction and indexing
- Dedicated PostgreSQL database

🎨 Canvas Service (scaler-canvas):
- Node positioning and visualization management
- Real-time collaboration state management
- Dedicated DynamoDB for nodes/edges + Redis for real-time state

💬 Chat Service (scaler-chat):
- RAG-based conversation management
- Context retrieval and response generation
- Chat history, sessions, and branching
- AI agent orchestration and workflow
- Dedicated PostgreSQL database + blob storage

🔎 Vector Service (scaler-vector):
- Vector similarity search and indexing
- Manages vector embeddings and their metadata
- Dedicated Qdrant vector database + Redis cache for performance

🔐 Authentication Service (scaler-auth):
- User registration, login, session management
- JWT token generation and validation
- OAuth integration and RBAC
- Permission evaluation and enforcement
- Dedicated PostgreSQL database

📊 Activity Service (scaler-activity):
- User behavior tracking and analytics
- System event logging and audit trails
- Engagement scoring and metrics
- Compliance and security monitoring
- Dedicated ClickHouse/TimescaleDB + Redis

📱 Notification Service (scaler-notifications):
- Email notifications and templates
- Push notifications and in-app messaging
- Notification preferences and delivery
- Event-driven notification triggers
- Dedicated PostgreSQL database + message queues

🤖 ML Service (scaler-ml):
- ML model training, versioning, and deployment
- Document clustering and 2D/3D coordinate calculation for visualization
- Embedding generation and batch processing
- Model serving and inference pipelines
- Performance monitoring and drift detection
- Dedicated PostgreSQL for metadata + S3 for models
```

### 4.2 API Gateway & Service Mesh Architecture

**RESTful API Gateway + Service Mesh Strategy:**

The platform employs a comprehensive API gateway strategy centered exclusively on RESTful APIs for public communication:

**Public API Gateway (Kong):**
- RESTful API endpoints for all public operations and CRUD functionality
- WebSocket gateway for real-time features and live updates
- Strict rate limiting, authentication, and authorization policies
- Enforced API versioning and comprehensive HTTP caching strategies
- Public endpoint exposure through a well-defined `/api/v1/*` path structure

**Service Mesh (Istio):**
Istio is selected as our service mesh to provide a comprehensive solution for connecting, securing, controlling, and observing our microservices. It was chosen over alternatives like Linkerd and Consul for its superior capabilities in key areas that align with our architecture.
- **Unmatched Traffic Management:** Granular traffic control for canary releases, A/B testing, and fault injection supports our independent deployment goals.
- **Deep Observability:** Automatically generates detailed metrics, logs, and distributed traces for all service-to-service traffic (including gRPC) without application code changes.
- **Kubernetes-Native Experience:** Its native integration with Kubernetes using CRDs provides a seamless workflow for our team.
- **Policy-Driven Security:** Enables a zero-trust network by default with automatic mTLS encryption and powerful, fine-grained authorization policies.
- **Resilience and Discovery:** Provides service discovery, intelligent load balancing, circuit breakers, and retry policies.

**Event Streaming (Apache Kafka):**
Apache Kafka is our chosen platform for event-driven communication, with a strong recommendation to use a managed service (e.g., AWS MSK, Confluent Cloud) to reduce operational complexity. Kafka's design as a durable, replayable commit log is perfectly aligned with our system's core requirements.
- **Event Sourcing & Replayability:** Kafka's persistent log is ideal for event sourcing, audit trails (`scaler-activity`), and replaying events to rebuild service state, supporting our "zero data loss" goal.
- **High Throughput & Scalability:** Horizontally scalable to handle massive event loads from services like `scaler-activity` and `scaler-ml`.
- **Durability & Reliability:** Provides strong durability guarantees with configurable replication, supporting dead-letter queues and retry mechanisms.
- **Rich Integration Ecosystem:** The mature ecosystem, especially **Kafka Connect**, allows for low-code integration with our diverse databases (PostgreSQL, ClickHouse, etc.), accelerating development.
- **Asynchronous Decoupling:** Ensures services remain loosely coupled, allowing teams to develop, deploy, and scale independently.

### 4.3 Microservice Allocation Framework

**Service Boundary Decision Criteria:**

The framework for determining service boundaries considers multiple dimensions:

**Domain Characteristics:**
- Business Capability: Core domain areas such as knowledge management, chat, search, visualization
- Data Ownership: Clear ownership of specific data entities and business logic
- Business Boundary: Well-defined functional boundaries with minimal overlap

**Technical Requirements:**
- Scaling Pattern: Different scaling needs (compute, storage, memory, GPU, concurrent users)
- Technology Stack: Optimization opportunities for specific languages (Go, Python, Node.js)
- Database Requirements: Optimal database choices (PostgreSQL, DynamoDB, ClickHouse, etc.)

**Operational Requirements:**
- Deployment Cadence: Independent release cycles and deployment flexibility
- Team Ownership: Dedicated team responsibility and accountability
- Monitoring Requirements: Service-specific SLAs and performance metrics

**Integration Patterns:**
- Event Publishing: Events that services publish to the system
- Event Consuming: Events that services consume from other services
- Synchronous API Requirements: Direct API dependencies between services

Service allocation follows a clear mapping to business capabilities, with fallback to creating new services for undefined domains.

## 5. RESTful + Event-Driven Communication Strategy

### 5.1 Multi-Protocol Communication Approach

Our microservices architecture employs layered communication patterns optimized for different interaction types:

**Frontend to API Gateway:**
- RESTful APIs as the primary interface for all operations
- WebSockets for real-time features and live updates

**Service to Service (Synchronous):**
- gRPC via Service Mesh (Istio) for high-performance internal communication

**Service to Service (Asynchronous):**
- Event Streaming via Kafka for loose coupling

**Service to External Systems:**
- REST APIs with circuit breakers and retry logic for resilience

### 5.2 Protocol Selection Rationale

#### 5.2.1 Frontend Communication: REST-First + WebSockets

**RESTful APIs for Standard Operations:**
RESTful APIs serve as the primary interface for all CRUD operations, providing clear resource-based endpoints for knowledge management, space operations, document handling, and search functionality. This approach offers excellent HTTP caching capabilities, broad client compatibility, and straightforward debugging through standard browser tools.

**WebSockets for Real-time Features:**
WebSocket connections handle streaming chat responses, live graph/canvas updates, and document processing status updates. This ensures low-latency real-time functionality while maintaining the simplicity of REST for standard operations.

**Benefits:**
- Native browser support with no additional client libraries required
- Excellent debugging tools available in browser DevTools and API testing tools
- Comprehensive HTTP caching for static data and improved performance
- Real-time capabilities via WebSockets for dynamic features
- Simple error handling and retry logic patterns

#### 5.2.2 Service-to-Service Communication: gRPC

**High-Performance Internal Communication:**
gRPC serves as the internal communication protocol between microservices, providing type-safe, high-performance binary communication with HTTP/2 multiplexing. Protocol Buffer definitions ensure API contract validation and enable automatic client/server code generation.

**Benefits:**
- High performance through binary protocol with HTTP/2 multiplexing
- Type safety through Protocol Buffers preventing API contract violations
- Built-in streaming support ideal for chat and processing updates
- Automatic code generation for client/server stubs
- Comprehensive observability with built-in metrics and tracing support

#### 5.2.3 External Service Communication: REST

**External Service Integration:**
All external service integrations use RESTful APIs enhanced with circuit breakers, retry logic, and comprehensive logging. Service-specific external client implementations handle integrations with vector databases, LLM services, cloud storage, and caching systems.

**Benefits:**
- Standard protocol support across all external services
- Extensive tooling and debugging support in the ecosystem
- Wide library availability across all programming languages
- Simple integration patterns with cloud services
- Enhanced observability through internal logging and analytics wrappers

## 6. Microservices Governance & Decision Guidelines

### 6.1 Service Allocation Decision Tree

**When implementing a new feature, follow this service allocation process:**

**Step 1: Frontend vs Backend Classification**
Determine if the feature is a UI/UX component or user interaction. If yes, implement in scaler-front. Otherwise, proceed to step 2.

**Step 2: Platform vs Business Logic Classification**
Identify if the feature involves platform infrastructure, orchestration, or cross-service tooling. If yes, implement in scaler (platform). Otherwise, proceed to step 3.

**Step 3: Business Domain Identification**
Map the feature to its primary business domain:
- Document/Content Management → scaler-knowledge
- Visualization/Hierarchical Graph/Tree → scaler-canvas  
- Conversation/AI → scaler-chat
- Search/Indexing → scaler-vector
- User Auth/Permissions → scaler-auth
- Analytics/Audit → scaler-activity
- Notifications/Messaging → scaler-notifications
- ML/Model Management → scaler-ml
- New Domain → Create new microservice

**Step 4: Cross-Cutting Concerns**
Handle cross-cutting concerns through established patterns:
- Shared Libraries for common functionality across services
- Events for asynchronous integration between services
- RESTful API composition for synchronous data aggregation

### 6.2 New Microservice Creation Guidelines

**Criteria for creating a new microservice:**

A new microservice should be created when a domain has distinct business capability, clear data ownership, and can operate independently. Additional factors that justify microservice creation include unique scaling requirements, specialized technology requirements, dedicated team ownership, or regulatory isolation needs.

**Examples of strong microservice candidates:**
- Payment processing requiring PCI compliance isolation
- Real-time collaboration engine with specialized WebSocket and event processing needs
- Video processing service requiring GPU resources and specialized technology stack
- Compliance/GDPR service needing regulatory isolation requirements

## 7. Success Metrics and Evolution Criteria

### 7.1 Performance Targets

**Real-time Features:**
- 3D Visualization Response: Less than 50ms for optimal user experience
- Chat Message Streaming: Less than 100ms for first token delivery
- Search Result Highlighting: Less than 200ms for responsive search
- Document Processing Updates: Less than 500ms for status notifications

**System Reliability:**
- API Gateway Uptime: 99.9% availability target
- Database Availability: 99.95% availability requirement
- Service Communication: Less than 1% error rate across all inter-service calls
- Data Consistency: Zero data loss tolerance with eventual consistency guarantees

### 7.2 Microservice Evolution Monitoring

**Key Evolution Indicators:**

The system monitors several metrics to identify when microservice boundaries should be reconsidered. Data ownership clarity measures the percentage of tables owned by a single service. Scaling divergence tracks the coefficient of variation in resource usage patterns across services. Team boundary formation indicates when dedicated team assignments emerge for specific domains.

Cross-service queries are counted to identify potential service boundary issues, while database contention is measured through lock wait time percentages. These metrics inform architectural evolution decisions and help maintain optimal service boundaries as the system grows and requirements change.

## 8. Conclusion

The **Full Microservices Architecture with Event-Driven Communication** provides a scalable and resilient foundation for our hierarchical text visualization system by:

1. **Service Autonomy:** Each microservice owns its data, deployment cycle, and technology stack
2. **Event-Driven Integration:** Asynchronous communication patterns ensure loose coupling and high availability
3. **Independent Scaling:** Services can scale based on specific demands (GPU for ML, concurrency for chat, storage for knowledge)
4. **Technology Optimization:** Go for performance-critical services, Python for ML, optimal databases per service
5. **Operational Excellence:** Service mesh, observability, and infrastructure automation enable reliable distributed operations

This architecture embraces the complexity of distributed systems while providing concrete patterns and frameworks to manage that complexity effectively:

- **RESTful API Composition** for efficient frontend data aggregation and cross-service integration
- **Event Sourcing and CQRS (Command Query Responsibility Segregation)** to ensure data consistency and create comprehensive audit trails by separating read and write operations.
- **Circuit Breakers and Retry Logic** for resilience against external service failures
- **Infrastructure as Code** for reproducible deployments and operational consistency
- **Service Mesh** for secure inter-service communication and observability

The decision represents a forward-looking approach to modern software architecture that prioritizes **horizontal scalability**, **technological flexibility**, and **team autonomy** while providing **robust operational patterns** for managing distributed system complexity.

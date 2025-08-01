# **Polyglot Backend Architecture Guide (Hybrid Model)**

This document outlines the architectural patterns for building backend services in **Go and Python**. It details how to support a modern Next.js frontend using a hybrid communication model: **gRPC** for service-to-service calls, **gRPC-web** for one-way client streams, and **WebSockets** for two-way interactive features, all orchestrated within an Istio service mesh.

## **Core Philosophy: The Right Tool for the Job**

Our architecture is built on a hybrid communication model:

* **gRPC-First Internally**: All internal service-to-service communication uses standard gRPC for maximum performance and type safety.  
* **Server Actions for Security**: The frontend's default communication pattern is Next.js Server Actions, which call backend gRPC services securely.  
* **Specialized Real-time Protocols**: We use the best tool for each real-time scenario:  
  * **gRPC-web** for efficient, one-way server-to-client data streams.  
  * **WebSockets** for complex, bi-directional client-server interaction.  
* **Istio-native**: Leverages Istio for traffic management, security, and observability.

## **When to Use Go vs Python**

**Go Services** \- Best for:

* High-performance, low-latency operations (e.g., User Service, API Gateway).  
* CPU-intensive computations.  
* Services requiring a minimal memory footprint.  
* Strong typing and compile-time safety.

**Python Services** \- Best for:

* AI/ML workloads and data processing (e.g., Knowledge Service, ML Model Serving).  
* Rapid prototyping and development.  
* Services with complex business logic.  
* Integration with Python-specific libraries (NumPy, Pandas, TensorFlow).

## **Project Structure**

### **Go Service Structure**

/go-service/  
├── api/proto/v1/        \# Protocol buffer definitions  
├── cmd/server/          \# Application entry point  
├── ent/                 \# Ent ORM schema and generated code  
├── internal/  
│   ├── config/          \# Configuration management  
│   ├── domain/          \# Business entities and interfaces  
│   ├── repository/      \# Data access layer (Ent-based)  
│   ├── service/         \# Business logic layer  
│   ├── server/          \# gRPC and WebSocket server handlers  
│   └── middleware/      \# Common middleware (e.g., auth)  
├── pkg/                 \# Public libraries  
├── deployments/         \# Docker, K8s configs  
└── go.mod

### **Python Service Structure**

/python-service/  
├── api/proto/v1/        \# Protocol buffer definitions  
├── app/  
│   ├── main.py          \# Application entry point  
│   ├── config/          \# Configuration management  
│   ├── domain/          \# Business entities and interfaces  
│   ├── repository/      \# Data access layer (SQLAlchemy/DynamoDB)  
│   ├── service/         \# Business logic layer  
│   ├── server/          \# gRPC and WebSocket server handlers  
│   └── middleware/      \# Common middleware (e.g., auth)  
├── tests/               \# Test files  
└── requirements.txt

## **Layer Responsibilities**

### **1\. API Layer (**api/proto/v1/**)**

* Define gRPC services and messages in .proto files.  
* Use Protocol Buffers for type-safe, language-agnostic API contracts.  
* This is the single source of truth for service interfaces.

### **2\. Domain Layer (**internal/domain/ **or** app/domain/**)**

* Contains core business entities and pure domain logic.  
* Defines repository interfaces (contracts) for data access.  
* Should have no dependencies on infrastructure (e.g., databases, gRPC).

### **3\. Repository Layer (**internal/repository/ **or** app/repository/**)**

* Implements the data access interfaces defined in the domain layer.  
* Handles all communication with the database using the chosen ORM.  
* Responsible for converting between database models and domain entities.

### **4\. Service Layer (**internal/service/ **or** app/service/**)**

* Implements the business logic and use cases.  
* Orchestrates domain entities and repository operations.  
* Handles business validation, rules, and transactions.

### **5\. Server Layer (**internal/server/ **or** app/server/**)**

* Implements the transport layer handlers (gRPC and WebSockets).  
* Maps incoming requests to the appropriate service layer calls.  
* Maps domain entities back to protobuf messages or WebSocket payloads.  
* Integrates authentication and authorization middleware.

## **Database Strategy**

### **Go Services**

* **Primary ORM**: Ent (entity framework).  
* **Schema Definition**: Define in ent/schema/ using Ent's type-safe Go code.  
* **Migrations**: Use Ent-generated migrations for schema evolution.  
* **Transactions**: Leverage Ent's built-in transaction support for data consistency.

### **Python Services**

* **Primary ORM**: SQLAlchemy 2.0+ with async support.  
* **Alternative**: DynamoDB with aioboto3 for NoSQL needs.  
* **Migrations**: Alembic for SQLAlchemy; manual or IaC for DynamoDB.  
* **Transactions**: Use SQLAlchemy's async transaction support.

## **Frontend-Backend Communication Patterns**

The backend must support three distinct frontend communication patterns.

### **1\. Request-Response via Server Actions**

* **Frontend Action**: A Next.js Server Action uses a native gRPC client (grpc-js) to make a standard RPC call.  
* **Backend Responsibility**: Expose a standard gRPC endpoint. This is the default for all CRUD operations and initial data loads.

### **2\. One-Way Streaming (Read-Only Views)**

* **Frontend Action**: The client uses a gRPC-web client to listen to a server-side stream.  
* **Backend Responsibility**: Expose a gRPC service that supports gRPC-web translation. This requires a proxy like **Envoy** or a server middleware. Ideal for pushing live data like stock tickers or read-only canvas updates.

### **3\. Two-Way Streaming (Interactive Features)**

* **Frontend Action**: The client opens a WebSocket connection for bi-directional communication.  
* **Backend Responsibility**: This requires a **new, dedicated WebSocket service**. This service handles the protocol upgrade, manages connection state, and broadcasts messages. Essential for collaborative canvases and chat.

## **Istio Service Mesh Integration**

* **gRPC & gRPC-web**: Use VirtualServices to route traffic. An EnvoyFilter or proxy configuration is necessary to enable gRPC-web to gRPC translation.  
* **WebSockets**: Use VirtualServices with WebSocket-specific configuration (websocket: true) to ensure proper routing of persistent connections.  
* **Security**: Enforce mTLS between all services for zero-trust security.  
* **Observability**: Leverage Istio for distributed tracing, metrics, and logging across all services.

## **Development Workflow**

1. **Define APIs**: Write .proto files for service contracts.  
2. **Generate Code**: Use protoc to generate language-specific gRPC code and grpc-web clients.  
3. **Implement Service**: Build domain, repository, service, and server layers. Choose the right communication pattern (gRPC, gRPC-web, WebSocket) for the server layer.  
4. **Configure Istio**: Set up VirtualServices, DestinationRules, and AuthorizationPolicies.  
5. **Deploy**: Services can be deployed independently regardless of language.

## **Tools and Libraries**

### **Go Tools**

* **gRPC**: google.golang.org/grpc  
* **gRPC-web Middleware**: github.com/improbable-eng/grpc-web  
* **WebSockets**: github.com/gorilla/websocket  
* **ORM**: entgo.io/ent  
* **Testing**: github.com/stretchr/testify  
* **Configuration**: github.com/spf13/viper

### **Python Tools**

* **gRPC**: grpcio, grpcio-tools  
* **gRPC-web Middleware**: grpcio-web  
* **WebSockets**: websockets library, or built-in support in frameworks like FastAPI.  
* **ORM**: sqlalchemy\[asyncio\], asyncpg  
* **NoSQL**: aioboto3 (DynamoDB)  
* **Migrations**: alembic  
* **Testing**: pytest, pytest-asyncio

## **Best Practices**

### **General Principles**

* Use interfaces for dependency injection.  
* Implement proper error handling with context.  
* Use context for cancellation and timeouts across service calls.  
* Keep functions small and focused on a single responsibility.  
* Strictly separate domain models from infrastructure concerns.

### **Go-Specific**

* Use Ent for type-safe database operations.  
* Separate domain models from Ent entities in the repository layer.  
* Use Ent's powerful query builder for complex queries.  
* Implement proper transaction handling with Ent's Tx client.

### **Python-Specific**

* Use type hints extensively for better code quality and documentation.  
* Implement proper async/await patterns for non-blocking I/O.  
* Use Pydantic or dataclasses for data validation and settings management.  
* Follow PEP 8 style guidelines.

### **Istio-Specific**

* Use appropriate timeout and retry policies in VirtualServices.  
* Implement circuit breakers for external or fragile dependencies.  
* Use AuthorizationPolicy to enforce fine-grained access control.  
* Leverage Istio's observability features for monitoring and debugging.

## **Code Generation**

### **Protocol Buffers**

\# Generate Go gRPC code  
protoc \--go\_out=. \--go\_opt=paths=source\_relative \\  
       \--go-grpc\_out=. \--go-grpc\_opt=paths=source\_relative \\  
       api/proto/v1/\*.proto

\# Generate Python gRPC code  
python \-m grpc\_tools.protoc \-I./api/proto/v1 \--python\_out=. \--grpc\_python\_out=. api/proto/v1/\*.proto

\# Generate gRPC-web client code for frontend  
protoc \-I./api/proto/v1 \--js\_out=import\_style=commonjs:frontend/src/api/generated/ \\  
       \--grpc-web\_out=import\_style=commonjs,mode=grpcwebtext:frontend/src/api/generated/ \\  
       api/proto/v1/\*.proto

### **Ent (Go)**

\# Generate Ent code from your schema definitions  
go generate ./ent

\# Create a migration file based on schema changes  
go run \-mod=mod entgo.io/ent/cmd/ent migrate diff  

# System Design and Architecture

**Document Version:** 1.0  
**Last Updated:** 2025-07-26  
**Target Audience:** Developers, DevOps Engineers, System Architects

## 1. Engineering Philosophy

- **Write human-understandable code.**
- **Let the code speak for itself.**
- **The core question of code review: "Can I maintain this code without any issues?"**
- **Embrace pure functions and log only what is necessary.**

## 2. Architecture Overview

This system uses a **Full Microservices Architecture** with service-specific databases, supported by a **multi-repository distribution strategy** and **event-driven communication** patterns.

### 2.1. Core Business Microservices

- **Knowledge Service (`ms_knowledge`):** Manages documents, content, spaces, and permissions.
- **Canvas Service (`ms_canvas`):** Handles node positioning and real-time collaboration.
- **Chat Service (`ms_chat`):** Manages RAG-based conversations and chat history.
- **Vector Service (`ms_vector`):** Provides vector similarity search and indexing.
- **Authentication Service (`ms_auth`):** Manages user authentication, sessions, and permissions.
- **Activity Service (`ms_activity`):** Tracks user behavior and system events.
- **Notification Service (`ms_notifications`):** Handles email, push, and in-app notifications.
- **ML Service (`ms_ml`):** Manages ML model training, versioning, and deployment.

### 2.2. Technology Choices

- **Frontend:** Next.js 15 with TypeScript and Server Actions.
- **Backend (Go):** High-performance services (e.g., User, Knowledge).
- **Backend (Python):** AI/ML workloads and data processing.
- **API Gateway:** Envoy Proxy for gRPC-Web translation.
- **Service Mesh:** Istio for traffic management, security, and observability.
- **Event Streaming:** Apache Kafka for asynchronous communication.
- **Databases:** PostgreSQL, DynamoDB, Redis, Qdrant, ClickHouse/TimescaleDB.

## 3. Communication Strategy

### 3.1. Frontend-Backend Communication

We use a hybrid model:

- **Next.js Server Actions (Default):** For secure, server-side data fetching and mutations using a server-side gRPC client.
- **gRPC-web (One-Way Streaming):** For efficient server-to-client data streams (e.g., read-only views).
- **WebSockets (Two-Way Streaming):** For interactive, bi-directional communication (e.g., collaborative editing, chat).

### 3.2. Service-to-Service Communication

- **gRPC:** For all internal, synchronous service-to-service communication.
- **Apache Kafka:** For asynchronous, event-driven communication between services.

## 4. Development Workflow

1. **Define APIs:** Write `.proto` files for service contracts.
2. **Generate Code:** Use `protoc` to generate language-specific gRPC code and gRPC-web clients.
3. **Implement Service:** Build domain, repository, service, and server layers.
4. **Configure Istio:** Set up `VirtualServices`, `DestinationRules`, and `AuthorizationPolicies`.
5. **Deploy:** Services can be deployed independently.

## 5. Frontend Development

- **Directory Structure:** Centralize all external communication logic under `src/api/`.
- **Server Actions:** Use for initial data loads and mutations.
- **Client-Side Streaming:** Use `gRPC-web` for one-way streams and `WebSockets` for two-way streams.
- **State Management:** Use a combination of server-side state and client-side state management libraries as needed.

## 6. Backend Development

- **Go Services:** Use Ent for type-safe database operations.
- **Python Services:** Use SQLAlchemy 2.0+ with async support.
- **Error Handling:** Implement proper error handling with context.
- **Testing:** Write unit and integration tests for all services.

## 7. gRPC-Web Implementation

- **Proxy:** Use Envoy to translate gRPC-web requests to standard gRPC.
- **Authentication:** Pass JWT tokens in the `authorization` header.
- **CORS:** Configure Envoy to handle CORS correctly.

## 8. Deployment

- **Orchestration:** Use Docker Compose for local development and Kubernetes for production.
- **CI/CD:** Implement a CI/CD pipeline to automate testing and deployment.

This document provides a high-level overview of the system architecture and design. For more detailed information, please refer to the specific guides in the `notes/dev/guides` directory.

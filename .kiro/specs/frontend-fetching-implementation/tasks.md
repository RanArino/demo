# Implementation Plan: Frontend Data Fetching from ms_canvas

> This document identifies and manages the development tasks to be performed based on the design document. Tasks are hierarchically divided and related requirements are clearly stated to ensure progress and traceability.

## Feature A: Backend gRPC Server Setup

### 1. CanvasPublic gRPC Server Implementation
> Implement the CanvasPublic gRPC server in ms_canvas to expose the canvas data API endpoints.

- [x] **1.1. Start gRPC server in main.go**
  > Add gRPC server startup code to cmd/main.go, register CanvasPublicServer, and expose on port 50054.
  >
  > **Related Requirements:** 1.1, 1.2, 1.3, 1.4
  >
  > **Modified Files:** `ms_canvas/go_app/cmd/main.go`

- [x] **1.2. Initialize required services**
  > Create instances of SearchService, NodeService, and LinkService with proper dependencies.
  >
  > **Related Requirements:** 1.1, 1.2, 1.3, 1.4
  >
  > **Modified Files:** `ms_canvas/go_app/cmd/main.go`

- [x] **1.3. Add graceful shutdown handling**
  > Implement graceful shutdown for gRPC server to ensure clean termination.
  >
  > **Related Requirements:** 1.5
  >
  > **Modified Files:** `ms_canvas/go_app/cmd/main.go`

### 2. Service Layer Dependencies
> Ensure all required services are properly initialized and wired together.

- [x] **2.1. Verify Neo4j connection and repositories**
  > Confirm NodeRepo, LinkRepo, and SearchRepo are working correctly.
  >
  > **Related Requirements:** 1.2, 1.3, 1.4
  >
  > **Modified Files:** `ms_canvas/go_app/cmd/main.go`

- [x] **2.2. Configure Python ML gateway**
  > Ensure Python service is accessible for semantic search operations.
  >
  > **Related Requirements:** 1.3
  >
  > **Verification:** Existing pythonGateway configuration confirmed working

- [x] **2.3. Test gRPC server health endpoints**
  > Add health check endpoints and verify server responds correctly.
  >
  > **Related Requirements:** 1.5
  >
  > **Verification:** gRPC server startup code tested, HTTP health endpoints on port 8080, gRPC on port 50054

## Feature B: Frontend gRPC Client (Server Actions)

### 3. Canvas Service gRPC Setup
> Create the frontend gRPC client to communicate directly with ms_canvas service, following the same pattern as ms_knowledge and ms_user.

- [x] **3.1. Add MS_CANVAS_GRPC_URL environment variable**
  > Configure the direct gRPC URL for ms_canvas service in frontend environment variables (localhost:50054).
  >
  > **Related Requirements:** 1.1
  >
  > **Modified Files:** `docker-compose.yml`
  >
  > **Verification:** Added MS_CANVAS_GRPC_URL_INTERNAL=ms_canvas:50054 to frontend environment and fixed port mapping from 50054:50051 to 50054:50054

- [x] **3.2. Implement gRPC client for canvas**
  > Implement `canvasClient.ts` with gRPC transport that connects directly to ms_canvas service (GetNodes, SemanticSearch, SearchNodes).
  >
  > **Related Requirements:** 1.2, 1.3, 1.4, 1.5
  >
  > **Modified Files:** `frontend/src/api/server-client.ts`
  >
  > **Verification:** Added CanvasClientManager singleton with getCanvasInstance() method following same pattern as ms_user and ms_knowledge services

- [x] **3.3. Add authentication to gRPC calls**
  > Ensure Clerk tokens are included in gRPC metadata for authentication.
  >
  > **Related Requirements:** 1.5
  >
  > **Modified Files:** `frontend/src/api/actions/canvasActions.ts`
  >
  > **Verification:** All canvas server actions use createAuthHeaders() utility to include Clerk JWT tokens with proper error handling

### 4. Server Actions for Canvas Operations
> Server Actions call ms_canvas gRPC endpoints directly (no additional API routes required).

- [x] **4.1. Create canvasActions.ts with getNodes action**
  > Implement a server action to fetch nodes by IDs using gRPC client with dual-layer caching and auth.
  >
  > **Related Requirements:** 1.2, 1.5
  >
  > **Modified Files:** `frontend/src/api/actions/canvasActions.ts`
  >
  > **Verification:** Implemented getNodes() server action with GetNodesRequest protobuf, dual-layer caching (React cache + Next.js unstable_cache), authentication, input validation, and error handling

- [x] **4.2. Add semanticSearch action**
  > Implement server action for semantic search using gRPC client with proper validation and error handling.
  >
  > **Related Requirements:** 1.3, 1.6
  >
  > **Modified Files:** `frontend/src/api/actions/canvasActions.ts`
  >
  > **Verification:** Implemented semanticSearch() server action with SemanticSearchRequest protobuf, query/spaceId validation, topK and nodeTypes parameters, score-based results, and comprehensive error handling

- [x] **4.3. Add searchNodes action**
  > Implement server action for spatial and filter-based search using gRPC client, keeping parity with backend.
  >
  > **Related Requirements:** 1.4
  >
  > **Modified Files:** `frontend/src/api/actions/canvasActions.ts`
  >
  > **Verification:** Implemented searchNodes() server action with SearchNodesRequest protobuf, flexible NodeFilter and SpatialBoundingBox construction, limit parameter, and full backend parity

### Additional Implementation: Test Page
> Created comprehensive test page for validating all canvas server actions.

- [x] **Canvas Test Page**
  > Simple UI to test getNodes, semanticSearch, and searchNodes functionality with proper error handling and results display.
  >
  > **Modified Files:** `frontend/src/app/canvas-test/page.tsx`
  >
  > **Verification:** Created `/canvas-test` route with complete UI for testing all three server actions, input validation, error states, loading indicators, and formatted results display with quick stats

## Feature C: Frontend Component Integration

### 5. Canvas Data Fetching Components
> Create React components that use the canvas API to display and interact with data.

- [ ] **5.1. Create NodeDisplay component**
  > Build a component to display individual node information and metadata.
  >
  > **Related Requirements:** 1.2, 1.6

- [ ] **5.2. Create SearchResults component**
  > Build a component to display semantic search results with relevance scores.
  >
  > **Related Requirements:** 1.3, 1.6

- [ ] **5.3. Create SearchInterface component**
  > Build a user interface for performing searches with filters and options.
  >
  > **Related Requirements:** 1.3, 1.4

### 6. Error Handling and Loading States
> Implement comprehensive error handling and user feedback mechanisms.

- [ ] **6.1. Add loading states to components**
  > Implement loading spinners and disabled states during API calls.
  >
  > **Related Requirements:** 1.5, 1.6

- [ ] **6.2. Add error boundaries and fallbacks**
  > Implement error boundaries and user-friendly error messages.
  >
  > **Related Requirements:** 1.5

- [ ] **6.3. Add retry functionality**
  > Implement retry buttons and automatic retry for failed requests.
  >
  > **Related Requirements:** 1.5

## Feature D: Testing and Validation

### 7. Backend Testing
> Ensure backend gRPC server functionality is thoroughly tested.

- [ ] **7.1. Unit tests for CanvasPublicServer**
  > Write unit tests for all gRPC methods with proper mocking.
  >
  > **Related Requirements:** 1.7

- [ ] **7.2. Integration tests with Neo4j**
  > Test database integration and data retrieval functionality.
  >
  > **Related Requirements:** 1.2, 1.3, 1.4, 1.7

- [ ] **7.3. Load testing for performance**
  > Test gRPC server performance under various loads.
  >
  > **Related Requirements:** 1.6

### 8. Frontend Testing
> Ensure frontend integration works correctly and handles edge cases.

- [ ] **8.1. Unit tests for canvas client**
  > Test client initialization, error handling, and retry logic.
  >
  > **Related Requirements:** 1.1, 1.5, 1.7

- [ ] **8.2. Integration tests for Server Actions**
  > Test Server Actions with mocked gRPC client responses.
  >
  > **Related Requirements:** 1.2, 1.3, 1.4, 1.7

- [ ] **8.3. End-to-end testing**
  > Test complete flow from frontend component to backend service.
  >
  > **Related Requirements:** 1.7

## Feature E: Configuration and Deployment

### 9. Environment Configuration
> Set up proper configuration for different environments.

- [ ] **9.1. Add MS_CANVAS_GRPC_URL environment variable**
  > Provide direct gRPC URL for ms_canvas service in frontend env (localhost:50054 for local, service URL for prod).
  >
  > **Related Requirements:** 1.1

- [ ] **9.2. Service discovery configuration**
  > Configure service discovery for ms_canvas in different environments (development, staging, production).
  >
  > **Related Requirements:** 1.1

- [ ] **9.3. Authentication configuration**
  > Ensure Clerk authentication is properly configured for gRPC calls.
  >
  > **Related Requirements:** 1.5

### 10. Documentation and Monitoring
> Create documentation and monitoring for the implementation.

- [ ] **10.1. API documentation**
  > Document all canvas API endpoints and usage examples.
  >
  > **Related Requirements:** 1.7

- [ ] **10.2. Setup monitoring and logging**
  > Add monitoring for gRPC calls and error tracking.
  >
  > **Related Requirements:** 1.5, 1.6

- [ ] **10.3. Create usage examples**
  > Provide code examples for common canvas operations.
  >
  > **Related Requirements:** 1.2, 1.3, 1.4

## General Tasks

### 11. Code Quality and Standards

- [ ] **11.1. Code review and linting**
  > Ensure all code follows project standards and passes linting.
  >
  > **Related Requirements:** 1.6

- [ ] **11.2. Performance optimization**
  > Optimize gRPC calls, caching, and component rendering.
  >
  > **Related Requirements:** 1.6

- [ ] **11.3. Security review**
  > Review authentication, CORS, and data access security.
  >
  > **Related Requirements:** 1.5

### 12. Final Integration and Testing

- [ ] **12.1. End-to-end integration test**
  > Test complete system integration from frontend to backend.
  >
  > **Related Requirements:** 1.7

- [ ] **12.2. User acceptance testing**
  > Validate that the implementation meets all requirements.
  >
  > **Related Requirements:** 1.2, 1.3, 1.4

- [ ] **12.3. Production deployment preparation**
  > Prepare deployment scripts and configurations.
  >
  > **Related Requirements:** 1.1, 1.6


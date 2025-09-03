# Authentication & Authorization Architecture Refactor Plan

Updated Date: 2025-09-03

Please again which approach is better; discussed in chat https://g.co/gemini/share/9bf5bd83a868
- give all user-related info (id, role, subscription-plan, etc.) to the Clerk's private metadata, which adds to custom JWT
- implement API Gateway pattern to validate Clerk JWT and revise to internal JWT, which includes all user-related info (id, role, subscription-plan, etc.)

## Current Situation Analysis

### Existing Architecture Problems

**1. Distributed Authentication Logic**
- Each microservice (ms_user, ms_knowledge) implements its own Clerk JWT validation
- Duplicated authentication code across services
- Inconsistent error handling and token validation logic

**2. Tight Service Coupling**
- ms_knowledge makes gRPC calls to ms_user for every ownership check
- Complex RBAC logic scattered in business logic methods
- Services are tightly coupled through authentication dependencies

**3. Security & Maintenance Issues**
- Multiple points of failure for authentication
- Difficult to update authentication logic consistently
- Each service needs Clerk secret keys
- Complex debugging when auth issues occur

### Current Flow Analysis

**User Registration Flow:**
1. User signs up via Clerk UI → JWT returned to frontend
2. Clerk sends webhook to ms_user → Creates user record with PENDING status
3. Frontend calls ms_user.ActivateUser → Updates profile, sets role in Clerk metadata
4. User can now access system with role-enabled JWT

**Authentication Flow:**
1. Frontend sends requests with Clerk JWT
2. Each service validates JWT independently using Clerk JWKS
3. Services extract user_id (clerk_id) and role from JWT
4. For ownership checks, services call ms_user.GetUser() via gRPC
5. Business logic performs authorization checks

## Proposed Refactor: API Gateway Pattern

### Architecture Goals

**1. Centralized Authentication**
- Single point of JWT validation at API Gateway
- Eliminate duplicated auth logic across services
- Consistent error handling and security policies

**2. Token Transformation**
- Convert external Clerk JWTs to internal JWTs at gateway
- Internal JWTs contain resolved user_id (UUID) and role
- Downstream services trust internal tokens without external calls

**3. Service Decoupling**
- Remove ms_user gRPC calls from business logic
- Services focus purely on business operations
- Clean separation of concerns

### Implementation Phases

## Phase 1: User Registration & Data Sync (No Changes Needed)

**Current webhook flow works well:**
- Clerk webhook → ms_user creates user record
- One-way data flow from Clerk to internal system
- No complex bidirectional dependencies

**Key Design Principle:** Keep this simple, unidirectional flow.

## Phase 2: API Gateway Implementation

### 2.1 Gateway Authentication Middleware

**Token Validation Flow:**
1. Client sends request with `Authorization: Bearer <clerk_jwt>`
2. Gateway validates Clerk JWT signature and expiry using Clerk JWKS
3. Gateway extracts `clerk_id` from JWT claims
4. Gateway queries ms_user database to resolve:
   - Internal `user_id` (UUID)
   - User `role` (user/admin)
   - User status (active/pending/deleted)

### 2.2 Internal JWT Generation

**Internal Token Structure:**
```json
{
  "sub": "internal_user_uuid",
  "role": "user|admin", 
  "clerk_id": "clerk_user_id",
  "exp": "expiry_timestamp",
  "iss": "api_gateway"
}
```

**Security Benefits:**
- Signed with internal secret key (not shared with external services)
- Contains resolved internal identifiers
- Short expiry (15-30 minutes)
- Cannot be forged by external parties

### 2.3 Request Forwarding

**Process:**
1. Replace original `Authorization` header with internal JWT
2. Forward request to appropriate downstream service
3. Downstream service validates internal JWT only
4. No external service calls needed for auth

## Phase 3: Microservice Refactoring

### 3.1 Remove Duplicated Auth Logic

**From ms_knowledge:**
- Remove Clerk JWT validation interceptor
- Remove ms_user gRPC client calls for ownership checks
- Simplify to internal JWT validation only

**From ms_user:**
- Keep Clerk validation for webhook endpoints only
- Remove from gRPC endpoints (will receive internal JWTs)

### 3.2 Simplified Authorization

**New Pattern:**
```go
// Before: Complex ownership check with external calls
if role != "admin" {
    user, err := userClient.GetUser(ctx, &userv1.GetUserRequest{})
    if err != nil { /* handle error */ }
    if space.OwnerID != user.Id { /* unauthorized */ }
}

// After: Simple context-based check
userID := ctx.Value("user_id").(string)
if role != "admin" && space.OwnerID != userID {
    return status.Error(codes.PermissionDenied, "not owner")
}
```

### 3.3 Context Injection

**Internal JWT Interceptor:**
- Validates internal JWT signature
- Extracts user_id, role, clerk_id
- Injects into request context
- Lightweight and fast

## Benefits of New Architecture

### 1. Performance Improvements
- Eliminate gRPC calls for every ownership check
- Faster request processing
- Reduced network latency
- Better scalability

### 2. Security Enhancements
- Single point of external token validation
- Reduced attack surface
- Consistent security policies
- Internal tokens cannot be forged

### 3. Development Experience
- Cleaner service code focused on business logic
- Easier testing (mock internal tokens)
- Simplified debugging
- Better separation of concerns

### 4. Operational Benefits
- Centralized authentication monitoring
- Easier to update auth logic
- Single point for rate limiting, logging
- Better observability

## Migration Strategy

### Step 1: Implement API Gateway
- Create new api_gateway service
- Implement Clerk JWT validation
- Add user resolution logic
- Generate internal JWTs

### Step 2: Update Service Interceptors
- Replace Clerk validation with internal JWT validation
- Update context injection logic
- Remove external service dependencies

### Step 3: Refactor Business Logic
- Remove ms_user gRPC calls from ownership checks
- Use context values directly
- Simplify authorization logic

### Step 4: Update Client Integration
- Point frontend to API Gateway endpoints
- Update service discovery configuration
- Test end-to-end flows

### Step 5: Cleanup
- Remove unused Clerk validation code
- Remove ms_user gRPC clients from services
- Update documentation and deployment configs

## Risk Mitigation

### 1. Gradual Migration
- Implement gateway alongside existing services
- Use feature flags for gradual rollout
- Maintain backward compatibility during transition

### 2. Monitoring & Observability
- Add comprehensive logging to gateway
- Monitor authentication success/failure rates
- Track performance metrics

### 3. Fallback Mechanisms
- Keep existing auth logic during migration
- Implement circuit breakers
- Plan rollback procedures

## Success Metrics

### Performance
- Reduce average request latency by 30-50%
- Eliminate 1-2 gRPC calls per ownership check
- Improve service throughput

### Code Quality
- Reduce authentication-related code by 60%
- Eliminate service-to-service auth dependencies
- Improve test coverage and maintainability

### Security
- Single point of token validation
- Consistent security policy enforcement
- Reduced credential distribution

This refactor transforms the system from a distributed authentication model to a centralized gateway pattern, significantly improving performance, security, and maintainability while reducing complexity across all microservices.
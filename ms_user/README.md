# MS_USER Service

The `ms_user` service is a gRPC-based microservice that handles user management, authentication, and authorization for the Knowledge Exploration and Information Structuring Platform.

## Overview

This service provides:
- User profile management (CRUD operations)
- Role-based access control (RBAC)
- Clerk.com integration for authentication
- User preferences management
- User status and profile completion checks

## Architecture

- **Protocol**: gRPC-First architecture
- **Authentication**: Clerk.com JWT tokens
- **Authorization**: Role-based access control (RBAC)
- **Database**: PostgreSQL with Ent ORM
- **Webhooks**: Clerk webhook handling for user events

## Quick Start

```bash
# Install dependencies
go mod download

# Run the service
go run ./cmd/main.go

# Build binary
go build -o ms_user ./cmd/main.go
```

## Configuration

The service requires the following environment variables:
- `CLERK_SECRET_KEY`: Clerk backend API key

- `DATABASE_URL`: PostgreSQL connection string
- `GRPC_SERVER_PORT`: gRPC server port (default: 50051)
- `WEBHOOK_SERVER_PORT`: HTTP webhook server port (default: 8080)

## Workflows

### User Registration & Profile Activation Flow

This diagram illustrates the end-to-end process from a new user signing up via Clerk to their profile being activated in the system. The creation is handled by a webhook, and the activation is handled by a gRPC call from the frontend after the user provides additional details.

```mermaid
sequenceDiagram
    participant Frontend
    participant Clerk
    participant ms_user_service as "ms_user (Webhook)"
    participant ms_user_service_grpc as "ms_user (gRPC)"
    participant Database
    participant Clerk_API as "Clerk (Management API)"

    Note over Frontend, Clerk_API: New User Registration & Activation

    Frontend->>Clerk: 1. User signs up via Clerk UI
    Clerk-->>Frontend: 2. Returns JWT (user is logged in)
    
    Clerk->>ms_user_service: 3. Sends 'user.created' webhook
    ms_user_service->>Database: 4. Create user record with PENDING status
    Database-->>ms_user_service: 5. User record created
    ms_user_service-->>Clerk: 6. Acknowledges webhook (HTTP 200)

    Frontend->>Frontend: 7. Redirects user to /profile-setup page
    Note over Frontend: User is prompted to enter Full Name and Username

    Frontend->>ms_user_service_grpc: 8. Calls ActivateUser gRPC method with fullName, username
    ms_user_service_grpc->>Database: 9. Find user by Clerk ID and update profile, set status to ACTIVE
    Database-->>ms_user_service_grpc: 10. User profile updated
    
    ms_user_service_grpc->>Clerk_API: 11. Set role="user" in Clerk publicMetadata
    Clerk_API-->>ms_user_service_grpc: 12. Role assignment confirmed
    
    ms_user_service_grpc-->>Frontend: 13. Returns success response
    Frontend->>Frontend: 14. Redirects user to /spaces
```

### Existing User Authentication Flow

```mermaid
sequenceDiagram
    participant Frontend
    participant Clerk
    participant AuthMiddleware
    participant AuthzMiddleware
    participant UserService
    participant Database

    Note over Frontend,Database: Existing User Sign-in Flow

    Frontend->>Clerk: 1. Sign in with credentials
    Clerk-->>Frontend: 2. JWT token (with role in publicMetadata)
    
    Frontend->>AuthMiddleware: 3. gRPC request with JWT
    AuthMiddleware->>AuthMiddleware: 4. Validate JWT & extract role
    AuthMiddleware->>AuthzMiddleware: 5. Pass user_id + role in context
    
    AuthzMiddleware->>AuthzMiddleware: 6. Check method permissions
    Note over AuthzMiddleware: Verify user has required permission
    
    AuthzMiddleware-->>UserService: 7. Allow request
    UserService->>Database: 8. Execute business logic
    Database-->>UserService: 9. Return data
    UserService-->>Frontend: 10. Response
```

### Role-Based Access Control (RBAC) Flow

```mermaid
flowchart TD
    A[gRPC Request] --> B[AuthInterceptor]
    B --> C{JWT Valid?}
    C -->|No| D[Return 401 Unauthenticated]
    C -->|Yes| E[Extract user_id + role from JWT]
    
    E --> F[AuthorizationInterceptor]
    F --> G{Method Exempt?}
    G -->|Yes CreateUser| H[Check user_id only]
    G -->|No| I[Check Required Permission]
    
    H --> J{user_id present?}
    J -->|No| D
    J -->|Yes| K[Allow Request]
    
    I --> L{Role has Permission?}
    L -->|No| M[Return 403 Permission Denied]
    L -->|Yes| K
    
    K --> N[Execute Business Logic]
    
    style G fill:#e1f5fe
    style H fill:#f3e5f5
    style I fill:#fff3e0
```

### User Profile Completion Check Flow

```mermaid
sequenceDiagram
    participant Frontend
    participant UserService
    participant Database

    Note over Frontend,Database: Profile Completion Status Check

    Frontend->>UserService: 1. CheckUserStatus()
    UserService->>Database: 2. GetByClerkID(user_id)
    
    alt User exists in database
        Database-->>UserService: 3a. User record found
        UserService->>UserService: 4a. Check profile completeness
        Note over UserService: Verify fullName, username, email are present
        
        alt Profile complete
            UserService-->>Frontend: 5a. ProfileCompleted: true, User: data
        else Profile incomplete
            UserService-->>Frontend: 5b. ProfileCompleted: false, RedirectURL: "/profile-completion"
        end
        
    else User not found
        Database-->>UserService: 3b. User not found
        UserService-->>Frontend: 5c. ProfileCompleted: false, NeedsRedirect: true, RedirectURL: "/profile-completion"
    end
```

### Webhook Event Processing Flow

```mermaid
sequenceDiagram
    participant Clerk
    participant WebhookService
    participant Database

    Note over Clerk,Database: Clerk Webhook Event Processing

    Clerk->>WebhookService: 1. POST /api/webhooks/clerk (e.g., user.created)
    Note over WebhookService: Verify webhook signature
    
    WebhookService->>WebhookService: 2. Parse event type
    
    alt user.created event
        WebhookService->>Database: 3a. Create user record with PENDING status
        Database-->>WebhookService: 4a. User created
        Note over WebhookService: This is the primary mechanism for user creation.
        
    else user.updated event
        WebhookService->>Database: 3b. Update user record
        Database-->>WebhookService: 4b. User updated
        
    else user.deleted event
        WebhookService->>Database: 3c. Delete user record
        Database-->>WebhookService: 4c. User deleted
    end
    
    WebhookService-->>Clerk: 5. HTTP 200 OK
```

## API Endpoints

### gRPC Methods

- `CreateUser(email, fullName, username)` - Create user profile
- `GetUser()` - Get current user profile
- `UpdateUser(email?, fullName?, username?)` - Update user profile
- `DeleteUser()` - Soft delete user
- `CheckUserStatus()` - Check profile completion status
- `UpdateUserPreferences(...)` - Update user preferences

### HTTP Endpoints

- `POST /api/webhooks/clerk` - Clerk webhook handler

## Permissions & Roles

### Roles
- `user` - Standard user role (default)
- `admin` - Administrator role

### Permissions
- `user:read` - Read user data
- `user:write` - Write user data
- `user:delete` - Delete user data
- `user:admin` - Administrative operations

### Method Permissions
- `CreateUser` - **EXEMPT** (requires authentication only)
- `GetUser` - `user:read`
- `UpdateUser` - `user:write`
- `DeleteUser` - `user:delete`
- `CheckUserStatus` - `user:read`
- `UpdateUserPreferences` - `user:write`

## Security Features

- JWT token validation via Clerk
- Role-based access control (RBAC)
- Webhook signature verification
- Input validation and sanitization
- Request size limits
- Audit logging for all operations

## Development

### Testing
```bash
go test ./...
```

### Linting
```bash
go vet ./...
```

### Database Migrations
```bash
go run ./cmd/main.go  # Auto-migration on startup
```
# Design

## 1. Current Architecture

- Protobuf definitions are located in `ms_user/api/proto` and `ms_knowledge/api/proto`.
- Code generation is configured in `frontend/buf.gen.yaml`.
- The current generation process creates both TypeScript and JavaScript files in `frontend/src/api/generated/v1`.
- The frontend application imports and uses the generated JavaScript objects, leading to type issues.

## 2. Proposed Architecture

The overall architecture will remain the same, but we will change the code generation and consumption part.

### 2.1. Code Generation

- We will modify the `frontend/buf.gen.yaml` to stop generating the old JavaScript files. The current configuration already generates the required TypeScript files, so the main task is to clean up the old generation steps if they exist, or just remove the old generated files.
- We will add a script to the `package.json` to run `buf generate` to make the generation process easier.

### 2.2. Code Refactoring

- We will identify all the files in the frontend that import from `.../generated/v1/*_pb.js`.
- We will refactor these files to import from the new TypeScript files (`.../generated/v1/*_pb.ts` and `.../generated/v1/*_connectweb.ts`).
- The refactoring will involve updating the way protobuf messages are instantiated and accessed. For example, instead of `new User()` and `user.setName()`, the new API might use `const user = { name: "test" } satisfies Partial<User>`. We need to check the generated typescript files for the exact usage.

### 2.3. Schema Mapping

This table outlines the primary mapping between the legacy manual types and the new generated Protobuf types.

| Manual Type (`/app/spaces/types/`) | Generated Type (`/api/generated/v1/`) | Notes |
| :--- | :--- | :--- |
| `Space` | `knowledge.v1.Space` | `userId` -> `ownerId`, `stats` object. |
| `ContentSource` | `knowledge.v1.ContentSource` | `processingStatus` -> `status` (enum). |
| `ContentSourceStatus` | `knowledge.v1.ContentSourceStatus` | Enum values require mapping. |
| `User` (Clerk) | `user.v1.User` | Requires mapping from Clerk object. |

**Note:** UI-specific types (e.g., `...Props`, `...State`) and shared utility types (e.g., `ErrorState`) will be preserved.

## 3. Data Models

The data models defined in the `.proto` files will not change.

## 4. API Interfaces

The gRPC-web API interfaces will not change. The client-side implementation will be updated to use the new `connect-web` TypeScript clients.

## 5. Testing Strategy

- After refactoring, we will perform a full build of the frontend application to catch any compile-time errors.
- We will run the application and manually test the features that use the refactored code to ensure they work as expected.
- We will check the browser console for any runtime errors.

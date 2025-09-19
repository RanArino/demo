# Neo4j Repository Unit Tests

This directory contains comprehensive unit tests for the Neo4j repository layer of the Canvas service.

## Test Coverage

### 1. NodeRepository Tests (`node_repository_test.go`)

**Integration Tests:**
- `TestNodeRepo_CreateChunkNodes_Integration` - Tests business logic for chunk node creation
- `TestNodeRepo_CreateContentNode_Validation` - Validates content node structure
- `TestNodeRepo_CreateClusterNode_Validation` - Validates cluster node creation logic
- `TestNodeRepo_UpdateValidations` - Tests update operations for all node types

**Database Interaction Tests:**
- `TestNodeRepo_CreateChunkNodes_DatabaseInteraction` - Validates Cypher query generation and parameter binding

**Key Features Tested:**
- CRUD operations for ContentNode, ChunkNode, and ClusterNode
- Update operations with partial field updates
- Soft delete functionality
- Embedding storage and retrieval
- Character offset tracking (start_position, end_position)
- Spatial coordinates handling
- ML metadata management

### 2. LinkRepository Tests (`link_repository_test.go`)

**Validation Tests:**
- `TestLinkRepo_CreateHierarchicalLinks_Validation` - Tests hierarchical parent-child relationships
- `TestLinkRepo_CreateSemanticLinks_Validation` - Tests semantic similarity links
- `TestLinkRepo_CreateStructuralLink_Validation` - Tests user-created structural links

**Database Interaction Tests:**
- `TestLinkRepo_CreateHierarchicalLinks_DatabaseInteraction` - Validates hierarchical link Cypher generation
- `TestLinkRepo_CreateSemanticLinks_DatabaseInteraction` - Validates semantic link batch operations
- `TestLinkRepo_CreateStructuralLink_DatabaseInteraction` - Validates structural link creation

**Key Features Tested:**
- HIERARCHICAL_PARENT relationship creation
- SEMANTIC_LINK with similarity scores and metadata
- STRUCTURAL_LINK with confidence scores and user attribution
- Batch operations for semantic links
- Metadata handling (exploration, style, semantic tags)
- Temporal tracking (created_at, updated_at, deleted_at)

### 3. Driver Tests (`driver_test.go`)

**Configuration Tests:**
- `TestNewDriver` - Tests driver initialization with various parameters
- `TestDriver_EnsureConstraints_Logic` - Validates constraint creation logic
- `TestDriver_EnsureIndexes_Logic` - Tests index creation including vector indexes

**Vector Index Tests:**
- Vector index configuration (384 dimensions, cosine similarity)
- High-dimensional embedding support
- Idempotent index creation with `IF NOT EXISTS`

**Error Handling Tests:**
- `TestDriver_ErrorHandling` - Tests error logging patterns
- Constraint and index creation error handling

**Performance & Integration Tests:**
- Session management patterns
- Database operation support
- Transaction type validation (write transactions for schema operations)

## Test Architecture

### Testing Approach

**1. Unit Tests with Mock Patterns:**
- Uses `testify/mock` for creating test doubles
- Avoids full Neo4j driver interface mocking (complex and brittle)
- Focuses on business logic validation without external dependencies

**2. Integration-Style Tests:**
- Tests that validate data structures and business rules
- Cypher query generation validation
- Parameter binding verification

**3. Validation Tests:**
- Domain model validation
- Required vs optional field handling
- Data type and constraint validation

### Mock Strategy

Rather than mocking the complex Neo4j driver interfaces directly, the tests use:

1. **Testable Repository Pattern**: Custom test interfaces that capture the essential database operations
2. **Cypher Validation**: Direct testing of generated Cypher queries and parameters
3. **Business Logic Isolation**: Testing domain logic separately from database concerns

### Test Data Patterns

- Uses `uuid.New()` for generating test IDs
- Employs helper functions (`stringPtr`, `intPtr`) for optional fields
- Comprehensive test cases covering both minimal and complete data structures

## Running Tests

```bash
# Run all Neo4j repository tests
go test ./internal/repository/neo4j/ -v

# Run with short flag (faster execution)
go test ./internal/repository/neo4j/ -v -short

# Run specific test
go test ./internal/repository/neo4j/ -v -run TestNodeRepo_CreateChunkNodes
```

## Test Coverage Statistics

The test suite covers:
- ✅ All CRUD operations for nodes and relationships
- ✅ Cypher query generation and parameter binding
- ✅ Error handling and edge cases
- ✅ Data validation and type safety
- ✅ Vector index configuration
- ✅ Temporal data handling
- ✅ Metadata management
- ✅ Soft delete operations
- ✅ Batch operations
- ✅ Schema management (constraints, indexes)

## Future Enhancements

Potential areas for additional testing:
1. **Performance Tests**: Benchmark tests for large batch operations
2. **Integration Tests**: Tests with actual Neo4j test containers
3. **Concurrent Operations**: Tests for concurrent read/write scenarios
4. **Schema Migration Tests**: Tests for database schema evolution
5. **Query Optimization Tests**: Tests for index usage and query performance
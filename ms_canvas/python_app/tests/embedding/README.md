# Embedding Service Tests

Comprehensive test suite for the Canvas Service embedding functionality using Neo4j GraphRAG.

## Test Structure

```
tests/embedding/
├── __init__.py                      # Package initialization
├── test_embedding_unit.py           # Unit tests for individual functions
├── test_embedding_integration.py    # Integration tests with real providers
├── test_embedding_performance.py    # Performance and scalability tests
├── test_all.py                     # Test runner and orchestration
└── README.md                       # This documentation
```

## Test Categories

### 1. Unit Tests (`test_embedding_unit.py`)

Tests individual components in isolation using mocks:

- **EmbeddingConfig**: Configuration dataclass functionality
- **EmbeddingResult**: Result dataclass functionality  
- **_get_neo4j_embedder**: Embedder factory function
- **embed_chunks**: Batch embedding functionality
- **embed_query**: Single query embedding functionality
- **Error handling**: Various failure scenarios

**Run unit tests:**
```bash
python tests/embedding/test_all.py unit
# OR
pytest tests/embedding/test_embedding_unit.py -v
```

### 2. Integration Tests (`test_embedding_integration.py`)

Tests with real embedding providers and gRPC integration:

- **Real Hugging Face embeddings**: Requires `INTEGRATION_TESTS=1`
- **Real OpenAI embeddings**: Requires `OPENAI_API_KEY`
- **gRPC servicer integration**: Tests through the gRPC interface
- **Embedding consistency**: Same input produces same output
- **Error scenarios**: Network failures, invalid configs

**Run integration tests:**
```bash
INTEGRATION_TESTS=1 python tests/embedding/test_all.py integration
# OR
INTEGRATION_TESTS=1 pytest tests/embedding/test_embedding_integration.py -v
```

**With OpenAI tests:**
```bash
OPENAI_API_KEY=your_key INTEGRATION_TESTS=1 python tests/embedding/test_all.py integration
```

### 3. Performance Tests (`test_embedding_performance.py`)

Tests performance characteristics and scalability:

- **Batch embedding performance**: Speed with large batches
- **Single query latency**: Response time for individual queries
- **Large text handling**: Performance with big documents
- **Memory efficiency**: Resource usage optimization
- **Scalability metrics**: Linear scaling verification

**Run performance tests:**
```bash
python tests/embedding/test_all.py performance
# OR
pytest tests/embedding/test_embedding_performance.py -v -m performance
```

### 4. Benchmark Tests (`test_embedding_performance.py`)

Detailed benchmarks for regression detection:

- **Throughput benchmarks**: Texts processed per second
- **Latency benchmarks**: Average response times
- **Resource usage**: Memory and CPU efficiency

**Run benchmark tests:**
```bash
RUN_BENCHMARKS=1 python tests/embedding/test_all.py benchmark
# OR
RUN_BENCHMARKS=1 pytest tests/embedding/test_embedding_performance.py -v -m benchmark
```

## Running All Tests

**Complete test suite:**
```bash
cd ms_canvas/python_app
python tests/embedding/test_all.py all
# OR
python tests/embedding/test_all.py  # default behavior
```

**With all optional tests:**
```bash
cd ms_canvas/python_app
INTEGRATION_TESTS=1 RUN_BENCHMARKS=1 OPENAI_API_KEY=your_key python tests/embedding/test_all.py all
```

**Quick verification (real embeddings):**
```bash
cd ms_canvas/python_app
python -c "from app.services.embedding import *; print('✅ Embedding service ready')"
```

## Test Configuration

### Environment Variables

- `INTEGRATION_TESTS=1`: Enables real embedding provider tests
- `OPENAI_API_KEY=your_key`: Enables OpenAI embedding tests
- `RUN_BENCHMARKS=1`: Enables benchmark tests

### Pytest Markers

- `@pytest.mark.performance`: Performance-related tests
- `@pytest.mark.benchmark`: Benchmark tests for regression detection
- `@pytest.mark.skipif`: Conditional test skipping

## Test Coverage

The test suite covers:

✅ **Functionality**
- All embedding providers (Hugging Face, OpenAI)
- Batch and single query operations
- Configuration handling
- Error scenarios

✅ **Integration**
- Neo4j GraphRAG library integration
- gRPC servicer functionality
- Real embedding provider APIs
- End-to-end workflows

✅ **Performance**
- Response time verification
- Memory usage optimization
- Scalability characteristics
- Throughput benchmarks

✅ **Error Handling**
- Invalid configurations
- Network failures
- Missing dependencies
- API errors

## Expected Performance Benchmarks

Based on the test suite, embedding service should meet:

- **Batch throughput**: >500 texts/second (mocked)
- **Query latency**: <10ms average (mocked)
- **Memory efficiency**: Uses float32 arrays
- **Scalability**: Linear scaling with batch size

## Continuous Integration

For CI/CD pipelines:

```bash
# Fast unit tests (always run)
pytest tests/embedding/test_embedding_unit.py -v

# Integration tests (when dependencies available)
INTEGRATION_TESTS=1 pytest tests/embedding/test_embedding_integration.py -v

# Performance regression tests
pytest tests/embedding/test_embedding_performance.py -v -m performance
```

## Troubleshooting

### Common Issues

1. **Import errors**: Install dependencies with `pip install -e .`
2. **Skipped integration tests**: Set `INTEGRATION_TESTS=1`
3. **OpenAI test failures**: Verify `OPENAI_API_KEY` is valid
4. **Performance test timeouts**: Check system resources

### Debug Mode

Run tests with verbose output and no capture:
```bash
pytest tests/embedding/ -v -s --tb=long
```

### Test Individual Functions

```bash
pytest tests/embedding/test_embedding_unit.py::TestEmbeddingConfig::test_default_config -v
```

## Contributing

When adding new tests:

1. **Follow naming conventions**: `test_[functionality]_[scenario]`
2. **Use appropriate markers**: `@pytest.mark.performance`, etc.
3. **Mock external dependencies**: Use `unittest.mock` for isolation
4. **Document test purpose**: Clear docstrings for complex tests
5. **Verify test independence**: Tests should not depend on each other
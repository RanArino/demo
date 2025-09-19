"""
Performance tests for the embedding service.
Tests embedding speed, memory usage, and scalability.
"""
import pytest
import time
import numpy as np
from unittest.mock import patch, Mock
from app.services.embedding import (
    EmbeddingConfig,
    embed_chunks,
    embed_query,
)


class TestEmbeddingPerformance:
    """Performance tests for embedding operations."""
    
    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_batch_embedding_performance(self, mock_get_embedder):
        """Test performance of batch embedding operations."""
        # Mock fast embedding response
        mock_embedder = Mock()
        embeddings = [[0.1] * 1536 for _ in range(100)]  # 100 embeddings of 1536 dimensions
        mock_embedder.embed_documents.return_value = embeddings
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        contents = [f"Test content {i}" for i in range(100)]
        
        start_time = time.time()
        result = embed_chunks(contents, config)
        end_time = time.time()
        
        duration = end_time - start_time
        
        # Performance assertions
        assert duration < 1.0  # Should complete within 1 second (mocked)
        assert result.vectors.shape == (100, 1536)
        assert result.dims == 1536
        
        # Verify batch processing
        mock_embedder.embed_documents.assert_called_once_with(contents)

    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_single_query_performance(self, mock_get_embedder):
        """Test performance of single query embedding."""
        mock_embedder = Mock()
        mock_embedder.embed_query.return_value = [0.1] * 1536
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        query = "What is the meaning of life?"
        
        start_time = time.time()
        result = embed_query(query, config)
        end_time = time.time()
        
        duration = end_time - start_time
        
        # Performance assertions
        assert duration < 0.1  # Should be very fast for single query (mocked)
        assert result.vectors.shape == (1, 1536)
        assert result.dims == 1536

    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_large_text_embedding(self, mock_get_embedder):
        """Test embedding of large text content."""
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [[0.1] * 1536]
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        
        # Create large text content (10KB)
        large_text = "This is a very long text. " * 400  # ~10KB
        contents = [large_text]
        
        start_time = time.time()
        result = embed_chunks(contents, config)
        end_time = time.time()
        
        duration = end_time - start_time
        
        # Performance assertions
        assert duration < 2.0  # Should handle large text efficiently
        assert result.vectors.shape == (1, 1536)

    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_concurrent_embedding_simulation(self, mock_get_embedder):
        """Simulate concurrent embedding requests."""
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [[0.1] * 1536]
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        
        # Simulate multiple concurrent requests
        requests = [["Request text"] for _ in range(10)]
        
        start_time = time.time()
        results = []
        for request in requests:
            result = embed_chunks(request, config)
            results.append(result)
        end_time = time.time()
        
        duration = end_time - start_time
        
        # Performance assertions
        assert duration < 1.0  # Should handle multiple requests efficiently
        assert len(results) == 10
        assert all(r.vectors.shape == (1, 1536) for r in results)

    @pytest.mark.performance
    def test_memory_efficient_embedding(self):
        """Test memory usage for embedding operations."""
        config = EmbeddingConfig()
        
        # Test that empty inputs don't consume unnecessary memory
        result = embed_chunks([], config)
        assert result.vectors.size == 0
        assert result.dims == 0
        
        result = embed_query("", config)
        assert result.vectors.size == 0
        assert result.dims == 0

    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_embedding_vector_dtype_efficiency(self, mock_get_embedder):
        """Test that embeddings use efficient data types."""
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [[0.1, 0.2, 0.3]]
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        result = embed_chunks(["test"], config)
        
        # Verify efficient float32 dtype
        assert result.vectors.dtype == np.float32
        assert result.vectors.nbytes == 12  # 3 floats * 4 bytes each


class TestScalabilityMetrics:
    """Test scalability characteristics of embedding operations."""
    
    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_linear_scaling_batch_size(self, mock_get_embedder):
        """Test that embedding time scales roughly linearly with batch size."""
        mock_embedder = Mock()
        
        def mock_embed_documents(contents):
            # Simulate processing time proportional to batch size
            time.sleep(0.001 * len(contents))
            return [[0.1] * 1536 for _ in contents]
        
        mock_embedder.embed_documents.side_effect = mock_embed_documents
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        
        # Test different batch sizes
        batch_sizes = [10, 20, 50]
        durations = []
        
        for batch_size in batch_sizes:
            contents = [f"Text {i}" for i in range(batch_size)]
            
            start_time = time.time()
            result = embed_chunks(contents, config)
            end_time = time.time()
            
            durations.append(end_time - start_time)
            assert result.vectors.shape[0] == batch_size
        
        # Verify roughly linear scaling (allowing for some variance)
        # Duration should roughly double from 10 to 20, and increase further to 50
        assert durations[1] > durations[0] * 1.5  # 20 > 10 * 1.5
        assert durations[2] > durations[1] * 2.0  # 50 > 20 * 2.0

    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_text_length_scaling(self, mock_get_embedder):
        """Test embedding performance with varying text lengths."""
        mock_embedder = Mock()
        
        def mock_embed_based_on_length(contents):
            total_chars = sum(len(content) for content in contents)
            # Simulate slight delay based on text length
            time.sleep(0.000001 * total_chars)
            return [[0.1] * 1536 for _ in contents]
        
        mock_embedder.embed_documents.side_effect = mock_embed_based_on_length
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        
        # Test different text lengths
        short_text = "Short text"
        medium_text = "This is a medium length text " * 10
        long_text = "This is a very long text " * 100
        
        texts = [short_text, medium_text, long_text]
        durations = []
        
        for text in texts:
            start_time = time.time()
            result = embed_chunks([text], config)
            end_time = time.time()
            
            durations.append(end_time - start_time)
            assert result.vectors.shape == (1, 1536)
        
        # Verify that longer texts take more time (but not excessively)
        assert durations[1] >= durations[0]  # Medium >= Short
        assert durations[2] >= durations[1]  # Long >= Medium
        assert durations[2] < durations[0] * 100  # But not 100x slower


class TestResourceUsage:
    """Test resource usage patterns."""
    
    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_embedder_reuse(self, mock_get_embedder):
        """Test that embedder instances are created efficiently."""
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [[0.1] * 1536]
        mock_embedder.embed_query.return_value = [0.1] * 1536
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        
        # Multiple embedding calls
        embed_chunks(["text1"], config)
        embed_chunks(["text2"], config)
        embed_query("query1", config)
        embed_query("query2", config)
        
        # Verify embedder creation
        assert mock_get_embedder.call_count == 4  # One per call
        
        # Verify all calls were made to the same embedder instance
        assert mock_embedder.embed_documents.call_count == 2
        assert mock_embedder.embed_query.call_count == 2

    @pytest.mark.performance
    def test_config_validation_performance(self):
        """Test that config validation is fast."""
        start_time = time.time()
        
        # Create many config instances
        configs = [EmbeddingConfig() for _ in range(1000)]
        
        end_time = time.time()
        duration = end_time - start_time
        
        assert duration < 0.1  # Config creation should be very fast
        assert len(configs) == 1000
        assert all(c.provider == "huggingface" for c in configs)

    @pytest.mark.performance
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_numpy_array_operations_performance(self, mock_get_embedder):
        """Test numpy array operations performance."""
        mock_embedder = Mock()
        # Large batch of embeddings
        embeddings = [[0.1] * 1536 for _ in range(500)]  # 500 embeddings, 1536 dims
        mock_embedder.embed_documents.return_value = embeddings
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        contents = [f"Text {i}" for i in range(500)]
        
        start_time = time.time()
        result = embed_chunks(contents, config)
        end_time = time.time()
        
        duration = end_time - start_time
        
        # Numpy operations should be fast
        assert duration < 0.5  # Array operations should be very fast
        assert result.vectors.shape == (500, 1536)
        assert result.vectors.dtype == np.float32


class TestBenchmarkSuite:
    """Benchmark tests for performance regression detection."""
    
    @pytest.mark.benchmark
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_embedding_throughput_benchmark(self, mock_get_embedder):
        """Benchmark embedding throughput."""
        mock_embedder = Mock()
        embeddings = [[0.1] * 1536 for _ in range(1000)]
        mock_embedder.embed_documents.return_value = embeddings
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        contents = [f"Benchmark text {i}" for i in range(1000)]
        
        start_time = time.time()
        result = embed_chunks(contents, config)
        end_time = time.time()
        
        duration = end_time - start_time
        throughput = len(contents) / duration  # texts per second
        
        # Performance benchmark assertions
        assert throughput > 500  # Should process at least 500 texts/second (mocked)
        assert result.vectors.shape == (1000, 1536)
        
        print(f"Embedding throughput: {throughput:.2f} texts/second")

    @pytest.mark.benchmark
    @patch('app.services.embedding._get_neo4j_embedder')  
    def test_query_latency_benchmark(self, mock_get_embedder):
        """Benchmark single query embedding latency."""
        mock_embedder = Mock()
        mock_embedder.embed_query.return_value = [0.1] * 1536
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        query = "Benchmark query embedding latency test"
        
        # Warm up
        embed_query(query, config)
        
        # Benchmark multiple runs
        durations = []
        for _ in range(10):
            start_time = time.time()
            result = embed_query(query, config)
            end_time = time.time()
            durations.append(end_time - start_time)
        
        avg_duration = sum(durations) / len(durations)
        
        # Latency benchmark assertions
        assert avg_duration < 0.01  # Should be under 10ms average (mocked)
        assert result.vectors.shape == (1, 1536)
        
        print(f"Average query embedding latency: {avg_duration*1000:.2f}ms")
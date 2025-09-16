"""
Integration tests for the embedding service.
Tests the service with real embedding providers and gRPC integration.
"""
import pytest
import numpy as np
import os
from unittest.mock import patch, Mock
from app.services.embedding import (
    EmbeddingConfig,
    embed_chunks,
    embed_query,
)
from app.server import CanvasInternalServicer
from app.proto.v1 import canvas_pb2


class TestEmbeddingIntegration:
    """Integration tests for embedding service."""
    
    @pytest.mark.skipif(
        "INTEGRATION_TESTS" not in os.environ,
        reason="Integration tests require INTEGRATION_TESTS env var"
    )
    def test_huggingface_embedding_integration(self):
        """Test real Hugging Face embedding."""
        config = EmbeddingConfig(provider="huggingface", model_id="all-MiniLM-L6-v2")
        contents = ["Hello world", "This is a test sentence"]
        
        result = embed_chunks(contents, config)
        
        assert result.vectors.shape[0] == 2  # Two embeddings
        assert result.vectors.shape[1] > 0   # Non-zero dimensions
        assert result.dims == result.vectors.shape[1]
        assert result.model_id == "all-MiniLM-L6-v2"
        assert len(result.model_version) > 0

    @pytest.mark.skipif(
        "OPENAI_API_KEY" not in os.environ,
        reason="OpenAI integration tests require OPENAI_API_KEY"
    )
    def test_openai_embedding_integration(self):
        """Test real OpenAI embedding."""
        config = EmbeddingConfig(provider="openai", model_id="text-embedding-3-small")
        contents = ["Hello world", "This is a test sentence"]
        
        result = embed_chunks(contents, config)
        
        assert result.vectors.shape[0] == 2
        assert result.vectors.shape[1] > 0
        assert result.dims == result.vectors.shape[1]
        assert result.model_id == "text-embedding-3-small"

    @pytest.mark.skipif(
        "INTEGRATION_TESTS" not in os.environ,
        reason="Integration tests require INTEGRATION_TESTS env var"
    )
    def test_query_embedding_integration(self):
        """Test real query embedding."""
        config = EmbeddingConfig(provider="huggingface", model_id="all-MiniLM-L6-v2")
        query = "What is machine learning?"
        
        result = embed_query(query, config)
        
        assert result.vectors.shape[0] == 1
        assert result.vectors.shape[1] > 0
        assert result.dims == result.vectors.shape[1]

    @pytest.mark.skipif(
        "INTEGRATION_TESTS" not in os.environ,
        reason="Integration tests require INTEGRATION_TESTS env var"
    )
    def test_embedding_consistency(self):
        """Test that same text produces same embeddings."""
        config = EmbeddingConfig(provider="huggingface", model_id="all-MiniLM-L6-v2")
        text = "Consistent embedding test"
        
        result1 = embed_query(text, config)
        result2 = embed_query(text, config)
        
        np.testing.assert_array_almost_equal(result1.vectors, result2.vectors, decimal=5)

    @pytest.mark.skipif(
        "INTEGRATION_TESTS" not in os.environ,
        reason="Integration tests require INTEGRATION_TESTS env var"
    )
    def test_different_texts_different_embeddings(self):
        """Test that different texts produce different embeddings."""
        config = EmbeddingConfig(provider="huggingface", model_id="all-MiniLM-L6-v2")
        
        result1 = embed_query("The cat sat on the mat", config)
        result2 = embed_query("Machine learning algorithms", config)
        
        # Embeddings should be different
        similarity = np.dot(result1.vectors[0], result2.vectors[0])
        assert similarity < 0.9  # Not too similar

    def test_batch_vs_individual_consistency(self):
        """Test batch embedding vs individual embeddings produce similar results."""
        with patch('app.services.embedding._get_neo4j_embedder') as mock_get_embedder:
            mock_embedder = Mock()
            mock_embedder.embed_documents.return_value = [
                [0.1, 0.2, 0.3],
                [0.4, 0.5, 0.6]
            ]
            mock_embedder.embed_query.side_effect = [
                [0.1, 0.2, 0.3],
                [0.4, 0.5, 0.6]
            ]
            mock_get_embedder.return_value = mock_embedder
            
            config = EmbeddingConfig()
            texts = ["text1", "text2"]
            
            # Batch embedding
            batch_result = embed_chunks(texts, config)
            
            # Individual embeddings
            individual_results = [embed_query(text, config) for text in texts]
            
            # Compare results
            for i, individual_result in enumerate(individual_results):
                np.testing.assert_array_almost_equal(
                    batch_result.vectors[i], 
                    individual_result.vectors[0],
                    decimal=5
                )


class TestGRPCIntegration:
    """Test embedding service through gRPC interface."""
    
    def setup_method(self):
        self.servicer = CanvasInternalServicer()
        self.context = Mock()

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_embed_chunks_grpc(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [
            [0.1, 0.2, 0.3],
            [0.4, 0.5, 0.6]
        ]
        mock_get_embedder.return_value = mock_embedder
        
        # Create request
        request = canvas_pb2.EmbedChunksRequest()
        request.contents.extend(["text1", "text2"])
        request.config.provider = "huggingface"
        request.config.model_id = "all-MiniLM-L6-v2"
        
        # Call servicer
        response = self.servicer.EmbedChunks(request, self.context)
        
        # Verify response
        assert len(response.flat_vectors) == 6  # 2 vectors * 3 dimensions
        assert response.dims == 3
        assert response.model_id == "all-MiniLM-L6-v2"
        
        # Verify vectors
        vectors = np.array(response.flat_vectors).reshape(2, 3)
        np.testing.assert_array_almost_equal(vectors[0], [0.1, 0.2, 0.3])
        np.testing.assert_array_almost_equal(vectors[1], [0.4, 0.5, 0.6])

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_embed_query_grpc(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_query.return_value = [0.1, 0.2, 0.3]
        mock_get_embedder.return_value = mock_embedder
        
        # Create request
        request = canvas_pb2.EmbedQueryRequest()
        request.text = "test query"
        request.config.provider = "huggingface"
        request.config.model_id = "all-MiniLM-L6-v2"
        
        # Call servicer
        response = self.servicer.EmbedQuery(request, self.context)
        
        # Verify response
        assert len(response.vector) == 3
        assert response.dims == 3
        assert response.model_id == "all-MiniLM-L6-v2"
        np.testing.assert_array_almost_equal(response.vector, [0.1, 0.2, 0.3])

    def test_embed_chunks_grpc_no_contents(self):
        # Create request with no contents
        request = canvas_pb2.EmbedChunksRequest()
        request.config.provider = "huggingface"
        
        # Call servicer
        response = self.servicer.EmbedChunks(request, self.context)
        
        # Verify error handling
        self.context.set_code.assert_called_with(pytest.importorskip("grpc").StatusCode.INVALID_ARGUMENT)

    def test_embed_query_grpc_empty_text(self):
        # Create request with empty text
        request = canvas_pb2.EmbedQueryRequest()
        request.text = ""
        request.config.provider = "huggingface"
        
        # Call servicer
        response = self.servicer.EmbedQuery(request, self.context)
        
        # Verify error handling
        self.context.set_code.assert_called_with(pytest.importorskip("grpc").StatusCode.INVALID_ARGUMENT)

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_embed_chunks_grpc_error_handling(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_documents.side_effect = Exception("Embedding failed")
        mock_get_embedder.return_value = mock_embedder
        
        # Create request
        request = canvas_pb2.EmbedChunksRequest()
        request.contents.extend(["text1"])
        request.config.provider = "huggingface"
        
        # Call servicer
        response = self.servicer.EmbedChunks(request, self.context)
        
        # Verify error handling
        self.context.set_code.assert_called_with(pytest.importorskip("grpc").StatusCode.INTERNAL)
        self.context.set_details.assert_called_with("Embedding failed: Embedding failed")

    def test_grpc_default_config(self):
        """Test that gRPC servicer uses correct defaults."""
        with patch('app.services.embedding._get_neo4j_embedder') as mock_get_embedder:
            mock_embedder = Mock()
            mock_embedder.embed_documents.return_value = [[0.1, 0.2]]
            mock_get_embedder.return_value = mock_embedder
            
            # Create request without config
            request = canvas_pb2.EmbedChunksRequest()
            request.contents.extend(["text"])
            
            # Call servicer
            response = self.servicer.EmbedChunks(request, self.context)
            
            # Verify default config was used
            args, kwargs = mock_get_embedder.call_args
            config = args[0]
            assert config.provider == "huggingface"
            assert config.model_id == "all-MiniLM-L6-v2"


class TestErrorScenarios:
    """Test various error scenarios in integration."""
    
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_model_loading_failure(self, mock_get_embedder):
        mock_get_embedder.side_effect = Exception("Model not found")
        
        config = EmbeddingConfig()
        
        with pytest.raises(Exception, match="Model not found"):
            embed_chunks(["test"], config)

    def test_invalid_config_types(self):
        """Test embedding with invalid configuration types."""
        # This should be handled gracefully by the system
        config = EmbeddingConfig(provider="huggingface")
        
        # Empty list should return empty result
        result = embed_chunks([], config)
        assert len(result.vectors) == 0

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_network_timeout_simulation(self, mock_get_embedder):
        """Simulate network timeout during embedding."""
        mock_embedder = Mock()
        mock_embedder.embed_documents.side_effect = TimeoutError("Request timed out")
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        
        with pytest.raises(RuntimeError, match="Failed to generate embeddings"):
            embed_chunks(["test"], config)
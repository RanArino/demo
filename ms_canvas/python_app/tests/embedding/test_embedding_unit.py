"""
Unit tests for the embedding service.
Tests individual components and functions.
"""
import pytest
import numpy as np
from unittest.mock import Mock, patch, MagicMock
from app.services.embedding import (
    EmbeddingConfig,
    EmbeddingResult,
    _get_neo4j_embedder,
    embed_chunks,
    embed_query,
)


class TestEmbeddingConfig:
    """Test EmbeddingConfig dataclass."""
    
    def test_default_config(self):
        config = EmbeddingConfig()
        assert config.provider == "huggingface"
        assert config.model_id == "all-MiniLM-L6-v2"
        assert config.model_version == ""

    def test_custom_config(self):
        config = EmbeddingConfig(
            provider="openai",
            model_id="text-embedding-3-small",
            model_version="v1.0"
        )
        assert config.provider == "openai"
        assert config.model_id == "text-embedding-3-small"
        assert config.model_version == "v1.0"


class TestEmbeddingResult:
    """Test EmbeddingResult dataclass."""
    
    def test_embedding_result_creation(self):
        vectors = np.array([[0.1, 0.2, 0.3], [0.4, 0.5, 0.6]])
        result = EmbeddingResult(
            vectors=vectors,
            dims=3,
            model_id="test-model",
            model_version="v1.0"
        )
        assert np.array_equal(result.vectors, vectors)
        assert result.dims == 3
        assert result.model_id == "test-model"
        assert result.model_version == "v1.0"


class TestGetNeo4jEmbedder:
    """Test _get_neo4j_embedder function."""
    
    @patch('app.services.embedding.SentenceTransformerEmbeddings')
    def test_huggingface_embedder(self, mock_st_embeddings):
        mock_embedder = Mock()
        mock_st_embeddings.return_value = mock_embedder
        
        config = EmbeddingConfig(provider="huggingface", model_id="test-model")
        result = _get_neo4j_embedder(config)
        
        mock_st_embeddings.assert_called_once_with(model="test-model")
        assert result == mock_embedder

    @patch('app.services.embedding.SentenceTransformerEmbeddings')
    def test_local_embedder(self, mock_st_embeddings):
        mock_embedder = Mock()
        mock_st_embeddings.return_value = mock_embedder
        
        config = EmbeddingConfig(provider="local", model_id="test-model")
        result = _get_neo4j_embedder(config)
        
        mock_st_embeddings.assert_called_once_with(model="test-model")
        assert result == mock_embedder

    @patch('app.services.embedding.OpenAIEmbeddings')
    @patch.dict('os.environ', {'OPENAI_API_KEY': 'test-key'})
    def test_openai_embedder(self, mock_openai_embeddings):
        mock_embedder = Mock()
        mock_openai_embeddings.return_value = mock_embedder
        
        config = EmbeddingConfig(provider="openai", model_id="text-embedding-3-small")
        result = _get_neo4j_embedder(config)
        
        mock_openai_embeddings.assert_called_once_with(
            api_key="test-key", 
            model="text-embedding-3-small"
        )
        assert result == mock_embedder

    @patch.dict('os.environ', {}, clear=True)
    def test_openai_embedder_no_api_key(self):
        config = EmbeddingConfig(provider="openai")
        
        with pytest.raises(ValueError, match="OPENAI_API_KEY environment variable is required"):
            _get_neo4j_embedder(config)

    def test_unsupported_provider(self):
        config = EmbeddingConfig(provider="unsupported")
        
        with pytest.raises(ValueError, match="Unsupported embedding provider: unsupported"):
            _get_neo4j_embedder(config)


class TestEmbedChunks:
    """Test embed_chunks function."""
    
    def test_empty_contents(self):
        config = EmbeddingConfig()
        result = embed_chunks([], config)
        
        assert len(result.vectors) == 0
        assert result.dims == 0
        assert result.model_id == config.model_id
        assert result.model_version == config.model_version

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_successful_embedding(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [
            [0.1, 0.2, 0.3],
            [0.4, 0.5, 0.6]
        ]
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        contents = ["test text 1", "test text 2"]
        result = embed_chunks(contents, config)
        
        mock_embedder.embed_documents.assert_called_once_with(contents)
        assert result.vectors.shape == (2, 3)
        assert result.dims == 3
        assert result.model_id == config.model_id
        assert np.allclose(result.vectors[0], [0.1, 0.2, 0.3])
        assert np.allclose(result.vectors[1], [0.4, 0.5, 0.6])

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_single_embedding_reshape(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [0.1, 0.2, 0.3]
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        contents = ["single text"]
        result = embed_chunks(contents, config)
        
        assert result.vectors.shape == (1, 3)
        assert result.dims == 3

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_embedding_failure(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_documents.side_effect = Exception("API Error")
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        contents = ["test text"]
        
        with pytest.raises(RuntimeError, match="Failed to generate embeddings: API Error"):
            embed_chunks(contents, config)


class TestEmbedQuery:
    """Test embed_query function."""
    
    def test_empty_query(self):
        config = EmbeddingConfig()
        result = embed_query("", config)
        
        assert len(result.vectors) == 0
        assert result.dims == 0
        assert result.model_id == config.model_id

    def test_whitespace_only_query(self):
        config = EmbeddingConfig()
        result = embed_query("   \n\t   ", config)
        
        assert len(result.vectors) == 0
        assert result.dims == 0

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_successful_query_embedding(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_query.return_value = [0.1, 0.2, 0.3]
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        text = "test query"
        result = embed_query(text, config)
        
        mock_embedder.embed_query.assert_called_once_with(text)
        assert result.vectors.shape == (1, 3)
        assert result.dims == 3
        assert np.allclose(result.vectors[0], [0.1, 0.2, 0.3])

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_query_embedding_failure(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_query.side_effect = Exception("API Error")
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        text = "test query"
        
        with pytest.raises(RuntimeError, match="Failed to generate query embedding: API Error"):
            embed_query(text, config)

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_empty_embedding_result(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_query.return_value = []
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig()
        text = "test query"
        result = embed_query(text, config)
        
        assert result.dims == 0


class TestConfigurationVariants:
    """Test different configuration combinations."""
    
    @patch('app.services.embedding._get_neo4j_embedder')
    def test_custom_model_version(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [[0.1, 0.2]]
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig(model_version="custom-v1.0")
        result = embed_chunks(["test"], config)
        
        assert result.model_version == "custom-v1.0"

    @patch('app.services.embedding._get_neo4j_embedder')
    def test_fallback_model_version(self, mock_get_embedder):
        mock_embedder = Mock()
        mock_embedder.embed_documents.return_value = [[0.1, 0.2]]
        mock_get_embedder.return_value = mock_embedder
        
        config = EmbeddingConfig(model_id="test-model", model_version="")
        result = embed_chunks(["test"], config)
        
        assert result.model_version == "test-model"
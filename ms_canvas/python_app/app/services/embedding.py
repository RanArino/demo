import os
import logging
from dataclasses import dataclass
from typing import List

import numpy as np
from sklearn.preprocessing import normalize

# Import provider classes from the providers package
try:
    from ..providers import (
        GeminiEmbeddingProvider,
        OpenAIEmbeddingProvider,
        LocalEmbeddingProvider,
        BaseEmbeddingProvider
    )
    PROVIDERS_AVAILABLE = True
except ImportError as e:
    logging.warning(f"Providers not available: {e}")
    PROVIDERS_AVAILABLE = False
    GeminiEmbeddingProvider = None
    OpenAIEmbeddingProvider = None
    LocalEmbeddingProvider = None
    BaseEmbeddingProvider = None


@dataclass
class EmbeddingConfig:
    provider: str = "gemini"
    model_id: str = "gemini-embedding-001"
    model_version: str = ""
    output_dimensionality: int = 1536
    task_type: str = "SEMANTIC_SIMILARITY"
    normalize_embeddings: bool = True


@dataclass
class EmbeddingResult:
    vectors: np.ndarray
    dims: int
    model_id: str
    model_version: str




def _get_embedder(config: EmbeddingConfig):
    """Get the appropriate embedder based on provider configuration"""
    if not PROVIDERS_AVAILABLE:
        raise RuntimeError("Embedding providers not available. Check provider dependencies.")

    if config.provider == "gemini":
        return GeminiEmbeddingProvider(
            model_name=config.model_id,
            api_key=os.getenv("GEMINI_API_KEY")
        )
    elif config.provider == "openai":
        return OpenAIEmbeddingProvider(
            model=config.model_id,
            api_key=os.getenv("OPENAI_API_KEY")
        )
    elif config.provider in ["huggingface", "local"]:
        # Use local embeddings provider
        return LocalEmbeddingProvider(model_name=config.model_id)
    else:
        raise ValueError(f"Unsupported embedding provider: {config.provider}")


def embed_chunks(contents: List[str], config: EmbeddingConfig) -> EmbeddingResult:
    if not contents:
        return EmbeddingResult(
            vectors=np.array([]),
            dims=0,
            model_id=config.model_id,
            model_version=config.model_version
        )

    embedder = _get_embedder(config)

    try:
        # Try batch embedding first (for providers that support it)
        if hasattr(embedder, 'embed_documents'):
            embeddings = embedder.embed_documents(
                contents,
                task_type=config.task_type,
                output_dimensionality=config.output_dimensionality
            )
        else:
            # Fallback to individual embedding
            embeddings = []
            for content in contents:
                embedding = embedder.embed_query(
                    content,
                    task_type=config.task_type,
                    output_dimensionality=config.output_dimensionality
                )
                embeddings.append(embedding)

        vectors_array = np.array(embeddings, dtype=np.float32)
        if vectors_array.ndim == 1:
            vectors_array = vectors_array.reshape(1, -1)

        # Normalize embeddings if requested
        if config.normalize_embeddings:
            vectors_array = normalize(vectors_array)

        dims = vectors_array.shape[1] if len(vectors_array) > 0 else 0
        model_version = config.model_version or config.model_id

        return EmbeddingResult(
            vectors=vectors_array,
            dims=dims,
            model_id=config.model_id,
            model_version=model_version
        )

    except Exception as e:
        logging.error(f"Embedding failed: {e}")
        raise RuntimeError(f"Failed to generate embeddings: {e}") from e


def embed_query(text: str, config: EmbeddingConfig) -> EmbeddingResult:
    if not text.strip():
        return EmbeddingResult(
            vectors=np.array([]),
            dims=0,
            model_id=config.model_id,
            model_version=config.model_version
        )

    embedder = _get_embedder(config)

    try:
        embedding = embedder.embed_query(
            text,
            task_type=config.task_type,
            output_dimensionality=config.output_dimensionality
        )

        vector_array = np.array([embedding], dtype=np.float32)

        # Normalize if requested
        if config.normalize_embeddings:
            vector_array = normalize(vector_array)

        dims = len(embedding) if embedding else 0
        model_version = config.model_version or config.model_id

        return EmbeddingResult(
            vectors=vector_array,
            dims=dims,
            model_id=config.model_id,
            model_version=model_version
        )

    except Exception as e:
        logging.error(f"Query embedding failed: {e}")
        raise RuntimeError(f"Failed to generate query embedding: {e}") from e
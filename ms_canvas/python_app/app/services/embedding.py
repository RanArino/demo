import os
import logging
from dataclasses import dataclass
from typing import List

import numpy as np
from neo4j_graphrag.embeddings import SentenceTransformerEmbeddings, OpenAIEmbeddings


@dataclass
class EmbeddingConfig:
    provider: str = "huggingface"
    model_id: str = "all-MiniLM-L6-v2"
    model_version: str = ""


@dataclass
class EmbeddingResult:
    vectors: np.ndarray
    dims: int
    model_id: str
    model_version: str


def _get_neo4j_embedder(config: EmbeddingConfig):
    if config.provider in ["huggingface", "local"]:
        return SentenceTransformerEmbeddings(model=config.model_id)
    elif config.provider == "openai":
        api_key = os.getenv("OPENAI_API_KEY")
        if not api_key:
            raise ValueError("OPENAI_API_KEY environment variable is required")
        return OpenAIEmbeddings(api_key=api_key, model=config.model_id)
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
    
    embedder = _get_neo4j_embedder(config)
    
    try:
        # Neo4j GraphRAG only has embed_query, so we need to embed each content individually
        embeddings = []
        for content in contents:
            embedding = embedder.embed_query(content)
            embeddings.append(embedding)
        
        vectors_array = np.array(embeddings, dtype=np.float32)
        if vectors_array.ndim == 1:
            vectors_array = vectors_array.reshape(1, -1)
        
        dims = vectors_array.shape[1] if len(vectors_array) > 0 else 0
        model_version = config.model_version or config.model_id
        
        return EmbeddingResult(
            vectors=vectors_array,
            dims=dims,
            model_id=config.model_id,
            model_version=model_version
        )
        
    except Exception as e:
        logging.error(f"Neo4j GraphRAG embedding failed: {e}")
        raise RuntimeError(f"Failed to generate embeddings: {e}") from e


def embed_query(text: str, config: EmbeddingConfig) -> EmbeddingResult:
    if not text.strip():
        return EmbeddingResult(
            vectors=np.array([]),
            dims=0,
            model_id=config.model_id,
            model_version=config.model_version
        )
    
    embedder = _get_neo4j_embedder(config)
    
    try:
        embedding = embedder.embed_query(text)
        
        vector_array = np.array([embedding], dtype=np.float32)
        dims = len(embedding) if embedding else 0
        model_version = config.model_version or config.model_id
        
        return EmbeddingResult(
            vectors=vector_array,
            dims=dims,
            model_id=config.model_id,
            model_version=model_version
        )
        
    except Exception as e:
        logging.error(f"Neo4j GraphRAG query embedding failed: {e}")
        raise RuntimeError(f"Failed to generate query embedding: {e}") from e
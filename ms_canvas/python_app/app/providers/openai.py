import os
import logging
from typing import List, Optional

from .base import BaseEmbeddingProvider


class OpenAIEmbeddingProvider(BaseEmbeddingProvider):
    """OpenAI embeddings provider"""

    def __init__(self, model: str = "text-embedding-ada-002", api_key: Optional[str] = None):
        self.model_name = model
        self.api_key = api_key or os.getenv("OPENAI_API_KEY")

        if not self.api_key:
            raise ValueError("OPENAI_API_KEY environment variable is required")

        try:
            from openai import OpenAI
            self.client = OpenAI(api_key=self.api_key)
            self.available = True
        except ImportError:
            logging.warning("OpenAI SDK not available. Install with: pip install openai")
            self.client = None
            self.available = False

    def embed_query(self, text: str, **kwargs) -> List[float]:
        """Embed a single text query"""
        if not self.available:
            raise RuntimeError("OpenAI provider not available. Install OpenAI SDK.")

        if not text.strip():
            raise ValueError("Text cannot be empty")

        try:
            response = self.client.embeddings.create(
                input=text,
                model=self.model_name
            )
            return response.data[0].embedding

        except Exception as e:
            logging.error(f"OpenAI embedding failed: {e}")
            raise RuntimeError(f"Failed to generate OpenAI embedding: {e}") from e

    def embed_documents(self, texts: List[str], **kwargs) -> List[List[float]]:
        """Embed multiple documents in batch"""
        if not self.available:
            raise RuntimeError("OpenAI provider not available. Install OpenAI SDK.")

        if not texts:
            return []

        try:
            response = self.client.embeddings.create(
                input=texts,
                model=self.model_name
            )
            return [data.embedding for data in response.data]

        except Exception as e:
            logging.error(f"OpenAI batch embedding failed: {e}")
            raise RuntimeError(f"Failed to generate OpenAI batch embeddings: {e}") from e

    def get_dimensions(self) -> int:
        """Get the dimensionality of embeddings from this provider"""
        # OpenAI ada-002 has 1536 dimensions
        return 1536

    def get_model_name(self) -> str:
        """Get the model name used by this provider"""
        return self.model_name

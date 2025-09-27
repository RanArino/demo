import os
import logging
from typing import List, Optional

from .base import BaseEmbeddingProvider


class GeminiEmbeddingProvider(BaseEmbeddingProvider):
    """Gemini embeddings provider using Google GenAI SDK"""

    def __init__(self, model_name: str = "gemini-embedding-001", api_key: Optional[str] = None):
        self.model_name = model_name
        self.api_key = api_key or os.getenv("GEMINI_API_KEY")

        if not self.api_key:
            raise ValueError("GEMINI_API_KEY environment variable is required")

        try:
            import google.generativeai as genai
            self.genai = genai
            self.genai.configure(api_key=self.api_key)
            self.model = self.genai.embed_content
            self.available = True
        except ImportError:
            logging.warning("Google GenAI SDK not available. Install with: pip install google-generativeai")
            self.genai = None
            self.model = None
            self.available = False

    def embed_query(self, text: str, task_type: str = "SEMANTIC_SIMILARITY",
                   output_dimensionality: int = 1536) -> List[float]:
        """Embed a single text query"""
        if not self.available:
            raise RuntimeError("Gemini provider not available. Install Google GenAI SDK.")

        if not text.strip():
            raise ValueError("Text cannot be empty")

        try:
            result = self.model(
                model=self.model_name,
                content=text,
                task_type=task_type,
                output_dimensionality=output_dimensionality
            )

            if "embedding" not in result:
                raise ValueError("No embedding returned from Gemini API")

            return result["embedding"]

        except Exception as e:
            logging.error(f"Gemini embedding failed: {e}")
            raise RuntimeError(f"Failed to generate Gemini embedding: {e}") from e

    def embed_documents(self, texts: List[str], task_type: str = "SEMANTIC_SIMILARITY",
                       output_dimensionality: int = 1536) -> List[List[float]]:
        """Embed multiple documents in batch"""
        if not self.available:
            raise RuntimeError("Gemini provider not available. Install Google GenAI SDK.")

        if not texts:
            return []

        embeddings = []
        for text in texts:
            embedding = self.embed_query(text, task_type, output_dimensionality)
            embeddings.append(embedding)

        return embeddings

    def get_dimensions(self) -> int:
        """Get the dimensionality of embeddings from this provider"""
        # For Gemini, this varies based on output_dimensionality parameter
        # Default is 1536, but can be 128-3072
        return 1536

    def get_model_name(self) -> str:
        """Get the model name used by this provider"""
        return self.model_name

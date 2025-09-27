import logging
from abc import ABC, abstractmethod
from typing import List, Optional


class BaseEmbeddingProvider(ABC):
    """Base interface for all embedding providers"""

    @abstractmethod
    def embed_query(self, text: str, **kwargs) -> List[float]:
        """Embed a single text query"""
        pass

    @abstractmethod
    def embed_documents(self, texts: List[str], **kwargs) -> List[List[float]]:
        """Embed multiple documents in batch"""
        pass

    @abstractmethod
    def get_dimensions(self) -> int:
        """Get the dimensionality of embeddings from this provider"""
        pass

    @abstractmethod
    def get_model_name(self) -> str:
        """Get the model name used by this provider"""
        pass

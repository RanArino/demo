# Disabled for now due to removal of sentence-transformers dependency
# import logging
# from typing import List

# from .base import BaseEmbeddingProvider


# class LocalEmbeddingProvider(BaseEmbeddingProvider):
#     """Local embeddings provider using sentence-transformers"""

#     def __init__(self, model_name: str = "all-MiniLM-L6-v2"):
#         self.model_name = model_name

#         try:
#             from sentence_transformers import SentenceTransformer
#             self.model = SentenceTransformer(model_name)
#             self.available = True
#         except ImportError:
#             logging.warning("sentence-transformers not available. Install with: pip install sentence-transformers")
#             self.model = None
#             self.available = False

#     def embed_query(self, text: str, **kwargs) -> List[float]:
#         """Embed a single text query"""
#         if not self.available:
#             raise RuntimeError("Local provider not available. Install sentence-transformers.")

#         if not text.strip():
#             raise ValueError("Text cannot be empty")

#         try:
#             embedding = self.model.encode(text)
#             return embedding.tolist()
#         except Exception as e:
#             logging.error(f"Local embedding failed: {e}")
#             raise RuntimeError(f"Failed to generate local embedding: {e}") from e

#     def embed_documents(self, texts: List[str], **kwargs) -> List[List[float]]:
#         """Embed multiple documents in batch"""
#         if not self.available:
#             raise RuntimeError("Local provider not available. Install sentence-transformers.")

#         if not texts:
#             return []

#         try:
#             embeddings = self.model.encode(texts)
#             return embeddings.tolist()
#         except Exception as e:
#             logging.error(f"Local batch embedding failed: {e}")
#             raise RuntimeError(f"Failed to generate local batch embeddings: {e}") from e

#     def get_dimensions(self) -> int:
#         """Get the dimensionality of embeddings from this provider"""
#         if not self.available:
#             return 0
#         return self.model.get_sentence_embedding_dimension()

#     def get_model_name(self) -> str:
#         """Get the model name used by this provider"""
#         return self.model_name

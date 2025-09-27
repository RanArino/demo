import logging
from typing import List, Iterator

from ..services.chunking import chunk_text, ChunkingConfig, Chunk as PyChunk
from ..services.embedding import embed_chunks, EmbeddingConfig


class ChunkEmbedding:
    def __init__(self, chunk: PyChunk, vector: List[float]):
        self.chunk = chunk
        self.vector = vector


class ChunkAndEmbedResult:
    def __init__(self, results: List[ChunkEmbedding], dims: int, model_id: str, model_version: str):
        self.results = results
        self.dims = dims
        self.model_id = model_id
        self.model_version = model_version


def chunk_and_embed(*, text: str | None, blob_url: str | None, chunk_cfg: ChunkingConfig, embed_cfg: EmbeddingConfig) -> ChunkAndEmbedResult:
    chunks = chunk_text(text=text, blob_url=blob_url, config=chunk_cfg)
    if not chunks:
        return ChunkAndEmbedResult(results=[], dims=0, model_id=embed_cfg.model_id, model_version=embed_cfg.model_version or embed_cfg.model_id)

    contents = [c.content for c in chunks]
    emb = embed_chunks(contents, embed_cfg)

    vectors = emb.vectors.tolist() if hasattr(emb.vectors, "tolist") else []
    if len(vectors) != len(chunks):
        logging.warning("Embedding count (%d) does not match chunks (%d)", len(vectors), len(chunks))
    n = min(len(vectors), len(chunks))
    results: List[ChunkEmbedding] = []
    for i in range(n):
        results.append(ChunkEmbedding(chunk=chunks[i], vector=vectors[i]))

    return ChunkAndEmbedResult(results=results, dims=emb.dims, model_id=emb.model_id, model_version=emb.model_version)


def chunk_and_embed_batched(*, text: str | None, blob_url: str | None, chunk_cfg: ChunkingConfig, embed_cfg: EmbeddingConfig, batch_size: int) -> Iterator[ChunkAndEmbedResult]:
    """
    Process text in batches, yielding ChunkAndEmbedResult for each batch.
    For demo phase, this returns a single batch with all results.
    In future, this can be extended to yield true streaming batches.
    """
    # For now, process all at once but structure for future batching
    chunks = chunk_text(text=text, blob_url=blob_url, config=chunk_cfg)
    if not chunks:
        yield ChunkAndEmbedResult(results=[], dims=0, model_id=embed_cfg.model_id, model_version=embed_cfg.model_version or embed_cfg.model_id)
        return

    # For demo phase, process all chunks at once but respect the logical batch structure
    contents = [c.content for c in chunks]
    emb = embed_chunks(contents, embed_cfg)

    vectors = emb.vectors.tolist() if hasattr(emb.vectors, "tolist") else []
    if len(vectors) != len(chunks):
        logging.warning("Embedding count (%d) does not match chunks (%d)", len(vectors), len(chunks))

    n = min(len(vectors), len(chunks))

    # Split results into batches
    for i in range(0, n, batch_size):
        batch_end = min(i + batch_size, n)
        batch_results: List[ChunkEmbedding] = []

        for j in range(i, batch_end):
            batch_results.append(ChunkEmbedding(chunk=chunks[j], vector=vectors[j]))

        yield ChunkAndEmbedResult(
            results=batch_results,
            dims=emb.dims,
            model_id=emb.model_id,
            model_version=emb.model_version
        )



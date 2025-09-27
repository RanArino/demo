import os
import logging
from concurrent import futures

import grpc
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

from .config import settings
from .services.chunking import ChunkingConfig
from .services.embedding import embed_query, EmbeddingConfig
from .pipelines.chunk_and_embed import chunk_and_embed, chunk_and_embed_batched
from .proto.private.v1 import (
    canvas_private_pb2 as canvas_pb2,
    canvas_private_pb2_grpc as canvas_pb2_grpc,
)   


def _setup_logging():
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")


def _setup_tracing():
    resource = Resource.create({"service.name": "ms_canvas_python"})
    provider = TracerProvider(resource=resource)
    processor = BatchSpanProcessor(OTLPSpanExporter())
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)


class CanvasInternalServicer(canvas_pb2_grpc.CanvasInternalServicer):
    def EmbedQuery(self, request, context):
        try:
            config = EmbeddingConfig(
                provider=request.config.provider or "huggingface",
                model_id=request.config.model_id or "all-MiniLM-L6-v2",
                model_version=request.config.model_version or ""
            )
            
            if not request.text.strip():
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("Empty query text provided")
                return canvas_pb2.EmbedQueryResponse()
            
            result = embed_query(request.text, config)
            
            response = canvas_pb2.EmbedQueryResponse()
            if len(result.vectors) > 0:
                response.vector.extend(result.vectors[0].tolist())
            response.dims = result.dims
            response.model_id = result.model_id
            response.model_version = result.model_version
            
            return response
            
        except Exception as e:
            logging.error(f"EmbedQuery failed: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Query embedding failed: {str(e)}")
            return canvas_pb2.EmbedQueryResponse()

    def ChunkEmbed(self, request, context):
        try:
            chunk_cfg = ChunkingConfig(
                target_tokens=request.chunking.target_tokens or settings.CANVAS_CHUNK_TARGET_TOKENS,
                overlap_percent=request.chunking.overlap_percent or settings.CANVAS_CHUNK_OVERLAP_PERCENT,
                tokenizer=request.chunking.tokenizer or settings.CANVAS_TOKENIZER,
            )
            embed_cfg = EmbeddingConfig(
                provider=request.embedding.provider or "huggingface",
                model_id=request.embedding.model_id or "all-MiniLM-L6-v2",
                model_version=request.embedding.model_version or "",
            )

            text = None
            blob_url = None
            if request.WhichOneof('source') == 'text':
                text = request.text
            elif request.WhichOneof('source') == 'blob_url':
                blob_url = request.blob_url

            # Check if batch_size is provided in request
            batch_size = request.batch_size if request.batch_size > 0 else settings.CANVAS_BATCH_SIZE

            # For demo phase, use regular unary response but structure for future batching
            if batch_size and batch_size < 1000:  # Use batching for reasonable batch sizes
                # For now, collect all batches into single response (demo phase)
                all_results = []
                dims = 0
                model_id = ""
                model_version = ""

                for batch_result in chunk_and_embed_batched(
                    text=text, blob_url=blob_url,
                    chunk_cfg=chunk_cfg, embed_cfg=embed_cfg,
                    batch_size=batch_size
                ):
                    all_results.extend(batch_result.results)
                    dims = batch_result.dims
                    model_id = batch_result.model_id
                    model_version = batch_result.model_version

                response = canvas_pb2.ChunkEmbedResponse()
                response.dims = dims
                response.model_id = model_id
                response.model_version = model_version
                for item in all_results:
                    res = response.results.add()
                    res.chunk.sequence_index = item.chunk.position
                    res.chunk.start_position = item.chunk.start_position
                    res.chunk.end_position = item.chunk.end_position
                    res.chunk.content = item.chunk.content
                    res.vector.extend(item.vector)
            else:
                # Use original non-batched approach for large batch sizes or when not specified
                result = chunk_and_embed(text=text, blob_url=blob_url, chunk_cfg=chunk_cfg, embed_cfg=embed_cfg)

                response = canvas_pb2.ChunkEmbedResponse()
                response.dims = result.dims
                response.model_id = result.model_id
                response.model_version = result.model_version
                for item in result.results:
                    res = response.results.add()
                    res.chunk.sequence_index = item.chunk.position
                    res.chunk.start_position = item.chunk.start_position
                    res.chunk.end_position = item.chunk.end_position
                    res.chunk.content = item.chunk.content
                    res.vector.extend(item.vector)

            return response
        except Exception as e:
            logging.error(f"ChunkEmbed failed: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"ChunkEmbed failed: {str(e)}")
            return canvas_pb2.ChunkEmbedResponse()
    
    def Healthz(self, request, context):
        response = canvas_pb2.HealthStatus()
        response.status = "OK"
        response.components["chunking"] = "healthy"
        response.components["embedding"] = "healthy"
        return response


def main():
    _setup_logging()
    _setup_tracing()

    host = os.getenv("CANVAS_PY_HOST", "0.0.0.0")
    port = int(os.getenv("CANVAS_PY_PORT", "50054"))
    max_msg = int(os.getenv("CANVAS_MAX_GRPC_MSG_BYTES", str(64 * 1024 * 1024)))

    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=4),
        options=[
            ("grpc.max_send_message_length", max_msg),
            ("grpc.max_receive_message_length", max_msg),
        ],
    )

    canvas_pb2_grpc.add_CanvasInternalServicer_to_server(CanvasInternalServicer(), server)

    server.add_insecure_port(f"{host}:{port}")
    logging.info("Python gRPC server listening on %s:%s", host, port)
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    main()



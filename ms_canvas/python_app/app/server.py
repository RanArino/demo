import os
import logging
from concurrent import futures

import grpc
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

from .services.chunking import chunk_text, ChunkingConfig
from .services.embedding import embed_chunks, embed_query, EmbeddingConfig
from .proto.v1 import canvas_pb2, canvas_pb2_grpc


def _setup_logging():
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")


def _setup_tracing():
    resource = Resource.create({"service.name": "ms_canvas_python"})
    provider = TracerProvider(resource=resource)
    processor = BatchSpanProcessor(OTLPSpanExporter())
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)


class CanvasInternalServicer(canvas_pb2_grpc.CanvasInternalServicer):
    
    def ChunkText(self, request, context):
        try:
            config = ChunkingConfig(
                target_tokens=request.config.target_tokens or 300,
                overlap_percent=request.config.overlap_percent or 10,
                tokenizer=request.config.tokenizer or "tiktoken:cl100k_base"
            )
            
            text = None
            blob_url = None
            
            if request.WhichOneof('source') == 'text':
                text = request.text
            elif request.WhichOneof('source') == 'blob_url':
                blob_url = request.blob_url
            
            chunks = chunk_text(text=text, blob_url=blob_url, config=config)
            
            response = canvas_pb2.ChunkTextResponse()
            for chunk in chunks:
                pb_chunk = response.chunks.add()
                pb_chunk.position = chunk.position
                pb_chunk.start_position = chunk.start_position
                pb_chunk.end_position = chunk.end_position
                pb_chunk.content = chunk.content
            
            return response
            
        except Exception as e:
            logging.error(f"ChunkText failed: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Text chunking failed: {str(e)}")
            return canvas_pb2.ChunkTextResponse()
    
    def EmbedChunks(self, request, context):
        try:
            config = EmbeddingConfig(
                provider=request.config.provider or "huggingface",
                model_id=request.config.model_id or "all-MiniLM-L6-v2",
                model_version=request.config.model_version or ""
            )
            
            contents = list(request.contents) if request.contents else []
            
            if not contents:
                context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                context.set_details("No contents provided for embedding")
                return canvas_pb2.EmbedChunksResponse()
            
            result = embed_chunks(contents, config)
            
            response = canvas_pb2.EmbedChunksResponse()
            response.flat_vectors.extend(result.vectors.flatten().tolist())
            response.dims = result.dims
            response.model_id = result.model_id
            response.model_version = result.model_version
            
            return response
            
        except Exception as e:
            logging.error(f"EmbedChunks failed: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Embedding failed: {str(e)}")
            return canvas_pb2.EmbedChunksResponse()
    
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
    port = int(os.getenv("CANVAS_PY_PORT", "50051"))
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



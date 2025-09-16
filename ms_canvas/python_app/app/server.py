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


def _setup_logging():
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")


def _setup_tracing():
    resource = Resource.create({"service.name": "ms_canvas_python"})
    provider = TracerProvider(resource=resource)
    processor = BatchSpanProcessor(OTLPSpanExporter())
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)


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

    # TODO: register generated CanvasInternalServicer here after codegen
    # add_CanvasInternalServicer_to_server(ServicerImpl(), server)

    server.add_insecure_port(f"{host}:{port}")
    logging.info("Python gRPC server listening on %s:%s", host, port)
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    main()



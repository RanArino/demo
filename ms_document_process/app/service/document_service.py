import logging
import time
from uuid import UUID
import hashlib

from app.domain.events import DocumentUploadedEvent, DocumentProcessedEvent
from app.repository.document_repository import DocumentRepository
from app.service.pdf_conversion_service import convert_document_to_markdown
from app.event.producer import KafkaProducer

logger = logging.getLogger(__name__)

# Encoding constant for consistent text encoding throughout the service
ENCODING = 'utf-8'


def _derive_processed_key_from_source(source_key: str) -> str:
    """Replace the filename extension with .md or append .md if none."""
    if not source_key:
        return "processed.md"
    # Find last '/' to isolate filename
    slash = source_key.rfind('/')
    if slash >= 0:
        prefix = source_key[:slash+1]
        filename = source_key[slash+1:]
    else:
        prefix = ""
        filename = source_key
    dot = filename.rfind('.')
    if dot > 0:
        md_name = filename[:dot] + '.md'
    else:
        md_name = filename + '.md'
    return prefix + md_name


class DocumentProcessService:
    def __init__(self, document_repository: DocumentRepository, kafka_producer: KafkaProducer):
        self.document_repository = document_repository
        self.kafka_producer = kafka_producer
        self.max_retries = 3
        self.initial_backoff_seconds = 0.5

    def _retry_with_backoff(self, operation_name: str, operation):
        attempt = 0
        backoff = self.initial_backoff_seconds
        last_exception = None
        while attempt < self.max_retries:
            try:
                return operation()
            except Exception as exc:
                last_exception = exc
                attempt += 1
                if attempt >= self.max_retries:
                    break
                logger.warning(
                    f"{operation_name} failed (attempt {attempt}/{self.max_retries}). Retrying in {backoff:.2f}s...",
                    exc_info=True,
                )
                time.sleep(backoff)
                backoff *= 2
        raise last_exception

    def process_document(self, event: DocumentUploadedEvent):
        logger.info(f"Processing document for content_source_id: {event.content_source_id}")
        try:
            # Determine key to fetch original (prefer object key if present)
            source_key = event.original_object_key if getattr(event, 'original_object_key', None) else event.original_blob_hash

            # Download the document from R2 (with retries)
            document_content = self._retry_with_backoff(
                "download_source_document",
                lambda: self.document_repository.download_source_document(source_key),
            )

            # Convert the document to Markdown
            # Assuming the source format is pdf for now. This might need to be more flexible.
            # Convert document (with retries to tolerate transient converter issues)
            markdown_content = self._retry_with_backoff(
                "convert_document_to_markdown",
                lambda: convert_document_to_markdown(document_content, "pdf"),
            )

            # Calculate the hash of the processed content (metadata)
            processed_blob_hash = hashlib.sha256(markdown_content.encode(ENCODING)).hexdigest()

            # Upload the processed document to R2 next to original (with retries)
            if '/' in source_key:
                processed_key = _derive_processed_key_from_source(source_key)
            else:
                # Legacy fallback: use the same key with .md appended
                processed_key = source_key + '.md'

            self._retry_with_backoff(
                "upload_processed_document",
                lambda: self.document_repository.upload_processed_document(
                    processed_key, markdown_content.encode(ENCODING)
                ),
            )

            # Produce a success event
            processed_event = DocumentProcessedEvent(
                content_source_id=event.content_source_id,
                processed_blob_hash=processed_blob_hash,
                status="PROCESSED",
                error_message=None,
            )
            self.kafka_producer.produce_document_processed_event(processed_event)
            logger.info(f"Successfully processed document for content_source_id: {event.content_source_id}")

        except Exception as e:
            logger.error(f"Processing failed for content_source_id: {event.content_source_id}: {e}", exc_info=True)
            failed_event = DocumentProcessedEvent(
                content_source_id=event.content_source_id,
                processed_blob_hash=None,
                status="FAILED",
                error_message=str(e),
            )
            self.kafka_producer.produce_document_processed_event(failed_event)

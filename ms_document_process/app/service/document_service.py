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

            # Calculate the hash of the processed content
            processed_blob_hash = hashlib.sha256(markdown_content.encode(ENCODING)).hexdigest()

            # Upload the processed document to R2
            # Upload the processed document (with retries)
            self._retry_with_backoff(
                "upload_processed_document",
                lambda: self.document_repository.upload_processed_document(
                    processed_blob_hash, markdown_content.encode(ENCODING)
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

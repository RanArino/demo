import hashlib
import logging
import time
from concurrent.futures import Future, ThreadPoolExecutor
from typing import Optional

from app.domain.events import DocumentProcessedEvent, DocumentUploadedEvent
from app.domain.insights import DocumentInsights
from app.event.producer import KafkaProducer
from app.repository.document_repository import DocumentRepository
from app.service.insights_service import DocumentInsightsService
from app.service.pdf_conversion_service import convert_document_to_markdown

logger = logging.getLogger(__name__)

# Encoding constant for consistent text encoding throughout the service
ENCODING = 'utf-8'


class DocumentProcessService:
    def __init__(
        self,
        document_repository: DocumentRepository,
        kafka_producer: KafkaProducer,
        insights_service: Optional[DocumentInsightsService] = None,
    ):
        self.document_repository = document_repository
        self.kafka_producer = kafka_producer
        self.insights_service = insights_service
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
            if not event.original_object_key:
                logger.error("Missing original_object_key for content_source_id: %s", event.content_source_id)
                raise ValueError(
                    f"original_object_key is required for processing content_source_id: {event.content_source_id}"
                )
            source_key = event.original_object_key

            document_content = self._retry_with_backoff(
                "download_source_document",
                lambda: self.document_repository.download_source_document(source_key),
            )

            conversion_workers = 2 + int(self.insights_service is not None)

            with ThreadPoolExecutor(max_workers=conversion_workers) as executor:
                conversion_future = executor.submit(
                    self._retry_with_backoff,
                    "convert_document_to_markdown",
                    lambda: convert_document_to_markdown(document_content, "pdf"),
                )

                upload_future = executor.submit(
                    self._upload_after_conversion,
                    conversion_future,
                    source_key,
                )

                insights_future: Optional[Future[Optional[DocumentInsights]]] = None
                if self.insights_service is not None:
                    insights_future = executor.submit(
                        self._insights_after_conversion,
                        conversion_future,
                        event.title,
                    )

                markdown_content = upload_future.result()
                processed_blob_hash = hashlib.sha256(markdown_content.encode(ENCODING)).hexdigest()
                insights = insights_future.result() if insights_future is not None else None

            processed_event = DocumentProcessedEvent(
                content_source_id=event.content_source_id,
                space_id=event.space_id,
                processed_blob_hash=processed_blob_hash,
                status="PROCESSED",
                error_message=None,
                title=event.title,
                summary=insights.summary if insights else None,
                keywords=insights.keywords if insights else None,
            )
            self.kafka_producer.produce_document_processed_event(processed_event)
            logger.info("Successfully processed document for content_source_id: %s", event.content_source_id)

        except Exception as e:
            logger.error("Processing failed for content_source_id: %s: %s", event.content_source_id, e, exc_info=True)
            failed_event = DocumentProcessedEvent(
                content_source_id=event.content_source_id,
                space_id=event.space_id,
                processed_blob_hash=None,
                status="FAILED",
                error_message=str(e),
                title=event.title,
                summary=None,
                keywords=None,
            )
            self.kafka_producer.produce_document_processed_event(failed_event)

    def _generate_insights_safe(
        self,
        document_text: str,
        title: Optional[str],
    ) -> Optional[DocumentInsights]:
        if self.insights_service is None:
            return None
        try:
            return self.insights_service.generate_insights(document_text=document_text, title=title)
        except Exception:
            logger.warning("Failed to generate document insights", exc_info=True)
            return None

    def _upload_after_conversion(
        self,
        conversion_future: Future[str],
        source_key: str,
    ) -> str:
        markdown_content = conversion_future.result()

        def upload_operation():
            return self.document_repository.upload_processed_document(
                source_key, markdown_content.encode(ENCODING)
            )

        self._retry_with_backoff(
            "upload_processed_document",
            upload_operation,
        )
        return markdown_content

    def _insights_after_conversion(
        self,
        conversion_future: Future[str],
        title: Optional[str],
    ) -> Optional[DocumentInsights]:
        markdown_content = conversion_future.result()
        return self._generate_insights_safe(markdown_content, title)

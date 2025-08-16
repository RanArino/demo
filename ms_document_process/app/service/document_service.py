import logging
from uuid import UUID
import hashlib

from app.domain.events import DocumentUploadedEvent, DocumentProcessedEvent
from app.repository.document_repository import DocumentRepository
from app.service.pdf_conversion_service import convert_document_to_markdown
from app.event.producer import KafkaProducer

logger = logging.getLogger(__name__)

class DocumentProcessService:
    def __init__(self, document_repository: DocumentRepository, kafka_producer: KafkaProducer):
        self.document_repository = document_repository
        self.kafka_producer = kafka_producer

    def process_document(self, event: DocumentUploadedEvent):
        logger.info(f"Processing document for content_source_id: {event.content_source_id}")
        try:
            # Download the document from R2
            document_content = self.document_repository.download_source_document(event.original_blob_hash)

            # Convert the document to Markdown
            # Assuming the source format is pdf for now. This might need to be more flexible.
            markdown_content = convert_document_to_markdown(document_content, "pdf")

            # Calculate the hash of the processed content
            processed_blob_hash = hashlib.sha256(markdown_content.encode('utf-8')).hexdigest()

            # Upload the processed document to R2
            self.document_repository.upload_processed_document(processed_blob_hash, markdown_content.encode('utf-8'))

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
            logger.error(f"Failed to process document for content_source_id: {event.content_source_id}", exc_info=True)
            # Produce a failure event
            processed_event = DocumentProcessedEvent(
                content_source_id=event.content_source_id,
                processed_blob_hash=None,
                status="FAILED",
                error_message=str(e),
            )
            self.kafka_producer.produce_document_processed_event(processed_event)

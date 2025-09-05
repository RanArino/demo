import unittest
from unittest.mock import MagicMock, patch
from uuid import uuid4
import hashlib

from app.domain.events import DocumentUploadedEvent, DocumentProcessedEvent
from app.service.document_service import DocumentProcessService

class TestDocumentProcessService(unittest.TestCase):

    def setUp(self):
        self.mock_document_repository = MagicMock()
        self.mock_kafka_producer = MagicMock()
        self.service = DocumentProcessService(
            document_repository=self.mock_document_repository,
            kafka_producer=self.mock_kafka_producer
        )

    @patch('app.service.document_service.convert_document_to_markdown')
    def test_process_document_success(self, mock_convert):
        # Arrange
        content_source_id = uuid4()
        space_id = uuid4()
        owner_id = uuid4()
        filename = "test-document.pdf"
        object_key = f"{owner_id}/spaces/{space_id}/content/{content_source_id}/{filename}"
        
        event = DocumentUploadedEvent(
            content_source_id=content_source_id,
            original_blob_hash='original_hash',
            space_id=space_id,
            original_object_key=object_key
        )
        document_content = b'pdf content'
        markdown_content = '# Markdown'
        processed_hash = hashlib.sha256(markdown_content.encode('utf-8')).hexdigest()

        self.mock_document_repository.download_source_document.return_value = document_content
        mock_convert.return_value = markdown_content

        # Act
        self.service.process_document(event)

        # Assert
        self.mock_document_repository.download_source_document.assert_called_once_with(object_key)
        mock_convert.assert_called_once_with(document_content, 'pdf')
        self.mock_document_repository.upload_processed_document.assert_called_once_with(
            object_key, markdown_content.encode('utf-8')
        )
        
        # Check that the success event was produced
        self.mock_kafka_producer.produce_document_processed_event.assert_called_once()
        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertIsInstance(produced_event, DocumentProcessedEvent)
        self.assertEqual(produced_event.status, "PROCESSED")
        self.assertEqual(produced_event.processed_blob_hash, processed_hash)

    def test_process_document_failure(self):
        # Arrange
        content_source_id = uuid4()
        space_id = uuid4()
        owner_id = uuid4()
        filename = "test-document.pdf"
        object_key = f"{owner_id}/spaces/{space_id}/content/{content_source_id}/{filename}"
        
        event = DocumentUploadedEvent(
            content_source_id=content_source_id,
            original_blob_hash='original_hash',
            space_id=space_id,
            original_object_key=object_key
        )
        self.mock_document_repository.download_source_document.side_effect = Exception("Download failed")

        # Act
        self.service.process_document(event)

        # Assert
        # Check that the failure event was produced
        self.mock_kafka_producer.produce_document_processed_event.assert_called_once()
        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertIsInstance(produced_event, DocumentProcessedEvent)
        self.assertEqual(produced_event.status, "FAILED")

    def test_process_document_missing_object_key(self):
        # Arrange - event without original_object_key
        event = DocumentUploadedEvent(
            content_source_id=uuid4(),
            original_blob_hash='original_hash',
            space_id=uuid4()
            # original_object_key is intentionally omitted
        )

        # Act
        self.service.process_document(event)

        # Assert - should produce failure event due to missing object key
        self.mock_kafka_producer.produce_document_processed_event.assert_called_once()
        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertIsInstance(produced_event, DocumentProcessedEvent)
        self.assertEqual(produced_event.status, "FAILED")
        self.assertIn("original_object_key is required", produced_event.error_message)

if __name__ == '__main__':
    unittest.main()

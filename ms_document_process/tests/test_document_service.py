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
        event = DocumentUploadedEvent(
            content_source_id=uuid4(),
            original_blob_hash='original_hash',
            space_id=uuid4()
        )
        document_content = b'pdf content'
        markdown_content = '# Markdown'
        processed_hash = hashlib.sha256(markdown_content.encode('utf-8')).hexdigest()

        self.mock_document_repository.download_source_document.return_value = document_content
        mock_convert.return_value = markdown_content

        # Act
        self.service.process_document(event)

        # Assert
        self.mock_document_repository.download_source_document.assert_called_once_with('original_hash')
        mock_convert.assert_called_once_with(document_content, 'pdf')
        self.mock_document_repository.upload_processed_document.assert_called_once_with(
            processed_hash, markdown_content.encode('utf-8')
        )
        
        # Check that the success event was produced
        self.mock_kafka_producer.produce_document_processed_event.assert_called_once()
        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertIsInstance(produced_event, DocumentProcessedEvent)
        self.assertEqual(produced_event.status, "PROCESSED")
        self.assertEqual(produced_event.processed_blob_hash, processed_hash)

    def test_process_document_failure(self):
        # Arrange
        event = DocumentUploadedEvent(
            content_source_id=uuid4(),
            original_blob_hash='original_hash',
            space_id=uuid4()
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
        self.assertIsNotNone(produced_event.error_message)

if __name__ == '__main__':
    unittest.main()

import unittest
from unittest.mock import MagicMock, patch
from uuid import uuid4
import hashlib

from app.domain.events import DocumentUploadedEvent, DocumentProcessedEvent
from app.service.document_service import DocumentProcessService
from app.service.insights_service import DocumentInsightsService


class TestDocumentProcessService(unittest.TestCase):

    def setUp(self):
        self.mock_document_repository = MagicMock()
        self.mock_kafka_producer = MagicMock()
        self.mock_insights_service = MagicMock(spec=DocumentInsightsService)
        self.service = DocumentProcessService(
            document_repository=self.mock_document_repository,
            kafka_producer=self.mock_kafka_producer,
            insights_service=self.mock_insights_service,
        )

    @patch('app.service.document_service.convert_document_to_markdown')
    def test_process_document_success(self, mock_convert):
        content_source_id = uuid4()
        space_id = uuid4()
        owner_id = uuid4()
        filename = "test-document.pdf"
        object_key = f"{owner_id}/spaces/{space_id}/content/{content_source_id}/{filename}"

        event = DocumentUploadedEvent(
            content_source_id=content_source_id,
            original_blob_hash='original_hash',
            space_id=space_id,
            original_object_key=object_key,
            title="Test Document",
        )
        document_content = b'pdf content'
        markdown_content = '# Markdown'
        processed_hash = hashlib.sha256(markdown_content.encode('utf-8')).hexdigest()

        self.mock_document_repository.download_source_document.return_value = document_content
        mock_convert.return_value = markdown_content
        self.mock_insights_service.generate_insights.return_value.summary = "summary"
        self.mock_insights_service.generate_insights.return_value.keywords = ["kw1", "kw2"]
        self.mock_document_repository.upload_processed_document.return_value = object_key

        self.service.process_document(event)

        self.mock_document_repository.download_source_document.assert_called_once_with(object_key)
        mock_convert.assert_called_once_with(document_content, 'pdf')
        self.mock_document_repository.upload_processed_document.assert_called_once_with(
            object_key, markdown_content.encode('utf-8')
        )
        self.mock_kafka_producer.produce_document_processed_event.assert_called_once()
        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertEqual(produced_event.processed_object_key, object_key)

        self.mock_insights_service.generate_insights.assert_called_once_with(document_text=markdown_content, title="Test Document")

        self.mock_kafka_producer.produce_document_processed_event.assert_called_once()
        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertIsInstance(produced_event, DocumentProcessedEvent)
        self.assertEqual(produced_event.status, "PROCESSED")
        self.assertEqual(produced_event.processed_blob_hash, processed_hash)
        self.assertEqual(produced_event.processed_object_key, object_key)
        self.assertEqual(produced_event.summary, "summary")
        self.assertEqual(produced_event.keywords, ["kw1", "kw2"])

    def test_process_document_failure(self):
        content_source_id = uuid4()
        space_id = uuid4()
        owner_id = uuid4()
        filename = "test-document.pdf"
        object_key = f"{owner_id}/spaces/{space_id}/content/{content_source_id}/{filename}"

        event = DocumentUploadedEvent(
            content_source_id=content_source_id,
            original_blob_hash='original_hash',
            space_id=space_id,
            original_object_key=object_key,
            title="Test Document",
        )
        self.mock_document_repository.download_source_document.side_effect = Exception("Download failed")
        self.mock_document_repository.upload_processed_document.return_value = object_key

        self.service.process_document(event)

        self.mock_kafka_producer.produce_document_processed_event.assert_called_once()
        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertIsInstance(produced_event, DocumentProcessedEvent)
        self.assertEqual(produced_event.status, "FAILED")
        self.assertIsNone(produced_event.summary)
        self.assertIsNone(produced_event.keywords)

    def test_process_document_missing_object_key(self):
        event = DocumentUploadedEvent(
            content_source_id=uuid4(),
            original_blob_hash='original_hash',
            space_id=uuid4(),
            title="Missing Key Document",
        )

        self.service.process_document(event)

        self.mock_kafka_producer.produce_document_processed_event.assert_called_once()
        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertIsInstance(produced_event, DocumentProcessedEvent)
        self.assertEqual(produced_event.status, "FAILED")
        self.assertIn("original_object_key is required", produced_event.error_message)

    @patch('app.service.document_service.convert_document_to_markdown')
    def test_process_document_success_without_insights(self, mock_convert):
        service = DocumentProcessService(
            document_repository=self.mock_document_repository,
            kafka_producer=self.mock_kafka_producer,
            insights_service=None,
        )
        content_source_id = uuid4()
        space_id = uuid4()
        owner_id = uuid4()
        filename = "test-document.pdf"
        object_key = f"{owner_id}/spaces/{space_id}/content/{content_source_id}/{filename}"
        event = DocumentUploadedEvent(
            content_source_id=content_source_id,
            original_blob_hash='original_hash',
            space_id=space_id,
            original_object_key=object_key,
            title="Test Document",
        )
        document_content = b'pdf content'
        markdown_content = '# Markdown'
        processed_hash = hashlib.sha256(markdown_content.encode('utf-8')).hexdigest()

        self.mock_document_repository.download_source_document.return_value = document_content
        mock_convert.return_value = markdown_content
        self.mock_document_repository.upload_processed_document.return_value = object_key

        service.process_document(event)

        produced_event = self.mock_kafka_producer.produce_document_processed_event.call_args[0][0]
        self.assertIsNone(produced_event.summary)
        self.assertIsNone(produced_event.keywords)
        self.assertEqual(produced_event.processed_blob_hash, processed_hash)
        self.assertEqual(produced_event.processed_object_key, object_key)


if __name__ == '__main__':
    unittest.main()

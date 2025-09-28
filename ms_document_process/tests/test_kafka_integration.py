import os
import unittest
import json
import threading
import time
from contextlib import contextmanager
from uuid import uuid4

from confluent_kafka import Producer, Consumer
from confluent_kafka.admin import AdminClient, NewTopic

from app.config.config import settings
from app.config.topics import get_document_uploaded_topic, get_document_processed_topic
from app.domain.events import DocumentUploadedEvent, DocumentProcessedEvent
from app.service.document_service import DocumentProcessService
from app.event.producer import KafkaProducer as ServiceKafkaProducer
from unittest.mock import patch


class InMemoryDocumentRepository:
    def __init__(self):
        self.storage = {}

    def download_source_document(self, original_blob_hash: str) -> bytes:
        # Return fixed bytes; in a real test, this would point to a fixture file
        return b"dummy pdf bytes"

    def upload_processed_document(self, processed_blob_hash: str, content_bytes: bytes) -> None:
        self.storage[processed_blob_hash] = content_bytes


class FakeInsightsService:
    def __init__(self):
        self.summary = "This is a summary"
        self.keywords = ["keyword1", "keyword2"]

    def generate_insights(self, document_text: str, title: str | None = None):
        return type("Insights", (), {"summary": self.summary, "keywords": self.keywords})()


@contextmanager
def kafka_topics(admin: AdminClient, topics: list[str]):
    try:
        new_topics = [NewTopic(t, num_partitions=1, replication_factor=1) for t in topics]
        # Ignore errors if already exist
        admin.create_topics(new_topics)
        time.sleep(0.5)
        yield
    finally:
        try:
            admin.delete_topics(topics)
        except Exception:
            pass


@unittest.skipUnless(os.getenv("RUN_KAFKA_INTEGRATION") == "1", "set RUN_KAFKA_INTEGRATION=1 to run against a real Kafka broker")
class TestKafkaIntegration(unittest.TestCase):
    def setUp(self):
        if not settings.kafka_brokers:
            self.skipTest("Kafka brokers not configured")

        settings.kafka_brokers = "localhost:9092"
        settings.kafka_security_protocol = "PLAINTEXT"
        settings.kafka_sasl_username = ""
        settings.kafka_sasl_password = ""

        self.admin = AdminClient({'bootstrap.servers': settings.kafka_brokers})
        self.source_topic = get_document_uploaded_topic()
        self.processed_topic = get_document_processed_topic()

        self.producer = Producer({'bootstrap.servers': settings.kafka_brokers})
        self.processed_consumer = Consumer({
            'bootstrap.servers': settings.kafka_brokers,
            'group.id': f'test-group-{uuid4()}',
            'auto.offset.reset': 'earliest'
        })
        self.processed_consumer.subscribe([self.processed_topic])

        self.repo = InMemoryDocumentRepository()
        self.service_producer = ServiceKafkaProducer()
        self.service = DocumentProcessService(self.repo, self.service_producer, FakeInsightsService())

    def tearDown(self):
        self.processed_consumer.close()

    def test_end_to_end_processes_message_and_emits_processed_event(self):
        with kafka_topics(self.admin, [self.source_topic, self.processed_topic]):
            content_source_id = uuid4()
            space_id = uuid4()
            owner_id = uuid4()
            filename = "test-document.pdf"
            
            event = DocumentUploadedEvent(
                content_source_id=content_source_id,
                original_blob_hash='original_hash',
                space_id=space_id,
                original_object_key=f"{owner_id}/spaces/{space_id}/content/{content_source_id}/{filename}",
                title="Integration Test Document",
            )

            # Patch conversion to avoid external PDF tooling
            with patch('app.service.document_service.convert_document_to_markdown', return_value="# MD"):
                # Instead of depending on the long-lived KafkaConsumer loop, directly invoke service logic
                # to validate producer wiring and message format to processed topic.
                self.service.process_document(event)

            # Assert the processed message is produced and consumable
            msg = self.processed_consumer.poll(10.0)
            self.assertIsNotNone(msg)
            self.assertIsNone(msg.error())

            consumed_event_data = json.loads(msg.value().decode('utf-8'))
            consumed_event = DocumentProcessedEvent(**consumed_event_data)

            self.assertEqual(consumed_event.content_source_id, event.content_source_id)
            self.assertEqual(consumed_event.status, "PROCESSED")
            self.assertIsNotNone(consumed_event.processed_blob_hash)
            self.assertEqual(consumed_event.processed_object_key, event.original_object_key)
            self.assertEqual(consumed_event.summary, "This is a summary")
            self.assertEqual(consumed_event.keywords, ["keyword1", "keyword2"])


if __name__ == '__main__':
    unittest.main()

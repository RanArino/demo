import json
import logging
import threading
import time
from uuid import UUID
from datetime import datetime
from confluent_kafka import Producer
from app.config.config import settings
from app.config.topics import get_document_processed_topic
from app.domain.events import DocumentProcessedEvent

logger = logging.getLogger(__name__)

class CustomJSONEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, UUID):
            return str(obj)
        elif isinstance(obj, datetime):
            return obj.isoformat()
        return super().default(obj)

class KafkaProducer:
    def __init__(self):
        producer_config = {
            'bootstrap.servers': settings.kafka_brokers,
            'security.protocol': settings.kafka_security_protocol,
        }
        if settings.kafka_security_protocol == 'SASL_SSL':
            producer_config.update({
                'sasl.mechanisms': 'PLAIN',
                'sasl.username': settings.kafka_sasl_username,
                'sasl.password': settings.kafka_sasl_password,
            })
        self.producer = Producer(producer_config)
        self.topic = get_document_processed_topic()

    def produce_document_processed_event(self, event: DocumentProcessedEvent):
        message = json.dumps(
            event.model_dump(),
            cls=CustomJSONEncoder
        ).encode('utf-8')

        delivery_error = None
        delivery_event = threading.Event()

        def delivery_report(err, msg):
            nonlocal delivery_error
            if err is not None:
                delivery_error = err
            delivery_event.set()

        max_retries = 3
        backoff = 0.5
        delivery_timeout_seconds = 3.0
        for attempt in range(1, max_retries + 1):
            delivery_error = None  # Reset error state before each attempt
            delivery_event.clear()
            try:
                self.producer.produce(self.topic, value=message, on_delivery=delivery_report)
                elapsed = 0.0
                poll_interval = 0.2
                while elapsed < delivery_timeout_seconds:
                    self.producer.poll(int(poll_interval * 1000))
                    if delivery_event.wait(timeout=0):
                        break
                    elapsed += poll_interval
                if not delivery_event.is_set():
                    raise TimeoutError("Kafka delivery callback timed out")
                if delivery_error is not None:
                    raise RuntimeError(str(delivery_error))
                logger.info(f"Produced message to topic {self.topic}")
                return
            except Exception:
                if attempt >= max_retries:
                    logger.error(
                        f"Failed to produce message to topic {self.topic} after {attempt} attempts",
                        exc_info=True,
                    )
                    break
                logger.warning(
                    f"Produce failed (attempt {attempt}/{max_retries}). Retrying in {backoff:.2f}s...",
                    exc_info=True,
                )
                time.sleep(backoff)
                backoff *= 2
            finally:
                # Ensure producer queue is drained to avoid buildup when retrying
                self.producer.poll(0)
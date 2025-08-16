import json
import logging
from confluent_kafka import Producer
from app.config.config import settings
from app.config.topics import get_document_processed_topic
from app.domain.events import DocumentProcessedEvent

logger = logging.getLogger(__name__)

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
        message = json.dumps(event.model_dump(), default=str).encode('utf-8')

        delivery_error = None

        def delivery_report(err, msg):
            nonlocal delivery_error
            if err is not None:
                delivery_error = err

        max_retries = 3
        backoff = 0.5
        for attempt in range(1, max_retries + 1):
            try:
                self.producer.produce(self.topic, value=message, on_delivery=delivery_report)
                self.producer.poll(0)
                self.producer.flush(5000)
                if delivery_error is None:
                    logger.info(f"Produced message to topic {self.topic}")
                    return
                else:
                    raise RuntimeError(str(delivery_error))
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
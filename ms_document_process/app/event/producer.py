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
        try:
            message = json.dumps(event.model_dump(), default=str).encode('utf-8')
            self.producer.produce(self.topic, value=message)
            self.producer.flush()
            logger.info(f"Produced message to topic {self.topic}: {message}")
        except Exception as e:
            logger.error(f"Failed to produce message to topic {self.topic}", exc_info=True)
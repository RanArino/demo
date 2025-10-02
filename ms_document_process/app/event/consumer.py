import json
import logging
from confluent_kafka import Consumer, KafkaException, KafkaError

from app.config.config import settings
from app.config.topics import get_document_uploaded_topic
from app.domain.events import DocumentUploadedEvent
from app.service.document_service import DocumentProcessService
from app.repository.document_repository import DocumentRepository
from app.event.producer import KafkaProducer
from app.infra.r2_client import R2Client
from app.service.insights_service import DocumentInsightsService

logger = logging.getLogger(__name__)

class KafkaConsumer:
    def __init__(self, document_service: DocumentProcessService):
        self.document_service = document_service
        consumer_config = {
            'bootstrap.servers': settings.kafka_brokers,
            'group.id': 'document-processor-group',
            'auto.offset.reset': 'earliest',
            'security.protocol': settings.kafka_security_protocol,
        }
        if settings.kafka_security_protocol == 'SASL_SSL':
            consumer_config.update({
                'sasl.mechanisms': 'PLAIN',
                'sasl.username': settings.kafka_sasl_username,
                'sasl.password': settings.kafka_sasl_password,
            })
        self.consumer = Consumer(consumer_config)
        self.topic = get_document_uploaded_topic()

    def consume(self):
        """
        Starts the consumer loop. This method blocks and continuously polls for messages.
        """
        self.consumer.subscribe([self.topic])
        logger.info(f"Subscribed to topic {self.topic}")
        while True:
            msg = self.consumer.poll(1.0)
            if msg is None:
                continue
            if msg.error():
                if msg.error().code() == KafkaError._PARTITION_EOF:
                    # End of partition event
                    logger.info(f"Reached end of partition for {msg.topic()} [{msg.partition()}]")
                    continue
                else:
                    # Log and continue to allow transient errors to heal
                    logger.error(f"Kafka error: {msg.error()}")
                    continue

            try:
                event_data = json.loads(msg.value().decode('utf-8'))
                document_uploaded_event = DocumentUploadedEvent(**event_data)
                # Process the message using the document service
                self.document_service.process_document(document_uploaded_event)
            except json.JSONDecodeError:
                logger.error(f"Failed to decode JSON message: {msg.value().decode('utf-8')}")
            except Exception as e:
                logger.error(f"Error processing message: {e}", exc_info=True)

        logger.info("Closing Kafka consumer.")
        self.consumer.close()

def validate_r2_configuration():
    """
    Validates that all required R2 configuration is present before starting the consumer.
    Raises ValueError if any required configuration is missing.
    """
    required_configs = {
        'R2_ENDPOINT': settings.r2_endpoint,
        'R2_ACCESS_KEY_ID': settings.r2_access_key_id,
        'R2_SECRET_ACCESS_KEY': settings.r2_secret_access_key,
        'R2_ACCOUNT_ID': settings.r2_account_id,
        'R2_BUCKET_SOURCE_NAME': settings.r2_bucket_source_name,
        'R2_BUCKET_PROCESSED_NAME': settings.r2_bucket_processed_name,
    }

    missing = [name for name, value in required_configs.items() if not value]

    if missing:
        raise ValueError(
            f"R2 configuration incomplete. Missing required environment variables: {', '.join(missing)}. "
            f"Please check your .env.local file and ensure all R2 credentials are set."
        )

    logger.info("R2 configuration validated successfully")

def start_consumer():
    """
    Initializes all dependencies and starts the Kafka consumer.
    This function is designed to be the entry point for the consumer process.
    """
    logging.basicConfig(level=logging.INFO)
    logger.info("Initializing consumer dependencies...")

    # Validate R2 configuration early to fail fast
    try:
        validate_r2_configuration()
    except ValueError as e:
        logger.critical(f"Configuration validation failed: {e}")
        return  # Exit early if configuration is invalid

    # Initialize dependencies
    try:
        r2_client = R2Client()
        doc_repository = DocumentRepository(r2_client)
        kafka_producer = KafkaProducer()
        insights_service = DocumentInsightsService()
        doc_service = DocumentProcessService(doc_repository, kafka_producer, insights_service)
        consumer = KafkaConsumer(doc_service)
    except Exception as e:
        logger.critical(f"Failed to initialize consumer dependencies: {e}", exc_info=True)
        return # Exit if dependencies fail

    logger.info("Dependencies initialized. Starting consumer...")
    consumer.consume()
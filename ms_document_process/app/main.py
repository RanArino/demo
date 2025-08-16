import logging
import multiprocessing
import time
from app.server.grpc_server import serve
from app.event.consumer import start_consumer

def run_grpc_server():
    """Target function to run the gRPC server in its own process."""
    logging.info("Starting gRPC server process...")
    serve()
    logging.info("gRPC server process finished.")

def run_kafka_consumer():
    """Target function to run the Kafka consumer in its own process."""
    logging.info("Starting Kafka consumer process...")
    start_consumer()
    logging.info("Kafka consumer process finished.")

def main():
    """
    Initializes and runs the microservice components in separate processes.
    """
    logging.basicConfig(level=logging.INFO)
    logging.info("Starting ms_document_process service...")

    # Create separate processes for the server and the consumer
    grpc_process = multiprocessing.Process(target=run_grpc_server)
    consumer_process = multiprocessing.Process(target=run_kafka_consumer)

    try:
        # Start both processes
        grpc_process.start()
        consumer_process.start()
        logging.info(f"gRPC server started in process ID: {grpc_process.pid}")
        logging.info(f"Kafka consumer started in process ID: {consumer_process.pid}")

        # Keep the main script alive and monitor child processes
        while True:
            if not grpc_process.is_alive():
                logging.error("gRPC server process has unexpectedly died. Terminating consumer.")
                consumer_process.terminate()
                break
            if not consumer_process.is_alive():
                logging.error("Kafka consumer process has unexpectedly died. Terminating server.")
                grpc_process.terminate()
                break
            time.sleep(1)

    except KeyboardInterrupt:
        logging.info("Shutdown signal received. Terminating processes gracefully...")
        grpc_process.terminate()
        consumer_process.terminate()

    finally:
        # Wait for processes to finish termination
        grpc_process.join()
        consumer_process.join()
        logging.info("All processes have been terminated. Application shut down.")

if __name__ == "__main__":
    main()

import grpc
from concurrent import futures
import logging
from app.config.config import settings

def serve():
    """
    Starts the gRPC server on the port specified in the settings.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # TODO: Add your generated service Servicer to the server.
    # For example:
    # from app.services.your_service_grpc import YourServiceServicer
    # from api.proto.v1 import your_service_pb2_grpc
    # your_service_pb2_grpc.add_YourServiceServicer_to_server(YourServiceServicer(), server)

    port = settings.grpc_port
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logging.info(f"gRPC server started, listening on port {port}")
    server.wait_for_termination()

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    serve()

from app.server.grpc_server import serve
from app.config.config import settings

if __name__ == "__main__":
    print("ms_document_process service started")
    print("Settings loaded successfully.")
    serve()
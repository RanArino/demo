from app.infra.r2_client import R2Client
from app.config.config import settings

class DocumentRepository:
    """
    Repository for accessing and managing documents in R2 storage.
    """
    def __init__(self, r2_client: R2Client):
        self.r2_client = r2_client
        # Prefer unified content bucket
        self.content_bucket = settings.r2_bucket_content_source or settings.r2_bucket_source_name

    def download_source_document(self, object_key: str) -> bytes:
        """
        Downloads the original document from the unified content bucket using the full object key.
        """
        return self.r2_client.download_file(self.content_bucket, object_key)

    def upload_processed_document(self, object_key: str, data: bytes, content_type: str = 'text/markdown') -> None:
        """
        Uploads the processed document (e.g., Markdown) to the same folder as the source (unified bucket).
        """
        self.r2_client.upload_file(self.content_bucket, object_key, data, content_type)

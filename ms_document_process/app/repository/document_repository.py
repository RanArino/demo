from app.infra.r2_client import R2Client
from app.config.config import settings

class DocumentRepository:
    """
    Repository for accessing and managing documents in R2 storage.
    """
    def __init__(self, r2_client: R2Client):
        self.r2_client = r2_client
        # Use separate buckets for source and processed documents
        self.source_bucket = settings.r2_bucket_source_name
        self.processed_bucket = settings.r2_bucket_processed_name

    def download_source_document(self, object_key: str) -> bytes:
        """
        Downloads the original document from the source bucket using the full object key.
        """
        return self.r2_client.download_file(self.source_bucket, object_key)

    def upload_processed_document(self, object_key: str, data: bytes, content_type: str = 'text/markdown') -> str:
        """
        Uploads the processed document (e.g., Markdown) to the processed bucket.
        """
        self.r2_client.upload_file(self.processed_bucket, object_key, data, content_type)
        return object_key

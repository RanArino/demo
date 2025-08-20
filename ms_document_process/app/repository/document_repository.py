from app.infra.r2_client import R2Client
from app.config.config import settings

class DocumentRepository:
    """
    Repository for accessing and managing documents in R2 storage.
    """
    def __init__(self, r2_client: R2Client):
        self.r2_client = r2_client
        self.source_bucket = settings.r2_bucket_source_name
        self.processed_bucket = settings.r2_bucket_processed_name

    def download_source_document(self, blob_hash: str) -> bytes:
        """
        Downloads the original document from the source bucket.
        """
        return self.r2_client.download_file(self.source_bucket, blob_hash)

    def upload_processed_document(self, blob_hash: str, data: bytes, content_type: str = 'text/markdown') -> None:
        """
        Uploads the processed document (e.g., Markdown) to the processed bucket.
        """
        self.r2_client.upload_file(self.processed_bucket, blob_hash, data, content_type)

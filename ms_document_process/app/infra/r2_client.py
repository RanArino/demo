import boto3
from botocore.client import Config as BotocoreConfig
from botocore.exceptions import ClientError
import logging
from app.config.config import settings

logger = logging.getLogger(__name__)

class R2Client:
    """
    A client for interacting with Cloudflare R2 storage.
    """
    def __init__(self):
        """
        Initializes the R2 client using credentials and settings.
        """
        if not all([settings.r2_endpoint, settings.r2_account_id, settings.r2_access_key_id, settings.r2_secret_access_key]):
            raise ValueError("R2 credentials, endpoint, and account ID must be configured.")

        try:
            self.s3_client = boto3.client(
                's3',
                endpoint_url=settings.r2_endpoint,
                aws_access_key_id=settings.r2_access_key_id,
                aws_secret_access_key=settings.r2_secret_access_key,
                config=BotocoreConfig(signature_version='s3v4'),
                region_name='auto'
            )
            logger.info("R2 client initialized successfully.")
        except Exception as e:
            logger.error(f"Failed to initialize R2 client: {e}")
            raise

    def download_file(self, bucket_name: str, key: str) -> bytes:
        """
        Downloads a file from the specified R2 bucket.
        """
        logger.info(f"Attempting to download file '{key}' from bucket '{bucket_name}'...")
        try:
            response = self.s3_client.get_object(Bucket=bucket_name, Key=key)
            file_content = response['Body'].read()
            logger.info(f"Successfully downloaded file '{key}'.")
            return file_content
        except ClientError as e:
            if e.response['Error']['Code'] == 'NoSuchKey':
                logger.error(f"File not found: {key} in bucket {bucket_name}")
                raise FileNotFoundError(f"The file '{key}' was not found in bucket '{bucket_name}'.")
            else:
                logger.error(f"Error downloading file '{key}': {e}")
                raise
        except Exception as e:
            logger.error(f"An unexpected error occurred during download: {e}")
            raise

    def upload_file(self, bucket_name: str, key: str, data: bytes, content_type: str = 'application/octet-stream') -> None:
        """
        Uploads data to the specified R2 bucket.
        """
        logger.info(f"Attempting to upload file '{key}' to bucket '{bucket_name}'...")
        try:
            self.s3_client.put_object(
                Bucket=bucket_name,
                Key=key,
                Body=data,
                ContentType=content_type
            )
            logger.info(f"Successfully uploaded file '{key}'.")
        except ClientError as e:
            logger.error(f"Error uploading file '{key}': {e}")
            raise
        except Exception as e:
            logger.error(f"An unexpected error occurred during upload: {e}")
            raise

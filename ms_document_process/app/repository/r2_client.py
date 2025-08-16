import boto3
from botocore.client import Config as BotocoreConfig
from botocore.exceptions import ClientError
import logging
from app.config.config import Settings

logger = logging.getLogger(__name__)

class R2Client:
    """
    A client for interacting with Cloudflare R2 storage.
    """
    def __init__(self, settings: Settings):
        """
        Initializes the R2 client using credentials and settings.
        """
        if not all([settings.r2_endpoint, settings.r2_account_id, settings.r2_access_key_id, settings.r2_secret_access_key]):
            raise ValueError("R2 credentials and account ID must be configured.")

        try:
            self.s3_client = boto3.client(
                's3',
                endpoint_url=settings.r2_endpoint,
                aws_access_key_id=settings.r2_access_key_id,
                aws_secret_access_key=settings.r2_secret_access_key,
                config=BotocoreConfig(signature_version='s3v4'),
                region_name='auto' # R2 typically uses 'auto'
            )
            logger.info("R2 client initialized successfully.")
        except Exception as e:
            logger.error(f"Failed to initialize R2 client: {e}")
            raise

    def download_file(self, bucket_name: str, key: str) -> bytes:
        """
        Downloads a file from the specified R2 bucket.

        Args:
            bucket_name: The name of the bucket.
            key: The key (object name) of the file to download.

        Returns:
            The file content as bytes.
        
        Raises:
            FileNotFoundError: If the file does not exist.
            Exception: For other download errors.
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

        Args:
            bucket_name: The name of the bucket.
            key: The key (object name) under which to store the data.
            data: The data to upload, as bytes.
            content_type: The MIME type of the content.
        
        Raises:
            Exception: For upload errors.
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

    def get_files(self, bucket_name: str, folder_name: str, limits: int = 10) -> list[str]:
        """
        Lists files in a specific folder within an R2 bucket. For testing purposes.

        Args:
            bucket_name: The name of the bucket.
            folder_name: The folder path (prefix) to search within.
            limits: The maximum number of file keys to return.

        Returns:
            A list of object keys (blob hashes).
        """
        logger.info(f"Listing files in folder '{folder_name}' of bucket '{bucket_name}' with a limit of {limits}...")
        try:
            response = self.s3_client.list_objects_v2(
                Bucket=bucket_name,
                Prefix=folder_name,
                MaxKeys=limits
            )
            keys = [item['Key'] for item in response.get('Contents', [])]
            logger.info(f"Found {len(keys)} files.")
            return keys
        except ClientError as e:
            logger.error(f"Error listing files in bucket '{bucket_name}': {e}")
            raise
        except Exception as e:
            logger.error(f"An unexpected error occurred while listing files: {e}")
            raise
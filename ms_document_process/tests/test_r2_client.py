import unittest
from unittest.mock import MagicMock, patch

from botocore.exceptions import ClientError

from app.infra.r2_client import R2Client


class TestR2Client(unittest.TestCase):
    def setUp(self):
        boto_patcher = patch("app.infra.r2_client.boto3.client")
        self.addCleanup(boto_patcher.stop)
        mock_boto_client = boto_patcher.start()
        self.mock_s3 = MagicMock()
        mock_boto_client.return_value = self.mock_s3

        settings_patcher = patch("app.infra.r2_client.settings")
        self.addCleanup(settings_patcher.stop)
        mocked_settings = settings_patcher.start()
        mocked_settings.r2_endpoint = "https://example.com"
        mocked_settings.r2_account_id = "acct"
        mocked_settings.r2_access_key_id = "key"
        mocked_settings.r2_secret_access_key = "secret"

        self.client = R2Client()

    def test_upload_file_success(self):
        data = b"content"
        self.client.upload_file("bucket", "key", data, content_type="text/plain")
        self.mock_s3.put_object.assert_called_once_with(
            Bucket="bucket",
            Key="key",
            Body=data,
            ContentType="text/plain",
        )

    def test_upload_file_client_error(self):
        error_response = {"Error": {"Code": "AccessDenied"}}
        self.mock_s3.put_object.side_effect = ClientError(error_response, "PutObject")
        with self.assertRaises(ClientError):
            self.client.upload_file("bucket", "key", b"data")

    def test_download_file_success(self):
        body = MagicMock()
        body.read.return_value = b"data"
        self.mock_s3.get_object.return_value = {"Body": body}

        content = self.client.download_file("bucket", "key")
        self.assertEqual(content, b"data")
        self.mock_s3.get_object.assert_called_once_with(Bucket="bucket", Key="key")

    def test_download_file_not_found(self):
        error_response = {"Error": {"Code": "NoSuchKey"}}
        self.mock_s3.get_object.side_effect = ClientError(error_response, "GetObject")
        with self.assertRaises(FileNotFoundError):
            self.client.download_file("bucket", "missing")

    def test_download_file_other_error(self):
        error_response = {"Error": {"Code": "AccessDenied"}}
        self.mock_s3.get_object.side_effect = ClientError(error_response, "GetObject")
        with self.assertRaises(ClientError):
            self.client.download_file("bucket", "key")


if __name__ == '__main__':
    unittest.main()

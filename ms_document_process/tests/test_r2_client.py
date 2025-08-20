import unittest
import os
from app.infra.r2_client import R2Client
from app.config.config import settings

class TestR2Client(unittest.TestCase):

    def setUp(self):
        """Set up the test client and check for credentials."""
        # This test is an integration test and requires R2 credentials to be set in the environment.
        if not all([settings.r2_account_id, settings.r2_access_key_id, settings.r2_secret_access_key]):
            self.skipTest("R2 credentials are not configured. Skipping R2 integration tests.")
        
        self.r2_client = R2Client()
        self.bucket_name = "scaler-demo"
        self.processed_bucket_name = "scaler-demo-document-process"
        self.upload_key = "integration-tests/test-upload.md"

    def tearDown(self):
        """Clean up any files created during the test."""
        try:
            # Attempt to delete the test file from the processed bucket
            self.r2_client.s3_client.delete_object(Bucket=self.processed_bucket_name, Key=self.upload_key)
            print(f"\nSuccessfully cleaned up test file '{self.upload_key}'.")
        except Exception as e:
            # Log if cleanup fails but don't fail the test
            print(f"\nCleanup failed for '{self.upload_key}': {e}")

    def test_upload_file_successfully(self):
        """
        Tests that a markdown text can be successfully uploaded to the processed R2 bucket.
        """
        # ARRANGE
        markdown_content = "# Test Header\n\nThis is a test markdown file."
        content_bytes = markdown_content.encode('utf-8')

        # ACT
        try:
            self.r2_client.upload_file(
                bucket_name=self.processed_bucket_name,
                key=self.upload_key,
                data=content_bytes,
                content_type='text/markdown'
            )
        except Exception as e:
            self.fail(f"An unexpected error occurred during upload: {e}")

        # ASSERT
        try:
            # Verify by downloading the file and checking its content
            downloaded_content = self.r2_client.download_file(self.processed_bucket_name, self.upload_key)
            self.assertEqual(downloaded_content, content_bytes)
            print(f"\nSuccessfully uploaded and verified '{self.upload_key}'.")
        except FileNotFoundError:
            self.fail("Uploaded file not found during verification.")
        except Exception as e:
            self.fail(f"An unexpected error occurred during verification: {e}")

    def test_download_file_successfully(self):
        """
        Tests that a file can be successfully downloaded from R2.
        """
        # ARRANGE
        # This key must exist in the 'scaler-demo' bucket for the test to pass.
        key = "integration-tests/679acdee-5431-4e18-9c3a-6f5e912f727c.pdf"

        # ACT
        try:
            file_content = self.r2_client.download_file(self.bucket_name, key)
        except FileNotFoundError:
            self.fail(f"Test file not found in R2: bucket='{self.bucket_name}', key='{key}'")
        except Exception as e:
            self.fail(f"An unexpected error occurred during download: {e}")

        # ASSERT
        self.assertIsNotNone(file_content)
        self.assertIsInstance(file_content, bytes)
        self.assertGreater(len(file_content), 0, "The downloaded file should not be empty.")
        print(f"\nSuccessfully downloaded '{key}'. Size: {len(file_content)} bytes.")

    def test_download_nonexistent_file(self):
        """
        Tests that downloading a nonexistent file raises FileNotFoundError.
        """
        # ARRANGE
        key = "this/key/definitely/does/not/exist.txt"

        # ACT & ASSERT
        with self.assertRaises(FileNotFoundError):
            self.r2_client.download_file(self.bucket_name, key)
        print(f"\nSuccessfully confirmed that downloading a non-existent file raises an error.")

if __name__ == '__main__':
    unittest.main()

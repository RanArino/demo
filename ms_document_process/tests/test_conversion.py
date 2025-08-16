import unittest
import os
from unittest.mock import patch, Mock

from app.service.pdf_conversion_service import convert_document_to_markdown
from app.service.converters.base import DocumentConverter


class TestDocumentConversionService(unittest.TestCase):

    def test_convert_document_to_markdown_with_mock_converter(self):
        # Arrange
        dummy_document_content = b"%PDF-1.4 dummy content"
        source_format = "pdf"
        expected_markdown = "Mocked markdown output"

        # Create a mock converter object that adheres to the DocumentConverter interface
        mock_converter = Mock(spec=DocumentConverter)
        mock_converter.convert.return_value = expected_markdown

        # Patch the factory function to return our mock converter
        with patch("app.service.pdf_conversion_service.get_document_converter") as mock_get_converter:
            mock_get_converter.return_value = mock_converter

            # Act
            result = convert_document_to_markdown(dummy_document_content, source_format)

            # Assert
            self.assertEqual(result, expected_markdown)
            # Verify that the factory was called
            mock_get_converter.assert_called_once()
            # Verify that the convert method on our mock was called with the correct content and format
            mock_converter.convert.assert_called_once_with(dummy_document_content, source_format)

    def test_convert_real_pdf_to_markdown(self):
        # Allow enabling real conversion via env var
        run_real = os.getenv("RUN_REAL_PDF_TESTS", "").lower() in ("1", "true", "yes")
        if not run_real:
            self.skipTest("Set RUN_REAL_PDF_TESTS=1 to enable real PDF conversion test.")

        # Ensure markitdown is importable in the active interpreter
        try:
            import importlib.util
            if importlib.util.find_spec("markitdown") is None:
                self.skipTest("markitdown not installed in the active interpreter; skipping real conversion test.")
        except Exception:
            self.skipTest("Unable to check for markitdown; skipping real conversion test.")

        # This test assumes test.pdf is available
        pdf_path_candidates = [
            "ms_document_process/test.pdf",
            "test.pdf",
        ]
        pdf_path = next((p for p in pdf_path_candidates if os.path.exists(p)), None)
        if not pdf_path:
            self.skipTest("test.pdf not found in expected locations, skipping real conversion test.")

        with open(pdf_path, "rb") as f:
            pdf_bytes = f.read()

        # Act 
        markdown_result = convert_document_to_markdown(pdf_bytes, "pdf")

        # Assert
        self.assertIsInstance(markdown_result, str)
        self.assertTrue(len(markdown_result) > 50)  # Output should be substantial
        print("\n--- Conversion Result Snippet ---")
        print(markdown_result[:500])
        print("--- End of Snippet ---")


if __name__ == '__main__':
    unittest.main()

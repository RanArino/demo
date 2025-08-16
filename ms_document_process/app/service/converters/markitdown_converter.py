import os
import tempfile
from typing import Any

from .base import DocumentConverter


class MarkitdownConverter(DocumentConverter):
    """Converts PDF to Markdown using the local markitdown library."""

    def __init__(self):
        self._markdown_converter = self._create_markdown_converter()

    def _create_markdown_converter(self):
        """Factory to create MarkItDown converter with lazy import."""
        try:
            from markitdown import MarkItDown  # type: ignore
        except ImportError as exc:
            raise ImportError(
                "markitdown is required for this converter. "
                "Install with extras: 'pip install markitdown[pdf]'"
            ) from exc
        return MarkItDown()

    def _extract_markdown_text(self, convert_result: Any) -> str:
        """Return a string markdown representation from various possible result types."""
        if isinstance(convert_result, str):
            return convert_result
        if hasattr(convert_result, "text_content") and isinstance(
            convert_result.text_content, str
        ):
            return convert_result.text_content
        if hasattr(convert_result, "markdown") and isinstance(convert_result.markdown, str):
            return convert_result.markdown
        if isinstance(convert_result, dict):
            for key in ("text_content", "markdown", "content", "text"):
                value = convert_result.get(key)
                if isinstance(value, str):
                    return value
        return str(convert_result)

    def convert(self, document_content: bytes, source_format: str) -> str:
        """
        Converts PDF content to Markdown using the MarkItDown library.
        """
        if source_format.lower() != "pdf":
            raise ValueError("MarkitdownConverter only supports PDF conversion.")

        # MarkItDown works with file paths, so we need to save the bytes to a temporary file.
        with tempfile.NamedTemporaryFile(delete=False, suffix=".pdf") as temp_pdf:
            temp_pdf.write(document_content)
            temp_pdf_path = temp_pdf.name

        try:
            # Convert the temporary PDF file to Markdown
            convert_result = self._markdown_converter.convert(temp_pdf_path)
            return self._extract_markdown_text(convert_result)
        finally:
            # Clean up the temporary file
            try:
                os.remove(temp_pdf_path)
            except FileNotFoundError:
                pass

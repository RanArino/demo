from functools import lru_cache

from .converters.base import DocumentConverter
from .converters.markitdown_converter import MarkitdownConverter


@lru_cache(maxsize=1)
def get_document_converter() -> DocumentConverter:
    """
    Factory function to get the configured document converter.
    Uses environment variables to decide which converter to instantiate.
    """
    # In the future, we can use an env var to switch implementations
    # converter_type = os.getenv("DOCUMENT_CONVERTER_TYPE", "markitdown")
    # if converter_type == "cloud":
    #     from .converters.cloud_converter import CloudConverter
    #     return CloudConverter()
    # elif converter_type == "markitdown":
    #     return MarkitdownConverter()
    # else:
    #     raise ValueError(f"Unknown document converter type: {converter_type}")

    # For now, we only have one implementation
    return MarkitdownConverter()


def convert_document_to_markdown(document_content: bytes, source_format: str) -> str:
    """
    Converts document content to Markdown using the configured converter.

    Args:
        document_content: The byte content of the document.
        source_format: The format of the source document (e.g., "pdf", "docx").

    Returns:
        The Markdown content as a string.
    """
    converter = get_document_converter()
    return converter.convert(document_content, source_format)

from abc import ABC, abstractmethod

class DocumentConverter(ABC):
    """Abstract base class for document to Markdown converters."""

    @abstractmethod
    def convert(self, document_content: bytes, source_format: str) -> str:
        """
        Converts document content to a Markdown string.

        Args:
            document_content: The byte content of the document.
            source_format: The format of the source document (e.g., "pdf", "docx").

        Returns:
            The Markdown content as a string.
        """
        pass

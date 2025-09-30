"""Base class for LLM providers."""
from abc import ABC, abstractmethod
from typing import Optional

from pkg.llm.types import InsightRequest, InsightResponse


class LLMProvider(ABC):
    """Abstract base class for LLM providers."""

    @abstractmethod
    def generate_insights(
        self,
        document_text: str,
        title: Optional[str] = None,
        summary_tokens: int = 512,
        keyword_count: int = 10,
    ) -> InsightResponse:
        """
        Generate insights (summary and keywords) from document text.

        Args:
            document_text: The document text to analyze
            title: Optional document title for context
            summary_tokens: Target token count for summary
            keyword_count: Number of keywords to extract

        Returns:
            InsightResponse containing summary and keywords

        Raises:
            RuntimeError: If insight generation fails
        """
        pass

    def _build_prompt(
        self,
        document_text: str,
        title: Optional[str],
        summary_tokens: int,
        keyword_count: int,
    ) -> str:
        """
        Build the prompt for insight generation.

        Args:
            document_text: The document text to analyze
            title: Optional document title for context
            summary_tokens: Target token count for summary
            keyword_count: Number of keywords to extract

        Returns:
            Formatted prompt string
        """
        heading = f"Title: {title}\n" if title else ""
        instructions = (
            "You are an assistant that summarizes documents and extracts concise keywords.\n"
            "Analyse the provided document and respond strictly with valid JSON matching the schema:\n"
            '{"summary": string, "keywords": string[]}.\n'
            "Guidelines:\n"
            f"- The summary should be around {summary_tokens} tokens and cover essential points.\n"
            f"- Return between 3 and {keyword_count} informative keywords.\n"
            "- Do not include markdown or explanations outside the JSON object.\n"
        )
        return f"{instructions}{heading}\nDocument:\n{document_text}"
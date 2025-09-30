from __future__ import annotations

import logging
from typing import Optional

from pkg.llm import LLMProvider, LLMProviderConfig, LLMProviderFactory, LLMProviderType

from app.config.config import settings
from app.domain.insights import DocumentInsights

logger = logging.getLogger(__name__)


class DocumentInsightsService:
    def __init__(self) -> None:
        """Initialize DocumentInsightsService with configured LLM provider."""
        self._provider = self._create_provider()
        # Backward compatibility: use legacy settings if new ones aren't set
        self._summary_tokens = settings.llm_summary_tokens or settings.gemini_summary_tokens
        self._keyword_count = settings.llm_keyword_count or settings.gemini_keyword_count

    def _create_provider(self) -> LLMProvider:
        """Create LLM provider based on configuration."""
        try:
            provider_type = LLMProviderType(settings.llm_provider.lower())
        except ValueError:
            logger.warning(
                f"Invalid provider '{settings.llm_provider}', falling back to openai"
            )
            provider_type = LLMProviderType.OPENAI

        config = LLMProviderConfig(
            provider_type=provider_type,
            openai_api_key=settings.openai_api_key,
            openai_model=settings.openai_model,
            vertex_project_id=settings.vertex_project_id,
            vertex_location=settings.vertex_location,
            vertex_model=settings.vertex_model,
            gemini_api_key=settings.gemini_api_key,
            gemini_model=settings.gemini_model,
            temperature=settings.llm_temperature,
            top_p=settings.llm_top_p,
        )

        return LLMProviderFactory.create_provider(config)

    def generate_insights(
        self, document_text: str, title: Optional[str] = None
    ) -> DocumentInsights:
        """
        Generate insights from document text using configured LLM provider.

        Args:
            document_text: The document text to analyze
            title: Optional document title for context

        Returns:
            DocumentInsights containing summary and keywords

        Raises:
            RuntimeError: If insight generation fails
        """
        try:
            response = self._provider.generate_insights(
                document_text=document_text,
                title=title,
                summary_tokens=self._summary_tokens,
                keyword_count=self._keyword_count,
            )
            return DocumentInsights.from_generation(response.summary, response.keywords)
        except Exception as err:
            logger.warning("Failed to generate document insights", exc_info=True)
            raise RuntimeError("Failed to generate document insights") from err

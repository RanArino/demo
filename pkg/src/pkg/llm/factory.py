"""Factory for creating LLM provider instances."""
import logging
from enum import Enum
from typing import Optional

from pkg.llm.base import LLMProvider
from pkg.llm.providers.gemini import GeminiProvider
from pkg.llm.providers.openai import OpenAIProvider
from pkg.llm.providers.vertex import VertexAIProvider

logger = logging.getLogger(__name__)


class LLMProviderType(str, Enum):
    """Available LLM provider types."""
    OPENAI = "openai"
    VERTEX = "vertex"
    GEMINI = "gemini"


class LLMProviderConfig:
    """Configuration for LLM providers."""

    def __init__(
        self,
        provider_type: LLMProviderType = LLMProviderType.OPENAI,
        # OpenAI config
        openai_api_key: Optional[str] = None,
        openai_model: str = "gpt-5-mini-2025-08-07",
        # Vertex AI config
        vertex_project_id: Optional[str] = None,
        vertex_location: str = "us-central1",
        vertex_model: str = "gemini-1.5-flash",
        # Gemini config
        gemini_api_key: Optional[str] = None,
        gemini_model: str = "models/gemini-1.5-flash",
        # Common config
        temperature: float = 0.2,
        top_p: float = 0.95,
    ):
        """
        Initialize LLM provider configuration.

        Args:
            provider_type: Type of provider to use (default: openai)
            openai_api_key: OpenAI API key
            openai_model: OpenAI model name
            vertex_project_id: GCP project ID for Vertex AI
            vertex_location: GCP location for Vertex AI
            vertex_model: Vertex AI model name
            gemini_api_key: Gemini API key
            gemini_model: Gemini model name
            temperature: Sampling temperature
            top_p: Nucleus sampling parameter
        """
        self.provider_type = provider_type
        self.openai_api_key = openai_api_key
        self.openai_model = openai_model
        self.vertex_project_id = vertex_project_id
        self.vertex_location = vertex_location
        self.vertex_model = vertex_model
        self.gemini_api_key = gemini_api_key
        self.gemini_model = gemini_model
        self.temperature = temperature
        self.top_p = top_p


class LLMProviderFactory:
    """Factory for creating LLM provider instances."""

    @staticmethod
    def create_provider(config: LLMProviderConfig) -> LLMProvider:
        """
        Create an LLM provider instance based on configuration.

        Args:
            config: Provider configuration

        Returns:
            Configured LLM provider instance

        Raises:
            ValueError: If provider type is invalid or required credentials are missing
        """
        if config.provider_type == LLMProviderType.OPENAI:
            if not config.openai_api_key:
                raise ValueError("openai_api_key is required for OpenAI provider")
            logger.info(f"Initializing OpenAI provider with model: {config.openai_model}")
            return OpenAIProvider(
                api_key=config.openai_api_key,
                model=config.openai_model,
                temperature=config.temperature,
                top_p=config.top_p,
            )

        elif config.provider_type == LLMProviderType.VERTEX:
            if not config.vertex_project_id:
                raise ValueError("vertex_project_id is required for Vertex AI provider")
            logger.info(
                f"Initializing Vertex AI provider with model: {config.vertex_model}, "
                f"project: {config.vertex_project_id}, location: {config.vertex_location}"
            )
            return VertexAIProvider(
                project_id=config.vertex_project_id,
                location=config.vertex_location,
                model=config.vertex_model,
                temperature=config.temperature,
                top_p=config.top_p,
            )

        elif config.provider_type == LLMProviderType.GEMINI:
            if not config.gemini_api_key:
                raise ValueError("gemini_api_key is required for Gemini provider")
            logger.info(f"Initializing Gemini provider with model: {config.gemini_model}")
            return GeminiProvider(
                api_key=config.gemini_api_key,
                model=config.gemini_model,
                temperature=config.temperature,
                top_p=config.top_p,
            )

        else:
            raise ValueError(f"Unsupported provider type: {config.provider_type}")
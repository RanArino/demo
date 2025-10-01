"""OpenAI provider implementation."""
import json
import logging
from typing import Optional

from openai import OpenAI
from openai.types.chat import ChatCompletion

from pkg.llm.base import LLMProvider
from pkg.llm.types import InsightResponse

logger = logging.getLogger(__name__)


class OpenAIProvider(LLMProvider):
    """OpenAI Chat Completions API provider."""

    def __init__(
        self,
        api_key: str,
        model: str = "gpt-5-mini-2025-08-07",
        temperature: float = 1.0,
        top_p: float = 1.0,
    ):
        """
        Initialize OpenAI provider.

        Args:
            api_key: OpenAI API key
            model: Model name (default: gpt-5-mini-2025-08-07)
            temperature: Sampling temperature (default: 1.0)
            top_p: Nucleus sampling parameter (default: 1.0)
        """
        if not api_key:
            raise ValueError("api_key is required for OpenAI provider")

        self.client = OpenAI(api_key=api_key)
        self.model = model
        self.temperature = temperature
        self.top_p = top_p

    def generate_insights(
        self,
        document_text: str,
        title: Optional[str] = None,
        summary_tokens: int = 512,
        keyword_count: int = 10,
    ) -> InsightResponse:
        """Generate insights using OpenAI Chat Completions API."""
        try:
            # Build kwargs for the API call
            kwargs = {
                "model": self.model,
                "messages": [
                    {
                        "role": "system",
                        "content": "You are a helpful assistant that generates document summaries and keywords in JSON format.",
                    },
                    {
                        "role": "user",
                        "content": self._build_prompt(
                            document_text, title, summary_tokens, keyword_count
                        ),
                    },
                ],
                "response_format": {"type": "json_object"},
            }

            # Only add temperature/top_p if not using default value of 1.0
            # (gpt-5-mini-2025-08-07 only supports temperature=1)
            if self.temperature != 1.0:
                kwargs["temperature"] = self.temperature
            if self.top_p != 1.0:
                kwargs["top_p"] = self.top_p

            response: ChatCompletion = self.client.chat.completions.create(**kwargs)

            return self._parse_response(response)

        except Exception as err:
            logger.warning("OpenAI API error while generating insights", exc_info=True)
            raise RuntimeError("Failed to generate document insights") from err

    def _parse_response(self, response: ChatCompletion) -> InsightResponse:
        """Parse OpenAI response to InsightResponse."""
        try:
            if not response.choices:
                logger.warning("Received empty choices from OpenAI API")
                return InsightResponse(summary="", keywords=[])

            message = response.choices[0].message
            if not message.content:
                logger.warning("Received empty content from OpenAI API")
                return InsightResponse(summary="", keywords=[])

            payload = json.loads(message.content)
            summary = payload.get("summary", "") if isinstance(payload, dict) else ""
            keywords = payload.get("keywords", []) if isinstance(payload, dict) else []

            # Clean keywords
            if isinstance(keywords, list):
                keywords = [kw.strip() for kw in keywords if isinstance(kw, str) and kw.strip()]
            else:
                keywords = []

            return InsightResponse(summary=summary.strip(), keywords=keywords)

        except Exception:
            logger.warning("Failed to parse OpenAI response", exc_info=True)
            return InsightResponse(summary="", keywords=[])
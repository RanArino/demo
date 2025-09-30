"""Google Gemini API provider implementation."""
import json
import logging
from typing import Any, Optional

import google.generativeai as genai
from google.api_core.exceptions import GoogleAPIError
from google.generativeai import types as genai_types

from pkg.llm.base import LLMProvider
from pkg.llm.types import InsightResponse

logger = logging.getLogger(__name__)


class GeminiProvider(LLMProvider):
    """Google Gemini API provider (for backward compatibility)."""

    def __init__(
        self,
        api_key: str,
        model: str = "models/gemini-1.5-flash",
        temperature: float = 0.2,
        top_p: float = 0.95,
    ):
        """
        Initialize Gemini provider.

        Args:
            api_key: Gemini API key
            model: Model name (default: models/gemini-1.5-flash)
            temperature: Sampling temperature (default: 0.2)
            top_p: Nucleus sampling parameter (default: 0.95)
        """
        if not api_key:
            raise ValueError("api_key is required for Gemini provider")

        genai.configure(api_key=api_key)
        self._model = genai.GenerativeModel(model)
        self.temperature = temperature
        self.top_p = top_p

    def generate_insights(
        self,
        document_text: str,
        title: Optional[str] = None,
        summary_tokens: int = 512,
        keyword_count: int = 10,
    ) -> InsightResponse:
        """Generate insights using Gemini API."""
        try:
            response = self._model.generate_content(
                contents=self._build_prompt(document_text, title, summary_tokens, keyword_count),
                generation_config=genai_types.GenerationConfig(
                    temperature=self.temperature,
                    top_p=self.top_p,
                    response_mime_type="application/json",
                    response_schema={
                        "type": "object",
                        "properties": {
                            "summary": {"type": "string"},
                            "keywords": {
                                "type": "array",
                                "items": {"type": "string"},
                                "min_items": min(3, keyword_count),
                                "max_items": keyword_count,
                            },
                        },
                        "required": ["summary", "keywords"],
                    },
                ),
            )
            return self._parse_response(response)

        except GoogleAPIError as err:
            logger.warning("Gemini API error while generating insights", exc_info=True)
            raise RuntimeError("Failed to generate document insights") from err

    def _parse_response(self, response: Any) -> InsightResponse:
        """Parse Gemini response to InsightResponse."""
        try:
            # Try parsed response first
            parsed = getattr(response, "parsed", None)
            if isinstance(parsed, dict):
                summary = parsed.get("summary", "")
                keywords = parsed.get("keywords", [])
                return self._clean_response(summary, keywords)

            # Fall back to text extraction
            raw_text = self._extract_text(response)
            if not raw_text or not raw_text.strip():
                logger.warning("Received empty response from Gemini API")
                return InsightResponse(summary="", keywords=[])

            payload = json.loads(raw_text)
            summary = payload.get("summary", "") if isinstance(payload, dict) else ""
            keywords = payload.get("keywords", []) if isinstance(payload, dict) else []
            return self._clean_response(summary, keywords)

        except Exception:
            logger.warning("Failed to parse Gemini response", exc_info=True)
            return InsightResponse(summary="", keywords=[])

    @staticmethod
    def _extract_text(response: Any) -> str:
        """Extract text from Gemini response."""
        try:
            text_value = getattr(response, "text")
        except (AttributeError, ValueError):
            text_value = None

        if text_value:
            return text_value

        candidates = getattr(response, "candidates", None)
        if candidates:
            for candidate in candidates:
                content = getattr(candidate, "content", None)
                if content:
                    for part in getattr(content, "parts", []):
                        text = getattr(part, "text", None)
                        if text:
                            return text
                        if hasattr(part, "function_response"):
                            function_response = getattr(part, "function_response")
                            if function_response and function_response.response:
                                try:
                                    return json.dumps(function_response.response)
                                except (TypeError, ValueError):
                                    logger.debug("Unable to serialize function response", exc_info=True)
                        if hasattr(part, "parsed") and part.parsed:
                            try:
                                return json.dumps(part.parsed)
                            except (TypeError, ValueError):
                                logger.debug("Unable to serialize parsed part", exc_info=True)
        return str(response)

    def _clean_response(self, summary: str, keywords) -> InsightResponse:
        """Clean and validate response data."""
        cleaned_keywords = []
        if keywords:
            if isinstance(keywords, str):
                # Handle string representation
                kw_str = keywords.strip()
                if kw_str.startswith('[') and kw_str.endswith(']'):
                    kw_str = kw_str[1:-1]
                cleaned_keywords = [kw.strip().strip('"\'') for kw in kw_str.split(',') if kw.strip()]
            elif isinstance(keywords, list):
                cleaned_keywords = [kw.strip() for kw in keywords if isinstance(kw, str) and kw.strip()]

        return InsightResponse(summary=summary.strip(), keywords=cleaned_keywords)
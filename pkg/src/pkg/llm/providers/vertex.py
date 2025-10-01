"""Vertex AI provider implementation."""
import json
import logging
from typing import Optional

from google.cloud import aiplatform
from vertexai.generative_models import GenerationConfig, GenerativeModel

from pkg.llm.base import LLMProvider
from pkg.llm.types import InsightResponse

logger = logging.getLogger(__name__)


class VertexAIProvider(LLMProvider):
    """Google Vertex AI provider."""

    def __init__(
        self,
        project_id: str,
        location: str = "us-central1",
        model: str = "gemini-1.5-flash",
        temperature: float = 0.2,
        top_p: float = 0.95,
    ):
        """
        Initialize Vertex AI provider.

        Args:
            project_id: GCP project ID
            location: GCP location (default: us-central1)
            model: Model name (default: gemini-1.5-flash)
            temperature: Sampling temperature (default: 0.2)
            top_p: Nucleus sampling parameter (default: 0.95)
        """
        if not project_id:
            raise ValueError("project_id is required for Vertex AI provider")

        aiplatform.init(project=project_id, location=location)
        self.model = GenerativeModel(model)
        self.temperature = temperature
        self.top_p = top_p

    def generate_insights(
        self,
        document_text: str,
        title: Optional[str] = None,
        summary_tokens: int = 512,
        keyword_count: int = 10,
    ) -> InsightResponse:
        """Generate insights using Vertex AI."""
        try:
            generation_config = GenerationConfig(
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
                            "minItems": min(3, keyword_count),
                            "maxItems": keyword_count,
                        },
                    },
                    "required": ["summary", "keywords"],
                },
            )

            response = self.model.generate_content(
                contents=self._build_prompt(document_text, title, summary_tokens, keyword_count),
                generation_config=generation_config,
            )

            return self._parse_response(response)

        except Exception as err:
            logger.warning("Vertex AI error while generating insights", exc_info=True)
            raise RuntimeError("Failed to generate document insights") from err

    def _parse_response(self, response) -> InsightResponse:
        """Parse Vertex AI response to InsightResponse."""
        try:
            # Try to get parsed response first
            parsed = getattr(response, "parsed", None)
            if isinstance(parsed, dict):
                summary = parsed.get("summary", "")
                keywords = parsed.get("keywords", [])
                return self._clean_response(summary, keywords)

            # Fall back to text parsing
            text = getattr(response, "text", None)
            if text:
                payload = json.loads(text)
                summary = payload.get("summary", "") if isinstance(payload, dict) else ""
                keywords = payload.get("keywords", []) if isinstance(payload, dict) else []
                return self._clean_response(summary, keywords)

            logger.warning("Received empty response from Vertex AI")
            return InsightResponse(summary="", keywords=[])

        except Exception:
            logger.warning("Failed to parse Vertex AI response", exc_info=True)
            return InsightResponse(summary="", keywords=[])

    def _clean_response(self, summary: str, keywords) -> InsightResponse:
        """Clean and validate response data."""
        cleaned_keywords = []
        if isinstance(keywords, list):
            cleaned_keywords = [kw.strip() for kw in keywords if isinstance(kw, str) and kw.strip()]
        elif isinstance(keywords, str):
            # Handle string representation of list
            kw_str = keywords.strip()
            if kw_str.startswith('[') and kw_str.endswith(']'):
                kw_str = kw_str[1:-1]
            cleaned_keywords = [kw.strip().strip('"\'') for kw in kw_str.split(',') if kw.strip()]

        return InsightResponse(summary=summary.strip(), keywords=cleaned_keywords)
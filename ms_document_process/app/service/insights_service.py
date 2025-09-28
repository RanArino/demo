from __future__ import annotations

import json
import logging
from typing import Iterable, Optional

import google.generativeai as genai
from google.api_core.exceptions import GoogleAPIError
from google.generativeai import types as genai_types

from app.config.config import settings
from app.domain.insights import DocumentInsights

logger = logging.getLogger(__name__)


class DocumentInsightsService:
    def __init__(self) -> None:
        if not settings.gemini_api_key:
            raise ValueError("gemini_api_key must be configured to enable document insights")

        genai.configure(api_key=settings.gemini_api_key)
        self._model = genai.GenerativeModel(settings.gemini_model)

    def generate_insights(self, document_text: str, title: Optional[str] = None) -> DocumentInsights:
        try:
            response = self._model.generate_content(
                contents=self._build_prompt(document_text, title),
                generation_config=genai_types.GenerationConfig(
                    temperature=0.2,
                    top_p=0.95,
                    response_mime_type="application/json",
                    response_schema={
                        "type": "object",
                        "properties": {
                            "summary": {"type": "string"},
                            "keywords": {
                                "type": "array",
                                "items": {"type": "string"},
                                "min_items": min(3, settings.gemini_keyword_count),
                                "max_items": settings.gemini_keyword_count,
                            },
                        },
                        "required": ["summary", "keywords"],
                    },
                ),
            )
        except GoogleAPIError as err:
            logger.warning("Gemini API error while generating insights", exc_info=True)
            raise RuntimeError("failed to generate document insights") from err

        return self._parse_response(response)

    def _build_prompt(self, document_text: str, title: Optional[str]) -> str:
        heading = f"Title: {title}\n" if title else ""
        instructions = (
            "You are an assistant that summarizes documents and extracts concise keywords.\n"
            "Analyse the provided document and respond strictly with valid JSON matching the schema:\n"
            "{\"summary\": string, \"keywords\": string[]}.\n"
            "Guidelines:\n"
            f"- The summary should be around {settings.gemini_summary_tokens} tokens and cover essential points.\n"
            f"- Return between 3 and {settings.gemini_keyword_count} informative keywords.\n"
            "- Do not include markdown or explanations outside the JSON object.\n"
        )
        return f"{instructions}{heading}\nDocument:\n{document_text}"

    def _parse_response(self, response: genai.types.Generation) -> DocumentInsights:
        try:
            parsed = getattr(response, "parsed", None)
            if isinstance(parsed, dict):
                summary = parsed.get("summary", "")
                keywords = parsed.get("keywords", [])
                return DocumentInsights.from_generation(summary, keywords)
            raw_text = self._extract_text(response)
            if not raw_text or not raw_text.strip():
                logger.warning("Received empty or whitespace-only response from Gemini API")
                return DocumentInsights(summary="", keywords=[])
            payload = json.loads(raw_text)
            summary = payload.get("summary", "") if isinstance(payload, dict) else ""
            keywords = payload.get("keywords", []) if isinstance(payload, dict) else []
            return DocumentInsights.from_generation(summary, keywords)
        except Exception:
            logger.warning("Failed to parse Gemini insights response", exc_info=True)
            return DocumentInsights(summary="", keywords=[])

    @staticmethod
    def _extract_text(response: genai.types.Generation) -> str:
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
                                    logger.debug("Unable to serialize function response part", exc_info=True)
                        if hasattr(part, "parsed") and part.parsed:
                            try:
                                return json.dumps(part.parsed)
                            except (TypeError, ValueError):
                                logger.debug("Unable to serialize parsed part", exc_info=True)
        return str(response)

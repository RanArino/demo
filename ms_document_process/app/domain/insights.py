from typing import Iterable, List

from pydantic import BaseModel, Field


class DocumentInsights(BaseModel):
    summary: str = Field(..., description="Summarized content (~300 tokens)")
    keywords: List[str] = Field(default_factory=list, description="Keywords describing the document")

    @classmethod
    def from_generation(cls, summary: str, keywords: Iterable[str]) -> "DocumentInsights":
        cleaned_summary = summary.strip()
        processed_keywords = []
        if keywords:
            kw_str_to_parse = None
            if isinstance(keywords, list) and len(keywords) == 1 and isinstance(keywords[0], str):
                kw_str_to_parse = keywords[0]
            elif isinstance(keywords, str):
                kw_str_to_parse = keywords

            if kw_str_to_parse:
                # Clean up string representation of list (e.g., '["kw1", "kw2"]')
                if kw_str_to_parse.startswith('[') and kw_str_to_parse.endswith(']'):
                    kw_str_to_parse = kw_str_to_parse[1:-1]
                
                # Split by comma, then clean up quotes and spaces from each item
                processed_keywords = [
                    kw.strip().strip("'\"") for kw in kw_str_to_parse.split(',') if kw.strip()
                ]
            elif isinstance(keywords, list):
                # It's already a list of strings, just clean each item.
                processed_keywords = [
                    kw.strip() for kw in keywords if isinstance(kw, str) and kw.strip()
                ]

        return cls(summary=cleaned_summary, keywords=processed_keywords)


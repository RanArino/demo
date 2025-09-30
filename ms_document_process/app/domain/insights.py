from typing import Iterable, List

from pydantic import BaseModel, Field


class DocumentInsights(BaseModel):
    summary: str = Field(..., description="Summarized content (~300 tokens)")
    keywords: List[str] = Field(default_factory=list, description="Keywords describing the document")

    @classmethod
    def from_generation(cls, summary: str, keywords: Iterable[str]) -> "DocumentInsights":
        cleaned_summary = summary.strip()
        cleaned_keywords = [kw.strip() for kw in keywords if isinstance(kw, str) and kw.strip()]
        return cls(summary=cleaned_summary, keywords=cleaned_keywords)


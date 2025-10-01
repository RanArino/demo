"""Shared types for LLM providers."""
from typing import List, Optional

from pydantic import BaseModel, Field


class InsightRequest(BaseModel):
    """
    Request for generating document insights.
    
    Usages:
        - ms_document_process : generate_insights
    """
    document_text: str = Field(..., description="The document text to analyze")
    title: Optional[str] = Field(None, description="Optional document title for context")
    summary_tokens: int = Field(512, description="Target token count for summary")
    keyword_count: int = Field(10, description="Number of keywords to extract")


class InsightResponse(BaseModel):
    """
    Response containing generated document insights.
    
    Usages:
        - ms_document_process : generate_insights
    """
    summary: str = Field(..., description="Summarized content")
    keywords: List[str] = Field(default_factory=list, description="Extracted keywords")
from pydantic import BaseModel
from uuid import UUID
from typing import Optional, Sequence

class DocumentUploadedEvent(BaseModel):
    content_source_id: UUID
    original_blob_hash: str
    space_id: UUID
    original_object_key: Optional[str] = None
    title: Optional[str] = None
    source: Optional[str] = None

class DocumentProcessedEvent(BaseModel):
    content_source_id: UUID
    space_id: UUID
    processed_blob_hash: Optional[str]
    processed_object_key: Optional[str] = None
    status: str  # e.g., "PROCESSED", "FAILED"
    error_message: Optional[str]
    keywords: Optional[Sequence[str]] = None
    summary: Optional[str] = None
    title: Optional[str] = None
    source: Optional[str] = None
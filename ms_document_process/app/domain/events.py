from pydantic import BaseModel
from uuid import UUID
from typing import Optional

class DocumentUploadedEvent(BaseModel):
    content_source_id: UUID
    original_blob_hash: str
    space_id: UUID
    original_object_key: Optional[str] = None

class DocumentProcessedEvent(BaseModel):
    content_source_id: UUID
    space_id: UUID
    processed_blob_hash: Optional[str]
    status: str  # e.g., "PROCESSED", "FAILED"
    error_message: Optional[str]
-- Update document_content table to add JSON serialization fields
ALTER TABLE document_content
ADD COLUMN IF NOT EXISTS tables_json JSONB,
ADD COLUMN IF NOT EXISTS images_json JSONB,
ADD COLUMN IF NOT EXISTS metadata_json JSONB; 
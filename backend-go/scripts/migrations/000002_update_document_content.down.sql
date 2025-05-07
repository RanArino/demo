-- Remove JSON serialization fields from document_content
ALTER TABLE document_content
DROP COLUMN IF EXISTS tables_json,
DROP COLUMN IF EXISTS images_json,
DROP COLUMN IF EXISTS metadata_json; 
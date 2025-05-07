-- Drop all tables in reverse order to handle references correctly
DROP TABLE IF EXISTS space_members;
DROP TABLE IF EXISTS document_permissions;
DROP TABLE IF EXISTS document_content;
DROP TABLE IF EXISTS document_space_assignments;
DROP TABLE IF EXISTS documents; 
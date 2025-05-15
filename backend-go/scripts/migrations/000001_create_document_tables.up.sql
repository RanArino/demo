-- Create tables for document service

-- Spaces table for organizing documents
CREATE TABLE IF NOT EXISTS spaces (
    id VARCHAR(36) PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    cover_image TEXT,
    keywords TEXT[],
    owner_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_by VARCHAR(36) NOT NULL,
    last_updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_updated_by VARCHAR(36) NOT NULL,
    document_count INTEGER DEFAULT 0,
    total_size_bytes BIGINT DEFAULT 0,
    visibility VARCHAR(20) NOT NULL DEFAULT 'private', -- 'private', 'shared', 'public'
    guest_access_enabled BOOLEAN DEFAULT FALSE,
    guest_access_expiry TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'archived', 'deleted'
    processing_status VARCHAR(20) DEFAULT NULL, -- 'processing', 'completed', 'failed'
    is_personal BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by VARCHAR(36)
);

-- Document metadata table
CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(36) PRIMARY KEY,
    owner_id VARCHAR(36) NOT NULL,
    original_filename TEXT NOT NULL,
    source VARCHAR(50) NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    processed_content_reference TEXT,
    original_file_reference TEXT,
    summary TEXT,
    keywords TEXT[],
    domain_metadata JSONB,
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by VARCHAR(36),
    error_details TEXT,
    processing_attempts INTEGER DEFAULT 0,
    last_processed_at TIMESTAMP WITH TIME ZONE,
    processing_completed_at TIMESTAMP WITH TIME ZONE
);

-- Document space assignments table for many-to-many relationship
CREATE TABLE IF NOT EXISTS document_space_assignments (
    document_id VARCHAR(36) NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    space_id VARCHAR(36) NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL,
    assigned_by VARCHAR(36) NOT NULL,
    PRIMARY KEY (document_id, space_id)
);

-- Document content table
CREATE TABLE IF NOT EXISTS document_content (
    document_id VARCHAR(36) PRIMARY KEY REFERENCES documents(id) ON DELETE CASCADE,
    markdown_content TEXT NOT NULL,
    storage_reference TEXT,
    raw_text TEXT,
    title TEXT,
    author TEXT,
    layout_info JSONB,
    tables_json JSONB,
    images_json JSONB,
    original_file_reference TEXT,
    metadata_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Document permissions table
CREATE TABLE IF NOT EXISTS document_permissions (
    id SERIAL PRIMARY KEY,
    document_id VARCHAR(36) NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL,
    role VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (document_id, user_id)
);

-- Space members table for space-level permissions
CREATE TABLE IF NOT EXISTS space_members (
    id SERIAL PRIMARY KEY,
    space_id VARCHAR(36) NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL,
    email VARCHAR(255),
    role VARCHAR(20) NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL,
    invited_by VARCHAR(36),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'pending', 'declined'
    last_access_at TIMESTAMP WITH TIME ZONE,
    custom_permissions JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (space_id, user_id)
);

-- Create indexes
CREATE INDEX idx_spaces_owner_id ON spaces(owner_id);
CREATE INDEX idx_spaces_visibility ON spaces(visibility);
CREATE INDEX idx_documents_owner_id ON documents(owner_id);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_doc_space_assignments_doc_id ON document_space_assignments(document_id);
CREATE INDEX idx_doc_space_assignments_space_id ON document_space_assignments(space_id);
CREATE INDEX idx_document_permissions_user_id ON document_permissions(user_id);
CREATE INDEX idx_space_members_user_id ON space_members(user_id);
CREATE INDEX idx_space_members_space_id ON space_members(space_id);
CREATE INDEX idx_space_members_email ON space_members(email);
CREATE INDEX idx_space_members_status ON space_members(status); 
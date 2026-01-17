-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For trigram similarity search

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    role VARCHAR(20) NOT NULL DEFAULT 'patron' CHECK (role IN ('admin', 'librarian', 'patron')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Collections table
CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    slug VARCHAR(255) UNIQUE NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT true,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    metadata JSONB,
    settings JSONB,
    document_count BIGINT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Documents table
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'active', 'rejected', 'archived')),
    
    -- File information
    original_filename VARCHAR(500) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    hash VARCHAR(64) UNIQUE NOT NULL, -- SHA-256
    
    -- Extracted content
    extracted_text TEXT,
    page_count INTEGER,
    language VARCHAR(10),
    
    -- Metadata
    metadata JSONB,
    
    -- Processing
    is_indexed BOOLEAN NOT NULL DEFAULT false,
    indexed_at TIMESTAMP,
    processing_error TEXT,
    
    -- Stats
    view_count BIGINT NOT NULL DEFAULT 0,
    download_count BIGINT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Document versions table
CREATE TABLE document_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    change_log TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(document_id, version_number)
);

-- Search queries table
CREATE TABLE search_queries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    query_text VARCHAR(500) NOT NULL,
    filters JSONB,
    result_count INTEGER,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Access logs table
CREATE TABLE access_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('view', 'download')),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================
-- INDEXES
-- ============================================

-- Users indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Collections indexes
CREATE INDEX idx_collections_owner_id ON collections(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_slug ON collections(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_owner_public ON collections(owner_id, is_public) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_deleted_at ON collections(deleted_at);
CREATE INDEX idx_collections_name ON collections(name) WHERE deleted_at IS NULL;

-- JSONB indexes for collections
CREATE INDEX idx_collections_metadata ON collections USING GIN(metadata);
CREATE INDEX idx_collections_settings ON collections USING GIN(settings);

-- Documents indexes
CREATE INDEX idx_documents_collection_id ON documents(collection_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_uploader_id ON documents(uploader_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_status ON documents(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_hash ON documents(hash);
CREATE INDEX idx_documents_is_indexed ON documents(is_indexed) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_deleted_at ON documents(deleted_at);
CREATE INDEX idx_documents_title ON documents(title) WHERE deleted_at IS NULL;

-- Composite indexes for common queries
CREATE INDEX idx_documents_collection_status ON documents(collection_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_uploader_created ON documents(uploader_id, created_at DESC) WHERE deleted_at IS NULL;

-- JSONB index for document metadata
CREATE INDEX idx_documents_metadata ON documents USING GIN(metadata);

-- Full-text search index on extracted_text
CREATE INDEX idx_documents_extracted_text ON documents USING GIN(to_tsvector('english', COALESCE(extracted_text, '')));

-- Trigram indexes for fuzzy search on title
CREATE INDEX idx_documents_title_trgm ON documents USING GIN(title gin_trgm_ops);

-- Document versions indexes
CREATE INDEX idx_document_versions_document_id ON document_versions(document_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_document_versions_created_by ON document_versions(created_by) WHERE deleted_at IS NULL;

-- Search queries indexes
CREATE INDEX idx_search_queries_user_id ON search_queries(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_search_queries_query_text ON search_queries(query_text);
CREATE INDEX idx_search_queries_created_at ON search_queries(created_at DESC);

-- Access logs indexes
CREATE INDEX idx_access_logs_document_id ON access_logs(document_id);
CREATE INDEX idx_access_logs_user_id ON access_logs(user_id);
CREATE INDEX idx_access_logs_created_at ON access_logs(created_at DESC);
CREATE INDEX idx_access_logs_action ON access_logs(action);

-- ============================================
-- COMMENTS (Documentation)
-- ============================================

COMMENT ON TABLE users IS 'System users with authentication and authorization';
COMMENT ON TABLE collections IS 'Groups of related documents';
COMMENT ON TABLE documents IS 'Digital documents/resources with file metadata';
COMMENT ON TABLE document_versions IS 'Version history for documents';
COMMENT ON TABLE search_queries IS 'Logged search queries for analytics';
COMMENT ON TABLE access_logs IS 'Document access tracking for analytics';

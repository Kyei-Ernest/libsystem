-- Phase 2: Fine-Grained Permissions
-- Adds document-level permissions and collection sharing

-- Document permissions table
CREATE TABLE IF NOT EXISTS document_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(20) NOT NULL CHECK (permission IN ('view', 'edit', 'delete', 'admin')),
    granted_by UUID NOT NULL REFERENCES users(id),
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(document_id, user_id, permission)
);

CREATE INDEX idx_doc_permissions_document ON document_permissions(document_id);
CREATE INDEX idx_doc_permissions_user ON document_permissions(user_id);
CREATE INDEX idx_doc_permissions_granted_by ON document_permissions(granted_by);

-- Collection sharing table
CREATE TABLE IF NOT EXISTS collection_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    shared_with_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(20) NOT NULL CHECK (permission IN ('view', 'edit', 'admin')),
    shared_by UUID NOT NULL REFERENCES users(id),
    shared_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(collection_id, shared_with_user_id)
);

CREATE INDEX idx_collection_shares_collection ON collection_shares(collection_id);
CREATE INDEX idx_collection_shares_user ON collection_shares(shared_with_user_id);
CREATE INDEX idx_collection_shares_shared_by ON collection_shares(shared_by);

-- Comments for documentation
COMMENT ON TABLE document_permissions IS 'Fine-grained access control for individual documents';
COMMENT ON TABLE collection_shares IS 'Share entire collections with specific users';
COMMENT ON COLUMN document_permissions.permission IS 'Permission level: view, edit, delete, admin';
COMMENT ON COLUMN collection_shares.permission IS 'Permission level: view (read-only), edit (add/remove docs), admin (full control)';

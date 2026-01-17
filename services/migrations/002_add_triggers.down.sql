-- Drop all triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_collections_updated_at ON collections;
DROP TRIGGER IF EXISTS update_documents_updated_at ON documents;
DROP TRIGGER IF EXISTS update_document_versions_updated_at ON document_versions;
DROP TRIGGER IF EXISTS update_search_queries_updated_at ON search_queries;

DROP TRIGGER IF EXISTS increment_doc_count ON documents;
DROP TRIGGER IF EXISTS decrement_doc_count_delete ON documents;
DROP TRIGGER IF EXISTS decrement_doc_count_soft_delete ON documents;
DROP TRIGGER IF EXISTS increment_doc_count_restore ON documents;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS increment_collection_document_count();
DROP FUNCTION IF EXISTS decrement_collection_document_count();

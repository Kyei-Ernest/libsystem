-- ============================================
-- TRIGGER: Auto-update updated_at timestamp
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_collections_updated_at BEFORE UPDATE ON collections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_document_versions_updated_at BEFORE UPDATE ON document_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_search_queries_updated_at BEFORE UPDATE ON search_queries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- TRIGGER: Auto-increment collection document count
-- ============================================

CREATE OR REPLACE FUNCTION increment_collection_document_count()
RETURNS TRIGGER AS $$
BEGIN
    -- Increment count when new document added
    UPDATE collections 
    SET document_count = document_count + 1
    WHERE id = NEW.collection_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION decrement_collection_document_count()
RETURNS TRIGGER AS $$
BEGIN
    -- Decrement count when document removed (soft or hard delete)
    UPDATE collections 
    SET document_count = document_count - 1
    WHERE id = OLD.collection_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Trigger when document inserted
CREATE TRIGGER increment_doc_count AFTER INSERT ON documents
    FOR EACH ROW EXECUTE FUNCTION increment_collection_document_count();

-- Trigger when document hard deleted
CREATE TRIGGER decrement_doc_count_delete AFTER DELETE ON documents
    FOR EACH ROW EXECUTE FUNCTION decrement_collection_document_count();

-- Trigger when document soft deleted (updated with deleted_at)
CREATE TRIGGER decrement_doc_count_soft_delete AFTER UPDATE OF deleted_at ON documents
    FOR EACH ROW 
    WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
    EXECUTE FUNCTION decrement_collection_document_count();

-- Trigger when document undeleted (deleted_at set to NULL)
CREATE TRIGGER increment_doc_count_restore AFTER UPDATE OF deleted_at ON documents
    FOR EACH ROW 
    WHEN (OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL)
    EXECUTE FUNCTION increment_collection_document_count();

-- ============================================
-- TRIGGER: Update collection view_count
-- ============================================
-- Note: This is logged via access_logs, but we can add a trigger
-- to auto-increment if desired. For now, application handles it.

COMMENT ON FUNCTION update_updated_at_column() IS 'Automatically updates the updated_at timestamp';
COMMENT ON FUNCTION increment_collection_document_count() IS 'Increments collection document count when document added';
COMMENT ON FUNCTION decrement_collection_document_count() IS 'Decrements collection document count when document removed';

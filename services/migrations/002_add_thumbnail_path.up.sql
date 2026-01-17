-- Add thumbnail_path to documents table
ALTER TABLE documents ADD COLUMN IF NOT EXISTS thumbnail_path VARCHAR(500);

-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS access_logs CASCADE;
DROP TABLE IF EXISTS search_queries CASCADE;
DROP TABLE IF EXISTS document_versions CASCADE;
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS collections CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "uuid-ossp";

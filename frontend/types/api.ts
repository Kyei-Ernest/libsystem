// ============================================
// LibSystem Frontend Type Definitions
// Matches Backend API Contracts
// ============================================

// ==================== User & Auth Types ====================

export interface User {
    id: string;
    email: string;
    username: string;
    first_name?: string;
    last_name?: string;
    role: UserRole;
    is_active: boolean;
    last_login_at?: string;
    created_at: string;
    updated_at: string;
}

export type UserRole = 'admin' | 'librarian' | 'patron' | 'archivist' | 'vendor';

// ==================== Document Types ====================
// Matches: Document Service (Port 8081)

export type FileType = 'pdf' | 'docx' | 'txt' | 'jpg' | 'jpeg' | 'png' | 'gif' | 'html' | 'htm';
export type IndexingStatus = 'pending' | 'completed' | 'failed';

export interface Document {
    id: string;
    title: string;
    description?: string;
    original_filename: string;
    file_name: string;
    file_type: FileType;
    file_size: number;
    file_path: string;
    thumbnail_path?: string;
    file_hash: string;
    collection_id: string;
    uploader_id: string;
    indexing_status: IndexingStatus;
    created_at: string;
    updated_at: string;
}

// ==================== Collection Types ====================
// Matches: Collection Service (Port 8082)

export interface CollectionSettings {
    allow_public_submissions: boolean;
    require_approval: boolean;
    allowed_file_types?: string[];
    max_file_size?: number;
}

export interface CollectionStats {
    document_count: number;
    view_count: number;
}

export interface Collection {
    id: string;
    name: string;
    description?: string;
    slug: string;
    owner_id: string;
    is_public: boolean;
    document_count?: number; // Optional count of documents in collection
    settings?: CollectionSettings;
    stats?: CollectionStats;
    created_at: string;
    updated_at: string;
}

// ==================== Search Types ====================
// Matches: Search Service (Port 8084)

export interface SearchFilters {
    file_type?: string;
    collection_id?: string;
    uploader_id?: string;
}

export interface SearchQuery {
    query: string;
    filters?: SearchFilters;
    from: number;
    size: number;
}

export interface SearchHit {
    document: Document;
    score: number;
    highlights?: Record<string, string[]>;
}

export interface SearchFacets {
    file_types: Record<string, number>;
    collections: Record<string, number>;
    uploaders?: Record<string, number>;
}

export interface SearchResult {
    hits: SearchHit[];
    total: number;
    facets?: SearchFacets;
    took: number; // milliseconds
}

// ==================== API Response Types ====================

export interface ApiError {
    code: string;
    message: string;
    details?: unknown;
}

export interface ApiResponse<T = unknown> {
    success: boolean;
    data?: T;
    error?: ApiError;
}

// ==================== Authentication Types ====================

export interface LoginRequest {
    email_or_username: string;
    password: string;
}

export interface RegisterRequest {
    email: string;
    username: string;
    password: string;
    first_name: string;
    last_name: string;
}

export interface AuthResponse {
    token: string;
    user: User;
}

// ==================== Pagination Types ====================

export interface PaginationParams {
    page: number;
    limit: number;
}

export interface PaginatedResponse<T> {
    data: T[];
    pagination: {
        page: number;
        page_size: number;
        total_items: number;
        total_pages: number;
    };
}

// ==================== Upload Types ====================

export interface UploadProgress {
    loaded: number;
    total: number;
    percentage: number;
}

export interface DocumentUploadRequest {
    file: File;
    title: string;
    collection_id: string;
    onProgress?: (progress: UploadProgress) => void;
}

export interface DocumentUploadResponse {
    document: Document;
    message: string;
}

// ==================== Form Types ====================

export interface LoginFormData {
    email: string;
    password: string;
}

export interface RegisterFormData {
    email: string;
    password: string;
    confirmPassword: string;
}

export interface CollectionFormData {
    name: string;
    description: string;
    is_public: boolean;
}

// ==================== UI State Types ====================

export interface LoadingState {
    isLoading: boolean;
    error: string | null;
}

export interface SortConfig {
    field: string;
    direction: 'asc' | 'desc';
}

export interface FilterState {
    search: string;
    fileType?: FileType;
    collectionId?: string;
    uploaderId?: string;
    indexingStatus?: IndexingStatus;
}

// ==================== Rate Limit Types ====================

export interface RateLimitInfo {
    limit: number;
    remaining: number;
    reset: number; // timestamp
}

export interface RateLimitError extends ApiError {
    retry_after: number; // seconds
}

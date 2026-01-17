import { apiClient } from './client';
import type { Document, PaginatedResponse, ApiResponse } from '@/types/api';

export interface DocumentUploadData {
    file: File;
    collection_id?: string;
    metadata?: {
        title?: string;
        description?: string;
        tags?: string[];
    };
    onProgress?: (progress: number) => void;
}

export const documentsApi = {
    /**
     * Upload a new document
     */
    async upload(data: DocumentUploadData): Promise<Document> {
        const formData = new FormData();
        formData.append('file', data.file);

        // Backend requires collection_id - use a placeholder if not provided
        const collectionId = data.collection_id || '00000000-0000-0000-0000-000000000000';
        formData.append('collection_id', collectionId);

        // Title is required
        const title = data.metadata?.title || data.file.name;
        formData.append('title', title);

        // Description is optional
        if (data.metadata?.description) {
            formData.append('description', data.metadata.description);
        }

        const document = await apiClient.post<Document>('/documents', formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
            onUploadProgress: (progressEvent) => {
                const total = progressEvent.total || 1;
                const progress = Math.round((progressEvent.loaded * 100) / total);
                if (data.onProgress) {
                    data.onProgress(progress);
                }
            },
        });

        return document;
    },

    /**
     * List all documents with pagination
     * Note: Uses raw axios because backend returns {success, data: [], pagination: {}} 
     * while apiClient.get() only extracts response.data.data
     */
    async list(params?: {
        page?: number;
        limit?: number;
        collection_id?: string;
        search?: string;
        file_type?: string;
        status?: string;
    }): Promise<PaginatedResponse<Document>> {
        const { data } = await apiClient.getInstance().get<ApiResponse<Document[]> & { pagination: PaginatedResponse<Document>['pagination'] }>('/documents', { params });

        if (!data.success) {
            throw new Error(data.error?.message || 'Request failed');
        }

        return {
            data: data.data || [],
            pagination: data.pagination || { page: 1, page_size: 10, total_items: 0, total_pages: 0 }
        };
    },

    /**
     * Get a single document by ID
     */
    async getById(id: string): Promise<Document> {
        return apiClient.get<Document>(`/documents/${id}`);
    },

    /**
     * Update document metadata
     */
    async update(id: string, data: Partial<Document>): Promise<Document> {
        return apiClient.put<Document>(`/documents/${id}`, data);
    },

    /**
     * Delete a document
     */
    async delete(id: string): Promise<void> {
        await apiClient.delete(`/documents/${id}`);
    },

    /**
    /**
     * View/preview a document
     */
    getViewUrl(id: string): string {
        return `/api/v1/documents/${id}/view`;
    },

    /**
     * Download a document
     */
    getDownloadUrl(id: string): string {
        return `/api/v1/documents/${id}/download`;
    },

    /**
     * Get document thumbnail URL
     */
    getThumbnailUrl(id: string): string {
        return `/api/v1/documents/${id}/thumbnail`;
    },
};

import { apiClient } from './client';
import type { Collection, CollectionSettings, PaginatedResponse, ApiResponse } from '@/types/api';

export interface CollectionCreateRequest {
    name: string;
    description?: string;
    is_public?: boolean;
    settings?: CollectionSettings;
}

export interface CollectionUpdateRequest {
    name?: string;
    description?: string;
    is_public?: boolean;
    settings?: CollectionSettings;
}

export interface CollectionListFilters {
    page?: number;
    page_size?: number;
    search?: string;
    owner_id?: string;
    is_public?: boolean;
}

export const collectionsApi = {
    /**
     * Create a new collection
     */
    async create(data: CollectionCreateRequest): Promise<Collection> {
        return apiClient.post<Collection>('/collections', data);
    },

    /**
     * Get a collection by ID
     */
    async getById(id: string): Promise<Collection> {
        return apiClient.get<Collection>(`/collections/${id}`);
    },

    /**
     * Get a collection by slug
     */
    async getBySlug(slug: string): Promise<Collection> {
        return apiClient.get<Collection>(`/collections/slug/${slug}`);
    },

    /**
     * List all collections with optional filters
     */
    async list(filters?: CollectionListFilters): Promise<PaginatedResponse<Collection>> {
        const { data } = await apiClient.getInstance().get<ApiResponse<Collection[]> & { pagination: PaginatedResponse<Collection>['pagination'] }>('/collections', { params: filters });

        if (!data.success) {
            throw new Error(data.error?.message || 'Request failed');
        }

        return {
            data: data.data || [],
            pagination: data.pagination || { page: 1, page_size: 20, total_items: 0, total_pages: 0 }
        };
    },

    /**
     * Update a collection
     */
    async update(id: string, data: CollectionUpdateRequest): Promise<Collection> {
        return apiClient.put<Collection>(`/collections/${id}`, data);
    },

    /**
     * Delete a collection
     */
    async delete(id: string): Promise<void> {
        await apiClient.delete(`/collections/${id}`);
    },

    /**
     * Get collection statistics
     */
    async getStats(id: string): Promise<{ document_count: number; view_count: number; created_at: string; updated_at: string }> {
        return apiClient.get(`/collections/${id}/stats`);
    },
};

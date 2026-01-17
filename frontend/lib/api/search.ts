import { apiClient } from './client';
import type { ApiResponse } from '@/types/api';

export interface SearchParams {
    query: string;
    from?: number;
    size?: number;
    file_type?: string;
    collection_id?: string;
}

export interface SearchHit {
    id: string;
    title: string;
    description?: string;
    original_filename: string;
    file_type: string;
    file_size: number;
    collection_id: string;
    uploader_id: string;
    created_at: string;
    score?: number; // Optional as backend doesn't currently return it in the flat structure
}

export interface SearchResult {
    hits: SearchHit[];
    total: number;
    facets?: Record<string, Record<string, number>>;
}

export const searchApi = {
    /**
     * Search for documents
     */
    async search(params: SearchParams): Promise<SearchResult> {
        const { data } = await apiClient.getInstance().get<ApiResponse<SearchResult>>('/search', {
            params: {
                q: params.query,
                from: params.from || 0,
                size: params.size || 10,
                file_type: params.file_type,
                collection_id: params.collection_id,
            },
        });

        if (!data.success || !data.data) {
            throw new Error(data.error?.message || 'Search failed');
        }

        return data.data;
    },

    /**
     * Advanced search
     */
    async advancedSearch(params: any): Promise<SearchResult> {
        const { data } = await apiClient.getInstance().post<ApiResponse<SearchResult>>('/search/advanced', params);

        if (!data.success || !data.data) {
            throw new Error(data.error?.message || 'Search failed');
        }

        return data.data;
    }
};

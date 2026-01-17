import { apiClient } from './client';
import type { User, PaginatedResponse, ApiResponse } from '@/types/api';

export interface UserListFilters {
    page?: number;
    page_size?: number;
    search?: string;
    role?: string;
    is_active?: boolean;
}

export interface UserUpdateRequest {
    first_name?: string;
    last_name?: string;
    role?: string;
    is_active?: boolean;
}

export const usersApi = {
    /**
     * List all users with optional filters
     */
    async list(filters?: UserListFilters): Promise<PaginatedResponse<User>> {
        const { data } = await apiClient.getInstance().get<ApiResponse<User[]> & { pagination: PaginatedResponse<User>['pagination'] }>('/users', { params: filters });

        if (!data.success) {
            throw new Error(data.error?.message || 'Request failed');
        }

        return {
            data: data.data || [],
            pagination: data.pagination || { page: 1, page_size: 20, total_items: 0, total_pages: 0 }
        };
    },

    /**
     * Get a user by ID
     */
    async getById(id: string): Promise<User> {
        return apiClient.get<User>(`/users/${id}`);
    },

    /**
     * Update a user
     */
    async update(id: string, data: UserUpdateRequest): Promise<User> {
        return apiClient.put<User>(`/users/${id}`, data);
    },

    /**
     * Delete a user
     */
    async delete(id: string): Promise<void> {
        await apiClient.delete(`/users/${id}`);
    },
};

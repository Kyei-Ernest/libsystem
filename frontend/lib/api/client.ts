import axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { ApiResponse } from '@/types/api';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ||
    (typeof window === 'undefined' ? 'http://localhost:8080/api/v1' : '/api/v1');

/**
 * Custom API Client for LibSystem
 * Handles authentication, error handling, and rate limiting
 */
class ApiClient {
    private client: AxiosInstance;

    constructor() {
        this.client = axios.create({
            baseURL: API_BASE_URL,
            headers: {
                'Content-Type': 'application/json',
            },
            timeout: 30000, // 30 seconds
        });

        this.setupInterceptors();
    }

    private setupInterceptors() {
        // Request interceptor: Add JWT token
        this.client.interceptors.request.use(
            (config) => {
                // Only add auth header on client side
                if (typeof window !== 'undefined') {
                    const token = localStorage.getItem('auth_token');
                    if (token && config.headers) {
                        config.headers.Authorization = `Bearer ${token}`;
                    }
                }
                return config;
            },
            (error) => Promise.reject(error)
        );

        // Response interceptor: Handle errors
        this.client.interceptors.response.use(
            (response) => response,
            (error: AxiosError<ApiResponse>) => {
                // Handle 401 Unauthorized
                if (error.response?.status === 401) {
                    if (typeof window !== 'undefined') {
                        localStorage.removeItem('auth_token');
                        // Only redirect if not already on login page
                        if (!window.location.pathname.includes('/login')) {
                            window.location.href = '/login';
                        }
                    }
                }

                // Handle 429 Rate Limit
                if (error.response?.status === 429) {
                    const retryAfter = error.response.headers['retry-after'];
                    console.warn(` Rate limit exceeded. Retry after ${retryAfter}s`);

                    // Add retry after info to error
                    if (error.response.data && error.response.data.error) {
                        error.response.data.error = {
                            ...error.response.data.error,
                            code: 'RATE_LIMIT_EXCEEDED',
                            message: `Rate limit exceeded. Please retry after ${retryAfter} seconds.`,
                        };
                    }
                }

                // Handle network errors
                if (!error.response) {
                    const networkError: ApiResponse = {
                        success: false,
                        error: {
                            code: 'NETWORK_ERROR',
                            message: 'Unable to connect to server. Please check your internet connection.',
                        },
                    };
                    error.response = {
                        data: networkError,
                        status: 0,
                        statusText: 'Network Error',
                        headers: {},
                        config: error.config as unknown,
                    } as AxiosResponse<ApiResponse>;
                }

                return Promise.reject(error);
            }
        );
    }

    /**
     * Get the Axios instance
     */
    getInstance(): AxiosInstance {
        return this.client;
    }

    /**
     * Helper: GET request
     */
    async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
        const response = await this.client.get<ApiResponse<T>>(url, config);
        if (!response.data.success) {
            throw new Error(response.data.error?.message || 'Request failed');
        }
        return response.data.data as T;
    }

    /**
     * Helper: POST request
     */
    async post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
        const response = await this.client.post<ApiResponse<T>>(url, data, config);
        if (!response.data.success) {
            throw new Error(response.data.error?.message || 'Request failed');
        }
        return response.data.data as T;
    }

    /**
     * Helper: PUT request
     */
    async put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
        const response = await this.client.put<ApiResponse<T>>(url, data, config);
        if (!response.data.success) {
            throw new Error(response.data.error?.message || 'Request failed');
        }
        return response.data.data as T;
    }

    /**
     * Helper: DELETE request
     */
    async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
        const response = await this.client.delete<ApiResponse<T>>(url, config);
        if (!response.data.success) {
            throw new Error(response.data.error?.message || 'Request failed');
        }
        return response.data.data as T;
    }
}

// Export singleton instance
export const apiClient = new ApiClient();

// Export raw Axios instance for advanced usage
export const axiosInstance = apiClient.getInstance();

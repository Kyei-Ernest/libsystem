import { apiClient } from './client';
import type {
    LoginRequest,
    AuthResponse,
    User,
    UserRole,
} from '@/types/api';

export interface RegisterRequest {
    email: string;
    username: string;
    password: string;
    first_name: string;
    last_name: string;
    role?: UserRole; // Optional role selection
}

export const authApi = {
    /**
     * Login with email and password
   */
    async login(credentials: { email: string; password: string }): Promise<AuthResponse> {
        return apiClient.post<AuthResponse>('/auth/login', {
            email_or_username: credentials.email,
            password: credentials.password
        });
    },

    /**
     * Register a new user
     */
    async register(data: RegisterRequest): Promise<AuthResponse> {
        return apiClient.post<AuthResponse>('/auth/register', data);
    },

    /**
     * Get current user profile
     */
    async getProfile(): Promise<User> {
        return apiClient.get<User>('/auth/me');
    },

    /**
     * Logout (client-side token removal)
     */
    logout(): void {
        if (typeof window !== 'undefined') {
            localStorage.removeItem('auth_token');
        }
    },
};

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { User } from '@/types/api';

interface AuthState {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    setAuth: (user: User, token: string) => void;
    clearAuth: () => void;
    updateUser: (user: Partial<User>) => void;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set, get) => ({
            user: null,
            token: null,
            isAuthenticated: false,

            setAuth: (user, token) => {
                // Store token in localStorage for API client
                if (typeof window !== 'undefined') {
                    localStorage.setItem('auth_token', token);
                }
                set({ user, token, isAuthenticated: true });
            },

            clearAuth: () => {
                // Remove token from localStorage
                if (typeof window !== 'undefined') {
                    localStorage.removeItem('auth_token');
                }
                set({ user: null, token: null, isAuthenticated: false });
            },

            updateUser: (updatedFields) => {
                const currentUser = get().user;
                if (currentUser) {
                    set({ user: { ...currentUser, ...updatedFields } });
                }
            },
        }),
        {
            name: 'auth-storage',
            storage: createJSONStorage(() => localStorage),
            partialize: (state) => ({
                user: state.user,
                token: state.token,
                isAuthenticated: state.isAuthenticated,
            }),
        }
    )
);

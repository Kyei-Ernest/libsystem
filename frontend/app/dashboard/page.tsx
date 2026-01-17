'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';

export default function DashboardHome() {
    const router = useRouter();
    const { user } = useAuthStore();

    useEffect(() => {
        if (!user) {
            router.push('/login');
            return;
        }

        // Redirect based on role
        switch (user.role) {
            case 'admin':
                router.push('/dashboard/admin');
                break;
            case 'librarian':
                router.push('/dashboard/librarian');
                break;
            case 'archivist':
                router.push('/dashboard/archivist');
                break;
            case 'vendor':
                router.push('/dashboard/vendor');
                break;
            case 'patron':
                router.push('/dashboard/patron');
                break;
            default:
                // Default to patron dashboard
                router.push('/dashboard/patron');
        }
    }, [user, router]);

    return (
        <div className="flex items-center justify-center min-h-screen">
            <p className="text-muted-foreground">Redirecting...</p>
        </div>
    );
}

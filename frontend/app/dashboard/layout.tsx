'use client';

import { useAuthStore } from '@/stores/auth-store';
import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import {
    LayoutDashboard,
    FileText,
    FolderOpen,
    Search,
    Settings,
    LogOut,
    Menu,
    Users,
    Shield
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger, SheetTitle, SheetDescription } from '@/components/ui/sheet';
import { cn } from '@/lib/utils';
import { useState } from 'react';
import { Badge } from '@/components/ui/badge';

// Role-based navigation configuration
const getNavigationForRole = (role?: string) => {
    // Role-specific overview link
    const roleOverview = role ? `/dashboard/${role}` : '/dashboard/patron';

    const baseNav = [
        { name: 'Overview', href: roleOverview, icon: LayoutDashboard, roles: ['admin', 'librarian', 'patron', 'archivist', 'vendor'] },
        { name: 'Documents', href: '/dashboard/documents', icon: FileText, roles: ['admin', 'librarian', 'patron', 'archivist', 'vendor'] },
        { name: 'Collections', href: '/dashboard/collections', icon: FolderOpen, roles: ['admin', 'librarian', 'patron', 'archivist', 'vendor'] },
        { name: 'Search', href: '/dashboard/search', icon: Search, roles: ['admin', 'librarian', 'patron', 'archivist', 'vendor'] },
    ];

    const adminOnlyNav = [
        { name: 'Users', href: '/dashboard/users', icon: Users, roles: ['admin'] },
        { name: 'Settings', href: '/dashboard/settings', icon: Settings, roles: ['admin'] },
    ];

    const allNav = [...baseNav, ...adminOnlyNav];

    // Filter based on role
    return allNav.filter(item => item.roles.includes(role || 'patron'));
};

const getRoleBadgeVariant = (role?: string) => {
    switch (role?.toLowerCase()) {
        case 'admin': return 'destructive';
        case 'librarian': return 'default';
        default: return 'outline';
    }
};

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const router = useRouter();
    const pathname = usePathname();
    const { user, clearAuth } = useAuthStore();
    const [isMobileOpen, setIsMobileOpen] = useState(false);

    const navigation = getNavigationForRole(user?.role);

    const handleLogout = () => {
        clearAuth();
        router.push('/login');
    };

    return (
        <div className="flex min-h-screen bg-gray-50">
            {/* Sidebar (Desktop) */}
            <aside className="hidden w-64 flex-col border-r bg-white md:flex">
                <div className="flex h-16 items-center px-6">
                    <Link href="/dashboard" className="flex items-center gap-2 font-bold text-xl">
                        <span>LibSystem</span>
                    </Link>
                </div>
                <nav className="flex-1 space-y-1 px-3 py-4">
                    {navigation.map((item) => {
                        const isActive = pathname === item.href;
                        return (
                            <Link
                                key={item.name}
                                href={item.href}
                                className={cn(
                                    'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                                    isActive
                                        ? 'bg-blue-50 text-blue-700'
                                        : 'text-gray-700 hover:bg-gray-100'
                                )}
                            >
                                <item.icon className="h-5 w-5" />
                                {item.name}
                            </Link>
                        );
                    })}
                </nav>
                <div className="border-t p-4">
                    <div className="flex items-center gap-3 rounded-md bg-gray-50 p-3">
                        <div className="flex-1 truncate">
                            <div className="flex items-center gap-2 mb-0.5">
                                <p className="text-sm font-medium truncate">{user?.username || 'User'}</p>
                                {user?.role && (
                                    <Badge variant={getRoleBadgeVariant(user.role)} className="h-5 px-1.5 text-[10px] uppercase">
                                        {user.role}
                                    </Badge>
                                )}
                            </div>
                            <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                        </div>
                    </div>
                    <Button
                        variant="ghost"
                        className="mt-2 w-full justify-start text-red-600 hover:bg-red-50 hover:text-red-700"
                        onClick={handleLogout}
                    >
                        <LogOut className="mr-2 h-4 w-4" />
                        Logout
                    </Button>
                </div>
            </aside>

            {/* Main Content */}
            <div className="flex flex-1 flex-col w-full overflow-x-hidden">
                {/* Header (Mobile) */}
                <header className="sticky top-0 z-10 flex h-16 items-center border-b bg-white px-4 md:hidden">
                    <Sheet open={isMobileOpen} onOpenChange={setIsMobileOpen}>
                        <SheetTrigger asChild>
                            <Button variant="ghost" size="icon" className="-ml-2">
                                <Menu className="h-6 w-6" />
                            </Button>
                        </SheetTrigger>
                        <SheetContent side="left" className="w-64 p-0">
                            <SheetTitle className="sr-only">Navigation Menu</SheetTitle>
                            <SheetDescription className="sr-only">
                                Main navigation menu for accessing dashboard features.
                            </SheetDescription>
                            <div className="flex h-16 items-center px-6 border-b">
                                <SheetTitle className="font-bold text-xl">LibSystem</SheetTitle>
                            </div>
                            <nav className="flex-1 space-y-1 px-3 py-4">
                                {navigation.map((item) => {
                                    const isActive = pathname === item.href;
                                    return (
                                        <Link
                                            key={item.name}
                                            href={item.href}
                                            onClick={() => setIsMobileOpen(false)}
                                            className={cn(
                                                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                                                isActive
                                                    ? 'bg-blue-50 text-blue-700'
                                                    : 'text-gray-700 hover:bg-gray-100'
                                            )}
                                        >
                                            <item.icon className="h-5 w-5" />
                                            {item.name}
                                        </Link>
                                    );
                                })}
                            </nav>
                            <div className="border-t p-4">
                                <div className="flex items-center gap-3 rounded-md bg-gray-50 p-3 mb-2">
                                    <div className="flex-1 truncate">
                                        <div className="flex items-center gap-2 mb-0.5">
                                            <p className="text-sm font-medium truncate">{user?.username || 'User'}</p>
                                            {user?.role && (
                                                <Badge variant={getRoleBadgeVariant(user.role)} className="h-5 px-1.5 text-[10px] uppercase">
                                                    {user.role}
                                                </Badge>
                                            )}
                                        </div>
                                        <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                                    </div>
                                </div>
                                <Button
                                    variant="ghost"
                                    className="w-full justify-start text-red-600 hover:bg-red-50 hover:text-red-700"
                                    onClick={handleLogout}
                                >
                                    <LogOut className="mr-2 h-4 w-4" />
                                    Logout
                                </Button>
                            </div>
                        </SheetContent>
                    </Sheet>
                    <span className="ml-4 font-semibold">Dashboard</span>
                </header>

                {/* Page Content */}
                <main className="flex-1 p-2 sm:p-4 md:p-8">
                    {children}
                </main>
            </div>
        </div>
    );
}

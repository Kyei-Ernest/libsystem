'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { usersApi } from '@/lib/api/users';
import { useAuthStore } from '@/stores/auth-store';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
    Table, TableBody, TableCell, TableHead, TableHeader, TableRow
} from '@/components/ui/table';
import {
    DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel,
    DropdownMenuSeparator, DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import {
    Search, User as UserIcon, MoreVertical, Shield, ShieldAlert,
    CheckCircle2, XCircle, Loader2, RefreshCw
} from 'lucide-react';
import { toast } from 'sonner';
import type { User } from '@/types/api';

export default function UsersPage() {
    const router = useRouter();
    const { user: currentUser } = useAuthStore();
    const [users, setUsers] = useState<User[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [totalUsers, setTotalUsers] = useState(0);

    useEffect(() => {
        loadUsers();
    }, [page, searchQuery]);

    const loadUsers = async () => {
        try {
            setIsLoading(true);
            const response = await usersApi.list({
                page,
                page_size: 10,
                search: searchQuery
            });
            setUsers(response.data);
            setTotalPages(response.pagination.total_pages);
            setTotalUsers(response.pagination.total_items);
        } catch (error) {
            console.error('Failed to load users:', error);
            toast.error('Failed to load users');
        } finally {
            setIsLoading(false);
        }
    };

    const handleRoleUpdate = async (userId: string, newRole: string) => {
        try {
            await usersApi.update(userId, { role: newRole });
            toast.success('User role updated successfully');
            loadUsers();
        } catch (error) {
            toast.error('Failed to update user role');
        }
    };

    const handleStatusUpdate = async (userId: string, isActive: boolean) => {
        try {
            await usersApi.update(userId, { is_active: isActive });
            toast.success(`User ${isActive ? 'activated' : 'deactivated'} successfully`);
            loadUsers();
        } catch (error) {
            toast.error('Failed to update user status');
        }
    };

    const getRoleBadgeColor = (role: string) => {
        switch (role) {
            case 'admin': return 'bg-red-100 text-red-800 border-red-200';
            case 'librarian': return 'bg-blue-100 text-blue-800 border-blue-200';
            case 'archivist': return 'bg-purple-100 text-purple-800 border-purple-200';
            case 'vendor': return 'bg-orange-100 text-orange-800 border-orange-200';
            default: return 'bg-green-100 text-green-800 border-green-200';
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">User Management</h2>
                    <p className="text-muted-foreground">Manage system access and roles</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" onClick={() => loadUsers()} disabled={isLoading}>
                        <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                        Refresh
                    </Button>
                </div>
            </div>

            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>All Users</CardTitle>
                            <CardDescription>
                                Total {totalUsers} registered users
                            </CardDescription>
                        </div>
                        <div className="relative w-full max-w-xs">
                            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search users..."
                                className="pl-9"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>User</TableHead>
                                    <TableHead>Role</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Joined</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {isLoading ? (
                                    <TableRow>
                                        <TableCell colSpan={5} className="h-24 text-center">
                                            <Loader2 className="h-6 w-6 animate-spin mx-auto text-muted-foreground" />
                                        </TableCell>
                                    </TableRow>
                                ) : users.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                                            No users found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    users.map((user) => (
                                        <TableRow key={user.id}>
                                            <TableCell>
                                                <div className="flex items-center gap-3">
                                                    <div className="h-8 w-8 rounded-full bg-gray-100 flex items-center justify-center">
                                                        <UserIcon className="h-4 w-4 text-gray-500" />
                                                    </div>
                                                    <div>
                                                        <div className="font-medium">
                                                            {user.first_name} {user.last_name}
                                                        </div>
                                                        <div className="text-xs text-muted-foreground">
                                                            {user.email}
                                                        </div>
                                                    </div>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <Badge
                                                    variant="outline"
                                                    className={getRoleBadgeColor(user.role)}
                                                >
                                                    {user.role}
                                                </Badge>
                                            </TableCell>
                                            <TableCell>
                                                {user.is_active ? (
                                                    <div className="flex items-center gap-2 text-sm text-green-600">
                                                        <CheckCircle2 className="h-4 w-4" />
                                                        Active
                                                    </div>
                                                ) : (
                                                    <div className="flex items-center gap-2 text-sm text-red-600">
                                                        <XCircle className="h-4 w-4" />
                                                        Inactive
                                                    </div>
                                                )}
                                            </TableCell>
                                            <TableCell className="text-muted-foreground text-sm">
                                                {new Date(user.created_at).toLocaleDateString()}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <Button variant="ghost" size="icon">
                                                            <MoreVertical className="h-4 w-4" />
                                                            <span className="sr-only">Open menu</span>
                                                        </Button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent align="end">
                                                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                        <DropdownMenuItem onClick={() => navigator.clipboard.writeText(user.id)}>
                                                            Copy ID
                                                        </DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuLabel>Change Role</DropdownMenuLabel>
                                                        <DropdownMenuItem onClick={() => handleRoleUpdate(user.id, 'admin')}>
                                                            Make Admin
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem onClick={() => handleRoleUpdate(user.id, 'librarian')}>
                                                            Make Librarian
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem onClick={() => handleRoleUpdate(user.id, 'patron')}>
                                                            Make Patron
                                                        </DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem
                                                            onClick={() => handleStatusUpdate(user.id, !user.is_active)}
                                                            className={user.is_active ? 'text-red-600' : 'text-green-600'}
                                                        >
                                                            {user.is_active ? 'Deactivate User' : 'Activate User'}
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </div>

                    {/* Simple Pagination */}
                    <div className="flex items-center justify-end space-x-2 py-4">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setPage((p) => Math.max(1, p - 1))}
                            disabled={page === 1 || isLoading}
                        >
                            Previous
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                            disabled={page === totalPages || isLoading}
                        >
                            Next
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}

'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import { documentsApi } from '@/lib/api/documents';
import { collectionsApi } from '@/lib/api/collections';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
    Shield, Users, Settings, Activity, AlertTriangle,
    Database, Server, Clock, XCircle, Loader2, FileText, Library
} from 'lucide-react';
import type { Document, Collection } from '@/types/api';

export default function AdminDashboardPage() {
    const router = useRouter();
    const { user } = useAuthStore();
    const [isAuthorized, setIsAuthorized] = useState<boolean | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [stats, setStats] = useState({
        totalDocuments: 0,
        totalCollections: 0,
        totalUsers: 0,
        systemUptime: '99.9%'
    });
    const [recentActivity, setRecentActivity] = useState<Document[]>([]);

    // Strict role protection - ONLY admins can access
    useEffect(() => {
        if (!user) {
            router.push('/login');
            return;
        }
        setIsAuthorized(user.role === 'admin');
    }, [user, router]);

    // Fetch real data
    useEffect(() => {
        if (isAuthorized) {
            loadDashboardData();
        }
    }, [isAuthorized]);

    const loadDashboardData = async () => {
        try {
            setIsLoading(true);

            // Fetch documents count
            const docsResponse = await documentsApi.list({ page: 1, limit: 5 });
            setStats(prev => ({ ...prev, totalDocuments: docsResponse.pagination?.total_items || 0 }));
            setRecentActivity(docsResponse.data || []);

            // Fetch collections count
            const collectionsResponse = await collectionsApi.list({ page: 1, page_size: 5 });
            setStats(prev => ({ ...prev, totalCollections: collectionsResponse.pagination?.total_items || 0 }));

        } catch (error) {
            console.error('Failed to load dashboard data:', error);
        } finally {
            setIsLoading(false);
        }
    };

    // Show loading while checking authorization
    if (isAuthorized === null) {
        return (
            <div className="flex items-center justify-center min-h-[50vh]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    // Access Denied
    if (!isAuthorized) {
        return (
            <div className="flex items-center justify-center min-h-[50vh]">
                <Card className="w-full max-w-md border-red-200">
                    <CardHeader className="text-center">
                        <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-4">
                            <XCircle className="h-8 w-8 text-red-600" />
                        </div>
                        <CardTitle className="text-red-900">Access Denied</CardTitle>
                        <CardDescription>This page is restricted to administrators only.</CardDescription>
                    </CardHeader>
                    <CardContent className="text-center">
                        <p className="text-sm text-muted-foreground mb-4">
                            Your role: <Badge variant="outline">{user?.role}</Badge>
                        </p>
                        <Button onClick={() => router.push(`/dashboard/${user?.role}`)}>Go to Your Dashboard</Button>
                    </CardContent>
                </Card>
            </div>
        );
    }

    const systemStats = [
        { title: 'Total Documents', value: stats.totalDocuments.toString(), icon: FileText, trend: 'up' },
        { title: 'Collections', value: stats.totalCollections.toString(), icon: Library, trend: 'up' },
        { title: 'Total Users', value: stats.totalUsers.toString(), icon: Users, trend: 'neutral' },
        { title: 'System Uptime', value: stats.systemUptime, icon: Server, trend: 'up' },
    ];

    const pendingApprovals = [
        { type: 'Librarian', count: 2, color: 'bg-blue-500' },
        { type: 'Archivist', count: 1, color: 'bg-purple-500' },
        { type: 'Vendor', count: 3, color: 'bg-orange-500' },
    ];

    return (
        <div className="space-y-4 sm:space-y-6">
            {/* Admin Header - Responsive */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-gradient-to-r from-red-600 to-red-700 text-white p-4 sm:p-6 rounded-lg shadow-lg">
                <div className="flex items-center gap-3 sm:gap-4">
                    <div className="w-12 h-12 sm:w-16 sm:h-16 bg-white/20 rounded-full flex items-center justify-center backdrop-blur flex-shrink-0">
                        <Shield className="h-6 w-6 sm:h-8 sm:w-8" />
                    </div>
                    <div>
                        <h2 className="text-xl sm:text-3xl font-bold">Admin Control Panel</h2>
                        <p className="text-red-100 text-sm sm:text-base">Welcome back, {user?.first_name || 'Admin'}!</p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <div className="flex items-center gap-2 bg-white/20 px-3 py-1.5 sm:px-4 sm:py-2 rounded-lg backdrop-blur">
                        <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
                        <span className="text-xs sm:text-sm font-medium">System Healthy</span>
                    </div>
                </div>
            </div>

            {/* System Stats - Responsive */}
            <div className="grid gap-3 sm:gap-4 grid-cols-2 lg:grid-cols-4">
                {systemStats.map((stat) => {
                    const Icon = stat.icon;
                    return (
                        <Card key={stat.title} className="border-l-4 border-l-red-500">
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
                                <Icon className="h-4 w-4 text-red-600" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-2xl font-bold">{isLoading ? '...' : stat.value}</div>
                            </CardContent>
                        </Card>
                    );
                })}
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Pending Approvals */}
                <Card className="border-t-4 border-t-yellow-500">
                    <CardHeader>
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="flex items-center gap-2">
                                    <AlertTriangle className="h-5 w-5 text-yellow-600" />
                                    Pending Approvals
                                </CardTitle>
                                <CardDescription>User registration requests</CardDescription>
                            </div>
                            <Badge variant="outline" className="bg-yellow-50 text-yellow-700 border-yellow-300">
                                {pendingApprovals.reduce((sum, item) => sum + item.count, 0)} Total
                            </Badge>
                        </div>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-3">
                            {pendingApprovals.map((item) => (
                                <div key={item.type} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                                    <div className="flex items-center gap-3">
                                        <div className={`w-3 h-3 ${item.color} rounded-full`} />
                                        <span className="font-medium">{item.type}</span>
                                    </div>
                                    <Badge variant="secondary">{item.count} pending</Badge>
                                </div>
                            ))}
                            <Button className="w-full mt-2 bg-yellow-600 hover:bg-yellow-700" variant="default">
                                Review All Approvals
                            </Button>
                        </div>
                    </CardContent>
                </Card>

                {/* Quick Actions */}
                <Card className="border-t-4 border-t-red-500">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Settings className="h-5 w-5 text-red-600" />
                            System Management
                        </CardTitle>
                        <CardDescription>Administrative actions</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="grid grid-cols-2 gap-3">
                            <Button
                                variant="outline"
                                className="h-24 flex-col gap-2 border-2 hover:border-red-500 hover:bg-red-50"
                                onClick={() => router.push('/dashboard/users')}
                            >
                                <Users className="h-6 w-6 text-red-600" />
                                <span className="text-sm font-medium">Manage Users</span>
                            </Button>
                            <Button
                                variant="outline"
                                className="h-24 flex-col gap-2 border-2 hover:border-red-500 hover:bg-red-50"
                                onClick={() => router.push('/dashboard/settings')}
                            >
                                <Settings className="h-6 w-6 text-red-600" />
                                <span className="text-sm font-medium">System Settings</span>
                            </Button>
                            <Button
                                variant="outline"
                                className="h-24 flex-col gap-2 border-2 hover:border-red-500 hover:bg-red-50"
                                onClick={() => router.push('/dashboard/documents')}
                            >
                                <Database className="h-6 w-6 text-red-600" />
                                <span className="text-sm font-medium">Documents</span>
                            </Button>
                            <Button
                                variant="outline"
                                className="h-24 flex-col gap-2 border-2 hover:border-red-500 hover:bg-red-50"
                                onClick={() => router.push('/dashboard/collections')}
                            >
                                <Activity className="h-6 w-6 text-red-600" />
                                <span className="text-sm font-medium">Collections</span>
                            </Button>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Recent Activity */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Clock className="h-5 w-5 text-gray-600" />
                        Recent Activity
                    </CardTitle>
                    <CardDescription>Latest system events</CardDescription>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8">
                            <Loader2 className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
                        </div>
                    ) : recentActivity.length === 0 ? (
                        <div className="text-center py-8 text-muted-foreground">
                            <Activity className="h-12 w-12 mx-auto mb-4 opacity-50" />
                            <p>No recent activity</p>
                        </div>
                    ) : (
                        <div className="space-y-3">
                            {recentActivity.map((doc) => (
                                <div
                                    key={doc.id}
                                    className="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 cursor-pointer"
                                    onClick={() => router.push(`/dashboard/documents/${doc.id}`)}
                                >
                                    <FileText className="h-4 w-4 text-red-600" />
                                    <span className="text-sm flex-1">Document uploaded: {doc.title || doc.original_filename}</span>
                                    <Badge variant="outline" className="text-xs">{doc.file_type}</Badge>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

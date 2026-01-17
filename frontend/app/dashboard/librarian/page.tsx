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
    Library, Upload, FolderPlus, FileCheck,
    TrendingUp, BookOpen, Tags, BarChart3, XCircle, Loader2
} from 'lucide-react';
import type { Document, Collection } from '@/types/api';

export default function LibrarianDashboardPage() {
    const router = useRouter();
    const { user } = useAuthStore();
    const [isAuthorized, setIsAuthorized] = useState<boolean | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [stats, setStats] = useState({
        totalDocuments: 0,
        totalCollections: 0,
        recentUploads: 0,
        pendingReview: 0
    });
    const [recentDocuments, setRecentDocuments] = useState<Document[]>([]);
    const [recentCollections, setRecentCollections] = useState<Collection[]>([]);

    // Strict role protection - ONLY librarians can access
    useEffect(() => {
        if (!user) {
            router.push('/login');
            return;
        }
        setIsAuthorized(user.role === 'librarian');
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

            // Fetch documents
            const docsResponse = await documentsApi.list({ page: 1, limit: 5 });
            setRecentDocuments(docsResponse.data || []);
            setStats(prev => ({ ...prev, totalDocuments: docsResponse.pagination?.total_items || 0 }));

            // Fetch collections
            const collectionsResponse = await collectionsApi.list({ page: 1, page_size: 5 });
            setRecentCollections(collectionsResponse.data || []);
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
                <Card className="w-full max-w-md border-blue-200">
                    <CardHeader className="text-center">
                        <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mb-4">
                            <XCircle className="h-8 w-8 text-blue-600" />
                        </div>
                        <CardTitle className="text-blue-900">Access Denied</CardTitle>
                        <CardDescription>This page is restricted to librarians only.</CardDescription>
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

    const contentStats = [
        { title: 'Total Documents', value: stats.totalDocuments.toString(), icon: BookOpen, color: 'text-blue-600' },
        { title: 'Collections', value: stats.totalCollections.toString(), icon: Library, color: 'text-indigo-600' },
        { title: 'This Month', value: '+24', icon: TrendingUp, color: 'text-green-600' },
        { title: 'Pending Review', value: '5', icon: FileCheck, color: 'text-orange-600' },
    ];

    return (
        <div className="space-y-6">
            {/* Librarian Header */}
            <div className="bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 text-white p-8 rounded-xl shadow-lg">
                <div className="flex items-center gap-4 mb-4">
                    <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center backdrop-blur">
                        <Library className="h-8 w-8" />
                    </div>
                    <div>
                        <h2 className="text-3xl font-bold">Content Management</h2>
                        <p className="text-blue-100">Welcome back, {user?.first_name || 'Librarian'}!</p>
                    </div>
                </div>

                <div className="grid grid-cols-4 gap-4 mt-6">
                    {contentStats.map((stat) => {
                        const Icon = stat.icon;
                        return (
                            <div key={stat.title} className="bg-white/10 backdrop-blur rounded-lg p-4">
                                <div className="flex items-center gap-2 mb-2">
                                    <Icon className="h-4 w-4" />
                                    <span className="text-sm opacity-90">{stat.title}</span>
                                </div>
                                <div className="text-2xl font-bold">{isLoading ? '...' : stat.value}</div>
                            </div>
                        );
                    })}
                </div>
            </div>

            {/* Quick Actions Grid */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card
                    className="cursor-pointer hover:shadow-lg transition-all border-2 hover:border-blue-300"
                    onClick={() => router.push('/dashboard/documents/upload')}
                >
                    <CardContent className="p-6">
                        <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl flex items-center justify-center mb-4">
                            <Upload className="h-6 w-6 text-white" />
                        </div>
                        <h3 className="font-semibold text-lg mb-1">Upload Document</h3>
                        <p className="text-sm text-muted-foreground">Add new content to the library</p>
                    </CardContent>
                </Card>

                <Card
                    className="cursor-pointer hover:shadow-lg transition-all border-2 hover:border-indigo-300"
                    onClick={() => router.push('/dashboard/collections/create')}
                >
                    <CardContent className="p-6">
                        <div className="w-12 h-12 bg-gradient-to-br from-indigo-500 to-indigo-600 rounded-xl flex items-center justify-center mb-4">
                            <FolderPlus className="h-6 w-6 text-white" />
                        </div>
                        <h3 className="font-semibold text-lg mb-1">Create Collection</h3>
                        <p className="text-sm text-muted-foreground">Organize documents</p>
                    </CardContent>
                </Card>

                <Card
                    className="cursor-pointer hover:shadow-lg transition-all border-2 hover:border-purple-300"
                    onClick={() => router.push('/dashboard/documents')}
                >
                    <CardContent className="p-6">
                        <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl flex items-center justify-center mb-4">
                            <BookOpen className="h-6 w-6 text-white" />
                        </div>
                        <h3 className="font-semibold text-lg mb-1">Browse Documents</h3>
                        <p className="text-sm text-muted-foreground">View all documents</p>
                    </CardContent>
                </Card>

                <Card
                    className="cursor-pointer hover:shadow-lg transition-all border-2 hover:border-green-300"
                    onClick={() => router.push('/dashboard/collections')}
                >
                    <CardContent className="p-6">
                        <div className="w-12 h-12 bg-gradient-to-br from-green-500 to-green-600 rounded-xl flex items-center justify-center mb-4">
                            <BarChart3 className="h-6 w-6 text-white" />
                        </div>
                        <h3 className="font-semibold text-lg mb-1">View Collections</h3>
                        <p className="text-sm text-muted-foreground">Manage collections</p>
                    </CardContent>
                </Card>
            </div>

            {/* Recent Activity */}
            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <FileCheck className="h-5 w-5 text-blue-600" />
                            Recent Documents
                        </CardTitle>
                        <CardDescription>Latest document uploads</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {isLoading ? (
                            <div className="text-center py-8">
                                <Loader2 className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
                            </div>
                        ) : recentDocuments.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground">
                                <Upload className="h-12 w-12 mx-auto mb-4 opacity-50" />
                                <p>No documents yet</p>
                                <Button
                                    className="mt-4 bg-blue-600 hover:bg-blue-700"
                                    onClick={() => router.push('/dashboard/documents/upload')}
                                >
                                    <Upload className="mr-2 h-4 w-4" />
                                    Upload Document
                                </Button>
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {recentDocuments.slice(0, 5).map((doc) => (
                                    <div
                                        key={doc.id}
                                        className="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 cursor-pointer"
                                        onClick={() => router.push(`/dashboard/documents/${doc.id}`)}
                                    >
                                        <BookOpen className="h-4 w-4 text-blue-600" />
                                        <span className="text-sm truncate flex-1">{doc.title || doc.original_filename}</span>
                                        <Badge variant="outline" className="text-xs">{doc.file_type}</Badge>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Library className="h-5 w-5 text-indigo-600" />
                            Recent Collections
                        </CardTitle>
                        <CardDescription>Latest created collections</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {isLoading ? (
                            <div className="text-center py-8">
                                <Loader2 className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
                            </div>
                        ) : recentCollections.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground">
                                <Library className="h-12 w-12 mx-auto mb-4 opacity-50" />
                                <p>No collections yet</p>
                                <Button
                                    className="mt-4 bg-indigo-600 hover:bg-indigo-700"
                                    onClick={() => router.push('/dashboard/collections/create')}
                                >
                                    <FolderPlus className="mr-2 h-4 w-4" />
                                    Create Collection
                                </Button>
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {recentCollections.slice(0, 5).map((col) => (
                                    <div
                                        key={col.id}
                                        className="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 cursor-pointer"
                                        onClick={() => router.push(`/dashboard/collections/${col.id}`)}
                                    >
                                        <Library className="h-4 w-4 text-indigo-600" />
                                        <span className="text-sm truncate flex-1">{col.name}</span>
                                        <Badge variant="outline" className="text-xs">{col.document_count || 0} docs</Badge>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}

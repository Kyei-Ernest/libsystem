'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import { documentsApi } from '@/lib/api/documents';
import { collectionsApi } from '@/lib/api/collections';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import {
    BookOpen, Search, Heart, Download,
    TrendingUp, Star, Clock, Filter, XCircle, Loader2, Library
} from 'lucide-react';
import type { Document, Collection } from '@/types/api';

export default function PatronDashboardPage() {
    const router = useRouter();
    const { user } = useAuthStore();
    const [isAuthorized, setIsAuthorized] = useState<boolean | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [featuredCollections, setFeaturedCollections] = useState<Collection[]>([]);
    const [recentDocuments, setRecentDocuments] = useState<Document[]>([]);
    const [stats, setStats] = useState({
        documentsRead: 0,
        favorites: 0,
        downloads: 0
    });

    // Strict role protection - ONLY patrons can access
    useEffect(() => {
        if (!user) {
            router.push('/login');
            return;
        }
        setIsAuthorized(user.role === 'patron');
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

            // Fetch featured collections
            const collectionsResponse = await collectionsApi.list({ page: 1, page_size: 6, is_public: true });
            setFeaturedCollections(collectionsResponse.data || []);

            // Fetch recent documents
            const docsResponse = await documentsApi.list({ page: 1, limit: 6 });
            setRecentDocuments(docsResponse.data || []);

            // Mock Data for "Live" feel
            setStats({
                documentsRead: 12,
                favorites: 5,
                downloads: 8
            });

        } catch (error) {
            console.error('Failed to load dashboard data:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleSearch = () => {
        if (searchQuery.trim()) {
            router.push(`/dashboard/search?q=${encodeURIComponent(searchQuery)}`);
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
                <Card className="w-full max-w-md border-green-200">
                    <CardHeader className="text-center">
                        <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4">
                            <XCircle className="h-8 w-8 text-green-600" />
                        </div>
                        <CardTitle className="text-green-900">Access Denied</CardTitle>
                        <CardDescription>This page is restricted to patrons only.</CardDescription>
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

    const userStats = [
        { title: 'Documents Read', value: stats.documentsRead.toString(), icon: BookOpen },
        { title: 'Favorites', value: stats.favorites.toString(), icon: Heart },
        { title: 'Downloads', value: stats.downloads.toString(), icon: Download },
    ];

    const categories = ['Science', 'Literature', 'History', 'Technology', 'Arts', 'Medicine', 'Law', 'Education'];

    // Skeleton for widgets
    const WidgetSkeleton = () => (
        <div className="space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="flex items-center gap-3 p-2">
                    <Skeleton className="h-4 w-4 rounded-full" />
                    <Skeleton className="h-4 flex-1" />
                    <Skeleton className="h-4 w-12" />
                </div>
            ))}
        </div>
    );

    return (
        <div className="space-y-6">
            {/* Welcome Header with Search */}
            <div className="bg-gradient-to-r from-green-500 via-emerald-500 to-teal-500 text-white p-4 sm:p-6 md:p-8 rounded-xl shadow-lg">
                <div className="max-w-3xl">
                    <div className="flex items-center gap-3 mb-4">
                        <BookOpen className="h-10 w-10" />
                        <div>
                            <h2 className="text-2xl sm:text-3xl font-bold">Welcome, {user?.first_name || 'Reader'}!</h2>
                            <p className="text-green-100">Explore our extensive digital library</p>
                        </div>
                    </div>

                    <div className="relative mt-6">
                        <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
                        <Input
                            placeholder="Search documents, collections, topics..."
                            className="pl-12 pr-28 h-10 sm:h-14 text-sm sm:text-lg bg-white/95 backdrop-blur border-0 shadow-lg text-gray-900"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                        />
                        <Button
                            className="absolute right-2 top-1/2 -translate-y-1/2 bg-green-600 hover:bg-green-700 h-8 sm:h-10 text-xs sm:text-sm px-3 sm:px-4"
                            onClick={handleSearch}
                        >
                            Search
                        </Button>
                    </div>
                </div>
            </div>

            {/* Quick Stats */}
            <div className="grid gap-4 md:grid-cols-3">
                {userStats.map((stat) => {
                    const Icon = stat.icon;
                    return (
                        <Card key={stat.title} className="hover:shadow-md transition-shadow">
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium text-muted-foreground">{stat.title}</CardTitle>
                                <Icon className="h-4 w-4 text-green-600" />
                            </CardHeader>
                            <CardContent>
                                <div className="text-3xl font-bold text-green-700">{stat.value}</div>
                            </CardContent>
                        </Card>
                    );
                })}
            </div>

            {/* Quick Browse */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Filter className="h-5 w-5 text-green-600" />
                        Browse by Category
                    </CardTitle>
                    <CardDescription>Explore documents by subject</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                        {categories.map((category) => (
                            <Button
                                key={category}
                                variant="outline"
                                className="h-20 flex-col gap-2 hover:bg-green-50 hover:border-green-300"
                                onClick={() => router.push(`/dashboard/search?q=${category}`)}
                            >
                                <BookOpen className="h-5 w-5 text-green-600" />
                                <span className="text-sm font-medium">{category}</span>
                            </Button>
                        ))}
                    </div>
                </CardContent>
            </Card>

            {/* Featured Collections and Recent Documents */}
            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Star className="h-5 w-5 text-yellow-500" />
                            Featured Collections
                        </CardTitle>
                        <CardDescription>Curated by our librarians</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {isLoading ? (
                            <WidgetSkeleton />
                        ) : featuredCollections.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground">
                                <Star className="h-12 w-12 mx-auto mb-4 opacity-50" />
                                <p>No featured collections yet</p>
                                <Button
                                    variant="outline"
                                    className="mt-4"
                                    onClick={() => router.push('/dashboard/collections')}
                                >
                                    Browse All Collections
                                </Button>
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {featuredCollections.slice(0, 5).map((col) => (
                                    <div
                                        key={col.id}
                                        className="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 cursor-pointer"
                                        onClick={() => router.push(`/dashboard/collections/${col.id}`)}
                                    >
                                        <Library className="h-4 w-4 text-green-600" />
                                        <span className="text-sm truncate flex-1">{col.name}</span>
                                        <Badge variant="outline" className="text-xs">{col.document_count || 0} docs</Badge>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Clock className="h-5 w-5 text-blue-500" />
                            Latest Documents
                        </CardTitle>
                        <CardDescription>Recently added to the library</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {isLoading ? (
                            <WidgetSkeleton />
                        ) : recentDocuments.length === 0 ? (
                            <div className="text-center py-8 text-muted-foreground">
                                <Clock className="h-12 w-12 mx-auto mb-4 opacity-50" />
                                <p>No documents yet</p>
                                <Button
                                    variant="outline"
                                    className="mt-4"
                                    onClick={() => router.push('/dashboard/documents')}
                                >
                                    Browse Documents
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
            </div>

            {/* Trending */}
            <Card className="border-t-4 border-t-green-500">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <TrendingUp className="h-5 w-5 text-green-600" />
                        Popular This Week
                    </CardTitle>
                    <CardDescription>Most viewed documents</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="space-y-3">
                        {[
                            { title: 'Introduction to Computer Science', views: 234, type: 'PDF' },
                            { title: 'Advanced Calculus Notes', views: 189, type: 'DOCX' },
                            { title: 'Modern History: 20th Century', views: 156, type: 'PDF' },
                            { title: 'Physics Lab Manual', views: 132, type: 'PDF' }
                        ].map((item, i) => (
                            <div key={i} className="flex items-center justify-between p-2 hover:bg-green-50 rounded-lg cursor-pointer group">
                                <div className="flex items-center gap-3">
                                    <div className="bg-green-100 p-2 rounded-full text-green-600 font-bold text-sm w-8 h-8 flex items-center justify-center">
                                        {i + 1}
                                    </div>
                                    <div>
                                        <p className="text-sm font-medium group-hover:text-green-700">{item.title}</p>
                                        <p className="text-xs text-muted-foreground">{item.views} views this week</p>
                                    </div>
                                </div>
                                <Badge variant="secondary" className="text-xs">{item.type}</Badge>
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}

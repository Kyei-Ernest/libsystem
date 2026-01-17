'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import { documentsApi } from '@/lib/api/documents';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
    Archive, Database, FileCheck, Tags,
    Clock, Shield, BookmarkCheck, HardDrive, XCircle, Loader2, BookOpen
} from 'lucide-react';
import type { Document } from '@/types/api';

export default function ArchivistDashboardPage() {
    const router = useRouter();
    const { user } = useAuthStore();
    const [isAuthorized, setIsAuthorized] = useState<boolean | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [stats, setStats] = useState({
        catalogedItems: 0,
        metadataRecords: 0,
        inPreservation: 0,
        storageUsed: '0 GB'
    });
    const [recentItems, setRecentItems] = useState<Document[]>([]);

    // Strict role protection - ONLY archivists can access
    useEffect(() => {
        if (!user) {
            router.push('/login');
            return;
        }
        setIsAuthorized(user.role === 'archivist');
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

            // Fetch documents (representing cataloged items)
            const docsResponse = await documentsApi.list({ page: 1, limit: 5 });
            setStats(prev => ({ ...prev, catalogedItems: docsResponse.pagination?.total_items || 0 }));
            setRecentItems(docsResponse.data || []);

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
                <Card className="w-full max-w-md border-purple-200">
                    <CardHeader className="text-center">
                        <div className="mx-auto w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mb-4">
                            <XCircle className="h-8 w-8 text-purple-600" />
                        </div>
                        <CardTitle className="text-purple-900">Access Denied</CardTitle>
                        <CardDescription>This page is restricted to archivists only.</CardDescription>
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

    const preservationStats = [
        { title: 'Cataloged Items', value: stats.catalogedItems.toString(), icon: FileCheck, color: 'bg-purple-100', textColor: 'text-purple-700' },
        { title: 'Metadata Records', value: stats.metadataRecords.toString(), icon: Tags, color: 'bg-pink-100', textColor: 'text-pink-700' },
        { title: 'In Preservation', value: stats.inPreservation.toString(), icon: Shield, color: 'bg-indigo-100', textColor: 'text-indigo-700' },
        { title: 'Storage Used', value: stats.storageUsed, icon: HardDrive, color: 'bg-violet-100', textColor: 'text-violet-700' },
    ];

    const workflows = [
        { name: 'Cataloging Queue', count: 0, priority: 'Medium' },
        { name: 'Format Migration', count: 0, priority: 'High' },
        { name: 'Quality Review', count: 0, priority: 'Low' },
    ];

    return (
        <div className="space-y-6">
            {/* Archivist Header */}
            <div className="relative overflow-hidden bg-gradient-to-br from-purple-600 via-fuchsia-600 to-pink-600 text-white p-4 sm:p-8 rounded-xl shadow-lg">
                <div className="absolute top-0 right-0 opacity-10">
                    <Archive className="h-32 w-32 sm:h-48 sm:w-48" />
                </div>
                <div className="relative z-10">
                    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 mb-6">
                        <div className="w-12 h-12 sm:w-16 sm:h-16 bg-white/20 rounded-2xl flex items-center justify-center backdrop-blur shrink-0">
                            <Archive className="h-6 w-6 sm:h-8 sm:w-8" />
                        </div>
                        <div>
                            <h2 className="text-2xl sm:text-3xl font-bold">Digital Preservation Center</h2>
                            <p className="text-purple-100 text-sm sm:text-base">Welcome back, {user?.first_name || 'Archivist'}!</p>
                        </div>
                    </div>

                    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
                        {preservationStats.map((stat) => {
                            const Icon = stat.icon;
                            return (
                                <div key={stat.title} className="bg-white/10 backdrop-blur rounded-lg p-3">
                                    <Icon className="h-5 w-5 mb-2 opacity-90" />
                                    <div className="text-xl sm:text-2xl font-bold">{isLoading ? '...' : stat.value}</div>
                                    <div className="text-xs opacity-75 mt-1">{stat.title}</div>
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>

            {/* Active Workflows */}
            <Card className="border-t-4 border-t-purple-500">
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle className="flex items-center gap-2">
                                <Clock className="h-5 w-5 text-purple-600" />
                                Active Workflows
                            </CardTitle>
                            <CardDescription>Preservation and cataloging queues</CardDescription>
                        </div>
                        <Badge className="bg-purple-100 text-purple-700 border-purple-300">
                            {workflows.reduce((sum, w) => sum + w.count, 0)} items
                        </Badge>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="space-y-3">
                        {workflows.map((workflow) => (
                            <div
                                key={workflow.name}
                                className="flex items-center justify-between p-4 bg-gradient-to-r from-purple-50 to-pink-50 rounded-lg border border-purple-200 hover:shadow-md transition-shadow cursor-pointer"
                            >
                                <div className="flex items-center gap-3">
                                    <div className="w-2 h-2 bg-purple-500 rounded-full animate-pulse" />
                                    <div>
                                        <div className="font-medium">{workflow.name}</div>
                                        <div className="text-xs text-muted-foreground">Priority: {workflow.priority}</div>
                                    </div>
                                </div>
                                <Badge variant="outline">{workflow.count} pending</Badge>
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>

            {/* Archival Tools */}
            <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                <Card
                    className="hover:shadow-lg transition-all cursor-pointer border-2 hover:border-purple-300"
                    onClick={() => router.push('/dashboard/documents/upload')}
                >
                    <CardContent className="p-6">
                        <div className="w-14 h-14 bg-gradient-to-br from-purple-500 to-purple-600 rounded-2xl flex items-center justify-center mb-4">
                            <FileCheck className="h-7 w-7 text-white" />
                        </div>
                        <h3 className="font-bold text-lg mb-2">Catalog Item</h3>
                        <p className="text-sm text-muted-foreground mb-4">Create detailed metadata records</p>
                        <Button className="w-full bg-purple-600 hover:bg-purple-700">Start Cataloging</Button>
                    </CardContent>
                </Card>

                <Card
                    className="hover:shadow-lg transition-all cursor-pointer border-2 hover:border-pink-300"
                    onClick={() => router.push('/dashboard/documents')}
                >
                    <CardContent className="p-6">
                        <div className="w-14 h-14 bg-gradient-to-br from-pink-500 to-pink-600 rounded-2xl flex items-center justify-center mb-4">
                            <Database className="h-7 w-7 text-white" />
                        </div>
                        <h3 className="font-bold text-lg mb-2">Preservation</h3>
                        <p className="text-sm text-muted-foreground mb-4">Manage format migration</p>
                        <Button className="w-full bg-pink-600 hover:bg-pink-700">Open Queue</Button>
                    </CardContent>
                </Card>

                <Card
                    className="hover:shadow-lg transition-all cursor-pointer border-2 hover:border-indigo-300"
                    onClick={() => router.push('/dashboard/collections')}
                >
                    <CardContent className="p-6">
                        <div className="w-14 h-14 bg-gradient-to-br from-indigo-500 to-indigo-600 rounded-2xl flex items-center justify-center mb-4">
                            <Tags className="h-7 w-7 text-white" />
                        </div>
                        <h3 className="font-bold text-lg mb-2">Collections</h3>
                        <p className="text-sm text-muted-foreground mb-4">Manage document collections</p>
                        <Button className="w-full bg-indigo-600 hover:bg-indigo-700">View Collections</Button>
                    </CardContent>
                </Card>
            </div>

            {/* Recent Cataloged Items */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <BookOpen className="h-5 w-5 text-purple-600" />
                        Recently Cataloged
                    </CardTitle>
                    <CardDescription>Latest archived items</CardDescription>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8">
                            <Loader2 className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
                        </div>
                    ) : recentItems.length === 0 ? (
                        <div className="text-center py-8 text-muted-foreground">
                            <Archive className="h-12 w-12 mx-auto mb-4 opacity-50" />
                            <p>No cataloged items yet</p>
                        </div>
                    ) : (
                        <div className="space-y-3">
                            {recentItems.map((doc) => (
                                <div
                                    key={doc.id}
                                    className="flex items-center gap-3 p-2 rounded-lg hover:bg-purple-50 cursor-pointer"
                                    onClick={() => router.push(`/dashboard/documents/${doc.id}`)}
                                >
                                    <BookOpen className="h-4 w-4 text-purple-600" />
                                    <span className="text-sm truncate flex-1">{doc.title || doc.original_filename}</span>
                                    <Badge variant="outline" className="text-xs">{doc.file_type}</Badge>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* Archival Standards */}
            <Card className="border-l-4 border-l-purple-500">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <BookmarkCheck className="h-5 w-5 text-purple-600" />
                        Archival Standards & Protocols
                    </CardTitle>
                    <CardDescription>Preservation best practices</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid md:grid-cols-3 gap-6 text-sm">
                        <div className="space-y-2">
                            <h4 className="font-semibold text-purple-700">Dublin Core</h4>
                            <p className="text-muted-foreground">15 core metadata elements for resource description</p>
                        </div>
                        <div className="space-y-2">
                            <h4 className="font-semibold text-pink-700">OAIS Model</h4>
                            <p className="text-muted-foreground">Reference model for long-term digital preservation</p>
                        </div>
                        <div className="space-y-2">
                            <h4 className="font-semibold text-indigo-700">PREMIS</h4>
                            <p className="text-muted-foreground">Preservation metadata strategies</p>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}

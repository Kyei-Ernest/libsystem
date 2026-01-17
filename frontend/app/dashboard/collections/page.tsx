'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { collectionsApi } from '@/lib/api/collections';
import { useAuthStore } from '@/stores/auth-store';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
    FolderOpen, Plus, Search, Lock, Globe, FileText,
    Loader2, MoreVertical, Edit, Trash2, Grid, List
} from 'lucide-react';
import type { Collection } from '@/types/api';
import { toast } from 'sonner';

// Collection color based on name
const getCollectionColor = (name: string) => {
    const colors = [
        'from-blue-500 to-indigo-600',
        'from-purple-500 to-pink-600',
        'from-green-500 to-emerald-600',
        'from-orange-500 to-red-600',
        'from-cyan-500 to-blue-600',
        'from-pink-500 to-rose-600',
    ];
    const index = name.charCodeAt(0) % colors.length;
    return colors[index];
};

export default function CollectionsPage() {
    const router = useRouter();
    const { user } = useAuthStore();
    const [collections, setCollections] = useState<Collection[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
    const [totalCollections, setTotalCollections] = useState(0);

    // Roles that can create collections
    const canCreateCollection = user?.role === 'admin' || user?.role === 'librarian' || user?.role === 'archivist';
    const canDelete = user?.role === 'admin';

    useEffect(() => {
        loadCollections();
    }, []);

    const loadCollections = async () => {
        try {
            setIsLoading(true);
            const response = await collectionsApi.list({
                page: 1,
                page_size: 50
            });

            setCollections(response.data || []);
            setTotalCollections(response.pagination?.total_items || 0);
        } catch (error) {
            console.error('Failed to load collections:', error);
            toast.error('Failed to load collections');
        } finally {
            setIsLoading(false);
        }
    };

    // Filter collections
    const filteredCollections = collections.filter(collection => {
        if (!searchQuery) return true;
        const query = searchQuery.toLowerCase();
        return collection.name.toLowerCase().includes(query) ||
            (collection.description?.toLowerCase().includes(query));
    });

    return (
        <div className="space-y-4 sm:space-y-6">
            {/* Header - Responsive */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div>
                    <h2 className="text-2xl sm:text-3xl font-bold tracking-tight">Collections</h2>
                    <p className="text-sm sm:text-base text-muted-foreground">
                        {totalCollections} collection{totalCollections !== 1 ? 's' : ''} in your library
                    </p>
                </div>
                {canCreateCollection && (
                    <Button
                        onClick={() => router.push('/dashboard/collections/create')}
                        className="w-full sm:w-auto"
                    >
                        <Plus className="mr-2 h-4 w-4" />
                        Create Collection
                    </Button>
                )}
            </div>

            {/* Search and Filters */}
            <Card>
                <CardContent className="pt-4 sm:pt-6">
                    <div className="flex flex-col sm:flex-row gap-3">
                        <div className="relative flex-1">
                            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                            <Input
                                placeholder="Search collections..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="pl-10"
                            />
                        </div>

                        {/* View Toggle */}
                        <div className="flex border rounded-md self-end sm:self-auto">
                            <Button
                                variant={viewMode === 'grid' ? 'default' : 'ghost'}
                                size="icon"
                                className="h-9 w-9 rounded-r-none"
                                onClick={() => setViewMode('grid')}
                            >
                                <Grid className="h-4 w-4" />
                            </Button>
                            <Button
                                variant={viewMode === 'list' ? 'default' : 'ghost'}
                                size="icon"
                                className="h-9 w-9 rounded-l-none"
                                onClick={() => setViewMode('list')}
                            >
                                <List className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Collections */}
            {isLoading ? (
                <div className="flex items-center justify-center py-12">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            ) : filteredCollections.length === 0 ? (
                <Card>
                    <CardContent className="py-12 text-center">
                        <FolderOpen className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
                        <h3 className="text-lg font-semibold mb-2">No collections found</h3>
                        <p className="text-muted-foreground mb-4 text-sm">
                            {searchQuery ? 'Try adjusting your search' : 'Create your first collection to organize documents'}
                        </p>
                        {!searchQuery && canCreateCollection && (
                            <Button onClick={() => router.push('/dashboard/collections/create')}>
                                <Plus className="mr-2 h-4 w-4" />
                                Create Collection
                            </Button>
                        )}
                    </CardContent>
                </Card>
            ) : viewMode === 'grid' ? (
                /* Grid View */
                <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {filteredCollections.map((collection) => (
                        <Card
                            key={collection.id}
                            className="group hover:shadow-lg transition-all cursor-pointer overflow-hidden"
                            onClick={() => router.push(`/dashboard/collections/${collection.id}`)}
                        >
                            {/* Visual Header */}
                            <div className={`h-24 sm:h-28 bg-gradient-to-br ${getCollectionColor(collection.name)} flex items-center justify-center relative`}>
                                <FolderOpen className="h-10 w-10 sm:h-12 sm:w-12 text-white/80" />

                                {/* Badge */}
                                <Badge
                                    variant="outline"
                                    className="absolute top-2 right-2 text-xs bg-white/90"
                                >
                                    {collection.is_public ? (
                                        <><Globe className="mr-1 h-3 w-3" /> Public</>
                                    ) : (
                                        <><Lock className="mr-1 h-3 w-3" /> Private</>
                                    )}
                                </Badge>
                            </div>

                            <CardContent className="p-3 sm:p-4">
                                <h3 className="font-semibold truncate text-sm sm:text-base">
                                    {collection.name}
                                </h3>
                                {collection.description && (
                                    <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                                        {collection.description}
                                    </p>
                                )}
                                <div className="flex items-center gap-2 mt-3 text-xs text-muted-foreground">
                                    <FileText className="h-3 w-3" />
                                    <span>{collection.document_count || 0} documents</span>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            ) : (
                /* List View */
                <div className="space-y-2">
                    {filteredCollections.map((collection) => (
                        <Card
                            key={collection.id}
                            className="hover:shadow-md transition-shadow cursor-pointer"
                            onClick={() => router.push(`/dashboard/collections/${collection.id}`)}
                        >
                            <CardContent className="p-3 sm:p-4">
                                <div className="flex items-center gap-3 sm:gap-4">
                                    {/* Icon */}
                                    <div className={`w-10 h-10 sm:w-12 sm:h-12 rounded-lg bg-gradient-to-br ${getCollectionColor(collection.name)} flex items-center justify-center flex-shrink-0`}>
                                        <FolderOpen className="h-5 w-5 sm:h-6 sm:w-6 text-white" />
                                    </div>

                                    {/* Info */}
                                    <div className="flex-1 min-w-0">
                                        <h3 className="font-semibold truncate text-sm sm:text-base">
                                            {collection.name}
                                        </h3>
                                        <div className="flex flex-wrap items-center gap-2 mt-1">
                                            <Badge
                                                variant="outline"
                                                className="text-xs"
                                            >
                                                {collection.is_public ? 'Public' : 'Private'}
                                            </Badge>
                                            <span className="text-xs text-muted-foreground">
                                                {collection.document_count || 0} docs
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}
        </div>
    );
}

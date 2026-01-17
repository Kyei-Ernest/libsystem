'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { documentsApi } from '@/lib/api/documents';
import { collectionsApi } from '@/lib/api/collections';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { DocumentThumbnail, getFileIcon } from '@/components/document-thumbnail';
import { toast } from 'sonner';
import {
    FileText, Download, Trash2, Upload, Search,
    Grid, List, Eye, ChevronLeft, ChevronRight
} from 'lucide-react';
import type { Document, Collection } from '@/types/api';
import { formatBytes } from '@/lib/utils';
import { useAuthStore } from '@/stores/auth-store';

// Get file type color
const getFileTypeColor = (fileType: string) => {
    const type = fileType?.toLowerCase() || '';
    if (type.includes('pdf')) return 'bg-red-100 text-red-700 border-red-200';
    if (type.includes('image') || type.includes('png') || type.includes('jpg')) return 'bg-green-100 text-green-700 border-green-200';
    if (type.includes('doc')) return 'bg-blue-100 text-blue-700 border-blue-200';
    if (type.includes('xls')) return 'bg-emerald-100 text-emerald-700 border-emerald-200';
    return 'bg-gray-100 text-gray-700 border-gray-200';
};

export default function DocumentsPage() {
    const router = useRouter();
    const { user } = useAuthStore();
    const [documents, setDocuments] = useState<Document[]>([]);
    const [collections, setCollections] = useState<Collection[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [totalItems, setTotalItems] = useState(0);
    const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

    // Filter states
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedCollection, setSelectedCollection] = useState<string>('all');
    const [selectedFileType, setSelectedFileType] = useState<string>('all');
    const pageSize = 12;

    // Check roles
    // Check roles
    const canUpload = user && user.role && user.role !== 'patron';
    const canDelete = user?.role === 'admin' || user?.role === 'librarian';

    // Debounced search and filter effect
    useEffect(() => {
        const timer = setTimeout(() => {
            loadDocuments();
        }, 500); // 500ms debounce

        return () => clearTimeout(timer);
    }, [currentPage, selectedCollection, selectedFileType, searchQuery]);

    const loadCollections = async () => {
        try {
            const response = await collectionsApi.list({ page: 1, page_size: 100 });
            setCollections(response.data || []);
        } catch (error) {
            console.error('Failed to load collections:', error);
        }
    };

    const loadDocuments = async () => {
        try {
            setIsLoading(true);
            const params: any = {
                page: currentPage,
                limit: pageSize,
                search: searchQuery || undefined,
                collection_id: selectedCollection !== 'all' ? selectedCollection : undefined,
                file_type: selectedFileType !== 'all' ? selectedFileType : undefined
            };

            const response = await documentsApi.list(params);
            setDocuments(response.data || []);
            setTotalPages(response.pagination?.total_pages || 1);
            setTotalItems(response.pagination?.total_items || 0);
        } catch (error) {
            console.error('Failed to load documents:', error);
            toast.error('Failed to load documents');
        } finally {
            setIsLoading(false);
        }
    };

    const handleDelete = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        if (!confirm('Are you sure you want to delete this document?')) return;

        try {
            await documentsApi.delete(id);
            toast.success('Document deleted successfully');
            loadDocuments();
        } catch (error) {
            console.error('Delete failed:', error);
            toast.error('Failed to delete document');
        }
    };

    const handleDownload = (id: string, filename: string, e: React.MouseEvent) => {
        e.stopPropagation();
        window.open(documentsApi.getDownloadUrl(id), '_blank');
    };

    // Initial load for collections only
    useEffect(() => {
        loadCollections();
    }, []);

    const fileTypes = [...new Set(documents.map(d => d.file_type?.split('/').pop() || 'unknown'))];

    // Helper for loading skeletons
    const DocumentsLoadingSkeleton = () => (
        <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="flex flex-col space-y-3">
                    <Skeleton className="h-[200px] w-full rounded-xl" />
                    <div className="space-y-2 px-1">
                        <Skeleton className="h-4 w-[250px]" />
                        <Skeleton className="h-4 w-[200px]" />
                    </div>
                </div>
            ))}
        </div>
    );

    return (
        <div className="space-y-4 sm:space-y-6">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div>
                    <h2 className="text-2xl sm:text-3xl font-bold tracking-tight">Documents</h2>
                    <p className="text-sm sm:text-base text-muted-foreground">
                        {totalItems} document{totalItems !== 1 ? 's' : ''} in your library
                    </p>
                </div>
                {canUpload && (
                    <Button onClick={() => router.push('/dashboard/documents/upload')} className="w-full sm:w-auto">
                        <Upload className="mr-2 h-4 w-4" />
                        Upload Document
                    </Button>
                )}
            </div>

            {/* Search and Filters */}
            <Card>
                <CardContent className="pt-4 sm:pt-6">
                    <div className="flex flex-col gap-3">
                        <div className="relative">
                            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                            <Input
                                placeholder="Search documents..."
                                value={searchQuery}
                                onChange={(e) => { setSearchQuery(e.target.value); setCurrentPage(1); }}
                                className="pl-10"
                            />
                        </div>
                        <div className="flex flex-col sm:flex-row gap-3">
                            <select
                                value={selectedCollection}
                                onChange={(e) => { setSelectedCollection(e.target.value); setCurrentPage(1); }}
                                className="w-full sm:w-[180px] h-10 px-3 py-2 rounded-md border text-sm bg-background ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                            >
                                <option value="all">All Collections</option>
                                {collections.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                            </select>
                            <select
                                value={selectedFileType}
                                onChange={(e) => { setSelectedFileType(e.target.value); setCurrentPage(1); }}
                                className="w-full sm:w-[140px] h-10 px-3 py-2 rounded-md border text-sm bg-background ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                            >
                                <option value="all">All Types</option>
                                <option value="pdf">PDF</option>
                                <option value="image">Image</option>
                                <option value="word">Word</option>
                                <option value="excel">Excel</option>
                                <option value="text">Text</option>
                            </select>
                            <div className="flex border rounded-md sm:ml-auto self-end sm:self-auto">
                                <Button variant={viewMode === 'grid' ? 'default' : 'ghost'} size="icon" className="h-9 w-9 rounded-r-none" onClick={() => setViewMode('grid')}>
                                    <Grid className="h-4 w-4" />
                                </Button>
                                <Button variant={viewMode === 'list' ? 'default' : 'ghost'} size="icon" className="h-9 w-9 rounded-l-none" onClick={() => setViewMode('list')}>
                                    <List className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Documents */}
            {isLoading ? (
                <DocumentsLoadingSkeleton />
            ) : documents.length === 0 ? (
                <Card>
                    <CardContent className="py-12 text-center">
                        <FileText className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
                        <h3 className="text-lg font-semibold mb-2">No documents found</h3>
                        <p className="text-muted-foreground mb-4 text-sm">
                            {searchQuery ? 'Try adjusting your search' : 'Upload your first document to get started'}
                        </p>
                        {!searchQuery && canUpload && (
                            <Button onClick={() => router.push('/dashboard/documents/upload')}>
                                <Upload className="mr-2 h-4 w-4" />
                                Upload Document
                            </Button>
                        )}
                    </CardContent>
                </Card>
            ) : viewMode === 'grid' ? (
                /* Grid View with Real Thumbnails */
                <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                    {documents.map((doc) => (
                        <Card
                            key={doc.id}
                            className="group hover:shadow-lg transition-all cursor-pointer overflow-hidden"
                            onClick={() => router.push(`/dashboard/documents/${doc.id}`)}
                        >
                            {/* Thumbnail with Real API Data */}
                            <div className="relative">
                                <DocumentThumbnail documentId={doc.id} fileType={doc.file_type} title={doc.title} />

                                {/* Hover Overlay */}
                                <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                                    <Button size="icon" variant="secondary" className="h-10 w-10" onClick={(e) => { e.stopPropagation(); router.push(`/dashboard/documents/${doc.id}`); }}>
                                        <Eye className="h-5 w-5" />
                                    </Button>
                                    <Button size="icon" variant="secondary" className="h-10 w-10" onClick={(e) => handleDownload(doc.id, doc.original_filename, e)}>
                                        <Download className="h-5 w-5" />
                                    </Button>
                                    {canDelete && (
                                        <Button size="icon" variant="destructive" className="h-10 w-10" onClick={(e) => handleDelete(doc.id, e)}>
                                            <Trash2 className="h-5 w-5" />
                                        </Button>
                                    )}
                                </div>

                                {/* File Type Badge */}
                                <Badge className={`absolute top-2 right-2 text-xs ${getFileTypeColor(doc.file_type)}`}>
                                    {doc.file_type?.split('/').pop()?.toUpperCase() || 'FILE'}
                                </Badge>
                            </div>

                            {/* Content */}
                            <CardContent className="p-3 sm:p-4">
                                <h3 className="font-semibold truncate text-sm sm:text-base">{doc.title || doc.original_filename}</h3>
                                <p className="text-xs sm:text-sm text-muted-foreground mt-1 truncate">{formatBytes(doc.file_size)}</p>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            ) : (
                /* List View */
                <div className="space-y-2">
                    {documents.map((doc) => (
                        <Card key={doc.id} className="hover:shadow-md transition-shadow cursor-pointer" onClick={() => router.push(`/dashboard/documents/${doc.id}`)}>
                            <CardContent className="p-3 sm:p-4">
                                <div className="flex items-center gap-3 sm:gap-4">
                                    <div className="w-10 h-10 sm:w-12 sm:h-12 flex-shrink-0">{getFileIcon(doc.file_type, 'fixed')}</div>
                                    <div className="flex-1 min-w-0">
                                        <h3 className="font-semibold truncate text-sm sm:text-base">{doc.title || doc.original_filename}</h3>
                                        <div className="flex flex-wrap items-center gap-2 mt-1">
                                            <Badge variant="outline" className="text-xs">{doc.file_type?.split('/').pop()?.toUpperCase()}</Badge>
                                            <span className="text-xs text-muted-foreground">{formatBytes(doc.file_size)}</span>
                                        </div>
                                    </div>
                                    <div className="flex gap-1 sm:gap-2">
                                        <Button variant="ghost" size="icon" className="h-8 w-8 sm:h-9 sm:w-9" onClick={(e) => handleDownload(doc.id, doc.original_filename, e)}>
                                            <Download className="h-4 w-4" />
                                        </Button>
                                        {canDelete && (
                                            <Button variant="ghost" size="icon" className="h-8 w-8 sm:h-9 sm:w-9 text-red-600" onClick={(e) => handleDelete(doc.id, e)}>
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        )}
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            {/* Pagination */}
            {totalPages > 1 && (
                <div className="flex flex-col sm:flex-row items-center justify-between gap-4 pt-4">
                    <p className="text-sm text-muted-foreground order-2 sm:order-1">Page {currentPage} of {totalPages}</p>
                    <div className="flex gap-2 order-1 sm:order-2">
                        <Button variant="outline" size="sm" onClick={() => setCurrentPage(p => Math.max(1, p - 1))} disabled={currentPage === 1}>
                            <ChevronLeft className="h-4 w-4 mr-1" /> Previous
                        </Button>
                        <Button variant="outline" size="sm" onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))} disabled={currentPage === totalPages}>
                            Next <ChevronRight className="h-4 w-4 ml-1" />
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
}

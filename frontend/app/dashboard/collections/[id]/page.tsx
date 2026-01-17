'use client';

import { use, useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { collectionsApi } from '@/lib/api/collections';
import { documentsApi } from '@/lib/api/documents';
import { useAuthStore } from '@/stores/auth-store';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import { ArrowLeft, Globe, Lock, FileText, Download, Trash2, Upload } from 'lucide-react';
import type { Collection, Document } from '@/types/api';
import { formatBytes } from '@/lib/utils';

export default function CollectionDetailPage({ params }: { params: Promise<{ id: string }> }) {
    const router = useRouter();
    // Unwrap params Promise for Next.js 16
    const { id } = use(params);
    const [collection, setCollection] = useState<Collection | null>(null);
    const [documents, setDocuments] = useState<Document[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [currentPage, setCurrentPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);

    useEffect(() => {
        loadCollectionAndDocuments();
    }, [id, currentPage]);

    const loadCollectionAndDocuments = async () => {
        try {
            setIsLoading(true);

            // Load collection details
            const collectionData = await collectionsApi.getById(id);
            setCollection(collectionData);

            // Load documents in this collection
            const documentsResponse = await documentsApi.list({
                collection_id: id,
                page: currentPage,
                limit: 10,
            });

            setDocuments(documentsResponse.data || []);
            setTotalPages(documentsResponse.pagination?.total_pages || 1);
        } catch (error) {
            console.error('Failed to load collection:', error);
            toast.error('Failed to load collection');
        } finally {
            setIsLoading(false);
        }
    };

    const { user } = useAuthStore();
    const canManage = user?.role === 'admin' || user?.role === 'librarian' || user?.role === 'archivist';

    const handleDeleteDocument = async (id: string) => {
        if (!confirm('Are you sure you want to delete this document?')) {
            return;
        }

        try {
            await documentsApi.delete(id);
            toast.success('Document deleted successfully');
            loadCollectionAndDocuments();
        } catch (error) {
            toast.error('Failed to delete document');
        }
    };

    const handleDownload = (id: string) => {
        window.open(documentsApi.getDownloadUrl(id), '_blank');
    };

    const getFileIcon = (fileType: string) => {
        const type = fileType.toLowerCase();
        const iconProps = { className: "h-10 w-10", strokeWidth: 1.5 };

        if (type === 'pdf') return <FileText {...iconProps} className="h-10 w-10 text-red-500" />;
        if (type === 'txt') return <FileText {...iconProps} className="h-10 w-10 text-gray-500" />;
        if (type === 'docx' || type === 'doc') return <FileText {...iconProps} className="h-10 w-10 text-blue-500" />;
        return <FileText {...iconProps} className="h-10 w-10 text-gray-400" />;
    };

    if (isLoading && !collection) {
        return (
            <div className="space-y-6">
                <Card>
                    <CardContent className="pt-6">
                        <p className="text-center text-muted-foreground">Loading collection...</p>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (!collection) {
        return (
            <div className="space-y-6">
                <Card>
                    <CardContent className="pt-6 text-center">
                        <h3 className="text-lg font-semibold mb-2">Collection not found</h3>
                        <Button onClick={() => router.push('/dashboard/collections')}>
                            Back to Collections
                        </Button>
                    </CardContent>
                </Card>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center gap-4">
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => router.push('/dashboard/collections')}
                >
                    <ArrowLeft className="h-5 w-5" />
                </Button>
                <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                        <h2 className="text-3xl font-bold tracking-tight">{collection.name}</h2>
                        {collection.is_public ? (
                            <Badge variant="outline">
                                <Globe className="mr-1 h-3 w-3" />
                                Public
                            </Badge>
                        ) : (
                            <Badge variant="outline">
                                <Lock className="mr-1 h-3 w-3" />
                                Private
                            </Badge>
                        )}
                    </div>
                    {collection.description && (
                        <p className="text-muted-foreground">{collection.description}</p>
                    )}
                </div>
                {canManage && (
                    <Button onClick={() => router.push(`/dashboard/documents/upload?collection_id=${collection.id}`)}>
                        <Upload className="mr-2 h-4 w-4" />
                        Upload to Collection
                    </Button>
                )}
            </div>

            {/* Documents List */}
            {isLoading ? (
                <Card>
                    <CardContent className="pt-6">
                        <p className="text-center text-muted-foreground">Loading documents...</p>
                    </CardContent>
                </Card>
            ) : documents.length === 0 ? (
                <Card>
                    <CardContent className="pt-6 text-center">
                        <FileText className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
                        <h3 className="text-lg font-semibold mb-2">No documents yet</h3>
                        <p className="text-muted-foreground mb-4">
                            Upload your first document to this collection
                        </p>
                        {canManage && (
                            <Button onClick={() => router.push(`/dashboard/documents/upload?collection_id=${collection.id}`)}>
                                <Upload className="mr-2 h-4 w-4" />
                                Upload Document
                            </Button>
                        )}
                    </CardContent>
                </Card>
            ) : (
                <div className="grid gap-4">
                    {documents.map((doc) => (
                        <Card key={doc.id} className="hover:shadow-md transition-shadow">
                            <CardHeader>
                                <div className="flex items-start gap-4">
                                    <div className="flex-shrink-0 mt-1 w-10 h-10">
                                        {doc.thumbnail_path ? (
                                            <img
                                                src={documentsApi.getThumbnailUrl(doc.id)}
                                                alt="Thumbnail"
                                                className="w-full h-full object-cover rounded shadow-sm"
                                            />
                                        ) : (
                                            getFileIcon(doc.file_type)
                                        )}
                                    </div>

                                    <div className="flex-1 min-w-0">
                                        <CardTitle className="text-lg truncate">{doc.title || doc.original_filename}</CardTitle>
                                        <CardDescription className="mt-1">
                                            {doc.title !== doc.original_filename && <>{doc.original_filename} • </>}
                                            {formatBytes(doc.file_size)} • {doc.file_type}
                                        </CardDescription>
                                    </div>

                                    <div className="flex gap-2 flex-shrink-0">
                                        <Button
                                            variant="outline"
                                            size="icon"
                                            onClick={() => handleDownload(doc.id)}
                                            title="Download"
                                        >
                                            <Download className="h-4 w-4" />
                                        </Button>
                                        {canManage && (
                                            <Button
                                                variant="outline"
                                                size="icon"
                                                onClick={() => handleDeleteDocument(doc.id)}
                                                title="Delete"
                                                className="text-red-600 hover:text-red-700"
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        )}
                                    </div>
                                </div>
                            </CardHeader>
                            {doc.description && (
                                <CardContent className="pt-0 pb-4">
                                    <p className="text-sm text-muted-foreground line-clamp-2">{doc.description}</p>
                                </CardContent>
                            )}
                        </Card>
                    ))}
                </div>
            )}

            {/* Pagination */}
            {totalPages > 1 && (
                <div className="flex justify-center gap-2">
                    <Button
                        variant="outline"
                        onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                        disabled={currentPage === 1}
                    >
                        Previous
                    </Button>
                    <div className="flex items-center gap-2 px-4">
                        <span className="text-sm text-muted-foreground">
                            Page {currentPage} of {totalPages}
                        </span>
                    </div>
                    <Button
                        variant="outline"
                        onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                        disabled={currentPage === totalPages}
                    >
                        Next
                    </Button>
                </div>
            )}
        </div>
    );
}

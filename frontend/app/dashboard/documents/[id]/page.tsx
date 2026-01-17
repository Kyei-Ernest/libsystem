'use client';

import { useState, useEffect } from 'react';
import { use } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { documentsApi } from '@/lib/api/documents';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import {
    ArrowLeft, Download, FileText, Share2, Printer,
    Loader2, File, Image as ImageIcon, FileArchive,
    Calendar, HardDrive, Tag
} from 'lucide-react';
import type { Document } from '@/types/api';
import { formatBytes } from '@/lib/utils';

// Dynamically import viewers to avoid SSR issues
const PDFViewer = dynamic(
    () => import('@/components/features/pdf-viewer'),
    { ssr: false, loading: () => <ViewerLoading /> }
);

const ImageViewer = dynamic(
    () => import('@/components/features/image-viewer'),
    { ssr: false, loading: () => <ViewerLoading /> }
);

const TextViewer = dynamic(
    () => import('@/components/features/text-viewer'),
    { ssr: false, loading: () => <ViewerLoading /> }
);

function ViewerLoading() {
    return (
        <Card>
            <CardContent className="flex items-center justify-center py-12">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </CardContent>
        </Card>
    );
}

// Get file icon based on type
const getFileIcon = (fileType: string) => {
    const type = fileType?.toLowerCase() || '';
    if (type.includes('pdf')) return <FileText className="h-16 w-16 text-red-500" />;
    if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'image'].some(ext => type.includes(ext)))
        return <ImageIcon className="h-16 w-16 text-green-500" />;
    if (['zip', 'rar', 'tar', 'gz'].some(ext => type.includes(ext)))
        return <FileArchive className="h-16 w-16 text-yellow-500" />;
    if (type.includes('doc'))
        return <FileText className="h-16 w-16 text-blue-500" />;
    return <File className="h-16 w-16 text-gray-500" />;
};

// Get file type badge color
const getFileTypeColor = (fileType: string) => {
    const type = fileType?.toLowerCase() || '';
    if (type.includes('pdf')) return 'bg-red-100 text-red-700 border-red-300';
    if (['png', 'jpg', 'jpeg', 'gif', 'webp'].some(ext => type.includes(ext)))
        return 'bg-green-100 text-green-700 border-green-300';
    if (type.includes('doc')) return 'bg-blue-100 text-blue-700 border-blue-300';
    return 'bg-gray-100 text-gray-700 border-gray-300';
};

// Determine file viewer type
function getFileViewerType(fileType: string, fileName: string): 'pdf' | 'image' | 'text' | 'none' {
    const type = fileType?.toLowerCase() || '';
    const name = fileName?.toLowerCase() || '';

    // PDF and Office files (Backend converts Office -> PDF for preview)
    // Checks for specific mime type parts or extensions
    if (type.includes('pdf') ||
        type.includes('word') || type.includes('doc') ||
        type.includes('sheet') || type.includes('xls') ||
        type.includes('presentation') || type.includes('ppt') ||
        type.includes('opendocument')) {
        return 'pdf';
    }

    // Image files
    if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp'].some(ext =>
        type.includes(ext) || name.endsWith(`.${ext}`)
    ) || type.includes('image/')) return 'image';

    // Text/code files
    // Use strict extension checking to avoid false positives (like 'docx' matching 'c')
    const textExtensions = [
        '.txt', '.md', '.markdown', '.json', '.xml', '.html', '.css', '.js', '.ts', '.jsx', '.tsx',
        '.py', '.go', '.java', '.c', '.cpp', '.h', '.sh', '.bash', '.yaml', '.yml', '.csv', '.log',
        '.sql', '.env'
    ];

    if (textExtensions.some(ext => name.endsWith(ext)) ||
        type.startsWith('text/') ||
        type === 'application/json' ||
        type === 'application/xml') {
        return 'text';
    }

    return 'none';
}

export default function DocumentViewPage({ params }: { params: Promise<{ id: string }> }) {
    const router = useRouter();
    const { id } = use(params);
    const [document, setDocument] = useState<Document | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        loadDocument();
    }, [id]);

    const loadDocument = async () => {
        try {
            setIsLoading(true);
            const doc = await documentsApi.getById(id);
            setDocument(doc);
        } catch (error) {
            console.error('Failed to load document:', error);
            toast.error('Failed to load document');
        } finally {
            setIsLoading(false);
        }
    };

    const handleDownload = () => {
        if (document) {
            window.open(documentsApi.getDownloadUrl(document.id), '_blank');
        }
    };

    const handleShare = async () => {
        if (document) {
            try {
                await navigator.clipboard.writeText(window.location.href);
                toast.success('Link copied to clipboard!');
            } catch {
                toast.error('Failed to copy link');
            }
        }
    };

    const handlePrint = () => window.print();

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[50vh]">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (!document) {
        return (
            <div className="space-y-6">
                <Card>
                    <CardContent className="py-12 text-center">
                        <FileText className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
                        <h3 className="text-lg font-semibold mb-2">Document not found</h3>
                        <p className="text-muted-foreground mb-4">This document may have been deleted or moved.</p>
                        <Button onClick={() => router.push('/dashboard/documents')}>
                            <ArrowLeft className="mr-2 h-4 w-4" />
                            Back to Documents
                        </Button>
                    </CardContent>
                </Card>
            </div>
        );
    }

    const fileUrl = documentsApi.getViewUrl(document.id);
    const viewerType = getFileViewerType(document.file_type, document.original_filename);

    return (
        <div className="space-y-4 sm:space-y-6">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-start gap-4">
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => router.back()}
                    className="self-start"
                >
                    <ArrowLeft className="h-5 w-5" />
                </Button>

                <div className="flex-1 min-w-0">
                    <div className="flex flex-wrap items-center gap-2 sm:gap-3 mb-2">
                        <h2 className="text-xl sm:text-3xl font-bold tracking-tight truncate">
                            {document.title || document.original_filename}
                        </h2>
                        <Badge className={`${getFileTypeColor(document.file_type)}`}>
                            {document.file_type?.toUpperCase() || 'FILE'}
                        </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground mb-2">
                        {document.original_filename} • {formatBytes(document.file_size)}
                    </p>
                    {/* Real URL Display - Removed for cleaner UI
                    <div className="flex items-center gap-2 max-w-md">
                        <code className="text-xs bg-muted p-1 rounded border flex-1 truncate select-all">
                            {fileUrl}
                        </code>
                    </div>
                    */}
                </div>

                {/* Actions */}
                <div className="flex flex-wrap gap-2">
                    <Button variant="outline" size="sm" onClick={handleDownload}>
                        <Download className="h-4 w-4 sm:mr-2" />
                        <span className="hidden sm:inline">Download</span>
                    </Button>
                    <Button variant="outline" size="sm" onClick={handleShare}>
                        <Share2 className="h-4 w-4 sm:mr-2" />
                        <span className="hidden sm:inline">Share</span>
                    </Button>
                    {viewerType !== 'none' && (
                        <Button variant="outline" size="sm" onClick={handlePrint}>
                            <Printer className="h-4 w-4 sm:mr-2" />
                            <span className="hidden sm:inline">Print</span>
                        </Button>
                    )}
                </div>
            </div>

            {/* Description */}
            {document.description && (
                <Card>
                    <CardContent className="p-4">
                        <h3 className="font-medium mb-2">Description</h3>
                        <p className="text-sm text-muted-foreground">{document.description}</p>
                    </CardContent>
                </Card>
            )}

            {/* Document Viewer */}
            {viewerType === 'pdf' && (
                <PDFViewer fileUrl={fileUrl} fileName={document.original_filename} />
            )}

            {viewerType === 'image' && (
                <ImageViewer fileUrl={fileUrl} fileName={document.original_filename} />
            )}

            {viewerType === 'text' && (
                <TextViewer fileUrl={fileUrl} fileName={document.original_filename} />
            )}

            {viewerType === 'none' && (
                <Card>
                    <CardContent className="py-12 text-center">
                        <div className="w-24 h-24 mx-auto mb-6 bg-gray-100 rounded-2xl flex items-center justify-center">
                            {getFileIcon(document.file_type)}
                        </div>
                        <h3 className="text-lg font-semibold mb-2">Preview not available</h3>
                        <CardDescription className="mb-6">
                            This file type ({document.file_type?.toUpperCase() || 'unknown'}) cannot be previewed in the browser.
                            <br />
                            Download the file to view it on your device.
                        </CardDescription>
                        <div className="flex flex-wrap justify-center gap-3">
                            <Button onClick={handleDownload}>
                                <Download className="mr-2 h-4 w-4" />
                                Download File
                            </Button>
                            <Button variant="outline" onClick={() => router.push('/dashboard/documents')}>
                                <ArrowLeft className="mr-2 h-4 w-4" />
                                Back to Documents
                            </Button>
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* Document Metadata */}
            <Card>
                <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">Document Details</CardTitle>
                </CardHeader>
                <CardContent className="p-4 pt-0">
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
                        <div className="flex items-start gap-2">
                            <Tag className="h-4 w-4 text-muted-foreground mt-0.5" />
                            <div>
                                <p className="text-muted-foreground text-xs">File Type</p>
                                <p className="font-medium">{document.file_type?.toUpperCase() || 'Unknown'}</p>
                            </div>
                        </div>
                        <div className="flex items-start gap-2">
                            <HardDrive className="h-4 w-4 text-muted-foreground mt-0.5" />
                            <div>
                                <p className="text-muted-foreground text-xs">File Size</p>
                                <p className="font-medium">{formatBytes(document.file_size)}</p>
                            </div>
                        </div>
                        <div className="flex items-start gap-2">
                            <Calendar className="h-4 w-4 text-muted-foreground mt-0.5" />
                            <div>
                                <p className="text-muted-foreground text-xs">Created</p>
                                <p className="font-medium">{new Date(document.created_at).toLocaleDateString()}</p>
                            </div>
                        </div>
                        <div className="flex items-start gap-2">
                            <File className="h-4 w-4 text-muted-foreground mt-0.5" />
                            <div>
                                <p className="text-muted-foreground text-xs">Original Name</p>
                                <p className="font-medium truncate">{document.original_filename}</p>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}

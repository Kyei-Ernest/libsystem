'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { documentsApi } from '@/lib/api/documents';
import { collectionsApi } from '@/lib/api/collections';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Progress } from '@/components/ui/progress';
import { toast } from 'sonner';
import { Upload, FileText, ArrowLeft } from 'lucide-react';
import type { Collection } from '@/types/api';

export default function DocumentUploadPage() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [file, setFile] = useState<File | null>(null);
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [collectionId, setCollectionId] = useState('');
    const [collections, setCollections] = useState<Collection[]>([]);
    const [isUploading, setIsUploading] = useState(false);
    const [uploadProgress, setUploadProgress] = useState(0);

    // Load collections on mount
    useEffect(() => {
        loadCollections();
        // Pre-select collection from URL if provided
        const urlCollectionId = searchParams.get('collection_id');
        if (urlCollectionId) {
            setCollectionId(urlCollectionId);
        }
    }, [searchParams]);

    const loadCollections = async () => {
        try {
            const response = await collectionsApi.list({ page: 1, page_size: 100 });
            setCollections(response.data || []);
        } catch (error) {
            console.error('Failed to load collections:', error);
        }
    };

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const selectedFile = e.target.files?.[0];
        if (selectedFile) {
            setFile(selectedFile);
            // Auto-fill title with filename if empty
            if (!title) {
                setTitle(selectedFile.name.replace(/\.[^/.]+$/, ''));
            }
        }
    };

    const handleUpload = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!file) {
            toast.error('Please select a file to upload');
            return;
        }

        if (!collectionId) {
            toast.error('Please select a collection');
            return;
        }

        try {
            setIsUploading(true);
            setUploadProgress(0);

            await documentsApi.upload({
                file,
                collection_id: collectionId,
                metadata: {
                    title: title || file.name,
                    description,
                },
                onProgress: (progress) => {
                    setUploadProgress(progress);
                },
            });

            // Ensure progress shows 100%
            setUploadProgress(100);
            toast.success('Document uploaded successfully!');

            // Artificial delay to let user see the 100% progress
            await new Promise(resolve => setTimeout(resolve, 1000));

            // Navigate back to the collection if we came from one
            const urlCollectionId = searchParams.get('collection_id');
            if (urlCollectionId) {
                router.push(`/dashboard/collections/${urlCollectionId}`);
            } else {
                router.push('/dashboard/documents');
            }
        } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to upload document';
            toast.error(message);
            setIsUploading(false); // Only reset on error, keep true on success until nav
        }
        // Remove finally block to prevent hiding progress during success delay
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-4">
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => router.back()}
                >
                    <ArrowLeft className="h-5 w-5" />
                </Button>
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Upload Document</h2>
                    <p className="text-muted-foreground">Add a new document to your library</p>
                </div>
            </div>

            <Card className="max-w-2xl">
                <CardHeader>
                    <CardTitle>Document Details</CardTitle>
                    <CardDescription>
                        Upload a file and provide metadata to organize it in your library
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleUpload} className="space-y-6">
                        {/* File Upload */}
                        <div className="space-y-2">
                            <Label htmlFor="file">File *</Label>
                            <div className="flex items-center gap-4">
                                <Input
                                    id="file"
                                    type="file"
                                    onChange={handleFileChange}
                                    disabled={isUploading}
                                    className="cursor-pointer"
                                    accept=".pdf,.doc,.docx,.txt,.md"
                                />
                                {file && (
                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                        <FileText className="h-4 w-4" />
                                        {(file.size / 1024 / 1024).toFixed(2)} MB
                                    </div>
                                )}
                            </div>
                            <p className="text-xs text-muted-foreground">
                                Supported formats: PDF, DOC, DOCX, TXT, MD (Max 100MB)
                            </p>
                        </div>

                        {/* Collection Selection */}
                        <div className="space-y-2">
                            <Label htmlFor="collection">Collection *</Label>
                            <select
                                id="collection"
                                value={collectionId}
                                onChange={(e) => setCollectionId(e.target.value)}
                                disabled={isUploading}
                                required
                                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                            >
                                <option value="">Select a collection</option>
                                {collections.map((collection) => (
                                    <option key={collection.id} value={collection.id}>
                                        {collection.name}
                                    </option>
                                ))}
                            </select>
                        </div>

                        {/* Title */}
                        <div className="space-y-2">
                            <Label htmlFor="title">Title</Label>
                            <Input
                                id="title"
                                value={title}
                                onChange={(e) => setTitle(e.target.value)}
                                placeholder="Enter document title"
                                disabled={isUploading}
                            />
                        </div>

                        {/* Description */}
                        <div className="space-y-2">
                            <Label htmlFor="description">Description</Label>
                            <textarea
                                id="description"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="Enter document description (optional)"
                                disabled={isUploading}
                                className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                            />
                        </div>

                        {/* Actions */}
                        {isUploading && (
                            <div className="space-y-2">
                                <div className="flex justify-between text-xs text-muted-foreground">
                                    <span>Uploading...</span>
                                    <span>{uploadProgress}%</span>
                                </div>
                                <Progress value={uploadProgress} />
                            </div>
                        )}

                        <div className="flex gap-4">
                            <Button
                                type="submit"
                                disabled={!file || !collectionId || isUploading}
                                className="flex-1"
                            >
                                <Upload className="mr-2 h-4 w-4" />
                                {isUploading ? 'Uploading...' : 'Upload Document'}
                            </Button>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => router.back()}
                                disabled={isUploading}
                            >
                                Cancel
                            </Button>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}

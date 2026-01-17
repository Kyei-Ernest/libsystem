'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { searchApi, type SearchHit } from '@/lib/api/search';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
    Search as SearchIcon, FileText, Loader2, Download, Eye,
    File, Image as ImageIcon, FileArchive, FileSpreadsheet, FileCode,
    ChevronLeft, ChevronRight, Filter
} from 'lucide-react';
import { toast } from 'sonner';
import { formatBytes } from '@/lib/utils';

// File type icon mapping
const getFileIcon = (fileType: string) => {
    const type = fileType?.toLowerCase() || '';
    if (type.includes('pdf')) return <FileText className="h-full w-full text-red-500" />;
    if (type.includes('image') || type.includes('png') || type.includes('jpg'))
        return <ImageIcon className="h-full w-full text-green-500" />;
    if (type.includes('zip') || type.includes('rar'))
        return <FileArchive className="h-full w-full text-yellow-500" />;
    if (type.includes('xls') || type.includes('csv'))
        return <FileSpreadsheet className="h-full w-full text-emerald-500" />;
    if (type.includes('doc'))
        return <FileText className="h-full w-full text-blue-500" />;
    return <File className="h-full w-full text-gray-500" />;
};

// Get file type color
const getFileTypeColor = (fileType: string) => {
    const type = fileType?.toLowerCase() || '';
    if (type.includes('pdf')) return 'bg-red-100 text-red-700';
    if (type.includes('image') || type.includes('png') || type.includes('jpg')) return 'bg-green-100 text-green-700';
    if (type.includes('doc')) return 'bg-blue-100 text-blue-700';
    return 'bg-gray-100 text-gray-700';
};

export default function SearchPage() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [query, setQuery] = useState(searchParams.get('q') || '');
    const [results, setResults] = useState<SearchHit[]>([]);
    const [total, setTotal] = useState(0);
    const [isSearching, setIsSearching] = useState(false);
    const [hasSearched, setHasSearched] = useState(false);
    const [currentPage, setCurrentPage] = useState(1);
    const pageSize = 12;

    // Auto-search if query param exists
    useEffect(() => {
        const q = searchParams.get('q');
        if (q) {
            setQuery(q);
            performSearch(q, 1);
        }
    }, [searchParams]);

    const performSearch = async (searchQuery: string, page: number) => {
        if (!searchQuery.trim()) return;

        try {
            setIsSearching(true);
            setHasSearched(true);
            setCurrentPage(page);

            const response = await searchApi.search({
                query: searchQuery.trim(),
                from: (page - 1) * pageSize,
                size: pageSize,
            });

            setResults(response.hits || []);
            setTotal(response.total || 0);
        } catch (error) {
            console.error('Search failed:', error);
            toast.error('Search failed. Please try again.');
            setResults([]);
            setTotal(0);
        } finally {
            setIsSearching(false);
        }
    };

    const handleSearch = (e?: React.FormEvent, page = 1) => {
        if (e) e.preventDefault();

        if (!query.trim()) {
            toast.error('Please enter a search query');
            return;
        }

        performSearch(query, page);
    };

    const handleDownload = (id: string, filename: string) => {
        const downloadUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8088'}/api/v1/documents/${id}/download`;
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

    const totalPages = Math.ceil(total / pageSize);

    return (
        <div className="space-y-4 sm:space-y-6">
            {/* Header */}
            <div>
                <h2 className="text-2xl sm:text-3xl font-bold tracking-tight">Search</h2>
                <p className="text-sm sm:text-base text-muted-foreground">
                    Find documents across your entire library
                </p>
            </div>

            {/* Search Box */}
            <Card>
                <CardContent className="pt-4 sm:pt-6">
                    <form onSubmit={handleSearch} className="flex flex-col sm:flex-row gap-3">
                        <div className="relative flex-1">
                            <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                            <Input
                                placeholder="Search by title, content, or keywords..."
                                value={query}
                                onChange={(e) => setQuery(e.target.value)}
                                className="pl-10 h-11"
                                disabled={isSearching}
                            />
                        </div>
                        <Button type="submit" disabled={isSearching} className="h-11">
                            {isSearching ? (
                                <>
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                    Searching...
                                </>
                            ) : (
                                <>
                                    <SearchIcon className="mr-2 h-4 w-4 sm:hidden" />
                                    Search
                                </>
                            )}
                        </Button>
                    </form>
                </CardContent>
            </Card>

            {/* Results */}
            {!hasSearched ? (
                <Card>
                    <CardContent className="py-12 text-center">
                        <div className="w-16 h-16 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
                            <SearchIcon className="h-8 w-8 text-muted-foreground" />
                        </div>
                        <h3 className="text-lg font-semibold mb-2">Start searching</h3>
                        <p className="text-muted-foreground text-sm">
                            Enter a query above to search through your documents
                        </p>
                    </CardContent>
                </Card>
            ) : isSearching ? (
                <div className="flex items-center justify-center py-12">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            ) : results.length === 0 ? (
                <Card>
                    <CardContent className="py-12 text-center">
                        <div className="w-16 h-16 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
                            <SearchIcon className="h-8 w-8 text-muted-foreground" />
                        </div>
                        <h3 className="text-lg font-semibold mb-2">No results found</h3>
                        <p className="text-muted-foreground text-sm">
                            Try different keywords or check your spelling
                        </p>
                    </CardContent>
                </Card>
            ) : (
                <>
                    {/* Results Count */}
                    <div className="flex items-center justify-between">
                        <p className="text-sm text-muted-foreground">
                            Found <strong>{total}</strong> result{total !== 1 ? 's' : ''} for "<strong>{query}</strong>"
                        </p>
                    </div>

                    {/* Results Grid */}
                    <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                        {results
                            .map((hit) => (
                                <Card
                                    key={hit.id}
                                    className="group hover:shadow-lg transition-all cursor-pointer overflow-hidden"
                                    onClick={() => router.push(`/dashboard/documents/${hit.id}`)}
                                >
                                    {/* Thumbnail Area */}
                                    <div className="relative aspect-[16/10] bg-gradient-to-br from-gray-100 to-gray-50 flex items-center justify-center">
                                        <div className="w-12 h-12 sm:w-16 sm:h-16">
                                            {getFileIcon(hit.file_type || 'unknown')}
                                        </div>

                                        {/* Hover Actions */}
                                        <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                                            <Button
                                                size="icon"
                                                variant="secondary"
                                                className="h-9 w-9"
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    router.push(`/dashboard/documents/${hit.id}`);
                                                }}
                                            >
                                                <Eye className="h-4 w-4" />
                                            </Button>
                                            <Button
                                                size="icon"
                                                variant="secondary"
                                                className="h-9 w-9"
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleDownload(hit.id || '', hit.original_filename || 'download');
                                                }}
                                            >
                                                <Download className="h-4 w-4" />
                                            </Button>
                                        </div>

                                        {/* File Type Badge */}
                                        <Badge
                                            className={`absolute top-2 right-2 text-xs ${getFileTypeColor(hit.file_type || '')}`}
                                        >
                                            {hit.file_type?.split('/').pop()?.toUpperCase() || 'FILE'}
                                        </Badge>

                                        {/* Relevance Score - Only show if available */}
                                        {hit.score && (
                                            <Badge
                                                variant="outline"
                                                className="absolute top-2 left-2 text-xs bg-white/90"
                                            >
                                                {(hit.score * 100).toFixed(0)}% match
                                            </Badge>
                                        )}
                                    </div>

                                    {/* Content */}
                                    <CardContent className="p-3 sm:p-4">
                                        <h3 className="font-semibold truncate text-sm sm:text-base">
                                            {hit.title || hit.original_filename || 'Untitled'}
                                        </h3>
                                        <p className="text-xs sm:text-sm text-muted-foreground mt-1 truncate">
                                            {hit.original_filename}
                                        </p>
                                        {hit.file_size && (
                                            <p className="text-xs text-muted-foreground mt-1">
                                                {formatBytes(hit.file_size)}
                                            </p>
                                        )}
                                    </CardContent>
                                </Card>
                            ))}
                    </div>

                    {/* Pagination */}
                    {totalPages > 1 && (
                        <div className="flex flex-col sm:flex-row items-center justify-between gap-4 pt-4">
                            <p className="text-sm text-muted-foreground order-2 sm:order-1">
                                Page {currentPage} of {totalPages}
                            </p>
                            <div className="flex gap-2 order-1 sm:order-2">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => handleSearch(undefined, currentPage - 1)}
                                    disabled={currentPage === 1 || isSearching}
                                >
                                    <ChevronLeft className="h-4 w-4 mr-1" />
                                    Previous
                                </Button>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => handleSearch(undefined, currentPage + 1)}
                                    disabled={currentPage === totalPages || isSearching}
                                >
                                    Next
                                    <ChevronRight className="h-4 w-4 ml-1" />
                                </Button>
                            </div>
                        </div>
                    )}
                </>
            )}
        </div>
    );
}

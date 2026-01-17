'use client';

import { useState } from 'react';
import { Document, Page, pdfjs } from 'react-pdf';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { ChevronLeft, ChevronRight, ZoomIn, ZoomOut, Download, Maximize2, Minimize2 } from 'lucide-react';

// Use local worker file to avoid CDN version issues
pdfjs.GlobalWorkerOptions.workerSrc = '/libsystem/pdf.worker.min.mjs';

interface PDFViewerProps {
    fileUrl: string;
    fileName?: string;
}

export default function PDFViewer({ fileUrl, fileName }: PDFViewerProps) {
    const [numPages, setNumPages] = useState<number>(0);
    const [pageNumber, setPageNumber] = useState<number>(1);
    const [scale, setScale] = useState<number>(0.6);
    const [isFullscreen, setIsFullscreen] = useState(false);

    function onDocumentLoadSuccess({ numPages }: { numPages: number }) {
        setNumPages(numPages);
        setPageNumber(1);
    }

    function changePage(offset: number) {
        setPageNumber(prevPageNumber => prevPageNumber + offset);
    }

    function previousPage() {
        changePage(-1);
    }

    function nextPage() {
        changePage(1);
    }

    function zoomIn() {
        setScale(prev => Math.min(prev + 0.2, 3.0));
    }

    function zoomOut() {
        setScale(prev => Math.max(prev - 0.2, 0.5));
    }

    function toggleFullscreen() {
        setIsFullscreen(!isFullscreen);
    }

    // Fullscreen View
    if (isFullscreen) {
        return (
            <div className="fixed inset-0 z-50 bg-background flex flex-col">
                {/* Toolbar */}
                <div className="flex items-center justify-between p-4 border-b bg-background shadow-sm z-10">
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                            <Button variant="outline" size="icon" onClick={previousPage} disabled={pageNumber <= 1}>
                                <ChevronLeft className="h-4 w-4" />
                            </Button>
                            <span className="text-sm font-medium">
                                Page {pageNumber} of {numPages || '--'}
                            </span>
                            <Button variant="outline" size="icon" onClick={nextPage} disabled={pageNumber >= numPages}>
                                <ChevronRight className="h-4 w-4" />
                            </Button>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button variant="outline" size="icon" onClick={zoomOut} disabled={scale <= 0.5}>
                                <ZoomOut className="h-4 w-4" />
                            </Button>
                            <span className="text-sm min-w-[3rem] text-center">{Math.round(scale * 100)}%</span>
                            <Button variant="outline" size="icon" onClick={zoomIn} disabled={scale >= 3.0}>
                                <ZoomIn className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                    <div className="flex items-center gap-2">
                        <Button variant="outline" onClick={() => window.open(fileUrl, '_blank')}>
                            <Download className="mr-2 h-4 w-4" />
                            Download
                        </Button>
                        <Button variant="default" onClick={toggleFullscreen}>
                            <Minimize2 className="mr-2 h-4 w-4" />
                            Exit Full Screen
                        </Button>
                    </div>
                </div>

                {/* PDF Content */}
                <div className="flex-1 overflow-auto bg-gray-100/50 p-8 flex justify-center">
                    <Document
                        file={fileUrl}
                        onLoadSuccess={onDocumentLoadSuccess}
                        loading={
                            <div className="flex items-center justify-center h-full">
                                <p className="text-muted-foreground">Loading PDF...</p>
                            </div>
                        }
                        error={
                            <div className="flex flex-col items-center justify-center h-full text-center">
                                <p className="text-destructive mb-2">Failed to load PDF</p>
                            </div>
                        }
                        className="shadow-lg"
                    >
                        <Page
                            pageNumber={pageNumber}
                            scale={scale}
                            renderTextLayer={false}
                            renderAnnotationLayer={false}
                            className="mb-8"
                        />
                    </Document>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {/* Controls */}
            <Card>
                <CardContent className="p-2 sm:p-4">
                    <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                        {/* Page Navigation */}
                        <div className="flex items-center justify-between sm:justify-start w-full sm:w-auto gap-2 bg-secondary/20 rounded-md p-1">
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                onClick={previousPage}
                                disabled={pageNumber <= 1}
                            >
                                <ChevronLeft className="h-4 w-4" />
                            </Button>
                            <span className="text-xs sm:text-sm font-medium min-w-[3rem] text-center">
                                {pageNumber} / {numPages || '--'}
                            </span>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                onClick={nextPage}
                                disabled={pageNumber >= numPages}
                            >
                                <ChevronRight className="h-4 w-4" />
                            </Button>
                        </div>

                        <div className="flex items-center justify-between sm:justify-end w-full sm:w-auto gap-2">
                            {/* Zoom Controls */}
                            <div className="flex items-center gap-1 bg-secondary/20 rounded-md p-1">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8"
                                    onClick={zoomOut}
                                    disabled={scale <= 0.5}
                                >
                                    <ZoomOut className="h-3 w-3 sm:h-4 sm:w-4" />
                                </Button>
                                <span className="text-xs sm:text-sm min-w-[2.5rem] sm:min-w-[3rem] text-center">
                                    {Math.round(scale * 100)}%
                                </span>
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8"
                                    onClick={zoomIn}
                                    disabled={scale >= 3.0}
                                >
                                    <ZoomIn className="h-3 w-3 sm:h-4 sm:w-4" />
                                </Button>
                            </div>

                            {/* Actions */}
                            <div className="flex items-center gap-1">
                                <Button
                                    variant="outline"
                                    size="icon"
                                    className="h-10 w-10"
                                    onClick={toggleFullscreen}
                                    title="Full Screen"
                                >
                                    <Maximize2 className="h-4 w-4" />
                                </Button>
                                <Button
                                    variant="outline"
                                    className="h-10 px-3 sm:px-4"
                                    onClick={() => window.open(fileUrl, '_blank')}
                                >
                                    <Download className="mr-0 sm:mr-2 h-4 w-4" />
                                    <span className="hidden sm:inline">Download</span>
                                </Button>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* PDF Display */}
            <div className="flex justify-center">
                <Card className="overflow-auto h-[calc(100vh-12rem)] w-full">
                    <CardContent className="p-4">
                        <Document
                            file={fileUrl}
                            onLoadSuccess={onDocumentLoadSuccess}
                            loading={
                                <div className="flex items-center justify-center p-8">
                                    <p className="text-muted-foreground">Loading PDF...</p>
                                </div>
                            }
                            error={
                                <div className="flex flex-col items-center justify-center p-8 text-center">
                                    <p className="text-destructive mb-2">Failed to load PDF</p>
                                    <p className="text-sm text-muted-foreground">
                                        The file may be corrupted or in an unsupported format
                                    </p>
                                </div>
                            }
                        >
                            <Page
                                pageNumber={pageNumber}
                                scale={scale}
                                renderTextLayer={false}
                                renderAnnotationLayer={false}
                            />
                        </Document>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}

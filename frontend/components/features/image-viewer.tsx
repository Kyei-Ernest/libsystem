'use client';

import { useState } from 'react';
import Image from 'next/image';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ZoomIn, ZoomOut, Download, RotateCw, Maximize2, Minimize2 } from 'lucide-react';

interface ImageViewerProps {
    fileUrl: string;
    fileName?: string;
    alt?: string;
}

export default function ImageViewer({ fileUrl, fileName, alt }: ImageViewerProps) {
    const [zoom, setZoom] = useState(100);
    const [rotation, setRotation] = useState(0);
    const [isFullscreen, setIsFullscreen] = useState(false);

    const handleZoomIn = () => setZoom(prev => Math.min(prev + 25, 300));
    const handleZoomOut = () => setZoom(prev => Math.max(prev - 25, 25));
    const handleRotate = () => setRotation(prev => (prev + 90) % 360);

    const handleDownload = () => {
        const link = document.createElement('a');
        link.href = fileUrl;
        link.download = fileName || 'image';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

    const handleFullscreen = () => {
        setIsFullscreen(!isFullscreen);
    };

    if (isFullscreen) {
        return (
            <div
                className="fixed inset-0 z-50 bg-black flex items-center justify-center"
                onClick={handleFullscreen}
            >
                {/* Controls */}
                <div className="absolute top-4 right-4 flex items-center gap-2 z-10" onClick={e => e.stopPropagation()}>
                    <Button variant="secondary" size="icon" onClick={handleZoomOut}>
                        <ZoomOut className="h-4 w-4" />
                    </Button>
                    <span className="text-white text-sm min-w-[50px] text-center">{zoom}%</span>
                    <Button variant="secondary" size="icon" onClick={handleZoomIn}>
                        <ZoomIn className="h-4 w-4" />
                    </Button>
                    <Button variant="secondary" size="icon" onClick={handleRotate}>
                        <RotateCw className="h-4 w-4" />
                    </Button>
                    <Button variant="secondary" size="icon" onClick={handleDownload}>
                        <Download className="h-4 w-4" />
                    </Button>
                    <Button variant="secondary" size="icon" onClick={handleFullscreen}>
                        <Minimize2 className="h-4 w-4" />
                    </Button>
                </div>

                {/* Image */}
                <div className="overflow-auto max-h-full max-w-full p-4" onClick={e => e.stopPropagation()}>
                    <Image
                        src={fileUrl}
                        alt={alt || fileName || 'Image'}
                        width={1200}
                        height={800}
                        className="max-w-none transition-transform duration-200"
                        style={{
                            transform: `scale(${zoom / 100}) rotate(${rotation}deg)`,
                            transformOrigin: 'center center'
                        }}
                        unoptimized
                    />
                </div>
            </div>
        );
    }

    return (
        <Card className="overflow-hidden">
            <CardHeader className="flex flex-row items-center justify-between p-3 sm:p-4 bg-gray-50 border-b">
                <CardTitle className="text-sm sm:text-base truncate">
                    {fileName || 'Image Preview'}
                </CardTitle>
                <div className="flex items-center gap-1 sm:gap-2">
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={handleZoomOut}
                        disabled={zoom <= 25}
                    >
                        <ZoomOut className="h-4 w-4" />
                    </Button>
                    <span className="text-xs min-w-[40px] text-center hidden sm:block">{zoom}%</span>
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={handleZoomIn}
                        disabled={zoom >= 300}
                    >
                        <ZoomIn className="h-4 w-4" />
                    </Button>
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={handleRotate}
                    >
                        <RotateCw className="h-4 w-4" />
                    </Button>
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={handleDownload}
                    >
                        <Download className="h-4 w-4" />
                    </Button>
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={handleFullscreen}
                    >
                        <Maximize2 className="h-4 w-4" />
                    </Button>
                </div>
            </CardHeader>
            <CardContent className="p-4 bg-gray-100 h-[calc(100vh-12rem)] flex flex-col">
                <div className="overflow-auto flex-1 flex items-center justify-center min-h-0">
                    <Image
                        src={fileUrl}
                        alt={alt || fileName || 'Image'}
                        width={800}
                        height={600}
                        className="max-w-none rounded-lg shadow-lg transition-transform duration-200 object-contain"
                        style={{
                            transform: `scale(${zoom / 100}) rotate(${rotation}deg)`,
                            transformOrigin: 'center center',
                            maxHeight: '100%',
                            maxWidth: '100%'
                        }}
                        unoptimized
                    />
                </div>
            </CardContent>
        </Card>
    );
}

'use client';

import { useState } from 'react';
import Image from 'next/image';
import { FileText, Image as ImageIcon, FileArchive, FileSpreadsheet, FileCode, File, Music, Video } from 'lucide-react';

interface DocumentThumbnailProps {
    documentId: string;
    fileType: string;
    title?: string;
    className?: string;
    size?: 'sm' | 'md' | 'lg';
}

// Map file types to icons
const getFileIcon = (fileType: string, size: number) => {
    const type = fileType?.toLowerCase() || '';
    const iconClass = `h-${size} w-${size}`;

    if (type.includes('pdf'))
        return <FileText className={iconClass} style={{ color: '#EF4444' }} />;
    if (type.includes('image') || ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].some(ext => type.includes(ext)))
        return <ImageIcon className={iconClass} style={{ color: '#22C55E' }} />;
    if (type.includes('zip') || type.includes('rar') || type.includes('tar') || type.includes('gz'))
        return <FileArchive className={iconClass} style={{ color: '#EAB308' }} />;
    if (type.includes('xls') || type.includes('csv') || type.includes('spreadsheet'))
        return <FileSpreadsheet className={iconClass} style={{ color: '#10B981' }} />;
    if (type.includes('doc') || type.includes('word') || type.includes('text'))
        return <FileText className={iconClass} style={{ color: '#3B82F6' }} />;
    if (type.includes('js') || type.includes('ts') || type.includes('json') || type.includes('html') || type.includes('css') || type.includes('code'))
        return <FileCode className={iconClass} style={{ color: '#8B5CF6' }} />;
    if (type.includes('audio') || type.includes('mp3') || type.includes('wav'))
        return <Music className={iconClass} style={{ color: '#F97316' }} />;
    if (type.includes('video') || type.includes('mp4') || type.includes('avi') || type.includes('mov'))
        return <Video className={iconClass} style={{ color: '#EC4899' }} />;

    return <File className={iconClass} style={{ color: '#6B7280' }} />;
};

// Get background gradient based on file type
const getBackgroundGradient = (fileType: string) => {
    const type = fileType?.toLowerCase() || '';

    if (type.includes('pdf')) return 'from-red-50 to-red-100';
    if (type.includes('image') || ['png', 'jpg', 'jpeg', 'gif', 'webp'].some(ext => type.includes(ext)))
        return 'from-green-50 to-green-100';
    if (type.includes('doc') || type.includes('word')) return 'from-blue-50 to-blue-100';
    if (type.includes('xls') || type.includes('csv')) return 'from-emerald-50 to-emerald-100';
    if (type.includes('zip') || type.includes('rar')) return 'from-yellow-50 to-yellow-100';
    if (type.includes('audio')) return 'from-orange-50 to-orange-100';
    if (type.includes('video')) return 'from-pink-50 to-pink-100';

    return 'from-gray-50 to-gray-100';
};

export function DocumentThumbnail({
    documentId,
    fileType,
    title = 'Document',
    className = '',
    size = 'md'
}: DocumentThumbnailProps) {
    const [imageError, setImageError] = useState(false);
    const [imageLoaded, setImageLoaded] = useState(false);

    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8088';
    const thumbnailUrl = `${apiUrl}/api/v1/documents/${documentId}/thumbnail`;

    // Determine if the file type might have a real thumbnail
    const isImageType = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'image'].some(
        ext => fileType?.toLowerCase().includes(ext)
    );
    const isPdfType = fileType?.toLowerCase().includes('pdf');
    const canHaveThumbnail = isImageType || isPdfType;

    // Size classes
    const sizeClasses = {
        sm: { container: 'h-16 w-16', icon: 8 },
        md: { container: 'h-24 w-24', icon: 12 },
        lg: { container: 'h-32 w-32', icon: 16 }
    };

    const { container: containerSize, icon: iconSize } = sizeClasses[size];

    return (
        <div
            className={`relative ${containerSize} bg-gradient-to-br ${getBackgroundGradient(fileType)} rounded-lg flex items-center justify-center overflow-hidden ${className}`}
        >
            {/* Try to load real thumbnail for images and PDFs */}
            {canHaveThumbnail && !imageError && (
                <Image
                    src={thumbnailUrl}
                    alt={title}
                    fill
                    className={`object-cover transition-opacity duration-300 ${imageLoaded ? 'opacity-100' : 'opacity-0'}`}
                    onLoad={() => setImageLoaded(true)}
                    onError={() => setImageError(true)}
                    unoptimized
                />
            )}

            {/* Fallback icon - show when image fails or for non-image files */}
            {(imageError || !canHaveThumbnail || !imageLoaded) && (
                <div className={`flex items-center justify-center ${imageLoaded && canHaveThumbnail ? 'hidden' : ''}`}>
                    {getFileIcon(fileType, iconSize)}
                </div>
            )}
        </div>
    );
}

// Larger thumbnail card for document grid
interface DocumentThumbnailCardProps {
    documentId: string;
    fileType: string;
    title?: string;
    className?: string;
}

export function DocumentThumbnailCard({
    documentId,
    fileType,
    title = 'Document',
    className = ''
}: DocumentThumbnailCardProps) {
    const [imageError, setImageError] = useState(false);
    const [imageLoaded, setImageLoaded] = useState(false);

    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8088';
    const thumbnailUrl = `${apiUrl}/api/v1/documents/${documentId}/thumbnail`;

    const isImageType = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'image'].some(
        ext => fileType?.toLowerCase().includes(ext)
    );
    const isPdfType = fileType?.toLowerCase().includes('pdf');
    const canHaveThumbnail = isImageType || isPdfType;

    return (
        <div
            className={`relative aspect-[4/3] bg-gradient-to-br ${getBackgroundGradient(fileType)} rounded-lg flex items-center justify-center overflow-hidden ${className}`}
        >
            {/* Real thumbnail */}
            {canHaveThumbnail && !imageError && (
                <Image
                    src={thumbnailUrl}
                    alt={title}
                    fill
                    className={`object-cover transition-opacity duration-300 ${imageLoaded ? 'opacity-100' : 'opacity-0'}`}
                    onLoad={() => setImageLoaded(true)}
                    onError={() => setImageError(true)}
                    unoptimized
                />
            )}

            {/* Fallback icon */}
            {(imageError || !canHaveThumbnail || !imageLoaded) && (
                <div className={`flex items-center justify-center ${imageLoaded && canHaveThumbnail ? 'hidden' : ''}`}>
                    {getFileIcon(fileType, 16)}
                </div>
            )}
        </div>
    );
}

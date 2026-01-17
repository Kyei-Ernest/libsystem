'use client';

import { useState } from 'react';
import Image from 'next/image';
import { documentsApi } from '@/lib/api/documents';
import {
    FileText,
    Image as ImageIcon,
    FileArchive,
    FileSpreadsheet,
    FileCode,
    File
} from 'lucide-react';

interface DocumentThumbnailProps {
    documentId: string;
    fileType: string;
    title?: string;
    className?: string; // Allow passing external styles
}

// File type icon mapping
const getFileIcon = (fileType: string, size = 'full') => {
    const type = fileType?.toLowerCase() || '';
    const sizeClass = size === 'full' ? 'h-full w-full' : 'h-10 w-10';

    if (type.includes('pdf')) return <FileText className={`${sizeClass} text-red-500`} />;
    if (type.includes('image') || type.includes('png') || type.includes('jpg') || type.includes('jpeg'))
        return <ImageIcon className={`${sizeClass} text-green-500`} />;
    if (type.includes('zip') || type.includes('rar') || type.includes('tar'))
        return <FileArchive className={`${sizeClass} text-yellow-500`} />;
    if (type.includes('xls') || type.includes('csv'))
        return <FileSpreadsheet className={`${sizeClass} text-emerald-500`} />;
    if (type.includes('doc') || type.includes('docx'))
        return <FileText className={`${sizeClass} text-blue-500`} />;
    if (type.includes('html') || type.includes('json') || type.includes('xml'))
        return <FileCode className={`${sizeClass} text-purple-500`} />;
    return <File className={`${sizeClass} text-gray-500`} />;
};

// Get background gradient based on file type
const getBackgroundGradient = (fileType: string) => {
    const type = fileType?.toLowerCase() || '';
    if (type.includes('pdf')) return 'from-red-50 to-red-100';
    if (type.includes('image') || ['png', 'jpg', 'jpeg', 'gif', 'webp'].some(ext => type.includes(ext)))
        return 'from-green-50 to-green-100';
    if (type.includes('doc') || type.includes('word')) return 'from-blue-50 to-blue-100';
    if (type.includes('xls') || type.includes('sheet')) return 'from-emerald-50 to-emerald-100';
    if (type.includes('ppt') || type.includes('presentation')) return 'from-orange-50 to-orange-100';
    return 'from-gray-50 to-gray-100';
};

// Check if file type can have a thumbnail
const canHaveThumbnail = (fileType: string) => {
    const type = fileType?.toLowerCase() || '';
    return type.includes('pdf') ||
        type.includes('image') ||
        ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].some(ext => type.includes(ext)) ||
        type.includes('word') || type.includes('doc') || type.includes('odt') ||
        type.includes('sheet') || type.includes('xls') || type.includes('ods') ||
        type.includes('presentation') || type.includes('ppt') || type.includes('odp') ||
        type.includes('text') || type.includes('txt');
};

export function DocumentThumbnail({ documentId, fileType, title, className = '' }: DocumentThumbnailProps) {
    const [imageError, setImageError] = useState(false);
    const [imageLoaded, setImageLoaded] = useState(false);

    const thumbnailUrl = documentsApi.getThumbnailUrl(documentId);
    const hasThumbnail = canHaveThumbnail(fileType);

    return (
        <div className={`relative aspect-[4/3] bg-gradient-to-br ${getBackgroundGradient(fileType)} flex items-center justify-center overflow-hidden ${className}`}>
            {/* Try loading real thumbnail for images and PDFs */}
            {hasThumbnail && !imageError && (
                <Image
                    src={thumbnailUrl}
                    alt={title || 'Document thumbnail'}
                    fill
                    className={`object-cover group-hover:scale-105 transition-all duration-300 ${imageLoaded ? 'opacity-100' : 'opacity-0'}`}
                    onLoad={() => setImageLoaded(true)}
                    onError={() => setImageError(true)}
                    unoptimized
                />
            )}

            {/* Fallback icon when no thumbnail or error */}
            {(imageError || !hasThumbnail || !imageLoaded) && (
                <div className={`w-16 h-16 sm:w-20 sm:h-20 transition-opacity ${imageLoaded && hasThumbnail ? 'opacity-0' : 'opacity-100'}`}>
                    {getFileIcon(fileType)}
                </div>
            )}
        </div>
    );
}

export { getFileIcon, getBackgroundGradient };

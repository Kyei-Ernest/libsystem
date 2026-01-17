'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2, Download, Copy, Check, WrapText, AlignLeft, Maximize2, Minimize2 } from 'lucide-react';
import { toast } from 'sonner';

interface TextViewerProps {
    fileUrl: string;
    fileName?: string;
}

export default function TextViewer({ fileUrl, fileName }: TextViewerProps) {
    const [content, setContent] = useState<string>('');
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [copied, setCopied] = useState(false);
    const [wordWrap, setWordWrap] = useState(true);

    const [isFullscreen, setIsFullscreen] = useState(false);

    useEffect(() => {
        loadContent();
    }, [fileUrl]);

    const loadContent = async () => {
        try {
            setIsLoading(true);
            setError(null);

            const response = await fetch(fileUrl);
            if (!response.ok) {
                throw new Error('Failed to load file');
            }

            const text = await response.text();
            setContent(text);
        } catch (err) {
            console.error('Failed to load text file:', err);
            setError('Failed to load file content');
        } finally {
            setIsLoading(false);
        }
    };

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(content);
            setCopied(true);
            toast.success('Content copied to clipboard');
            setTimeout(() => setCopied(false), 2000);
        } catch {
            toast.error('Failed to copy content');
        }
    };

    const handleDownload = () => {
        window.open(fileUrl, '_blank');
    };

    const toggleFullscreen = () => {
        setIsFullscreen(!isFullscreen);
    };

    // Determine syntax highlighting class based on file extension
    const getLanguageClass = () => {
        if (!fileName) return '';
        const ext = fileName.split('.').pop()?.toLowerCase();
        switch (ext) {
            case 'js':
            case 'jsx':
                return 'language-javascript';
            case 'ts':
            case 'tsx':
                return 'language-typescript';
            case 'json':
                return 'language-json';
            case 'html':
                return 'language-html';
            case 'css':
                return 'language-css';
            case 'py':
                return 'language-python';
            case 'go':
                return 'language-go';
            case 'md':
                return 'language-markdown';
            case 'xml':
                return 'language-xml';
            case 'yaml':
            case 'yml':
                return 'language-yaml';
            default:
                return '';
        }
    };

    if (isLoading) {
        return (
            <Card>
                <CardContent className="flex items-center justify-center py-12">
                    <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </CardContent>
            </Card>
        );
    }

    if (error) {
        return (
            <Card>
                <CardContent className="py-12 text-center">
                    <p className="text-destructive mb-4">{error}</p>
                    <Button variant="outline" onClick={loadContent}>
                        Try Again
                    </Button>
                </CardContent>
            </Card>
        );
    }

    if (isFullscreen) {
        return (
            <div className="fixed inset-0 z-50 bg-background flex flex-col">
                <div className="flex items-center justify-between p-4 border-b bg-background shadow-sm">
                    <h2 className="text-lg font-semibold">{fileName || 'Text Viewer'}</h2>
                    <div className="flex items-center gap-2">
                        <Button
                            variant="outline"
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => setWordWrap(!wordWrap)}
                            title={wordWrap ? 'Disable word wrap' : 'Enable word wrap'}
                        >
                            {wordWrap ? <AlignLeft className="h-4 w-4" /> : <WrapText className="h-4 w-4" />}
                        </Button>
                        <Button
                            variant="outline"
                            size="icon"
                            className="h-8 w-8"
                            onClick={handleCopy}
                        >
                            {copied ? <Check className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
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
                            variant="default"
                            size="icon"
                            className="h-8 w-8"
                            onClick={toggleFullscreen}
                        >
                            <Minimize2 className="h-4 w-4" />
                        </Button>
                    </div>
                </div>
                <div className="flex-1 overflow-auto bg-gray-900">
                    <pre
                        className={`p-8 text-sm font-mono text-gray-100 ${wordWrap ? 'whitespace-pre-wrap break-words' : 'whitespace-pre'} ${getLanguageClass()}`}
                    >
                        <code>{content}</code>
                    </pre>
                </div>
                <div className="px-4 py-2 bg-background border-t text-xs text-muted-foreground flex items-center justify-between">
                    <span>{content.split('\n').length} lines</span>
                    <span>{content.length.toLocaleString()} characters</span>
                </div>
            </div>
        );
    }

    return (
        <Card className="overflow-hidden">
            <CardHeader className="flex flex-row items-center justify-between p-3 sm:p-4 bg-gray-50 border-b">
                <CardTitle className="text-sm sm:text-base">
                    {fileName || 'Text File'}
                </CardTitle>
                <div className="flex items-center gap-2">
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setWordWrap(!wordWrap)}
                        title={wordWrap ? 'Disable word wrap' : 'Enable word wrap'}
                    >
                        {wordWrap ? <AlignLeft className="h-4 w-4" /> : <WrapText className="h-4 w-4" />}
                    </Button>
                    <Button
                        variant="outline"
                        size="icon"
                        className="h-8 w-8"
                        onClick={handleCopy}
                    >
                        {copied ? <Check className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
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
                        onClick={toggleFullscreen}
                    >
                        <Maximize2 className="h-4 w-4" />
                    </Button>
                </div>
            </CardHeader>
            <CardContent className="p-0">
                <div className="h-[calc(100vh-12rem)] overflow-auto">
                    <pre
                        className={`p-4 text-sm font-mono bg-gray-900 text-gray-100 ${wordWrap ? 'whitespace-pre-wrap break-words' : 'whitespace-pre'} ${getLanguageClass()}`}
                    >
                        <code>{content}</code>
                    </pre>
                </div>
                <div className="px-4 py-2 bg-gray-100 border-t text-xs text-muted-foreground flex items-center justify-between">
                    <span>{content.split('\n').length} lines</span>
                    <span>{content.length.toLocaleString()} characters</span>
                </div>
            </CardContent>
        </Card>
    );
}

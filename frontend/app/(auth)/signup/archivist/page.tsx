'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { authApi } from '@/lib/api/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { toast } from 'sonner';
import { User, Mail, Lock, Archive, AlertCircle } from 'lucide-react';

export default function ArchivistSignupPage() {
    const router = useRouter();
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string>('');

    const [formData, setFormData] = useState({
        email: '',
        username: '',
        password: '',
        first_name: '',
        last_name: '',
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            await authApi.register({
                ...formData,
                role: 'archivist',
            });

            toast.success('Application submitted! Awaiting approval.');
            router.push('/login?message=approval_pending');
        } catch (err: any) {
            setError(err.response?.data?.error?.message || 'Registration failed');
            toast.error('Registration failed');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 via-white to-pink-50 py-12 px-4 sm:px-6 lg:px-8">
            <Card className="w-full max-w-md shadow-xl">
                <CardHeader className="space-y-4 text-center">
                    <div className="mx-auto w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center">
                        <Archive className="h-8 w-8 text-purple-600" />
                    </div>
                    <div>
                        <CardTitle className="text-3xl font-bold">Apply as Archivist</CardTitle>
                        <CardDescription className="text-base mt-2">
                            Specialized cataloging and digital preservation
                        </CardDescription>
                    </div>

                    <Alert className="bg-purple-50 border-purple-200">
                        <AlertCircle className="h-4 w-4 text-purple-600" />
                        <AlertDescription className="text-purple-900">
                            Your application will be reviewed by an administrator
                        </AlertDescription>
                    </Alert>
                </CardHeader>

                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {error && (
                            <Alert variant="destructive">
                                <AlertDescription>{error}</AlertDescription>
                            </Alert>
                        )}

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="first_name">First Name</Label>
                                <Input
                                    id="first_name"
                                    value={formData.first_name}
                                    onChange={(e) =>
                                        setFormData({ ...formData, first_name: e.target.value })
                                    }
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="last_name">Last Name</Label>
                                <Input
                                    id="last_name"
                                    value={formData.last_name}
                                    onChange={(e) =>
                                        setFormData({ ...formData, last_name: e.target.value })
                                    }
                                    required
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="email">Professional Email</Label>
                            <div className="relative">
                                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input
                                    id="email"
                                    type="email"
                                    className="pl-10"
                                    value={formData.email}
                                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                    required
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="username">Username</Label>
                            <div className="relative">
                                <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input
                                    id="username"
                                    className="pl-10"
                                    value={formData.username}
                                    onChange={(e) =>
                                        setFormData({ ...formData, username: e.target.value })
                                    }
                                    required
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="password">Password</Label>
                            <div className="relative">
                                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input
                                    id="password"
                                    type="password"
                                    className="pl-10"
                                    value={formData.password}
                                    onChange={(e) =>
                                        setFormData({ ...formData, password: e.target.value })
                                    }
                                    required
                                />
                            </div>
                        </div>

                        <div className="bg-purple-50 p-4 rounded-lg border border-purple-200">
                            <h4 className="font-semibold text-sm text-purple-900 mb-2">Archivist Capabilities:</h4>
                            <ul className="text-xs text-purple-700 space-y-1">
                                <li>✓ Advanced metadata management</li>
                                <li>✓ Digital preservation workflows</li>
                                <li>✓ Historical document cataloging</li>
                                <li>✓ Quality control and verification</li>
                            </ul>
                        </div>

                        <Button type="submit" className="w-full bg-purple-600 hover:bg-purple-700" disabled={isLoading}>
                            {isLoading ? 'Submitting Application...' : 'Submit Application'}
                        </Button>

                        <div className="text-center text-sm space-y-2">
                            <div>
                                <span className="text-muted-foreground">Already have an account? </span>
                                <Link href="/login" className="text-purple-600 hover:underline font-medium">
                                    Sign in
                                </Link>
                            </div>
                            <div>
                                <Link href="/register" className="text-muted-foreground hover:text-foreground">
                                    ← Choose a different role
                                </Link>
                            </div>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}

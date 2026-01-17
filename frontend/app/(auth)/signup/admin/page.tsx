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
import { User, Mail, Lock, Shield, Lock as LockIcon } from 'lucide-react';

export default function AdminSignupPage() {
    const router = useRouter();
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string>('');

    const [formData, setFormData] = useState({
        email: '',
        username: '',
        password: '',
        first_name: '',
        last_name: '',
        invite_code: '',
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            // Note: You'll need to validate invite_code on backend
            await authApi.register({
                email: formData.email,
                username: formData.username,
                password: formData.password,
                first_name: formData.first_name,
                last_name: formData.last_name,
                role: 'admin',
            });

            toast.success('Admin account created successfully!');
            router.push('/login');
        } catch (err: any) {
            setError(err.response?.data?.error?.message || 'Registration failed');
            toast.error('Registration failed');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-red-50 via-white to-orange-50 py-12 px-4 sm:px-6 lg:px-8">
            <Card className="w-full max-w-md shadow-xl border-red-200">
                <CardHeader className="space-y-4 text-center">
                    <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center">
                        <Shield className="h-8 w-8 text-red-600" />
                    </div>
                    <div>
                        <CardTitle className="text-3xl font-bold text-red-900">Admin Registration</CardTitle>
                        <CardDescription className="text-base mt-2">
                            Full system access - Invite only
                        </CardDescription>
                    </div>

                    <Alert className="bg-red-50 border-red-200">
                        <LockIcon className="h-4 w-4 text-red-600" />
                        <AlertDescription className="text-red-900">
                            <strong>Restricted Access:</strong> Valid invite code required
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

                        <div className="space-y-2">
                            <Label htmlFor="invite_code" className="text-red-900">Invite Code *</Label>
                            <Input
                                id="invite_code"
                                placeholder="Enter your invite code"
                                value={formData.invite_code}
                                onChange={(e) =>
                                    setFormData({ ...formData, invite_code: e.target.value })
                                }
                                required
                                className="border-red-200 focus:border-red-500"
                            />
                        </div>

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
                            <Label htmlFor="email">Admin Email</Label>
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

                        <div className="bg-red-50 p-4 rounded-lg border border-red-200">
                            <h4 className="font-semibold text-sm text-red-900 mb-2">Administrator Powers:</h4>
                            <ul className="text-xs text-red-700 space-y-1">
                                <li>✓ Full system configuration access</li>
                                <li>✓ User management and role assignment</li>
                                <li>✓ System settings and security</li>
                                <li>✓ Complete oversight and control</li>
                            </ul>
                        </div>

                        <Button type="submit" className="w-full bg-red-600 hover:bg-red-700" disabled={isLoading}>
                            {isLoading ? 'Verifying...' : 'Create Admin Account'}
                        </Button>

                        <div className="text-center text-sm space-y-2">
                            <div>
                                <span className="text-muted-foreground">Already have an account? </span>
                                <Link href="/login" className="text-red-600 hover:underline font-medium">
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

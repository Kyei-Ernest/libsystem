'use client';

import { use } from 'react';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import Link from 'next/link';
import { authApi } from '@/lib/api/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { toast } from 'sonner';
import { BookOpen, Shield, User } from 'lucide-react';
import type { UserRole } from '@/types/api';

interface RoleInfo {
    title: string;
    description: string;
    icon: any;
    color: string;
    requiresApproval: boolean;
}

const roleInfo: Record<UserRole, RoleInfo> = {
    admin: {
        title: 'Administrator',
        description: 'Full system access and control. Invite-only.',
        icon: Shield,
        color: 'text-red-600',
        requiresApproval: true,
    },
    librarian: {
        title: 'Librarian',
        description: 'Manage collections and upload documents. Requires approval.',
        icon: BookOpen,
        color: 'text-blue-600',
        requiresApproval: true,
    },
    patron: {
        title: 'Patron',
        description: 'Browse and access library documents.',
        icon: User,
        color: 'text-green-600',
        requiresApproval: false,
    },
    archivist: {
        title: 'Archivist',
        description: 'Specialized cataloging and metadata management.',
        icon: BookOpen,
        color: 'text-purple-600',
        requiresApproval: true,
    },
    vendor: {
        title: 'Vendor',
        description: 'Upload documents only. Requires verification.',
        icon: User,
        color: 'text-orange-600',
        requiresApproval: true,
    },
};

export default function RoleRegistrationPage({
    params,
}: {
    params: Promise<{ role: string }>;
}) {
    const { role: roleParam } = use(params);
    const role = roleParam as UserRole;
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

    const info = roleInfo[role];

    if (!info) {
        router.push('/register');
        return null;
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            await authApi.register({
                ...formData,
                role,
            });

            if (info.requiresApproval) {
                toast.success('Registration submitted! Awaiting approval.');
                router.push('/login?message=approval_pending');
            } else {
                toast.success('Registration successful!');
                router.push('/login');
            }
        } catch (err: any) {
            setError(err.response?.data?.error?.message || 'Registration failed');
            toast.error('Registration failed');
        } finally {
            setIsLoading(false);
        }
    };

    const Icon = info.icon;

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <Card className="w-full max-w-md">
                <CardHeader className="space-y-4">
                    <div className="flex items-center justify-center">
                        <Icon className={`h-12 w-12 ${info.color}`} />
                    </div>
                    <div className="text-center">
                        <CardTitle className="text-2xl">Register as {info.title}</CardTitle>
                        <CardDescription>{info.description}</CardDescription>
                    </div>

                    {info.requiresApproval && (
                        <Alert>
                            <AlertDescription>
                                Your registration will be reviewed by an administrator before activation.
                            </AlertDescription>
                        </Alert>
                    )}
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
                            <Label htmlFor="email">Email</Label>
                            <Input
                                id="email"
                                type="email"
                                value={formData.email}
                                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                required
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="username">Username</Label>
                            <Input
                                id="username"
                                value={formData.username}
                                onChange={(e) =>
                                    setFormData({ ...formData, username: e.target.value })
                                }
                                required
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="password">Password</Label>
                            <Input
                                id="password"
                                type="password"
                                value={formData.password}
                                onChange={(e) =>
                                    setFormData({ ...formData, password: e.target.value })
                                }
                                required
                            />
                        </div>

                        <Button type="submit" className="w-full" disabled={isLoading}>
                            {isLoading ? 'Registering...' : 'Register'}
                        </Button>

                        <div className="text-center text-sm">
                            <span className="text-muted-foreground">Already have an account? </span>
                            <Link href="/login" className="text-primary hover:underline">
                                Sign in
                            </Link>
                        </div>

                        <div className="text-center text-sm">
                            <Link href="/register" className="text-muted-foreground hover:underline">
                                ← Choose a different role
                            </Link>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}

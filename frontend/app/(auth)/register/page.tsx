'use client';

import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { BookOpen, Library, Shield, Archive, Package } from 'lucide-react';

const roles = [
    {
        id: 'patron',
        name: 'Patron',
        description: 'Browse and access library resources',
        icon: BookOpen,
        color: 'green',
        bgColor: 'bg-green-100',
        textColor: 'text-green-600',
        borderColor: 'hover:border-green-300',
        href: '/signup/patron',
        instant: true,
    },
    {
        id: 'librarian',
        name: 'Librarian',
        description: 'Manage collections and content',
        icon: Library,
        color: 'blue',
        bgColor: 'bg-blue-100',
        textColor: 'text-blue-600',
        borderColor: 'hover:border-blue-300',
        href: '/signup/librarian',
        instant: false,
    },
    {
        id: 'archivist',
        name: 'Archivist',
        description: 'Specialized cataloging and preservation',
        icon: Archive,
        color: 'purple',
        bgColor: 'bg-purple-100',
        textColor: 'text-purple-600',
        borderColor: 'hover:border-purple-300',
        href: '/signup/archivist',
        instant: false,
    },
    {
        id: 'vendor',
        name: 'Vendor',
        description: 'Upload and distribute content',
        icon: Package,
        color: 'orange',
        bgColor: 'bg-orange-100',
        textColor: 'text-orange-600',
        borderColor: 'hover:border-orange-300',
        href: '/signup/vendor',
        instant: false,
    },
    {
        id: 'admin',
        name: 'Administrator',
        description: 'Full system access (invite only)',
        icon: Shield,
        color: 'red',
        bgColor: 'bg-red-100',
        textColor: 'text-red-600',
        borderColor: 'hover:border-red-300',
        href: '/signup/admin',
        instant: false,
    },
];

export default function RegisterPage() {
    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-5xl mx-auto">
                <div className="text-center mb-12">
                    <h1 className="text-4xl font-bold text-gray-900 mb-4">
                        Join LibSystem
                    </h1>
                    <p className="text-lg text-gray-600">
                        Choose your role to get started
                    </p>
                </div>

                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 mb-8">
                    {roles.map((role) => {
                        const Icon = role.icon;
                        return (
                            <Card
                                key={role.id}
                                className={`transition-all hover:shadow-lg ${role.borderColor} cursor-pointer`}
                            >
                                <CardHeader>
                                    <div className={`w-12 h-12 ${role.bgColor} rounded-lg flex items-center justify-center mb-4`}>
                                        <Icon className={`h-6 w-6 ${role.textColor}`} />
                                    </div>
                                    <CardTitle className="text-xl">{role.name}</CardTitle>
                                    <CardDescription>{role.description}</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <Link href={role.href}>
                                        <Button className="w-full" variant="outline">
                                            {role.instant ? 'Sign Up Now' : 'Apply Now'}
                                        </Button>
                                    </Link>
                                    {!role.instant && (
                                        <p className="text-xs text-muted-foreground mt-2 text-center">
                                            Requires approval
                                        </p>
                                    )}
                                </CardContent>
                            </Card>
                        );
                    })}
                </div>

                <div className="text-center">
                    <p className="text-sm text-gray-600">
                        Already have an account?{' '}
                        <Link href="/login" className="text-blue-600 hover:underline font-medium">
                            Sign in
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    );
}

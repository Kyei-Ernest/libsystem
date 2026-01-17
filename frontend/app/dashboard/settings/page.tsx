'use client';

import { useState } from 'react';
import { useAuthStore } from '@/stores/auth-store';
import { usersApi } from '@/lib/api/users';
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { toast } from 'sonner';
import { Loader2, Save, User as UserIcon, Bell, Shield } from 'lucide-react';

export default function SettingsPage() {
    const { user, updateUser } = useAuthStore();
    const [isLoading, setIsLoading] = useState(false);
    const [formData, setFormData] = useState({
        first_name: user?.first_name || '',
        last_name: user?.last_name || '',
    });

    if (!user) return null;

    const handleUpdateProfile = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            setIsLoading(true);
            const updatedUser = await usersApi.update(user.id, formData);

            // Update local store
            // Note: In a real app, we might need to refresh the token if claims change,
            // but for name changes, updating the user object in store is usually enough.
            updateUser({ ...user, ...updatedUser });

            toast.success('Profile updated successfully');
        } catch (error: any) {
            console.error('Failed to update profile:', error);
            const message = error.response?.data?.error?.message || 'Failed to update profile';
            toast.error(message);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="space-y-6 max-w-4xl mx-auto">
            <div>
                <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
                <p className="text-muted-foreground">Manage your account and preferences</p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Profile Information */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <UserIcon className="h-5 w-5 text-blue-500" />
                            Profile Information
                        </CardTitle>
                        <CardDescription>Update your personal details</CardDescription>
                    </CardHeader>
                    <form onSubmit={handleUpdateProfile}>
                        <CardContent className="space-y-4">
                            <div className="grid gap-2">
                                <Label>Email</Label>
                                <Input value={user.email} disabled className="bg-muted" />
                                <p className="text-[0.8rem] text-muted-foreground">Email cannot be changed</p>
                            </div>
                            <div className="grid gap-2">
                                <Label>Username</Label>
                                <Input value={user.username} disabled className="bg-muted" />
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="grid gap-2">
                                    <Label htmlFor="firstName">First Name</Label>
                                    <Input
                                        id="firstName"
                                        value={formData.first_name}
                                        onChange={(e) => setFormData(prev => ({ ...prev, first_name: e.target.value }))}
                                    />
                                </div>
                                <div className="grid gap-2">
                                    <Label htmlFor="lastName">Last Name</Label>
                                    <Input
                                        id="lastName"
                                        value={formData.last_name}
                                        onChange={(e) => setFormData(prev => ({ ...prev, last_name: e.target.value }))}
                                    />
                                </div>
                            </div>
                            <div className="grid gap-2">
                                <Label>Role</Label>
                                <div className="flex items-center gap-2 border p-2 rounded-md bg-muted">
                                    <Shield className="h-4 w-4 text-muted-foreground" />
                                    <span className="capitalize text-sm font-medium">{user.role}</span>
                                </div>
                            </div>
                        </CardContent>
                        <CardFooter>
                            <Button type="submit" disabled={isLoading}>
                                {isLoading ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        Saving...
                                    </>
                                ) : (
                                    <>
                                        <Save className="mr-2 h-4 w-4" />
                                        Save Changes
                                    </>
                                )}
                            </Button>
                        </CardFooter>
                    </form>
                </Card>

                {/* Preferences */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Bell className="h-5 w-5 text-yellow-500" />
                            Preferences
                        </CardTitle>
                        <CardDescription>Manage your notifications and display settings</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-6">
                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="email-notifications" className="flex flex-col space-y-1">
                                <span>Email Notifications</span>
                                <span className="font-normal leading-snug text-muted-foreground">
                                    Receive emails about new documents and updates
                                </span>
                            </Label>
                            <Switch id="email-notifications" defaultChecked />
                        </div>
                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="marketing-emails" className="flex flex-col space-y-1">
                                <span>Marketing Emails</span>
                                <span className="font-normal leading-snug text-muted-foreground">
                                    Receive news and special offers
                                </span>
                            </Label>
                            <Switch id="marketing-emails" />
                        </div>
                        <div className="flex items-center justify-between space-x-2">
                            <Label htmlFor="theme" className="flex flex-col space-y-1">
                                <span>Dark Mode</span>
                                <span className="font-normal leading-snug text-muted-foreground">
                                    Toggle system theme (coming soon)
                                </span>
                            </Label>
                            <Switch id="theme" disabled />
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}

'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
    Users, Activity, Smartphone, Shield, 
    Loader2, Power, PowerOff, Wifi, WifiOff,
    RefreshCw, TrendingUp, Edit2, Check, X
} from 'lucide-react';

interface Stats {
    total_users: number;
    active_users: number;
    wa_enabled_users: number;
    active_wa_connections: number;
    admin_count: number;
}

interface User {
    id: number;
    username: string;
    role: string;
    schema_name: string;
    is_active: boolean;
    wa_enabled: boolean;
    wa_connected: boolean;
    created_at: string;
    daily_limit: number;
    monthly_limit: number;
}

export default function AdminDashboardPage() {
    const [stats, setStats] = useState<Stats | null>(null);
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(true);
    const [actionLoading, setActionLoading] = useState<number | null>(null);
    const [editingLimits, setEditingLimits] = useState<{userId: number, daily: number, monthly: number} | null>(null);

    const fetchStats = async () => {
        try {
            const { data } = await api.get('/admin/stats');
            setStats(data);
        } catch (error: any) {
            if (error.response?.status === 403) {
                alert('Admin access required');
            }
        }
    };

    const fetchUsers = async () => {
        try {
            const { data } = await api.get('/admin/users');
            setUsers(data || []);
        } catch (error) {
            console.error('Failed to fetch users', error);
        } finally {
            setLoading(false);
        }
    };

    const toggleUserStatus = async (userId: number, currentStatus: boolean) => {
        setActionLoading(userId);
        try {
            await api.put(`/admin/users/${userId}/status`, { is_active: !currentStatus });
            fetchUsers();
            fetchStats();
        } catch (error: any) {
            alert(error.response?.data?.error || 'Failed to update user');
        } finally {
            setActionLoading(null);
        }
    };

    const toggleWAEnabled = async (userId: number, currentStatus: boolean) => {
        setActionLoading(userId);
        try {
            await api.put(`/admin/users/${userId}/whatsapp`, { wa_enabled: !currentStatus });
            fetchUsers();
            fetchStats();
        } catch (error: any) {
            alert(error.response?.data?.error || 'Failed to update user');
        } finally {
            setActionLoading(null);
        }
    };

    const disconnectWA = async (userId: number) => {
        if (!confirm('Disconnect this user\'s WhatsApp?')) return;
        setActionLoading(userId);
        try {
            await api.post(`/admin/users/${userId}/disconnect-wa`);
            fetchUsers();
            fetchStats();
        } catch (error: any) {
            alert(error.response?.data?.error || 'Failed to disconnect');
        } finally {
            setActionLoading(null);
        }
    };

    const updateLimits = async () => {
        if (!editingLimits) return;
        setActionLoading(editingLimits.userId);
        try {
            await api.put(`/admin/users/${editingLimits.userId}/limits`, {
                daily_limit: editingLimits.daily,
                monthly_limit: editingLimits.monthly
            });
            setEditingLimits(null);
            fetchUsers();
        } catch (error: any) {
            alert(error.response?.data?.error || 'Failed to update limits');
        } finally {
            setActionLoading(null);
        }
    };

    useEffect(() => {
        fetchStats();
        fetchUsers();
    }, []);

    if (loading) {
        return (
            <div className="flex items-center justify-center h-[60vh]">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight flex items-center gap-2">
                        <Shield className="h-8 w-8 text-primary" />
                        Admin Dashboard
                    </h2>
                    <p className="text-gray-500">Platform statistics and user management</p>
                </div>
                <Button variant="outline" onClick={() => { fetchStats(); fetchUsers(); }}>
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Refresh
                </Button>
            </div>

            {/* Stats Cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
                <Card className="bg-gradient-to-br from-blue-500 to-blue-600 text-white">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium opacity-90">Total Users</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <span className="text-3xl font-bold">{stats?.total_users || 0}</span>
                            <Users className="h-8 w-8 opacity-50" />
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-gradient-to-br from-green-500 to-green-600 text-white">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium opacity-90">Active Users</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <span className="text-3xl font-bold">{stats?.active_users || 0}</span>
                            <Activity className="h-8 w-8 opacity-50" />
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-gradient-to-br from-emerald-500 to-emerald-600 text-white">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium opacity-90">WA Enabled</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <span className="text-3xl font-bold">{stats?.wa_enabled_users || 0}</span>
                            <Smartphone className="h-8 w-8 opacity-50" />
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-gradient-to-br from-purple-500 to-purple-600 text-white">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium opacity-90">Connected Now</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <span className="text-3xl font-bold">{stats?.active_wa_connections || 0}</span>
                            <Wifi className="h-8 w-8 opacity-50" />
                        </div>
                    </CardContent>
                </Card>

                <Card className="bg-gradient-to-br from-orange-500 to-orange-600 text-white">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium opacity-90">Admins</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <span className="text-3xl font-bold">{stats?.admin_count || 0}</span>
                            <Shield className="h-8 w-8 opacity-50" />
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Activity Chart Placeholder */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <TrendingUp className="h-5 w-5" />
                        Platform Overview
                    </CardTitle>
                    <CardDescription>User and WhatsApp activity summary</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="h-32 flex items-center justify-center bg-gray-50 rounded-lg">
                        <div className="text-center">
                            <div className="flex items-center gap-6 text-gray-500">
                                <div className="text-center">
                                    <div className="text-2xl font-bold text-green-600">
                                        {stats?.active_users ? Math.round((stats.active_users / stats.total_users) * 100) : 0}%
                                    </div>
                                    <div className="text-xs">Active Rate</div>
                                </div>
                                <div className="h-12 w-px bg-gray-200" />
                                <div className="text-center">
                                    <div className="text-2xl font-bold text-purple-600">
                                        {stats?.active_wa_connections || 0}/{stats?.wa_enabled_users || 0}
                                    </div>
                                    <div className="text-xs">Connected/Enabled</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Users Table */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Users className="h-5 w-5" />
                        User Management
                    </CardTitle>
                    <CardDescription>Manage user accounts and WhatsApp access</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="border-b bg-gray-50">
                                    <th className="text-left p-3 font-medium">ID</th>
                                    <th className="text-left p-3 font-medium">Username</th>
                                    <th className="text-left p-3 font-medium">Role</th>
                                    <th className="text-left p-3 font-medium">Schema</th>
                                    <th className="text-left p-3 font-medium">Status</th>
                                    <th className="text-left p-3 font-medium">WhatsApp</th>
                                    <th className="text-left p-3 font-medium">Limits</th>
                                    <th className="text-left p-3 font-medium">Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {users.map(user => (
                                    <tr key={user.id} className="border-b hover:bg-gray-50">
                                        <td className="p-3 font-mono text-gray-500">{user.id}</td>
                                        <td className="p-3">
                                            <span className="font-medium">{user.username}</span>
                                        </td>
                                        <td className="p-3">
                                            <Badge variant={user.role === 'admin' ? 'default' : 'secondary'}>
                                                {user.role}
                                            </Badge>
                                        </td>
                                        <td className="p-3">
                                            <code className="text-xs bg-gray-100 px-2 py-1 rounded">
                                                {user.schema_name || 'N/A'}
                                            </code>
                                        </td>
                                        <td className="p-3">
                                            {user.is_active ? (
                                                <Badge className="bg-green-100 text-green-700 hover:bg-green-100">
                                                    Active
                                                </Badge>
                                            ) : (
                                                <Badge variant="destructive">Disabled</Badge>
                                            )}
                                        </td>
                                        <td className="p-3">
                                            <div className="flex items-center gap-2">
                                                {user.wa_enabled ? (
                                                    <Badge className="bg-emerald-100 text-emerald-700 hover:bg-emerald-100">
                                                        Enabled
                                                    </Badge>
                                                ) : (
                                                    <Badge variant="outline">Disabled</Badge>
                                                )}
                                                {user.wa_connected && (
                                                    <span title="Connected">
                                                        <Wifi className="h-4 w-4 text-green-500" />
                                                    </span>
                                                )}
                                            </div>
                                        </td>
                                        <td className="p-3">
                                            {editingLimits?.userId === user.id ? (
                                                <div className="flex items-center gap-2">
                                                    <div className="flex flex-col gap-1">
                                                        <input
                                                            type="number"
                                                            value={editingLimits.daily}
                                                            onChange={(e) => setEditingLimits({...editingLimits, daily: parseInt(e.target.value) || 0})}
                                                            className="w-20 px-2 py-1 text-xs border rounded"
                                                            placeholder="Daily"
                                                        />
                                                        <input
                                                            type="number"
                                                            value={editingLimits.monthly}
                                                            onChange={(e) => setEditingLimits({...editingLimits, monthly: parseInt(e.target.value) || 0})}
                                                            className="w-20 px-2 py-1 text-xs border rounded"
                                                            placeholder="Monthly"
                                                        />
                                                    </div>
                                                    <div className="flex flex-col gap-1">
                                                        <Button variant="ghost" size="sm" onClick={updateLimits} className="h-6 w-6 p-0">
                                                            <Check className="h-3 w-3 text-green-600" />
                                                        </Button>
                                                        <Button variant="ghost" size="sm" onClick={() => setEditingLimits(null)} className="h-6 w-6 p-0">
                                                            <X className="h-3 w-3 text-red-500" />
                                                        </Button>
                                                    </div>
                                                </div>
                                            ) : (
                                                <div className="flex items-center gap-2">
                                                    <div className="text-xs">
                                                        <div>{user.daily_limit}/day</div>
                                                        <div className="text-gray-400">{user.monthly_limit}/mo</div>
                                                    </div>
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => setEditingLimits({userId: user.id, daily: user.daily_limit, monthly: user.monthly_limit})}
                                                        className="h-6 w-6 p-0"
                                                    >
                                                        <Edit2 className="h-3 w-3 text-gray-400" />
                                                    </Button>
                                                </div>
                                            )}
                                        </td>
                                        <td className="p-3">
                                            <div className="flex items-center gap-1">
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={() => toggleUserStatus(user.id, user.is_active)}
                                                    disabled={actionLoading === user.id || user.role === 'admin'}
                                                    title={user.is_active ? 'Disable account' : 'Enable account'}
                                                >
                                                    {actionLoading === user.id ? (
                                                        <Loader2 className="h-4 w-4 animate-spin" />
                                                    ) : user.is_active ? (
                                                        <PowerOff className="h-4 w-4 text-red-500" />
                                                    ) : (
                                                        <Power className="h-4 w-4 text-green-500" />
                                                    )}
                                                </Button>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    onClick={() => toggleWAEnabled(user.id, user.wa_enabled)}
                                                    disabled={actionLoading === user.id}
                                                    title={user.wa_enabled ? 'Disable WhatsApp' : 'Enable WhatsApp'}
                                                >
                                                    {user.wa_enabled ? (
                                                        <Smartphone className="h-4 w-4 text-emerald-500" />
                                                    ) : (
                                                        <Smartphone className="h-4 w-4 text-gray-300" />
                                                    )}
                                                </Button>
                                                {user.wa_connected && (
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => disconnectWA(user.id)}
                                                        disabled={actionLoading === user.id}
                                                        title="Force disconnect WhatsApp"
                                                    >
                                                        <WifiOff className="h-4 w-4 text-red-500" />
                                                    </Button>
                                                )}
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}

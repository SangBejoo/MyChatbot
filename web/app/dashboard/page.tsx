'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import api from '@/lib/api';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { 
  Menu, Database, Settings, Smartphone, 
  ArrowRight, CheckCircle, XCircle, Loader2,
  TrendingUp, Zap, MessageSquare
} from 'lucide-react';

interface QuotaStatus {
  daily_limit: number;
  monthly_limit: number;
  today_sent: number;
  month_sent: number;
  daily_remaining: number;
  monthly_remaining: number;
  daily_percent: number;
  monthly_percent: number;
}

interface DashboardStats {
  menu_count: number;
  table_count: number;
  config_count: number;
  wa_connected: boolean;
  wa_phone: string;
  wa_name: string;
  schema_name: string;
  quota?: QuotaStatus;
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.get('/dashboard/stats')
      .then(res => setStats(res.data))
      .catch(console.error)
      .finally(() => setLoading(false));
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
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
          <p className="text-gray-500">Welcome to your bot management portal</p>
        </div>
        <div className="text-right">
          <p className="text-xs text-gray-400">Tenant Schema</p>
          <code className="text-sm bg-gray-100 px-2 py-1 rounded font-mono">{stats?.schema_name}</code>
        </div>
      </div>
      
      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {/* Menus */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Bot Menus</CardTitle>
            <Menu className="h-5 w-5 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.menu_count || 0}</div>
            <p className="text-xs text-gray-400 mt-1">Interactive menu configurations</p>
            <Link href="/dashboard/menus">
              <Button variant="link" className="p-0 h-auto mt-2 text-blue-600">
                Manage <ArrowRight className="h-3 w-3 ml-1" />
              </Button>
            </Link>
          </CardContent>
        </Card>

        {/* Datasets */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Datasets</CardTitle>
            <Database className="h-5 w-5 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.table_count || 0}</div>
            <p className="text-xs text-gray-400 mt-1">Data tables uploaded</p>
            <Link href="/dashboard/data">
              <Button variant="link" className="p-0 h-auto mt-2 text-emerald-600">
                View Data <ArrowRight className="h-3 w-3 ml-1" />
              </Button>
            </Link>
          </CardContent>
        </Card>

        {/* Configs */}
        <Card className="hover:shadow-lg transition-shadow">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Configurations</CardTitle>
            <Settings className="h-5 w-5 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.config_count || 0}</div>
            <p className="text-xs text-gray-400 mt-1">Bot settings configured</p>
            <Link href="/dashboard/settings">
              <Button variant="link" className="p-0 h-auto mt-2 text-purple-600">
                Configure <ArrowRight className="h-3 w-3 ml-1" />
              </Button>
            </Link>
          </CardContent>
        </Card>

        {/* WhatsApp Status */}
        <Card className={`hover:shadow-lg transition-shadow ${stats?.wa_connected ? 'ring-2 ring-green-200' : ''}`}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">WhatsApp</CardTitle>
            <Smartphone className={`h-5 w-5 ${stats?.wa_connected ? 'text-green-500' : 'text-gray-400'}`} />
          </CardHeader>
          <CardContent>
            {stats?.wa_connected ? (
              <>
                <div className="flex items-center gap-2">
                  <CheckCircle className="h-5 w-5 text-green-500" />
                  <span className="text-lg font-bold text-green-700">Connected</span>
                </div>
                <p className="text-xs text-gray-500 mt-1 font-mono">+{stats.wa_phone}</p>
              </>
            ) : (
              <>
                <div className="flex items-center gap-2">
                  <XCircle className="h-5 w-5 text-gray-400" />
                  <span className="text-lg font-medium text-gray-500">Not Connected</span>
                </div>
                <p className="text-xs text-gray-400 mt-1">Connect your WhatsApp</p>
              </>
            )}
            <Link href="/dashboard/whatsapp">
              <Button variant="link" className="p-0 h-auto mt-2 text-blue-600">
                {stats?.wa_connected ? 'Manage' : 'Connect'} <ArrowRight className="h-3 w-3 ml-1" />
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>

      {/* Message Quota */}
      {stats?.quota && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MessageSquare className="h-5 w-5 text-blue-500" />
              Message Quota
            </CardTitle>
            <CardDescription>Your WhatsApp message usage and limits</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-2">
              {/* Daily Usage */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="font-medium">Daily Usage</span>
                  <span className={stats.quota.daily_percent >= 80 ? 'text-red-600 font-medium' : 'text-gray-500'}>
                    {stats.quota.today_sent} / {stats.quota.daily_limit}
                  </span>
                </div>
                <div className="h-3 bg-gray-100 rounded-full overflow-hidden">
                  <div 
                    className={`h-full rounded-full transition-all ${
                      stats.quota.daily_percent >= 80 ? 'bg-red-500' : 
                      stats.quota.daily_percent >= 50 ? 'bg-yellow-500' : 'bg-green-500'
                    }`}
                    style={{ width: `${Math.min(stats.quota.daily_percent, 100)}%` }}
                  />
                </div>
                <p className="text-xs text-gray-400">
                  {stats.quota.daily_remaining > 0 
                    ? `${stats.quota.daily_remaining} messages remaining today`
                    : '⚠️ Daily limit reached'
                  }
                </p>
              </div>

              {/* Monthly Usage */}
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="font-medium">Monthly Usage</span>
                  <span className={stats.quota.monthly_percent >= 80 ? 'text-red-600 font-medium' : 'text-gray-500'}>
                    {stats.quota.month_sent} / {stats.quota.monthly_limit}
                  </span>
                </div>
                <div className="h-3 bg-gray-100 rounded-full overflow-hidden">
                  <div 
                    className={`h-full rounded-full transition-all ${
                      stats.quota.monthly_percent >= 80 ? 'bg-red-500' : 
                      stats.quota.monthly_percent >= 50 ? 'bg-yellow-500' : 'bg-green-500'
                    }`}
                    style={{ width: `${Math.min(stats.quota.monthly_percent, 100)}%` }}
                  />
                </div>
                <p className="text-xs text-gray-400">
                  {stats.quota.monthly_remaining > 0 
                    ? `${stats.quota.monthly_remaining} messages remaining this month`
                    : '⚠️ Monthly limit reached'
                  }
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Quick Actions */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Zap className="h-5 w-5 text-yellow-500" />
              Quick Actions
            </CardTitle>
            <CardDescription>Common tasks to manage your bot</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3">
            <Link href="/dashboard/menus">
              <Button variant="outline" className="w-full justify-start">
                <Menu className="h-4 w-4 mr-2" /> Create New Menu
              </Button>
            </Link>
            <Link href="/dashboard/data">
              <Button variant="outline" className="w-full justify-start">
                <Database className="h-4 w-4 mr-2" /> Upload Dataset (CSV)
              </Button>
            </Link>
            <Link href="/dashboard/whatsapp">
              <Button variant="outline" className="w-full justify-start">
                <Smartphone className="h-4 w-4 mr-2" /> 
                {stats?.wa_connected ? 'View WhatsApp Status' : 'Connect WhatsApp'}
              </Button>
            </Link>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-blue-500" />
              System Overview
            </CardTitle>
            <CardDescription>Your bot infrastructure status</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="h-3 w-3 rounded-full bg-green-500 animate-pulse" />
                  <span className="font-medium">Bot Engine</span>
                </div>
                <span className="text-sm text-green-600">Online</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center gap-3">
                  <div className={`h-3 w-3 rounded-full ${stats?.wa_connected ? 'bg-green-500 animate-pulse' : 'bg-gray-300'}`} />
                  <span className="font-medium">WhatsApp Service</span>
                </div>
                <span className={`text-sm ${stats?.wa_connected ? 'text-green-600' : 'text-gray-400'}`}>
                  {stats?.wa_connected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="h-3 w-3 rounded-full bg-green-500 animate-pulse" />
                  <span className="font-medium">Database</span>
                </div>
                <span className="text-sm text-green-600">Healthy</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}


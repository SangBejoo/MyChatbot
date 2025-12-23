'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2, Smartphone, QrCode, LogOut, RefreshCw, CheckCircle, XCircle } from 'lucide-react';

interface WhatsAppStatus {
    connected: boolean;
    initialized: boolean;
    phone: string;
    name: string;
    hasQR: boolean;
}

export default function WhatsAppPage() {
    const [status, setStatus] = useState<WhatsAppStatus | null>(null);
    const [loading, setLoading] = useState(true);
    const [connecting, setConnecting] = useState(false);
    const [loggingOut, setLoggingOut] = useState(false);
    const [qrUrl, setQrUrl] = useState<string | null>(null);
    const [qrError, setQrError] = useState<string | null>(null);

    const fetchStatus = useCallback(async () => {
        try {
            const { data } = await api.get('/whatsapp/status');
            setStatus(data);
            return data;
        } catch (error) {
            console.error('Failed to fetch status', error);
            setStatus({ connected: false, initialized: false, phone: '', name: '', hasQR: false });
            return null;
        } finally {
            setLoading(false);
        }
    }, []);

    const fetchQR = useCallback(async () => {
        try {
            setQrError(null);
            const response = await api.get('/whatsapp/qr', { responseType: 'blob' });
            
            // Check if it's text (already logged in or error)
            if (response.data.type?.includes('text')) {
                const text = await response.data.text();
                if (text === 'Already logged in') {
                    setQrUrl(null);
                    fetchStatus();
                } else {
                    setQrError(text);
                }
                return;
            }
            
            // Create blob URL for image
            const url = URL.createObjectURL(response.data);
            setQrUrl(url);
        } catch (error: any) {
            // 202 Accepted means QR not ready yet
            if (error.response?.status === 202) {
                setQrError('Generating QR code... Please wait.');
            } else {
                setQrError('Failed to load QR code');
            }
        }
    }, [fetchStatus]);

    const handleConnect = async () => {
        setConnecting(true);
        try {
            await api.post('/whatsapp/connect');
            // Start polling for QR
            await fetchQR();
        } catch (error) {
            console.error('Failed to connect', error);
        } finally {
            setConnecting(false);
        }
    };

    const handleLogout = async () => {
        if (!confirm('Are you sure you want to disconnect WhatsApp?')) return;
        
        setLoggingOut(true);
        try {
            await api.post('/whatsapp/logout');
            setQrUrl(null);
            fetchStatus();
        } catch (error) {
            console.error('Failed to logout', error);
        } finally {
            setLoggingOut(false);
        }
    };

    // Initial load
    useEffect(() => {
        fetchStatus();
    }, [fetchStatus]);

    // Auto-refresh QR when not connected
    useEffect(() => {
        if (status && !status.connected && status.initialized) {
            fetchQR();
            const interval = setInterval(fetchQR, 5000); // Refresh QR every 5 seconds
            return () => clearInterval(interval);
        }
    }, [status, fetchQR]);

    // Poll status when waiting for connection
    useEffect(() => {
        if (status && !status.connected && status.initialized) {
            const interval = setInterval(fetchStatus, 3000);
            return () => clearInterval(interval);
        }
    }, [status, fetchStatus]);

    // Cleanup blob URLs
    useEffect(() => {
        return () => {
            if (qrUrl) URL.revokeObjectURL(qrUrl);
        };
    }, [qrUrl]);

    if (loading) {
        return (
            <div className="flex items-center justify-center h-[60vh]">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div>
                <h2 className="text-3xl font-bold tracking-tight">WhatsApp Integration</h2>
                <p className="text-gray-500">Connect your WhatsApp to receive and send messages</p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Status Card */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Smartphone className="h-5 w-5" />
                            Connection Status
                        </CardTitle>
                        <CardDescription>
                            Your WhatsApp business account connection
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            <div className="flex items-center gap-3">
                                {status?.connected ? (
                                    <>
                                        <div className="h-3 w-3 rounded-full bg-green-500 animate-pulse" />
                                        <span className="font-medium text-green-700">Connected</span>
                                    </>
                                ) : (
                                    <>
                                        <div className="h-3 w-3 rounded-full bg-gray-300" />
                                        <span className="text-gray-500">Not Connected</span>
                                    </>
                                )}
                            </div>

                            {status?.connected && (
                                <div className="bg-green-50 rounded-lg p-4 space-y-2">
                                    <div className="flex items-center gap-2">
                                        <CheckCircle className="h-5 w-5 text-green-600" />
                                        <span className="font-medium">{status.name || 'WhatsApp User'}</span>
                                    </div>
                                    <div className="text-sm text-green-700">
                                        Phone: <span className="font-mono">+{status.phone}</span>
                                    </div>
                                </div>
                            )}

                            {!status?.connected && (
                                <div className="bg-gray-50 rounded-lg p-4">
                                    <div className="flex items-center gap-2 text-gray-500">
                                        <XCircle className="h-5 w-5" />
                                        <span>No WhatsApp account linked</span>
                                    </div>
                                </div>
                            )}

                            <div className="flex gap-2">
                                {status?.connected ? (
                                    <Button 
                                        variant="destructive" 
                                        onClick={handleLogout}
                                        disabled={loggingOut}
                                    >
                                        {loggingOut ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <LogOut className="h-4 w-4 mr-2" />}
                                        Disconnect
                                    </Button>
                                ) : (
                                    <Button onClick={handleConnect} disabled={connecting}>
                                        {connecting ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <QrCode className="h-4 w-4 mr-2" />}
                                        Connect WhatsApp
                                    </Button>
                                )}
                                <Button variant="outline" onClick={fetchStatus}>
                                    <RefreshCw className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* QR Code Card */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <QrCode className="h-5 w-5" />
                            QR Code Login
                        </CardTitle>
                        <CardDescription>
                            Scan with WhatsApp &rarr; Linked Devices &rarr; Link a Device
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="flex flex-col items-center justify-center min-h-[300px]">
                            {status?.connected ? (
                                <div className="text-center space-y-2">
                                    <CheckCircle className="h-16 w-16 text-green-500 mx-auto" />
                                    <p className="font-medium text-green-700">Already Connected!</p>
                                    <p className="text-sm text-gray-500">Your WhatsApp is linked to this account</p>
                                </div>
                            ) : qrUrl ? (
                                <div className="space-y-4 text-center">
                                    <img 
                                        src={qrUrl} 
                                        alt="WhatsApp QR Code" 
                                        className="w-64 h-64 border rounded-lg shadow-lg"
                                    />
                                    <p className="text-sm text-gray-500">QR code refreshes automatically</p>
                                </div>
                            ) : qrError ? (
                                <div className="text-center space-y-3">
                                    <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
                                    <p className="text-gray-500">{qrError}</p>
                                    <Button variant="outline" size="sm" onClick={fetchQR}>
                                        <RefreshCw className="h-4 w-4 mr-2" />
                                        Retry
                                    </Button>
                                </div>
                            ) : (
                                <div className="text-center space-y-3">
                                    <QrCode className="h-16 w-16 text-gray-300 mx-auto" />
                                    <p className="text-gray-500">Click "Connect WhatsApp" to generate QR code</p>
                                </div>
                            )}
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Instructions */}
            <Card>
                <CardHeader>
                    <CardTitle>How to Connect</CardTitle>
                </CardHeader>
                <CardContent>
                    <ol className="list-decimal list-inside space-y-2 text-gray-600">
                        <li>Click the <strong>"Connect WhatsApp"</strong> button above</li>
                        <li>Open WhatsApp on your phone</li>
                        <li>Tap <strong>Menu (⋮)</strong> or <strong>Settings</strong> &rarr; <strong>Linked Devices</strong></li>
                        <li>Tap <strong>"Link a Device"</strong></li>
                        <li>Point your phone camera at the QR code above</li>
                        <li>Wait for the connection to be established</li>
                    </ol>
                </CardContent>
            </Card>
        </div>
    );
}

'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card';
import { Loader2, LogOut, RefreshCw, Smartphone } from 'lucide-react';

export default function ChatPage() {
  const [status, setStatus] = useState<'connected' | 'disconnected' | 'loading'>('loading');
  const [qrUrl, setQrUrl] = useState<string | null>(null);
  const [userInfo, setUserInfo] = useState<{ phone: string; name: string } | null>(null);

  const checkStatus = useCallback(async () => {
    try {
      const { data } = await api.get('/whatsapp/status');
      if (data.connected) {
        setStatus('connected');
        setUserInfo({ phone: data.phone, name: data.name });
      } else {
        setStatus('disconnected');
        setUserInfo(null);
      }
    } catch (error) {
      console.error('Failed to check status', error);
      setStatus('disconnected');
    }
  }, []);

  const fetchQR = useCallback(async () => {
    try {
      // First try to connect
      await api.post('/whatsapp/connect').catch(() => {});
      
      const response = await api.get('/whatsapp/qr', { responseType: 'blob' });
      
      // Check if it's text (already logged in or error)
      if (response.data.type?.includes('text')) {
        const text = await response.data.text();
        if (text === 'Already logged in') {
          checkStatus();
        }
        return;
      }
      
      // Create blob URL for image
      const url = URL.createObjectURL(response.data);
      setQrUrl((prev) => {
        if (prev) URL.revokeObjectURL(prev);
        return url;
      });
    } catch (error) {
      console.error('Failed to fetch QR', error);
    }
  }, [checkStatus]);

  useEffect(() => {
    checkStatus();
  }, [checkStatus]);

  useEffect(() => {
    if (status === 'disconnected') {
      fetchQR();
      const interval = setInterval(() => {
        fetchQR();
        checkStatus();
      }, 5000);
      return () => clearInterval(interval);
    }
  }, [status, fetchQR, checkStatus]);

  // Cleanup blob URLs
  useEffect(() => {
    return () => {
      if (qrUrl) URL.revokeObjectURL(qrUrl);
    };
  }, [qrUrl]);

  const handleLogout = async () => {
    try {
      await api.post('/whatsapp/logout');
      setStatus('disconnected');
      setUserInfo(null);
      setQrUrl(null);
      checkStatus();
    } catch (error) {
      alert('Failed to logout');
    }
  };

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col items-center justify-center p-4">
      <Card className="w-full max-w-md shadow-xl text-center">
        <CardHeader>
          <CardTitle className="flex items-center justify-center gap-2">
            <Smartphone className="h-6 w-6" /> WhatsApp Connection
          </CardTitle>
          <CardDescription>Scan to connect your bot</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-center justify-center min-h-[300px]">
          {status === 'loading' && <Loader2 className="h-10 w-10 animate-spin text-gray-400" />}
          
          {status === 'disconnected' && (
            <div className="space-y-4">
                <div className="border-4 border-gray-900 rounded-lg p-2 inline-block bg-white">
                    {qrUrl ? (
                      <img 
                          src={qrUrl} 
                          alt="WhatsApp QR Code" 
                          className="w-64 h-64 object-contain"
                      />
                    ) : (
                      <div className="w-64 h-64 flex items-center justify-center">
                        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
                      </div>
                    )}
                </div>
                <p className="text-sm text-gray-500">
                    Open WhatsApp on your phone <br/> Go to <b>Settings &gt; Linked Devices</b> <br/> and scan this code.
                </p>
                <div className="text-xs text-gray-400">Refreshes automatically...</div>
            </div>
          )}

          {status === 'connected' && (
             <div className="space-y-6">
                <div className="bg-green-100 text-green-700 p-6 rounded-full inline-block">
                    <Smartphone className="h-16 w-16" />
                </div>
                <div>
                    <h3 className="text-xl font-bold text-green-700">Connected!</h3>
                    <p className="text-gray-500">Your bot is active and listening.</p>
                </div>
                {userInfo && (
                  <div className="bg-gray-100 p-4 rounded-lg text-left">
                    <p className="text-sm text-gray-500 mb-1">Linked Account:</p>
                    <p className="font-bold text-gray-900">{userInfo.name || 'Unknown Name'}</p>
                    <p className="text-sm text-gray-600 font-mono">+{userInfo.phone}</p>
                  </div>
                )}
             </div>
          )}
        </CardContent>
        <CardFooter className="justify-center">
            {status === 'connected' && (
                <Button variant="destructive" onClick={handleLogout} className="w-full">
                    <LogOut className="mr-2 h-4 w-4" /> Disconnect Device
                </Button>
            )}
            {status === 'disconnected' && (
                 <Button variant="ghost" onClick={() => checkStatus()} size="sm">
                    <RefreshCw className="mr-2 h-3 w-3" /> Check Status
                 </Button>
            )}
        </CardFooter>
      </Card>
    </div>
  );
}


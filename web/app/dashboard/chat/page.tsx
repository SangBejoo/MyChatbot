'use client';

import { useEffect, useState,  useCallback } from 'react';
import api from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card';
import { Loader2, LogOut, RefreshCw, Smartphone } from 'lucide-react';

export default function ChatPage() {
  const [status, setStatus] = useState<'connected' | 'disconnected' | 'loading'>('loading');
  const [qrKey, setQrKey] = useState(0); // Used to force refresh QR image
  const [userInfo, setUserInfo] = useState<{ phone: string; name: string } | null>(null);

  const checkStatus = useCallback(async () => {
    try {
      const res = await fetch('http://localhost:8080/whatsapp/status');
      const data = await res.json();
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
  }, []); // Empty dependencies since it doesn't depend on props/state

  useEffect(() => {
    checkStatus();
    const interval = setInterval(() => {
        if (status !== 'connected') {
            checkStatus();
            setQrKey(prev => prev + 1); // Refresh QR every few seconds to prevent expiration awareness drift
        }
    }, 5000);
    return () => clearInterval(interval);
  }, [status, checkStatus]);

  const handleLogout = async () => {
    try {
      await fetch('http://localhost:8080/whatsapp/logout', { method: 'POST' });
      setStatus('disconnected');
      setUserInfo(null);
      checkStatus(); // Force check
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
                    {/* QR Code Image - Add timestamp to force reload */}
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img 
                        src={`http://localhost:8080/whatsapp/qr?t=${qrKey}`} 
                        alt="WhatsApp QR Code" 
                        className="w-64 h-64 object-contain"
                    />
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

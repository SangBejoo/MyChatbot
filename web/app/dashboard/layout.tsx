'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Sidebar } from '@/components/layout/sidebar';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [authorized, setAuthorized] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    // Simple client-side auth check
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }
    
    // Decode JWT to get role
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const userIsAdmin = payload.role === 'admin';
      setIsAdmin(userIsAdmin);
      
      // Admin can only access /dashboard/admin/*
      // Regular users cannot access /dashboard/admin/*
      if (userIsAdmin && !pathname.startsWith('/dashboard/admin')) {
        router.push('/dashboard/admin');
        return;
      }
      
      if (!userIsAdmin && pathname.startsWith('/dashboard/admin')) {
        router.push('/dashboard');
        return;
      }
      
      setAuthorized(true);
    } catch {
      router.push('/login');
    }
  }, [router, pathname]);

  if (!authorized) {
    return null; // Or a spinner
  }

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Sidebar */}
      <Sidebar isAdmin={isAdmin} />

      {/* Main Content */}
      <main className="flex-1 ml-64 p-6 h-screen overflow-hidden">
        <div className="h-full w-full overflow-auto">
          {children}
        </div>
      </main>
    </div>
  );
}


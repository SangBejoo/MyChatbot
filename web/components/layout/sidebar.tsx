'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';
import { LayoutDashboard, Settings, MessageSquare, LogOut, Database, Menu, Smartphone, Shield, Users } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface SidebarProps {
  isAdmin?: boolean;
}

export function Sidebar({ isAdmin = false }: SidebarProps) {
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = () => {
    localStorage.removeItem('token');
    router.push('/login');
  };

  // User menu items
  const userMenuItems = [
    { href: '/dashboard', label: 'Overview', icon: LayoutDashboard },
    { href: '/dashboard/chat', label: 'Chat', icon: MessageSquare },
    { href: '/dashboard/data', label: 'Datasets', icon: Database },
    { href: '/dashboard/menus', label: 'Menus', icon: Menu },
    { href: '/dashboard/whatsapp', label: 'WhatsApp', icon: Smartphone },
    { href: '/dashboard/settings', label: 'Settings', icon: Settings },
  ];

  // Admin menu items
  const adminMenuItems = [
    { href: '/dashboard/admin', label: 'Overview', icon: LayoutDashboard },
    { href: '/dashboard/admin', label: 'Users', icon: Users },
  ];

  const menuItems = isAdmin ? adminMenuItems : userMenuItems;

  return (
    <div className="flex bg-gray-900 text-white w-64 flex-col h-screen fixed">
      <div className="p-6">
        <h1 className="text-2xl font-bold tracking-wider flex items-center gap-2">
          {isAdmin ? (
            <>
              <Shield className="h-6 w-6 text-orange-400" />
              ADMIN
            </>
          ) : (
            'BOT DASH'
          )}
        </h1>
        {isAdmin && (
          <p className="text-xs text-gray-400 mt-1">System Administrator</p>
        )}
      </div>
      
      <nav className="flex-1 px-4 space-y-2">
        {menuItems.map((item, index) => {
          const Icon = item.icon;
          const isActive = pathname === item.href;
          return (
            <Link 
              key={`${item.href}-${index}`} 
              href={item.href}
              className={cn(
                "flex items-center gap-3 px-4 py-3 rounded-lg transition-colors",
                isActive ? "bg-blue-600 text-white" : "text-gray-400 hover:bg-gray-800 hover:text-white"
              )}
            >
              <Icon size={20} />
              <span className="font-medium">{item.label}</span>
            </Link>
          );
        })}
      </nav>

      <div className="p-4 border-t border-gray-800">
        <Button 
          variant="ghost" 
          className="w-full justify-start gap-3 text-red-400 hover:text-red-300 hover:bg-red-950/20"
          onClick={handleLogout}
        >
          <LogOut size={20} />
          <span>Logout</span>
        </Button>
      </div>
    </div>
  );
}


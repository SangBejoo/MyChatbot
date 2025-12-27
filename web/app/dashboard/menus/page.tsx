'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Menu, Plus, Trash2, Save, Loader2, GripVertical, Eye, EyeOff } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import TelegramPreview from '@/components/TelegramPreview';

interface MenuItem {
    label: string;
    action: string;
    payload: string;
}

interface MenuDTO {
    id: number;
    slug: string;
    title: string;
    items: MenuItem[]; // Parsed JSON
}

interface DynamicTable {
    table_name: string;
    display_name: string;
}

export default function MenuManagerPage() {
    const [menus, setMenus] = useState<MenuDTO[]>([]);
    const [tables, setTables] = useState<DynamicTable[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    
    // Create/Edit State
    const [editingMenu, setEditingMenu] = useState<MenuDTO | null>(null);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [showPreview, setShowPreview] = useState(true);
    
    // New Menu State
    const [newSlug, setNewSlug] = useState('');
    const [newTitle, setNewTitle] = useState('');

    const fetchMenus = async () => {
        setLoading(true);
        try {
            const { data } = await api.get('/menus');
            // Backend sends "items" as interface{}, need to ensure it's array
            const formatted = (data || []).map((m: any) => ({
                ...m,
                items: Array.isArray(m.items) ? m.items : [] 
            }));
            setMenus(formatted);
        } catch (error) {
            console.error('Failed to fetch menus', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchTables = async () => {
        try {
            const { data } = await api.get('/tables');
            setTables(data || []);
        } catch (error) {
            console.error('Failed to fetch tables', error);
        }
    };

    const handleCreate = async () => {
        setSaving(true);
        try {
            await api.post('/menus', {
                slug: newSlug,
                title: newTitle,
                items: [] // Start empty
            });
            fetchMenus();
            setNewTitle('');
            setNewSlug('');
            setIsCreateOpen(false);
            alert('Menu Created! Now select it to add buttons.');
        } catch (error) {
            alert('Failed to create menu');
        } finally {
            setSaving(false);
        }
    };

    const handleUpdate = async () => {
        if (!editingMenu) return;
        setSaving(true);
        try {
            await api.put(`/menus/${editingMenu.slug}`, {
                title: editingMenu.title,
                items: editingMenu.items
            });
            fetchMenus();
            alert('Menu Saved!');
        } catch (error) {
            alert('Failed to update menu');
        } finally {
            setSaving(false);
        }
    };

    const handleDelete = async (slug: string) => {
        if (!confirm('Are you sure?')) return;
        try {
            await api.delete(`/menus/${slug}`);
            fetchMenus();
            if (editingMenu?.slug === slug) setEditingMenu(null);
        } catch (error) {
            alert('Failed to delete');
        }
    };

    // Item Management
    const addItem = () => {
        if (!editingMenu) return;
        setEditingMenu({
            ...editingMenu,
            items: [...editingMenu.items, { label: 'New Button', action: 'reply', payload: '' }]
        });
    };

    const updateItem = (index: number, field: keyof MenuItem, value: string) => {
        if (!editingMenu) return;
        const newItems = [...editingMenu.items];
        newItems[index] = { ...newItems[index], [field]: value };
        
        // Reset payload if action changes to view_table or calculate_from_table to force selection
        if (field === 'action' && (value === 'view_table' || value === 'calculate_from_table')) {
             newItems[index].payload = tables.length > 0 ? tables[0].display_name : '';
        }
        
        setEditingMenu({ ...editingMenu, items: newItems });
    };

    const removeItem = (index: number) => {
        if (!editingMenu) return;
        const newItems = [...editingMenu.items];
        newItems.splice(index, 1);
        setEditingMenu({ ...editingMenu, items: newItems });
    };

    useEffect(() => {
        fetchMenus();
        fetchTables();
    }, []);

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Menu Manager</h2>
                    <p className="text-gray-500">Design your bot's buttons and interactions</p>
                </div>
                
                <div className="flex gap-2">
                    <Button variant="outline" onClick={() => setShowPreview(!showPreview)}>
                        {showPreview ? <EyeOff className="mr-2 h-4 w-4" /> : <Eye className="mr-2 h-4 w-4" />}
                        {showPreview ? 'Hide Preview' : 'Show Preview'}
                    </Button>
                <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                    <DialogTrigger asChild>
                        <Button>
                            <Plus className="mr-2 h-4 w-4" /> New Menu
                        </Button>
                    </DialogTrigger>
                    <DialogContent className="max-w-md">
                        <DialogHeader>
                            <DialogTitle>Create New Menu</DialogTitle>
                        </DialogHeader>
                        <div className="space-y-4">
                            <div>
                                <label className="text-sm font-medium">Menu ID (Slug)</label>
                                <Input 
                                    placeholder="main_menu" 
                                    value={newSlug}
                                    onChange={(e) => setNewSlug(e.target.value)}
                                />
                                <p className="text-xs text-gray-400 mt-1">Use 'main_menu' for the start menu.</p>
                            </div>
                            <div>
                                <label className="text-sm font-medium">Display Title</label>
                                <Input 
                                    placeholder="Main Menu" 
                                    value={newTitle}
                                    onChange={(e) => setNewTitle(e.target.value)}
                                />
                            </div>
                            <Button onClick={handleCreate} disabled={saving} className="w-full">
                                {saving ? <Loader2 className="animate-spin" /> : 'Create Menu'}
                            </Button>
                        </div>
                    </DialogContent>
                </Dialog>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-12 h-[calc(100vh-200px)]">
                {/* List - 4 Cols */}
                <div className="md:col-span-4 flex flex-col gap-4 overflow-y-auto pr-2">
                    {loading ? <Loader2 className="animate-spin mx-auto" /> : menus.map(menu => (
                        <Card 
                            key={menu.id} 
                            className={`cursor-pointer transition-all ${editingMenu?.id === menu.id ? 'border-primary bg-primary/5 shadow-md' : 'hover:bg-gray-50'}`}
                            onClick={() => setEditingMenu(menu)}
                        >
                            <CardHeader className="flex flex-row items-center justify-between p-4">
                                <div className="space-y-1">
                                    <CardTitle className="text-base font-semibold">{menu.title}</CardTitle>
                                    <div className="text-xs text-gray-400 font-mono bg-gray-100 px-2 py-0.5 rounded w-fit">{menu.slug}</div>
                                </div>
                                <Menu className={`h-5 w-5 ${editingMenu?.id === menu.id ? 'text-primary' : 'text-gray-300'}`} />
                            </CardHeader>
                            <CardContent className="p-4 pt-0">
                                <div className="text-xs text-gray-500">{menu.items.length} Buttons</div>
                            </CardContent>
                        </Card>
                    ))}
                </div>

                {/* Editor - Adjusts based on preview visibility */}
                <Card className={`${showPreview ? 'md:col-span-5' : 'md:col-span-8'} flex flex-col h-full overflow-hidden transition-all`}>
                    <CardHeader className="border-b bg-gray-50/50">
                        <div className="flex justify-between items-center">
                            <CardTitle>
                                {editingMenu ? `Editing: ${editingMenu.title}` : 'Select a menu to edit'}
                            </CardTitle>
                            {editingMenu && (
                                <div className="flex gap-2">
                                     <Button variant="destructive" size="sm" onClick={() => handleDelete(editingMenu.slug)}>
                                        <Trash2 className="h-4 w-4 mr-2" /> Delete
                                    </Button>
                                    <Button onClick={handleUpdate} disabled={saving}>
                                        <Save className="h-4 w-4 mr-2" /> Save Changes
                                    </Button>
                                </div>
                            )}
                        </div>
                    </CardHeader>
                    
                    <CardContent className="flex-1 overflow-y-auto p-6 bg-gray-50/30">
                        {!editingMenu ? (
                            <div className="flex flex-col items-center justify-center h-full text-gray-300 gap-4">
                                <Menu className="h-16 w-16" />
                                <p>Select a menu from the left to start building</p>
                            </div>
                        ) : (
                            <div className="space-y-6">
                                <div className="border p-4 rounded-lg bg-white shadow-sm">
                                    <label className="text-sm font-medium mb-1 block">Menu Title</label>
                                    <Input 
                                        value={editingMenu.title}
                                        onChange={(e) => setEditingMenu({...editingMenu, title: e.target.value})}
                                        className="max-w-md"
                                    />
                                </div>

                                <div className="space-y-4">
                                    <div className="flex items-center justify-between">
                                        <h3 className="font-medium flex items-center gap-2">
                                            Buttons 
                                            <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded-full">{editingMenu.items.length}</span>
                                        </h3>
                                        <Button size="sm" variant="outline" onClick={addItem}>
                                            <Plus className="h-4 w-4 mr-2" /> Add Button
                                        </Button>
                                    </div>

                                    {editingMenu.items.length === 0 && (
                                        <div className="text-center py-10 border-2 border-dashed rounded-lg text-gray-400">
                                            No buttons yet. Click "Add Button" to start.
                                        </div>
                                    )}

                                    {editingMenu.items.map((item, idx) => (
                                        <div key={idx} className="flex gap-4 items-start p-4 bg-white border rounded-lg shadow-sm group hover:border-primary/50 transition-colors">
                                            <div className="mt-3 text-gray-300 cursor-move">
                                                <GripVertical className="h-4 w-4" />
                                            </div>
                                            
                                            <div className="flex-1 grid grid-cols-1 md:grid-cols-2 gap-4">
                                                {/* Label */}
                                                <div>
                                                    <label className="text-xs font-medium text-gray-500 mb-1 block">Button Label</label>
                                                    <Input 
                                                        value={item.label}
                                                        onChange={(e) => updateItem(idx, 'label', e.target.value)}
                                                        placeholder="e.g. Check Stock"
                                                    />
                                                </div>

                                                {/* Action */}
                                                <div>
                                                    <label className="text-xs font-medium text-gray-500 mb-1 block">Action Type</label>
                                                    <div className="flex gap-2">
                                                        <Select 
                                                            value={['reply', 'view_table', 'calculate_from_table'].includes(item.action) ? item.action : '_custom'} 
                                                            onValueChange={(val) => {
                                                                if (val === '_custom') {
                                                                    updateItem(idx, 'action', '');
                                                                } else {
                                                                    updateItem(idx, 'action', val);
                                                                }
                                                            }}
                                                        >
                                                            <SelectTrigger className="flex-1">
                                                                <SelectValue />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                <SelectItem value="reply">💬 Send Reply</SelectItem>
                                                                <SelectItem value="view_table">📊 View Dataset</SelectItem>
                                                                <SelectItem value="calculate_from_table">🧮 Calculate from Dataset</SelectItem>
                                                                <SelectItem value="_custom">✏️ Custom...</SelectItem>
                                                            </SelectContent>
                                                        </Select>
                                                        {!['reply', 'view_table', 'calculate_from_table'].includes(item.action) && (
                                                            <Input 
                                                                value={item.action}
                                                                onChange={(e) => updateItem(idx, 'action', e.target.value)}
                                                                placeholder="custom_action"
                                                                className="flex-1"
                                                            />
                                                        )}
                                                    </div>
                                                </div>

                                                {/* Payload (with type toggle) */}
                                                <div className="md:col-span-2">
                                                    <div className="flex items-center justify-between mb-1">
                                                        <label className="text-xs font-medium text-gray-500">Payload / Response</label>
                                                        <div className="flex gap-1 text-xs">
                                                            <button
                                                                type="button"
                                                                onClick={() => updateItem(idx, 'payload', '')}
                                                                className={`px-2 py-0.5 rounded ${!item.payload.startsWith('$table:') ? 'bg-primary text-white' : 'bg-gray-100 text-gray-500'}`}
                                                            >
                                                                Text
                                                            </button>
                                                            <button
                                                                type="button"
                                                                onClick={() => updateItem(idx, 'payload', '$table:' + (tables[0]?.table_name || ''))}
                                                                className={`px-2 py-0.5 rounded ${item.payload.startsWith('$table:') ? 'bg-primary text-white' : 'bg-gray-100 text-gray-500'}`}
                                                            >
                                                                Dataset
                                                            </button>
                                                        </div>
                                                    </div>
                                                    
                                                    {item.payload.startsWith('$table:') ? (
                                                        <Select 
                                                            value={item.payload.replace('$table:', '')} 
                                                            onValueChange={(val) => updateItem(idx, 'payload', '$table:' + val)}
                                                        >
                                                            <SelectTrigger>
                                                                <SelectValue placeholder="Select a dataset..." />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                {tables.map(t => (
                                                                    <SelectItem key={t.table_name} value={t.table_name}>
                                                                        📊 {t.display_name}
                                                                    </SelectItem>
                                                                ))}
                                                                {tables.length === 0 && <SelectItem value="none" disabled>No datasets found</SelectItem>}
                                                            </SelectContent>
                                                        </Select>
                                                    ) : (
                                                        <Input 
                                                            value={item.payload}
                                                            onChange={(e) => updateItem(idx, 'payload', e.target.value)}
                                                            placeholder={item.action === 'reply' ? 'What should the bot say?' : 'Payload value...'}
                                                        />
                                                    )}
                                                </div>
                                            </div>

                                            <Button 
                                                variant="ghost" 
                                                size="icon" 
                                                className="text-gray-400 hover:text-red-500 mt-1"
                                                onClick={() => removeItem(idx)}
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </CardContent>
                </Card>

                {/* Telegram Preview - 3 Cols */}
                {showPreview && (
                    <div className="md:col-span-3 flex flex-col gap-4">
                        <div className="text-sm font-medium text-gray-600 flex items-center gap-2">
                            <Eye className="h-4 w-4" /> Live Preview
                        </div>
                        <TelegramPreview 
                            menuTitle={editingMenu?.title || 'Welcome! 👋'}
                            items={editingMenu?.items || []}
                        />
                    </div>
                )}
            </div>
        </div>
    );
}

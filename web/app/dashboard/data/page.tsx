'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Database, FileUp, Loader2, RefreshCw, Trash2, Search, Edit2, Save, X } from 'lucide-react';
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area"

interface DynamicTable {
    id: number;
    table_name: string;
    display_name: string;
    created_at: string;
}

export default function DataManagerPage() {
    const [tables, setTables] = useState<DynamicTable[]>([]);
    const [loading, setLoading] = useState(true);
    const [tablesLoading, setTablesLoading] = useState(false);
    const [uploading, setUploading] = useState(false);
    
    // Selection State
    const [selectedTable, setSelectedTable] = useState<DynamicTable | null>(null);
    const [tableData, setTableData] = useState<any[]>([]);
    const [dataLoading, setDataLoading] = useState(false);

    // Upload State
    const [displayName, setDisplayName] = useState('');
    const [file, setFile] = useState<File | null>(null);
    const [isUploadOpen, setIsUploadOpen] = useState(false);

    // Edit State
    const [editingRowId, setEditingRowId] = useState<number | null>(null);
    const [editedData, setEditedData] = useState<Record<string, any>>({});

    const fetchTables = async () => {
        setTablesLoading(true);
        try {
            const { data } = await api.get('/tables');
            setTables(data || []);
        } catch (error) {
            console.error('Failed to fetch tables', error);
        } finally {
            setTablesLoading(false);
            setLoading(false);
        }
    };

    const fetchTableData = async (table: DynamicTable) => {
        setDataLoading(true);
        setSelectedTable(table);
        setEditingRowId(null);
        try {
            const { data } = await api.get(`/tables/${table.table_name}/data`);
            setTableData(data || []);
        } catch (error) {
            console.error('Failed to fetch table data', error);
        } finally {
            setDataLoading(false);
        }
    };

    const handleUpload = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!file || !displayName) return;

        setUploading(true);
        const formData = new FormData();
        formData.append('display_name', displayName);
        formData.append('file', file);

        try {
            await api.post('/tables/import', formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            fetchTables();
            setDisplayName('');
            setFile(null);
            setIsUploadOpen(false);
        } catch (error) {
            alert('Import Failed');
            console.error(error);
        } finally {
            setUploading(false);
        }
    };

    const handleDeleteTable = async () => {
        if (!selectedTable) return;
        if (!confirm(`Delete "${selectedTable.display_name}" and all its data?`)) return;

        try {
            await api.delete(`/tables/${selectedTable.table_name}`);
            setSelectedTable(null);
            setTableData([]);
            fetchTables();
        } catch (error) {
            alert('Failed to delete table');
            console.error(error);
        }
    };

    const startEdit = (row: any) => {
        setEditingRowId(row.id);
        setEditedData({ ...row });
    };

    const cancelEdit = () => {
        setEditingRowId(null);
        setEditedData({});
    };

    const saveEdit = async () => {
        if (!selectedTable || editingRowId === null) return;
        
        try {
            await api.put(`/tables/${selectedTable.table_name}/row`, {
                row_id: editingRowId,
                data: editedData
            });
            // Refresh data
            fetchTableData(selectedTable);
            setEditingRowId(null);
        } catch (error) {
            alert('Failed to save changes');
            console.error(error);
        }
    };

    const handleDeleteRow = async (rowId: number) => {
        if (!selectedTable) return;
        if (!confirm('Delete this row?')) return;

        try {
            await api.delete(`/tables/${selectedTable.table_name}/row`, {
                data: { row_id: rowId }
            });
            fetchTableData(selectedTable);
        } catch (error) {
            alert('Failed to delete row');
            console.error(error);
        }
    };

    useEffect(() => {
        fetchTables();
    }, []);

    return (
        <div className="h-full flex flex-col gap-4">
            <div className="flex items-center justify-between shrink-0">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Data Manager</h2>
                    <p className="text-gray-500">Import CSVs and manage your dynamic datasets</p>
                </div>
                
                <Dialog open={isUploadOpen} onOpenChange={setIsUploadOpen}>
                    <DialogTrigger asChild>
                        <Button>
                            <FileUp className="mr-2 h-4 w-4" /> Import CSV
                        </Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Import New Data Table</DialogTitle>
                        </DialogHeader>
                        <form onSubmit={handleUpload} className="space-y-4">
                            <div>
                                <label className="text-sm font-medium">Dataset Name</label>
                                <Input 
                                    placeholder="e.g. Warehouse Inventory" 
                                    value={displayName}
                                    onChange={(e) => setDisplayName(e.target.value)}
                                    required
                                />
                            </div>
                            <div>
                                <label className="text-sm font-medium">CSV File</label>
                                <Input 
                                    type="file" 
                                    accept=".csv"
                                    onChange={(e) => setFile(e.target.files?.[0] || null)}
                                    required
                                />
                            </div>
                            <Button type="submit" disabled={uploading} className="w-full">
                                {uploading ? 'Uploading...' : 'Start Import'}
                            </Button>
                        </form>
                    </DialogContent>
                </Dialog>
            </div>

            {/* Top Section: Dataset Selector */}
            <Card className="shrink-0 bg-gray-50/50">
                <CardHeader className="py-3 px-4 flex flex-row items-center justify-between space-y-0">
                    <CardTitle className="text-sm font-medium flex items-center gap-2">
                        <Database className="h-4 w-4" /> 
                        Your Datasets 
                        <span className="text-gray-400 font-normal">({tables.length})</span>
                    </CardTitle>
                    <Button variant="ghost" size="sm" onClick={fetchTables} disabled={tablesLoading}>
                        <RefreshCw className={`h-4 w-4 ${tablesLoading ? 'animate-spin' : ''}`} />
                    </Button>
                </CardHeader>
                <CardContent className="py-2 px-2">
                     {tables.length === 0 ? (
                        <div className="p-4 text-center text-sm text-gray-500">No datasets uploaded yet. Import a CSV to get started.</div>
                     ) : (
                        <ScrollArea className="w-full whitespace-nowrap pb-2">
                            <div className="flex w-max space-x-2 p-2">
                                {tables.map((t) => (
                                    <button
                                        key={t.id}
                                        onClick={() => fetchTableData(t)}
                                        className={`inline-flex items-center rounded-full border px-3 py-1 text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 ${
                                            selectedTable?.id === t.id 
                                            ? 'border-transparent bg-primary text-primary-foreground shadow' 
                                            : 'border-transparent bg-white hover:bg-gray-100 text-gray-900 shadow-sm'
                                        }`}
                                    >
                                        <Database className="mr-2 h-3.5 w-3.5" />
                                        {t.display_name}
                                    </button>
                                ))}
                            </div>
                            <ScrollBar orientation="horizontal" />
                        </ScrollArea>
                     )}
                </CardContent>
            </Card>

            {/* Bottom Section: Excel-like Full Width Preview */}
            <Card className="flex-1 min-h-0 overflow-hidden flex flex-col shadow-md border-t-4 border-t-primary/20">
                <CardHeader className="py-3 px-4 border-b bg-white shrink-0">
                    <div className="flex items-center justify-between">
                        <CardTitle className="flex items-center gap-2">
                            {selectedTable ? (
                                <>
                                    <span className="text-primary">{selectedTable.display_name}</span>
                                    <span className="text-xs text-gray-400 font-normal font-mono bg-gray-100 px-2 rounded">
                                        {selectedTable.table_name}
                                    </span>
                                </>
                            ) : (
                                <span className="text-gray-400">Select a dataset above to preview</span>
                            )}
                        </CardTitle>
                        {selectedTable && (
                            <div className="flex items-center gap-3">
                                <span className="text-sm text-gray-500">{tableData.length} Rows</span>
                                <Button variant="destructive" size="sm" onClick={handleDeleteTable}>
                                    <Trash2 className="h-4 w-4 mr-1" /> Delete Table
                                </Button>
                            </div>
                        )}
                    </div>
                </CardHeader>
                
                {dataLoading && (
                    <div className="flex-1 flex items-center justify-center bg-white/50">
                        <Loader2 className="h-8 w-8 animate-spin text-primary" />
                    </div>
                )}
                
                {!dataLoading && !selectedTable && (
                    <div className="flex-1 flex flex-col items-center justify-center text-gray-300">
                        <Database className="h-16 w-16 mb-4 opacity-20" />
                        <p>No dataset selected</p>
                    </div>
                )}

                {!dataLoading && selectedTable && (
                    <div className="flex-1 w-full overflow-auto relative bg-white">
                        <table className="w-full min-w-max caption-bottom text-sm text-left border-collapse">
                            <thead className="sticky top-0 z-20 bg-gray-100 shadow-sm">
                                <tr>
                                    <th className="h-10 px-2 text-center align-middle font-medium text-muted-foreground bg-gray-50 border-b border-r whitespace-nowrap w-20">
                                        Actions
                                    </th>
                                    {tableData.length > 0 && Object.keys(tableData[0]).filter(k => k !== 'id').map((key) => (
                                        <th key={key} className="h-10 px-4 text-left align-middle font-medium text-muted-foreground bg-gray-50 border-b border-r last:border-r-0 whitespace-nowrap">
                                            {key}
                                        </th>
                                    ))}
                                </tr>
                            </thead>
                            <tbody>
                                {tableData.map((row) => (
                                    <tr key={row.id} className="hover:bg-blue-50/50 border-b transition-colors">
                                        <td className="p-1 text-center border-r">
                                            {editingRowId === row.id ? (
                                                <div className="flex gap-1 justify-center">
                                                    <Button size="icon" variant="ghost" className="h-7 w-7 text-green-600" onClick={saveEdit}>
                                                        <Save className="h-3.5 w-3.5" />
                                                    </Button>
                                                    <Button size="icon" variant="ghost" className="h-7 w-7 text-gray-500" onClick={cancelEdit}>
                                                        <X className="h-3.5 w-3.5" />
                                                    </Button>
                                                </div>
                                            ) : (
                                                <div className="flex gap-1 justify-center">
                                                    <Button size="icon" variant="ghost" className="h-7 w-7 text-blue-500" onClick={() => startEdit(row)}>
                                                        <Edit2 className="h-3.5 w-3.5" />
                                                    </Button>
                                                    <Button size="icon" variant="ghost" className="h-7 w-7 text-red-500" onClick={() => handleDeleteRow(row.id)}>
                                                        <Trash2 className="h-3.5 w-3.5" />
                                                    </Button>
                                                </div>
                                            )}
                                        </td>
                                        {Object.entries(row).filter(([k]) => k !== 'id').map(([key, val]: [string, any]) => (
                                            <td key={key} className="p-2 align-middle border-r last:border-r-0 whitespace-nowrap font-mono text-xs max-w-[300px] overflow-hidden text-ellipsis">
                                                {editingRowId === row.id ? (
                                                    <Input 
                                                        value={editedData[key] ?? ''} 
                                                        onChange={(e) => setEditedData({...editedData, [key]: e.target.value})}
                                                        className="h-7 text-xs"
                                                    />
                                                ) : (
                                                    <CellViewer value={val} />
                                                )}
                                            </td>
                                        ))}
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </Card>
        </div>
    );
}

function CellViewer({ value }: { value: any }) {
    const [open, setOpen] = useState(false);
    const content = String(value ?? '');
    const isLong = content.length > 50;

    if (!isLong) return <span>{content}</span>;

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <div className="cursor-pointer group flex items-center justify-between gap-2 hover:bg-gray-100 p-1 rounded -m-1">
                    <span className="truncate">{content.substring(0, 50)}...</span>
                    <Search className="h-3 w-3 text-gray-300 group-hover:text-blue-500 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
            </DialogTrigger>
            <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Cell Content</DialogTitle>
                </DialogHeader>
                <div className="p-4 bg-gray-50 rounded-lg border font-mono text-sm whitespace-pre-wrap break-words">
                    {content}
                </div>
            </DialogContent>
        </Dialog>
    );
}

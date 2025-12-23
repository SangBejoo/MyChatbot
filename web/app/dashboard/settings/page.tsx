'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import { Loader2, Save, Bot, MessageSquare, Sparkles } from 'lucide-react';

interface BotConfig {
    key: string;
    value: string;
    updated_at: string;
}

export default function SettingsPage() {
    const [configs, setConfigs] = useState<BotConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    
    // Local state for edits
    const [welcomeMsg, setWelcomeMsg] = useState('');
    const [defaultReply, setDefaultReply] = useState('');
    const [aiPrompt, setAiPrompt] = useState('');

    useEffect(() => {
        fetchConfigs();
    }, []);

    const fetchConfigs = async () => {
        setLoading(true);
        try {
            const { data } = await api.get('/config');
            if (data) {
                setConfigs(data);
                const w = data.find((c: BotConfig) => c.key === 'welcome_message');
                const d = data.find((c: BotConfig) => c.key === 'default_reply');
                const a = data.find((c: BotConfig) => c.key === 'ai_system_prompt');
                if (w) setWelcomeMsg(w.value);
                if (d) setDefaultReply(d.value);
                if (a) setAiPrompt(a.value);
            }
        } catch (error) {
            console.error('Failed to fetch configs', error);
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            await api.post('/config', { key: 'welcome_message', value: welcomeMsg });
            await api.post('/config', { key: 'default_reply', value: defaultReply });
            await api.post('/config', { key: 'ai_system_prompt', value: aiPrompt });
            alert('Settings saved!');
        } catch (error) {
            alert('Failed to save settings');
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return <div className="flex h-full items-center justify-center"><Loader2 className="animate-spin" /></div>;
    }

    return (
        <div className="space-y-6 max-w-3xl">
            <h2 className="text-3xl font-bold tracking-tight">Bot Settings</h2>
            
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2"><MessageSquare className="h-5 w-5" /> Messages</CardTitle>
                    <CardDescription>Configure default bot messages</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="welcome">Welcome Message</Label>
                        <Input 
                            id="welcome" 
                            value={welcomeMsg}
                            onChange={(e) => setWelcomeMsg(e.target.value)}
                            placeholder="Welcome! How can I help you?"
                        />
                        <p className="text-xs text-gray-500">Sent when a user starts the chat.</p>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="default">Default Reply (Fallback)</Label>
                        <Input 
                            id="default" 
                            value={defaultReply}
                            onChange={(e) => setDefaultReply(e.target.value)}
                            placeholder="I did not understand that."
                        />
                        <p className="text-xs text-gray-500">Sent when the bot matches no intent.</p>
                    </div>
                </CardContent>
            </Card>

            <Card className="border-blue-200 bg-blue-50/30">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2"><Sparkles className="h-5 w-5 text-blue-600" /> AI Behavior</CardTitle>
                    <CardDescription>Configure how the AI assistant responds</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="aiPrompt">System Prompt</Label>
                        <textarea 
                            id="aiPrompt" 
                            value={aiPrompt}
                            onChange={(e) => setAiPrompt(e.target.value)}
                            placeholder={`You are a helpful assistant for [Your Business].\n\nRules:\n- Answer ONLY based on the provided context/database.\n- If the question is outside your knowledge, politely say "I can only help with questions about [topic]."\n- Keep responses short and professional.\n- Never make up information.`}
                            className="w-full min-h-[200px] p-3 text-sm border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-y font-mono bg-white"
                        />
                        <p className="text-xs text-gray-500">
                            This prompt controls the AI's personality and rules. Use placeholders like <code className="bg-gray-200 px-1 rounded">{'{{context}}'}</code> for dynamic data.
                        </p>
                    </div>
                    
                    <div className="bg-blue-100 border border-blue-200 p-3 rounded-lg text-sm">
                        <strong className="text-blue-800">💡 Tips for preventing hallucination:</strong>
                        <ul className="list-disc list-inside mt-2 text-blue-700 space-y-1">
                            <li>Include "Only answer based on provided context"</li>
                            <li>Add "If unsure, say 'I don't have that information'"</li>
                            <li>Specify "Do not make up prices, dates, or facts"</li>
                        </ul>
                    </div>
                </CardContent>
            </Card>

            <Button onClick={handleSave} disabled={saving} size="lg" className="w-full">
                {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                <Save className="mr-2 h-4 w-4" />
                Save All Settings
            </Button>
        </div>
    );
}

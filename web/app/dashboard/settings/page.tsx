'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import { Loader2, Save, MessageSquare, Sparkles, Bot, CheckCircle, XCircle, Eye, EyeOff } from 'lucide-react';

interface BotConfig {
    key: string;
    value: string;
    updated_at: string;
}

interface TelegramStatus {
    has_token: boolean;
    connected: boolean;
    bot_name: string;
}

export default function SettingsPage() {
    const [configs, setConfigs] = useState<BotConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    
    // Bot message settings
    const [welcomeMsg, setWelcomeMsg] = useState('');
    const [defaultReply, setDefaultReply] = useState('');
    const [aiPrompt, setAiPrompt] = useState('');
    
    // Telegram settings
    const [tgStatus, setTgStatus] = useState<TelegramStatus | null>(null);
    const [tgToken, setTgToken] = useState('');
    const [tgLoading, setTgLoading] = useState(false);
    const [showToken, setShowToken] = useState(false);
    const [tgValidating, setTgValidating] = useState(false);
    const [validatedBotName, setValidatedBotName] = useState('');

    useEffect(() => {
        fetchConfigs();
        fetchTelegramStatus();
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

    const fetchTelegramStatus = async () => {
        try {
            const { data } = await api.get('/telegram/status');
            setTgStatus(data);
        } catch (error) {
            console.error('Failed to fetch Telegram status', error);
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

    const validateToken = async () => {
        if (!tgToken.trim()) return;
        setTgValidating(true);
        setValidatedBotName('');
        try {
            const { data } = await api.post('/telegram/validate', { token: tgToken });
            if (data.valid) {
                setValidatedBotName(data.bot_name);
            } else {
                alert('Invalid token: ' + data.error);
            }
        } catch (error: any) {
            alert('Validation failed: ' + (error.response?.data?.error || 'Unknown error'));
        } finally {
            setTgValidating(false);
        }
    };

    const saveTelegramToken = async () => {
        setTgLoading(true);
        try {
            await api.post('/telegram/token', { token: tgToken });
            setTgToken('');
            setValidatedBotName('');
            await fetchTelegramStatus();
            alert('Telegram token saved!');
        } catch (error: any) {
            alert('Failed to save token: ' + (error.response?.data?.error || 'Unknown error'));
        } finally {
            setTgLoading(false);
        }
    };

    const connectTelegram = async () => {
        setTgLoading(true);
        try {
            const { data } = await api.post('/telegram/connect');
            await fetchTelegramStatus();
            alert(`Bot connected: ${data.bot_name}`);
        } catch (error: any) {
            alert('Failed to connect: ' + (error.response?.data?.error || 'Unknown error'));
        } finally {
            setTgLoading(false);
        }
    };

    const disconnectTelegram = async () => {
        setTgLoading(true);
        try {
            await api.post('/telegram/disconnect');
            await fetchTelegramStatus();
        } catch (error) {
            alert('Failed to disconnect');
        } finally {
            setTgLoading(false);
        }
    };

    if (loading) {
        return <div className="flex h-full items-center justify-center"><Loader2 className="animate-spin" /></div>;
    }

    return (
        <div className="space-y-6 max-w-3xl">
            <h2 className="text-3xl font-bold tracking-tight">Bot Settings</h2>
            
            {/* Telegram Bot Configuration */}
            <Card className="border-purple-200 bg-gradient-to-br from-purple-50/50 to-blue-50/50">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Bot className="h-5 w-5 text-purple-600" /> Telegram Bot
                    </CardTitle>
                    <CardDescription>Configure your own Telegram bot for receiving messages</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    {/* Status */}
                    <div className="flex items-center gap-3 p-3 bg-white rounded-lg border">
                        {tgStatus?.connected ? (
                            <>
                                <CheckCircle className="h-5 w-5 text-green-500" />
                                <div>
                                    <p className="font-medium text-green-700">Connected</p>
                                    <p className="text-sm text-gray-500">@{tgStatus.bot_name}</p>
                                </div>
                                <Button 
                                    variant="outline" 
                                    size="sm" 
                                    className="ml-auto"
                                    onClick={disconnectTelegram}
                                    disabled={tgLoading}
                                >
                                    {tgLoading && <Loader2 className="mr-2 h-3 w-3 animate-spin" />}
                                    Disconnect
                                </Button>
                            </>
                        ) : tgStatus?.has_token ? (
                            <>
                                <XCircle className="h-5 w-5 text-yellow-500" />
                                <div>
                                    <p className="font-medium text-yellow-700">Token saved but not connected</p>
                                    <p className="text-sm text-gray-500">Click connect to start your bot</p>
                                </div>
                                <Button 
                                    size="sm" 
                                    className="ml-auto bg-purple-600 hover:bg-purple-700"
                                    onClick={connectTelegram}
                                    disabled={tgLoading}
                                >
                                    {tgLoading && <Loader2 className="mr-2 h-3 w-3 animate-spin" />}
                                    Connect
                                </Button>
                            </>
                        ) : (
                            <>
                                <XCircle className="h-5 w-5 text-gray-400" />
                                <div>
                                    <p className="font-medium text-gray-700">No token configured</p>
                                    <p className="text-sm text-gray-500">Get a token from @BotFather on Telegram</p>
                                </div>
                            </>
                        )}
                    </div>

                    {/* Token Input */}
                    <div className="space-y-2">
                        <Label htmlFor="tgToken">Bot Token</Label>
                        <div className="flex gap-2">
                            <div className="relative flex-1">
                                <Input 
                                    id="tgToken"
                                    type={showToken ? "text" : "password"}
                                    value={tgToken}
                                    onChange={(e) => { setTgToken(e.target.value); setValidatedBotName(''); }}
                                    placeholder="123456789:ABCdefGHIjklmNOPQRstuvWXYz"
                                    className="pr-10"
                                />
                                <button 
                                    type="button"
                                    onClick={() => setShowToken(!showToken)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                                >
                                    {showToken ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </button>
                            </div>
                            <Button 
                                variant="outline" 
                                onClick={validateToken}
                                disabled={tgValidating || !tgToken.trim()}
                            >
                                {tgValidating && <Loader2 className="mr-2 h-3 w-3 animate-spin" />}
                                Validate
                            </Button>
                        </div>
                        {validatedBotName && (
                            <p className="text-sm text-green-600 flex items-center gap-1">
                                <CheckCircle className="h-4 w-4" /> Valid token for {validatedBotName}
                            </p>
                        )}
                        <p className="text-xs text-gray-500">
                            Get your token from <a href="https://t.me/BotFather" target="_blank" className="text-blue-600 hover:underline">@BotFather</a> on Telegram.
                        </p>
                    </div>

                    {/* Save Token Button */}
                    {tgToken && validatedBotName && (
                        <Button 
                            onClick={saveTelegramToken}
                            disabled={tgLoading}
                            className="w-full bg-purple-600 hover:bg-purple-700"
                        >
                            {tgLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            <Save className="mr-2 h-4 w-4" />
                            Save Token & Connect
                        </Button>
                    )}
                </CardContent>
            </Card>
            
            {/* Messages Card */}
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

            {/* AI Behavior Card */}
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
                            className="w-full min-h-[150px] p-3 text-sm border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-y font-mono bg-white"
                        />
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


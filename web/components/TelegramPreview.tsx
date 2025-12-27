'use client';

import { useState } from 'react';
import { MenuItem } from './types';

interface TelegramPreviewProps {
    menuTitle: string;
    items: MenuItem[];
    onButtonClick?: (item: MenuItem) => void;
}

interface ChatMessage {
    type: 'bot' | 'user';
    content: string;
    buttons?: MenuItem[];
}

export default function TelegramPreview({ menuTitle, items, onButtonClick }: TelegramPreviewProps) {
    const [conversation, setConversation] = useState<ChatMessage[]>([]);
    const [selectedButton, setSelectedButton] = useState<MenuItem | null>(null);

    const handleButtonClick = (item: MenuItem) => {
        // Add user click as a message
        setSelectedButton(item);
        
        // Create bot response based on action type
        let botResponse = '';
        switch (item.action) {
            case 'reply':
                botResponse = item.payload || 'No response configured';
                break;
            case 'view_table':
                botResponse = `📊 Fetching data from ${item.payload}...`;
                break;
            default:
                botResponse = `Action: ${item.action}\nPayload: ${item.payload || 'none'}`;
        }

        setConversation([
            { type: 'user', content: item.label },
            { type: 'bot', content: botResponse }
        ]);

        onButtonClick?.(item);
    };

    const resetPreview = () => {
        setConversation([]);
        setSelectedButton(null);
    };

    const timeNow = new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false });

    return (
        <div className="w-full max-w-md mx-auto bg-[#0e1621] rounded-2xl overflow-hidden shadow-2xl border border-[#1e2c3a]">
            {/* Header */}
            <div className="bg-[#17212b] px-4 py-3 flex items-center gap-3 border-b border-[#1e2c3a]">
                <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                    <span className="text-white text-lg">🤖</span>
                </div>
                <div className="flex-1">
                    <div className="text-white font-medium text-sm">Your Bot</div>
                    <div className="text-[#6c7883] text-xs">online</div>
                </div>
                {conversation.length > 0 && (
                    <button 
                        onClick={resetPreview}
                        className="text-[#6c7883] text-xs hover:text-white transition-colors"
                    >
                        ↻ Reset
                    </button>
                )}
            </div>

            {/* Chat Area */}
            <div className="p-4 min-h-[350px] max-h-[400px] overflow-y-auto space-y-3 bg-[#0e1621]">
                {/* User /start command */}
                <div className="flex justify-end">
                    <div className="bg-[#2b5278] text-white rounded-2xl rounded-tr-sm px-4 py-2 max-w-[80%] shadow">
                        <p className="text-sm">/start</p>
                        <span className="text-[10px] text-[#8babc7] float-right mt-1">{timeNow} ✓✓</span>
                    </div>
                </div>

                {/* Bot Welcome Message with Menu */}
                <div className="flex justify-start">
                    <div className="bg-[#182533] text-[#f5f5f5] rounded-2xl rounded-tl-sm px-4 py-2 max-w-[85%] shadow">
                        <p className="text-sm font-medium">{menuTitle || 'Welcome! 👋'}</p>
                        <p className="text-sm mt-1 text-[#8b9ba7]">Choose an option:</p>
                        <span className="text-[10px] text-[#6c7883] float-right mt-1">{timeNow}</span>
                    </div>
                </div>

                {/* Dynamic Menu Buttons */}
                {items.length > 0 ? (
                    <div className="flex flex-wrap gap-2 pl-2">
                        {items.map((item, idx) => (
                            <button
                                key={idx}
                                onClick={() => handleButtonClick(item)}
                                className={`text-sm px-4 py-2 rounded-lg transition-colors duration-200 shadow-md ${
                                    selectedButton?.label === item.label 
                                        ? 'bg-[#3a6a94] text-white ring-2 ring-blue-400' 
                                        : 'bg-[#2b5278] hover:bg-[#3a6a94] text-white'
                                }`}
                            >
                                {item.label}
                            </button>
                        ))}
                    </div>
                ) : (
                    <div className="text-center py-4 text-[#6c7883] text-sm">
                        No buttons configured yet.<br/>
                        Add buttons to see them here.
                    </div>
                )}

                {/* Dynamic Conversation based on button clicks */}
                {conversation.map((msg, idx) => (
                    <div key={idx} className={`flex ${msg.type === 'user' ? 'justify-end' : 'justify-start'}`}>
                        <div className={`rounded-2xl px-4 py-2 max-w-[80%] shadow ${
                            msg.type === 'user' 
                                ? 'bg-[#2b5278] text-white rounded-tr-sm' 
                                : 'bg-[#182533] text-[#f5f5f5] rounded-tl-sm'
                        }`}>
                            <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                            <span className={`text-[10px] float-right mt-1 ${
                                msg.type === 'user' ? 'text-[#8babc7]' : 'text-[#6c7883]'
                            }`}>
                                {timeNow} {msg.type === 'user' && '✓✓'}
                            </span>
                        </div>
                    </div>
                ))}
            </div>

            {/* Input Area */}
            <div className="bg-[#17212b] px-4 py-3 flex items-center gap-3 border-t border-[#1e2c3a]">
                <div className="flex-1 bg-[#242f3d] rounded-full px-4 py-2">
                    <input 
                        type="text" 
                        placeholder="Type a message..." 
                        className="bg-transparent text-[#f5f5f5] text-sm w-full outline-none placeholder-[#6c7883]"
                        disabled
                    />
                </div>
                <button className="w-10 h-10 rounded-full bg-[#2b5278] flex items-center justify-center hover:bg-[#3a6a94] transition-colors">
                    <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                        <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
                    </svg>
                </button>
            </div>
        </div>
    );
}

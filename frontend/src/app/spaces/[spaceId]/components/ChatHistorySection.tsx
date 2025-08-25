'use client';

import { MessageCircle, GitBranch, Clock, User, Bot } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ChatHistoryProps {
  spaceId: string;
  className?: string;
}

interface ChatHistoryItem {
  id: string;
  title: string;
  lastMessage: string;
  messageCount: number;
  timestamp: Date;
  type: 'conversation' | 'branch';
  participants: ('user' | 'assistant')[];
  level: number;
  isActive?: boolean;
}

export default function ChatHistorySection({ spaceId, className }: ChatHistoryProps) {
  // Mock data - replace with actual data when available
  const mockChatHistory: ChatHistoryItem[] = [
    {
      id: '1',
      title: 'Document Analysis',
      lastMessage: 'Could you summarize the key points?',
      messageCount: 12,
      timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000),
      type: 'conversation',
      participants: ['user', 'assistant'],
      level: 0,
      isActive: true
    },
    {
      id: '2',
      title: 'Follow-up Questions',
      lastMessage: 'What about the methodology section?',
      messageCount: 5,
      timestamp: new Date(Date.now() - 4 * 60 * 60 * 1000),
      type: 'branch',
      participants: ['user', 'assistant'],
      level: 1
    },
    {
      id: '3',
      title: 'Research Insights',
      lastMessage: 'Based on the data patterns...',
      messageCount: 8,
      timestamp: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000),
      type: 'conversation',
      participants: ['user', 'assistant'],
      level: 0
    }
  ];

  const formatTimeAgo = (date: Date) => {
    const now = new Date();
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60));
    
    if (diffInMinutes < 60) {
      return `${diffInMinutes}m ago`;
    } else if (diffInMinutes < 1440) {
      return `${Math.floor(diffInMinutes / 60)}h ago`;
    } else {
      return `${Math.floor(diffInMinutes / 1440)}d ago`;
    }
  };

  const handleChatClick = (chatId: string) => {
    // TODO: Navigate to specific chat when functionality is implemented
    console.log('Navigate to chat:', chatId);
  };

  return (
    <div className={cn("bg-white", className)}>
      {/* Header */}
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center gap-2 mb-1">
          <GitBranch className="h-4 w-4 text-gray-600" />
          <h3 className="text-sm font-medium text-gray-900">Chat History</h3>
        </div>
        <p className="text-xs text-gray-600">
          Conversation threads in this space
        </p>
      </div>

      {/* Chat History Tree */}
      <div className="p-4">
        {mockChatHistory.length > 0 ? (
          <div className="space-y-1">
            {mockChatHistory.map((chat, index) => (
              <div key={chat.id} className="relative">
                {/* Tree visualization lines */}
                {chat.level > 0 && (
                  <>
                    {/* Horizontal line */}
                    <div 
                      className="absolute left-2 top-4 w-3 h-px bg-gray-300"
                      style={{ left: `${chat.level * 12 - 4}px` }}
                    />
                    {/* Vertical line connection */}
                    <div 
                      className="absolute top-0 bottom-4 w-px bg-gray-300"
                      style={{ left: `${(chat.level - 1) * 12 + 8}px` }}
                    />
                  </>
                )}

                {/* Chat item */}
                <div
                  onClick={() => handleChatClick(chat.id)}
                  className={cn(
                    "relative flex items-start gap-2 p-2 rounded-lg cursor-pointer transition-colors",
                    "hover:bg-gray-50",
                    chat.isActive ? "bg-blue-50 border border-blue-200" : "",
                    chat.level > 0 ? "ml-3" : ""
                  )}
                  style={{ marginLeft: `${chat.level * 12}px` }}
                >
                  {/* Icon */}
                  <div className={cn(
                    "flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs",
                    chat.type === 'branch' ? "bg-orange-100 text-orange-600" : "bg-blue-100 text-blue-600"
                  )}>
                    {chat.type === 'branch' ? (
                      <GitBranch className="h-3 w-3" />
                    ) : (
                      <MessageCircle className="h-3 w-3" />
                    )}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between mb-1">
                      <h4 className={cn(
                        "text-sm font-medium truncate",
                        chat.isActive ? "text-blue-900" : "text-gray-900"
                      )}>
                        {chat.title}
                      </h4>
                      <div className="flex items-center gap-1 text-xs text-gray-500">
                        <span>{chat.messageCount}</span>
                        <MessageCircle className="h-3 w-3" />
                      </div>
                    </div>
                    
                    <p className="text-xs text-gray-600 truncate mb-1">
                      {chat.lastMessage}
                    </p>
                    
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1">
                        {chat.participants.map((participant, i) => (
                          <div 
                            key={i}
                            className={cn(
                              "w-3 h-3 rounded-full flex items-center justify-center",
                              participant === 'user' ? "bg-gray-200" : "bg-blue-200"
                            )}
                          >
                            {participant === 'user' ? (
                              <User className="h-2 w-2 text-gray-600" />
                            ) : (
                              <Bot className="h-2 w-2 text-blue-600" />
                            )}
                          </div>
                        ))}
                      </div>
                      
                      <div className="flex items-center gap-1 text-xs text-gray-500">
                        <Clock className="h-3 w-3" />
                        <span>{formatTimeAgo(chat.timestamp)}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <div className="w-10 h-10 bg-gray-100 rounded-full flex items-center justify-center mb-3">
              <GitBranch className="h-5 w-5 text-gray-400" />
            </div>
            <p className="text-sm text-gray-500 mb-1">No chat history yet</p>
            <p className="text-xs text-gray-400">
              Start a conversation to see your chat history
            </p>
          </div>
        )}
      </div>

      {/* Future: Add pagination or load more */}
      {mockChatHistory.length > 0 && (
        <div className="px-4 pb-4">
          <button className="text-xs text-blue-600 hover:text-blue-700 font-medium">
            Load more conversations
          </button>
        </div>
      )}
    </div>
  );
}
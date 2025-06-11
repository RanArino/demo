'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { 
  Send, 
  Bot, 
  User, 
  FileText,
  MessageCircle,
  Trash2,
  Download,
  Copy,
  ThumbsUp,
  ThumbsDown
} from 'lucide-react';
import { MockNode, getNodesByLayer } from '../mock-data';

interface Message {
  id: string;
  type: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  relatedNodes?: string[];
  feedback?: 'positive' | 'negative';
}

interface ChatInterfaceProps {
  nodes: MockNode[];
  selectedNodeId?: string | null;
}

export default function ChatInterface({ nodes, selectedNodeId }: ChatInterfaceProps) {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      type: 'assistant',
      content: 'Hello! I can help you explore and understand the documents in this knowledge space. What would you like to know?',
      timestamp: new Date(),
    }
  ]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Mock RAG response generator
  const generateResponse = (query: string): { content: string; relatedNodes: string[] } => {
    const documentNodes = getNodesByLayer(2);
    const clusterNodes = getNodesByLayer(1);
    const chunkNodes = getNodesByLayer(0);
    
    // Simple keyword matching for demo
    const queryLower = query.toLowerCase();
    const keywords = queryLower.split(' ').filter(word => word.length > 2);
    
    // Find relevant nodes based on keywords
    const relevantNodes = [...documentNodes, ...clusterNodes, ...chunkNodes].filter(node =>
      keywords.some(keyword =>
        node.action_data.textual_data.title.toLowerCase().includes(keyword) ||
        node.action_data.textual_data.summary.toLowerCase().includes(keyword) ||
        node.action_data.textual_data.keywords.some(k => k.toLowerCase().includes(keyword))
      )
    );

    if (relevantNodes.length === 0) {
      return {
        content: "I couldn't find specific information related to your query in the current documents. Could you try rephrasing your question or asking about topics like AI ethics, machine learning, or natural language processing?",
        relatedNodes: []
      };
    }

    const topRelevantNodes = relevantNodes.slice(0, 3);
    const nodeTypes = topRelevantNodes.map(node => {
      switch (node.content_entity_type) {
        case 'content_source': return 'document';
        case 'chunk_cluster': return 'cluster';
        case 'content_chunk': return 'chunk';
        default: return 'node';
      }
    });

    const responses = [
      `Based on the ${nodeTypes.join(', ')} I found, here's what I can tell you about "${query}":`,
      `I found relevant information in ${topRelevantNodes.length} ${topRelevantNodes.length === 1 ? 'source' : 'sources'} related to your question:`,
      `Let me share insights from the knowledge base about "${query}":`,
    ];

    let response = responses[Math.floor(Math.random() * responses.length)] + '\n\n';

    topRelevantNodes.forEach((node, index) => {
      response += `**${node.action_data.textual_data.title}**\n`;
      response += `${node.action_data.textual_data.summary}\n\n`;
      
      if (node.action_data.textual_data.quote) {
        response += `"${node.action_data.textual_data.quote}"\n\n`;
      }
    });

    response += `\nThis information comes from ${topRelevantNodes.length} ${topRelevantNodes.length === 1 ? 'source' : 'sources'} in the knowledge space. Would you like me to elaborate on any specific aspect?`;

    return {
      content: response,
      relatedNodes: topRelevantNodes.map(node => node.id)
    };
  };

  const handleSendMessage = async () => {
    if (!inputValue.trim() || isLoading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: inputValue,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsLoading(true);

    // Simulate API delay
    setTimeout(() => {
      const { content, relatedNodes } = generateResponse(inputValue);
      
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content,
        timestamp: new Date(),
        relatedNodes,
      };

      setMessages(prev => [...prev, assistantMessage]);
      setIsLoading(false);
    }, 1000 + Math.random() * 2000);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const clearChat = () => {
    setMessages([{
      id: '1',
      type: 'assistant',
      content: 'Hello! I can help you explore and understand the documents in this knowledge space. What would you like to know?',
      timestamp: new Date(),
    }]);
  };

  const copyMessage = (content: string) => {
    navigator.clipboard.writeText(content);
  };

  const provideFeedback = (messageId: string, feedback: 'positive' | 'negative') => {
    setMessages(prev => prev.map(msg => 
      msg.id === messageId 
        ? { ...msg, feedback: msg.feedback === feedback ? undefined : feedback }
        : msg
    ));
  };

  const formatTimestamp = (timestamp: Date) => {
    return timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const getNodeTitle = (nodeId: string) => {
    const node = nodes.find(n => n.id === nodeId);
    return node?.action_data.textual_data.title || 'Unknown Node';
  };

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <MessageCircle className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-semibold">Document Chat</h2>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={clearChat}>
              <Trash2 className="w-4 h-4" />
            </Button>
            <Button variant="outline" size="sm">
              <Download className="w-4 h-4" />
            </Button>
          </div>
        </div>
        <p className="text-sm text-gray-600 mt-1">
          Ask questions about the documents and explore connections
        </p>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((message) => (
          <div key={message.id} className={`flex gap-3 ${message.type === 'user' ? 'justify-end' : ''}`}>
            {message.type === 'assistant' && (
              <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center flex-shrink-0">
                <Bot className="w-4 h-4 text-blue-600" />
              </div>
            )}
            
            <div className={`max-w-[80%] ${message.type === 'user' ? 'order-2' : ''}`}>
              <Card className={`${
                message.type === 'user' 
                  ? 'bg-blue-600 text-white' 
                  : 'bg-white'
              }`}>
                <CardContent className="p-3">
                  <div className="whitespace-pre-wrap text-sm">
                    {message.content}
                  </div>
                  
                  {/* Related nodes for assistant messages */}
                  {message.type === 'assistant' && message.relatedNodes && message.relatedNodes.length > 0 && (
                    <div className="mt-3 pt-3 border-t border-gray-200">
                      <p className="text-xs font-medium text-gray-600 mb-2">Related sources:</p>
                      <div className="space-y-1">
                        {message.relatedNodes.map((nodeId) => (
                          <div key={nodeId} className="flex items-center gap-2 text-xs">
                            <FileText className="w-3 h-3 text-gray-400" />
                            <span className="text-gray-600">{getNodeTitle(nodeId)}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  <div className="flex items-center justify-between mt-2 pt-2 border-t border-gray-200">
                    <span className={`text-xs ${
                      message.type === 'user' ? 'text-blue-100' : 'text-gray-500'
                    }`}>
                      {formatTimestamp(message.timestamp)}
                    </span>
                    
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0"
                        onClick={() => copyMessage(message.content)}
                      >
                        <Copy className="w-3 h-3" />
                      </Button>
                      
                      {message.type === 'assistant' && (
                        <>
                          <Button
                            variant="ghost"
                            size="sm"
                            className={`h-6 w-6 p-0 ${
                              message.feedback === 'positive' ? 'text-green-600' : ''
                            }`}
                            onClick={() => provideFeedback(message.id, 'positive')}
                          >
                            <ThumbsUp className="w-3 h-3" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className={`h-6 w-6 p-0 ${
                              message.feedback === 'negative' ? 'text-red-600' : ''
                            }`}
                            onClick={() => provideFeedback(message.id, 'negative')}
                          >
                            <ThumbsDown className="w-3 h-3" />
                          </Button>
                        </>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {message.type === 'user' && (
              <div className="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center flex-shrink-0 order-3">
                <User className="w-4 h-4 text-gray-600" />
              </div>
            )}
          </div>
        ))}

        {isLoading && (
          <div className="flex gap-3">
            <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center flex-shrink-0">
              <Bot className="w-4 h-4 text-blue-600" />
            </div>
            <Card className="bg-white">
              <CardContent className="p-3">
                <div className="flex space-x-1">
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
                </div>
              </CardContent>
            </Card>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input area */}
      <div className="p-4 border-t">
        <div className="flex gap-2">
          <Input
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Ask about the documents..."
            disabled={isLoading}
            className="flex-1"
          />
          <Button 
            onClick={handleSendMessage}
            disabled={!inputValue.trim() || isLoading}
            size="sm"
          >
            <Send className="w-4 h-4" />
          </Button>
        </div>
        
        {/* Quick suggestions */}
        <div className="flex flex-wrap gap-2 mt-2">
          {[
            "What are the main topics?",
            "Explain AI ethics",
            "Show learning methods",
            "Compare approaches"
          ].map((suggestion) => (
            <Button
              key={suggestion}
              variant="outline"
              size="sm"
              className="text-xs"
              onClick={() => setInputValue(suggestion)}
              disabled={isLoading}
            >
              {suggestion}
            </Button>
          ))}
        </div>
      </div>
    </div>
  );
} 
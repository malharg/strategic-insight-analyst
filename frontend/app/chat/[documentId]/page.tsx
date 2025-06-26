"use client";

import { useEffect, useState } from "react";
import { authenticatedFetch } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/context/AuthContext";
import { useRouter } from "next/navigation";
import ReactMarkdown from 'react-markdown'; // npm install react-markdown
import React from "react";

type Message = {
  type: 'user' | 'ai';
  content: string;
};

export default function ChatPage({ params }: { params: Promise<{ documentId: string }> }) {
  const { documentId } = React.use(params);
  const { user, loading } = useAuth();
  const router = useRouter();
  const [messages, setMessages] = useState<Message[]>([]);
  const [query, setQuery] = useState("");
  const [isSending, setIsSending] = useState(false);

   

  useEffect(() => {
    console.log("ChatPage mounted. Document ID from params:", documentId);
  }, [documentId]);






  if (loading) return <p>Loading...</p>;
  if (!user) {
    router.push("/");
    return null;
  }
  
  const handleSendQuery = async () => {
    if (!query.trim()) return;

    const userMessage: Message = { type: 'user', content: query };
    setMessages(prev => [...prev, userMessage]);
    setQuery("");
    setIsSending(true);

    try {
      const payload = { documentId, query };
      
        // =========================================================================
        // DEBUGGING CODE
        // Log the payload right before sending it
      console.log("Sending payload to /api/chat:", payload);
        // =========================================================================
      const data = await authenticatedFetch("/api/chat", {
        method: 'POST',
        body: JSON.stringify({ documentId, query }),
      });
      const aiMessage: Message = { type: 'ai', content: data.response };
      setMessages(prev => [...prev, aiMessage]);
    } catch (error) {
    // Also log the error from the backend if possible
      console.error("Error from backend:", error);
      const errorMessage: Message = { type: 'ai', content: "Sorry, I couldn't process that request." };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsSending(false);
    }
  };

  return (
    <div className="container mx-auto p-4 flex flex-col h-screen">
      <h1 className="text-2xl font-bold mb-4">Analyze Document</h1>
      <div className="flex-grow overflow-y-auto border p-4 rounded-lg space-y-4 bg-white">
        {messages.map((msg, index) => (
          <div key={index} className={`flex ${msg.type === 'user' ? 'justify-end' : 'justify-start'}`}>
              <div className={`p-3 rounded-lg max-w-lg ${msg.type === 'user' ? 'bg-blue-500 text-white' : 'bg-gray-200 text-black'}`}>
                <ReactMarkdown>{msg.content}</ReactMarkdown>
              </div>
          </div>
        ))}
        {isSending && (
            <div className="flex justify-start">
                <div className="p-3 rounded-lg bg-gray-200 text-black">
                    <p>Thinking...</p>
                </div>
            </div>
        )}
      </div>
      <div className="mt-4 flex gap-2">
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="e.g., Summarize the key strategic initiatives..."
          onKeyDown={(e) => e.key === 'Enter' && !isSending && handleSendQuery()}
          disabled={isSending}
        />
        <Button onClick={handleSendQuery} disabled={isSending}>
          {isSending ? "..." : "Send"}
        </Button>
      </div>
    </div>
  );
}
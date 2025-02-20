'use client';

import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';

interface Task {
  title: string;
  description: string;
  priority: string;
  subtasks?: string[];
}

interface AIResponse {
  response: string;
  tasks?: Task[];
}

export default function TaskAIChat() {
  const [message, setMessage] = useState('');
  const [response, setResponse] = useState<AIResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [mode, setMode] = useState<'general' | 'breakdown'>('general');
  const { token } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!message.trim() || isLoading) return;

    setIsLoading(true);
    try {
      const response = await fetch('http://localhost:8080/api/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ 
          message,
          type: mode
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to get AI response');
      }

      const data = await response.json();
      setResponse(data);
    } catch (error) {
      console.error('Error getting AI response:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex space-x-4">
        <button
          type="button"
          onClick={() => setMode('general')}
          className={`px-4 py-2 rounded-lg font-medium ${
            mode === 'general'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
          }`}
        >
          General Assistance
        </button>
        <button
          type="button"
          onClick={() => setMode('breakdown')}
          className={`px-4 py-2 rounded-lg font-medium ${
            mode === 'breakdown'
              ? 'bg-blue-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
          }`}
        >
          Task Breakdown
        </button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="message" className="block text-sm font-medium text-gray-900 mb-2">
            {mode === 'breakdown' 
              ? 'Describe your task for AI breakdown...'
              : 'Ask about task management...'}
          </label>
          <textarea
            id="message"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            rows={4}
            className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 placeholder-gray-500 focus:border-blue-500 focus:ring-blue-500"
            placeholder={mode === 'breakdown'
              ? 'Describe your task in detail for AI-powered breakdown...'
              : 'How can I organize my tasks better?'}
          />
        </div>
        <button
          type="submit"
          disabled={isLoading || !message.trim()}
          className="w-full rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
        >
          {isLoading ? 'Thinking...' : mode === 'breakdown' ? 'Break Down Task' : 'Send'}
        </button>
      </form>

      {response && (
        <div className="rounded-lg border border-gray-200 bg-gray-50 p-6 space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">AI Assistant Response:</h3>
          <div className="prose max-w-none">
            <p className="whitespace-pre-wrap text-gray-800">{response.response}</p>
          </div>

          {response.tasks && response.tasks.length > 0 && (
            <div className="mt-6 space-y-4">
              <h4 className="text-md font-semibold text-gray-900">Suggested Task Breakdown:</h4>
              {response.tasks.map((task, index) => (
                <div key={index} className="bg-white rounded-lg border border-gray-200 p-4">
                  <h5 className="font-medium text-gray-900">{task.title}</h5>
                  <p className="mt-1 text-gray-600">{task.description}</p>
                  <div className="mt-2">
                    <span className={`inline-block px-2 py-1 text-sm rounded-full ${
                      task.priority === 'high' 
                        ? 'bg-red-100 text-red-800'
                        : task.priority === 'medium'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-blue-100 text-blue-800'
                    }`}>
                      {task.priority} priority
                    </span>
                  </div>
                  {task.subtasks && task.subtasks.length > 0 && (
                    <div className="mt-3">
                      <h6 className="text-sm font-medium text-gray-900 mb-2">Subtasks:</h6>
                      <ul className="list-disc pl-5 space-y-1">
                        {task.subtasks.map((subtask, idx) => (
                          <li key={idx} className="text-gray-600">{subtask}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

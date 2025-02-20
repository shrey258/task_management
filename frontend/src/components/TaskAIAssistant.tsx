'use client';

import { useState } from 'react';
import { useAuth } from '@/context/AuthContext';

interface TaskSuggestion {
  title: string;
  description: string;
  priority: string;
  tags: string[];
}

interface TaskAnalysis {
  priority: string;
  reasoning: string;
  suggestedTags: string[];
  timeEstimate: string;
}

export default function TaskAIAssistant() {
  const [description, setDescription] = useState('');
  const [title, setTitle] = useState('');
  const [suggestions, setSuggestions] = useState<TaskSuggestion[]>([]);
  const [analysis, setAnalysis] = useState<TaskAnalysis | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const { token } = useAuth();

  const generateSuggestions = async () => {
    if (!description.trim() || isLoading) return;
    setIsLoading(true);

    try {
      const response = await fetch('http://localhost:8080/api/ai/suggest', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ description: description.trim() }),
      });

      if (!response.ok) {
        throw new Error('Failed to get suggestions');
      }

      const data = await response.json();
      setSuggestions(data.suggestions);
    } catch (error) {
      console.error('Error:', error);
      alert('Failed to generate suggestions. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const analyzeTask = async () => {
    if (!title.trim() || !description.trim() || isLoading) return;
    setIsLoading(true);

    try {
      const response = await fetch('http://localhost:8080/api/ai/analyze', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          title: title.trim(),
          description: description.trim(),
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to analyze task');
      }

      const data = await response.json();
      setAnalysis(data);
    } catch (error) {
      console.error('Error:', error);
      alert('Failed to analyze task. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-6 p-6 bg-white rounded-lg shadow-lg">
      <div>
        <h2 className="text-2xl font-bold mb-4">AI Task Assistant</h2>
        
        {/* Task Description Input */}
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">Task Description</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              rows={4}
              placeholder="Describe your task or project..."
            />
          </div>
          
          <button
            onClick={generateSuggestions}
            disabled={isLoading || !description.trim()}
            className="w-full px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
          >
            {isLoading ? 'Generating...' : 'Generate Task Suggestions'}
          </button>
        </div>

        {/* Display Suggestions */}
        {suggestions.length > 0 && (
          <div className="mt-6">
            <h3 className="text-lg font-semibold mb-3">Suggested Tasks</h3>
            <div className="space-y-4">
              {suggestions.map((suggestion, index) => (
                <div key={index} className="p-4 border rounded-md">
                  <h4 className="font-medium">{suggestion.title}</h4>
                  <p className="text-gray-600 mt-1">{suggestion.description}</p>
                  <div className="mt-2">
                    <span className="inline-block px-2 py-1 text-sm bg-blue-100 text-blue-800 rounded">
                      Priority: {suggestion.priority}
                    </span>
                    {suggestion.tags.map((tag, i) => (
                      <span key={i} className="ml-2 inline-block px-2 py-1 text-sm bg-gray-100 text-gray-800 rounded">
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Task Analysis Section */}
        <div className="mt-8 pt-8 border-t">
          <h3 className="text-lg font-semibold mb-4">Task Analysis</h3>
          
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Task Title</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                placeholder="Enter task title..."
              />
            </div>
            
            <button
              onClick={analyzeTask}
              disabled={isLoading || !title.trim() || !description.trim()}
              className="w-full px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:opacity-50"
            >
              {isLoading ? 'Analyzing...' : 'Analyze Task'}
            </button>
          </div>

          {/* Display Analysis Results */}
          {analysis && (
            <div className="mt-4 p-4 border rounded-md bg-gray-50">
              <div className="space-y-3">
                <div>
                  <span className="font-medium">Priority:</span>
                  <span className="ml-2 inline-block px-2 py-1 text-sm bg-blue-100 text-blue-800 rounded">
                    {analysis.priority}
                  </span>
                </div>
                <div>
                  <span className="font-medium">Reasoning:</span>
                  <p className="mt-1 text-gray-600">{analysis.reasoning}</p>
                </div>
                <div>
                  <span className="font-medium">Time Estimate:</span>
                  <span className="ml-2 text-gray-600">{analysis.timeEstimate}</span>
                </div>
                <div>
                  <span className="font-medium">Suggested Tags:</span>
                  <div className="mt-2">
                    {analysis.suggestedTags.map((tag, index) => (
                      <span key={index} className="mr-2 inline-block px-2 py-1 text-sm bg-gray-100 text-gray-800 rounded">
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

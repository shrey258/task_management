'use client';

import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import TaskForm from '@/components/TaskForm';
import TaskAIChat from '@/components/TaskAIChat';
import { Task } from '@/types';

export default function CreateTaskPage() {
  const router = useRouter();
  const { token } = useAuth();

  const handleSubmit = async (taskData: Partial<Task>) => {
    try {
      const response = await fetch('http://localhost:8080/api/tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(taskData),
      });

      if (!response.ok) {
        throw new Error('Failed to create task');
      }

      router.push('/dashboard/tasks');
    } catch (error) {
      console.error('Error creating task:', error);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div className="bg-white rounded-lg shadow-lg p-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-6">Create New Task</h1>
          <TaskForm
            onSubmit={handleSubmit}
            onCancel={() => router.push('/dashboard/tasks')}
          />
        </div>

        <div className="bg-white rounded-lg shadow-lg p-6">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">AI Task Assistant</h2>
          <p className="text-gray-600 mb-4">
            Get help with task planning and organization. Ask questions about task management
            best practices or get suggestions for breaking down complex tasks.
          </p>
          <TaskAIChat />
        </div>
      </div>
    </div>
  );
}

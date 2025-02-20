'use client';

import { createContext, useContext, useEffect, useRef, useState } from 'react';
import { useAuth } from './AuthContext';
import { Task } from '@/types';

interface WebSocketContextType {
  sendMessage: (message: any) => void;
  isConnected: boolean;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

interface WebSocketMessage {
  type: string;
  payload: any;
}

interface WebSocketProviderProps {
  children: React.ReactNode;
  onTaskUpdate?: (task: Task) => void;
  onTaskCreate?: (task: Task) => void;
  onTaskDelete?: (taskId: string) => void;
}

export function WebSocketProvider({ 
  children,
  onTaskUpdate,
  onTaskCreate,
  onTaskDelete
}: WebSocketProviderProps) {
  const { token, isAuthenticated } = useAuth();
  const ws = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const reconnectTimeout = useRef<NodeJS.Timeout>();
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5;

  const connect = () => {
    if (!token || !isAuthenticated) {
      console.log('WebSocket: Not connecting - user not authenticated');
      return;
    }

    if (ws.current?.readyState === WebSocket.OPEN) {
      console.log('WebSocket: Already connected');
      return;
    }

    if (reconnectAttempts.current >= maxReconnectAttempts) {
      console.log('WebSocket: Max reconnection attempts reached');
      return;
    }

    try {
      // Include token in the WebSocket URL
      const socket = new WebSocket(`ws://localhost:8080/ws?token=${token}`);
      ws.current = socket;

      socket.onopen = () => {
        console.log('WebSocket: Connected');
        setIsConnected(true);
        reconnectAttempts.current = 0;
        if (reconnectTimeout.current) {
          clearTimeout(reconnectTimeout.current);
          reconnectTimeout.current = undefined;
        }
      };

      socket.onclose = (event) => {
        console.log('WebSocket: Disconnected', event.code, event.reason);
        setIsConnected(false);
        
        // Only attempt to reconnect if we're still authenticated
        if (isAuthenticated && token && reconnectAttempts.current < maxReconnectAttempts) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000);
          console.log(`WebSocket: Reconnecting in ${delay}ms (attempt ${reconnectAttempts.current + 1}/${maxReconnectAttempts})`);
          
          if (!reconnectTimeout.current) {
            reconnectTimeout.current = setTimeout(() => {
              reconnectTimeout.current = undefined;
              reconnectAttempts.current++;
              connect();
            }, delay);
          }
        }
      };

      socket.onerror = (error) => {
        console.error('WebSocket: Error:', error);
      };

      socket.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          
          switch (message.type) {
            case 'task_updated':
              onTaskUpdate?.(message.payload);
              break;
            case 'task_created':
              onTaskCreate?.(message.payload);
              break;
            case 'task_deleted':
              onTaskDelete?.(message.payload);
              break;
            default:
              console.log('WebSocket: Unknown message type:', message.type);
          }
        } catch (error) {
          console.error('WebSocket: Error processing message:', error);
        }
      };
    } catch (error) {
      console.error('WebSocket: Error creating connection:', error);
      reconnectAttempts.current++;
    }
  };

  useEffect(() => {
    if (isAuthenticated && token) {
      connect();
    } else {
      // Clean up connection if user becomes unauthenticated
      if (ws.current) {
        ws.current.close();
        ws.current = null;
      }
      setIsConnected(false);
      reconnectAttempts.current = 0;
      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current);
        reconnectTimeout.current = undefined;
      }
    }

    return () => {
      if (ws.current) {
        ws.current.close();
        ws.current = null;
      }
      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current);
        reconnectTimeout.current = undefined;
      }
    };
  }, [isAuthenticated, token]);

  const sendMessage = (message: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket: Cannot send message - not connected');
    }
  };

  return (
    <WebSocketContext.Provider value={{ sendMessage, isConnected }}>
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocket() {
  const context = useContext(WebSocketContext);
  if (context === undefined) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
}

import { useState, useEffect, useRef, useCallback } from 'react';

const getWebSocketURL = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  return `${protocol}//${host}/ws/translation`;
};

// Valid message types from server
const VALID_MESSAGE_TYPES = [
  't_room_created',
  't_opponent_joined',
  't_game_start',
  't_countdown',
  't_round_start',
  't_time_update',
  't_opponent_submitted',
  't_evaluating',
  't_round_result',
  't_game_over',
  't_opponent_ready',
  't_rematch_start',
  't_opponent_left',
  't_opponent_next_ready',
  't_error'
];

export default function useTranslationWS() {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState(null);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const messageHandlersRef = useRef(new Map());

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const ws = new WebSocket(getWebSocketURL());

    ws.onopen = () => {
      setIsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Validate message type
        if (!data.type || !VALID_MESSAGE_TYPES.includes(data.type)) {
          return;
        }
        
        setLastMessage(data);
        
        // Call registered handlers for this message type
        const handler = messageHandlersRef.current.get(data.type);
        if (handler) {
          handler(data);
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    wsRef.current = ws;
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const sendMessage = useCallback((message) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  const onMessage = useCallback((type, handler) => {
    messageHandlersRef.current.set(type, handler);
    return () => {
      messageHandlersRef.current.delete(type);
    };
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    lastMessage,
    connect,
    disconnect,
    sendMessage,
    onMessage,
  };
}

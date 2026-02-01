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

// Connection states
export const ConnectionState = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting',
  CONNECTED: 'connected',
  RECONNECTING: 'reconnecting',
};

const MAX_RECONNECT_ATTEMPTS = 3;
const RECONNECT_DELAY = 2000; // 2 seconds

export default function useTranslationWS() {
  const [isConnected, setIsConnected] = useState(false);
  const [connectionState, setConnectionState] = useState(ConnectionState.DISCONNECTED);
  const [lastMessage, setLastMessage] = useState(null);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttemptsRef = useRef(0);
  const messageHandlersRef = useRef(new Map());
  const intentionalDisconnectRef = useRef(false);
  const lastMessageRef = useRef(null); // 재연결 시 복원할 상태

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    intentionalDisconnectRef.current = false;
    setConnectionState(ConnectionState.CONNECTING);

    const ws = new WebSocket(getWebSocketURL());

    ws.onopen = () => {
      setIsConnected(true);
      setConnectionState(ConnectionState.CONNECTED);
      reconnectAttemptsRef.current = 0; // 연결 성공 시 재연결 카운터 리셋
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Validate message type
        if (!data.type || !VALID_MESSAGE_TYPES.includes(data.type)) {
          return;
        }
        
        setLastMessage(data);
        lastMessageRef.current = data;
        
        // Call all registered handlers for this message type
        const handlers = messageHandlersRef.current.get(data.type);
        if (handlers) {
          handlers.forEach(handler => handler(data));
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    ws.onclose = (event) => {
      setIsConnected(false);
      wsRef.current = null;
      
      // 의도적 연결 해제가 아니면 재연결 시도
      if (!intentionalDisconnectRef.current) {
        if (reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
          setConnectionState(ConnectionState.RECONNECTING);
          reconnectAttemptsRef.current++;
          
          console.log(`WebSocket 재연결 시도 ${reconnectAttemptsRef.current}/${MAX_RECONNECT_ATTEMPTS}`);
          
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, RECONNECT_DELAY);
        } else {
          setConnectionState(ConnectionState.DISCONNECTED);
          // 재연결 실패 알림
          const handlers = messageHandlersRef.current.get('t_error');
          if (handlers) {
            handlers.forEach(handler => handler({ 
              type: 't_error', 
              message: '서버 연결이 끊어졌습니다. 페이지를 새로고침해주세요.' 
            }));
          }
        }
      } else {
        setConnectionState(ConnectionState.DISCONNECTED);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    wsRef.current = ws;
  }, []);

  const disconnect = useCallback(() => {
    intentionalDisconnectRef.current = true;
    
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
    setConnectionState(ConnectionState.DISCONNECTED);
    reconnectAttemptsRef.current = 0;
  }, []);

  const sendMessage = useCallback((message) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
      return true;
    }
    return false;
  }, []);

  const onMessage = useCallback((type, handler) => {
    // 해당 타입의 핸들러 배열이 없으면 생성
    if (!messageHandlersRef.current.has(type)) {
      messageHandlersRef.current.set(type, []);
    }
    // 핸들러를 배열에 추가
    messageHandlersRef.current.get(type).push(handler);
    
    // cleanup 함수: 해당 핸들러만 배열에서 제거
    return () => {
      const handlers = messageHandlersRef.current.get(type);
      if (handlers) {
        const index = handlers.indexOf(handler);
        if (index > -1) {
          handlers.splice(index, 1);
        }
        // 배열이 비었으면 Map에서 제거
        if (handlers.length === 0) {
          messageHandlersRef.current.delete(type);
        }
      }
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
    connectionState,
    lastMessage,
    connect,
    disconnect,
    sendMessage,
    onMessage,
  };
}

import { useState, useEffect, useCallback } from 'react';
import useWebSocket from '../../../hooks/useWebSocket';
import SpeedClickBattle from './SpeedClickBattle';
import './BattleLobby.css';

// 승자 점수를 랭킹에 저장하는 함수
const saveWinnerScore = async (nickname, score) => {
  try {
    await fetch('/api/scores', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        game: 'speed-click-battle',
        nickname,
        score,
      }),
    });
  } catch (error) {
    console.error('Failed to save score:', error);
  }
};

function BattleLobby({ onBack }) {
  const [nickname, setNickname] = useState('');
  const [roomCode, setRoomCode] = useState('');
  const [inputRoomCode, setInputRoomCode] = useState('');
  const [lobbyState, setLobbyState] = useState('input'); // input, waiting, countdown, playing, finished
  const [opponentNickname, setOpponentNickname] = useState('');
  const [countdown, setCountdown] = useState(0);
  const [error, setError] = useState('');
  const [gameData, setGameData] = useState(null);
  const [rematchRequested, setRematchRequested] = useState(false);
  const [opponentReady, setOpponentReady] = useState(false);
  const [playerIndex, setPlayerIndex] = useState(0); // 0: player1, 1: player2

  const { isConnected, connect, disconnect, sendMessage, onMessage } = useWebSocket();

  // Setup message handlers
  useEffect(() => {
    const unsubscribers = [
      onMessage('room_created', (data) => {
        setRoomCode(data.roomCode);
        setLobbyState('waiting');
        setError('');
        setPlayerIndex(0); // 방 만든 사람은 player1
      }),

      onMessage('opponent_joined', (data) => {
        setOpponentNickname(data.nickname);
      }),

      onMessage('countdown', (data) => {
        setLobbyState('countdown');
        setCountdown(data.count);
      }),

      onMessage('game_start', (data) => {
        setLobbyState('playing');
        setGameData({ duration: data.duration });
      }),

      onMessage('game_end', (data) => {
        setLobbyState('finished');
        setGameData(prev => ({ ...prev, result: data }));
        setRematchRequested(false);
        setOpponentReady(false);
        
        // 승자의 점수를 랭킹에 저장
        if (data.result === 'win') {
          saveWinnerScore(data.winnerNickname, data.myScore);
        }
      }),

      onMessage('rematch_start', () => {
        setLobbyState('countdown');
        setGameData(null);
        setRematchRequested(false);
        setOpponentReady(false);
      }),

      onMessage('opponent_ready', () => {
        setOpponentReady(true);
      }),

      onMessage('opponent_left', () => {
        setError('상대방이 나갔습니다.');
        setLobbyState('input');
        setRoomCode('');
        setOpponentNickname('');
      }),

      onMessage('error', (data) => {
        setError(data.message);
      }),
    ];

    return () => {
      unsubscribers.forEach(unsub => unsub());
    };
  }, [onMessage]);

  const handleCreateRoom = useCallback(() => {
    if (!nickname.trim()) {
      setError('닉네임을 입력해주세요.');
      return;
    }
    if (nickname.length > 10) {
      setError('닉네임은 10자 이하로 입력해주세요.');
      return;
    }
    setError('');
    connect();
    // Wait for connection then send create message
    setTimeout(() => {
      sendMessage({ type: 'create', nickname: nickname.trim() });
    }, 500);
  }, [nickname, connect, sendMessage]);

  const handleJoinRoom = useCallback(() => {
    if (!nickname.trim()) {
      setError('닉네임을 입력해주세요.');
      return;
    }
    if (nickname.length > 10) {
      setError('닉네임은 10자 이하로 입력해주세요.');
      return;
    }
    if (!inputRoomCode.trim()) {
      setError('방 코드를 입력해주세요.');
      return;
    }
    if (inputRoomCode.length !== 6) {
      setError('방 코드는 6자리입니다.');
      return;
    }
    setError('');
    setRoomCode(inputRoomCode.toUpperCase());
    setPlayerIndex(1); // 참가자는 player2
    connect();
    setTimeout(() => {
      sendMessage({ 
        type: 'join', 
        roomCode: inputRoomCode.toUpperCase(), 
        nickname: nickname.trim() 
      });
    }, 500);
  }, [nickname, inputRoomCode, connect, sendMessage]);

  const handleCancel = useCallback(() => {
    sendMessage({ type: 'leave' });
    disconnect();
    setLobbyState('input');
    setRoomCode('');
    setOpponentNickname('');
    setError('');
  }, [sendMessage, disconnect]);

  const handleBack = useCallback(() => {
    disconnect();
    onBack();
  }, [disconnect, onBack]);

  const handleRematch = useCallback(() => {
    sendMessage({ type: 'ready_rematch' });
    setRematchRequested(true);
  }, [sendMessage]);

  const handleLeaveGame = useCallback(() => {
    sendMessage({ type: 'leave' });
    disconnect();
    setLobbyState('input');
    setRoomCode('');
    setOpponentNickname('');
    setGameData(null);
  }, [sendMessage, disconnect]);

  // Render game screen
  if (lobbyState === 'playing') {
    return (
      <SpeedClickBattle
        nickname={nickname}
        opponentNickname={opponentNickname}
        playerIndex={playerIndex}
        sendMessage={sendMessage}
        onMessage={onMessage}
      />
    );
  }

  // Render finished screen
  if (lobbyState === 'finished' && gameData?.result) {
    const { result } = gameData;
    return (
      <div className="battle-lobby">
        <div className="result-screen">
          <h2 className={`result-title ${result.result}`}>
            {result.result === 'win' && '승리!'}
            {result.result === 'lose' && '패배...'}
            {result.result === 'draw' && '무승부!'}
          </h2>
          
          <div className="result-scores">
            <div className="result-player">
              <span className="result-nickname">{nickname}</span>
              <span className="result-score">{result.myScore}점</span>
            </div>
            <span className="vs">vs</span>
            <div className="result-player">
              <span className="result-nickname">{opponentNickname}</span>
              <span className="result-score">{result.opponentScore}점</span>
            </div>
          </div>

          {result.winnerNickname && (
            <p className="winner-text">{result.winnerNickname} 승리!</p>
          )}

          <div className="result-buttons">
            <button 
              className={`btn-rematch ${rematchRequested ? 'requested' : ''}`} 
              onClick={handleRematch}
              disabled={rematchRequested}
            >
              {rematchRequested ? '대기 중...' : '재대결'}
            </button>
            <button className="btn-leave" onClick={handleLeaveGame}>
              나가기
            </button>
          </div>
          
          {opponentReady && !rematchRequested && (
            <p className="opponent-ready-text">상대가 재대결을 원합니다!</p>
          )}
          {rematchRequested && !opponentReady && (
            <p className="waiting-opponent-text">상대의 응답을 기다리는 중...</p>
          )}
        </div>
      </div>
    );
  }

  // Render countdown screen
  if (lobbyState === 'countdown') {
    return (
      <div className="battle-lobby">
        <div className="countdown-screen">
          <p className="vs-text">vs {opponentNickname}</p>
          <div className="countdown-number">{countdown}</div>
        </div>
      </div>
    );
  }

  // Render waiting screen
  if (lobbyState === 'waiting') {
    return (
      <div className="battle-lobby">
        <div className="waiting-screen">
          <h2>대기 중...</h2>
          <div className="room-code-display">
            <span className="label">방 코드</span>
            <span className="code">{roomCode}</span>
          </div>
          <p className="hint">친구에게 이 코드를 공유하세요!</p>
          <div className="waiting-spinner"></div>
          <button className="btn-cancel" onClick={handleCancel}>
            취소
          </button>
        </div>
      </div>
    );
  }

  // Render input screen
  return (
    <div className="battle-lobby">
      <div className="lobby-content">
        <h2 className="lobby-title">대결 모드</h2>

        {error && <div className="error-message">{error}</div>}

        <div className="input-group">
          <label>닉네임</label>
          <input
            type="text"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            placeholder="닉네임 입력 (최대 10자)"
            maxLength={10}
          />
        </div>

        <button className="btn-create" onClick={handleCreateRoom}>
          방 만들기
        </button>

        <div className="divider">
          <span>또는</span>
        </div>

        <div className="input-group">
          <label>방 코드</label>
          <input
            type="text"
            value={inputRoomCode}
            onChange={(e) => setInputRoomCode(e.target.value.toUpperCase())}
            placeholder="6자리 코드 입력"
            maxLength={6}
          />
        </div>

        <button className="btn-join" onClick={handleJoinRoom}>
          참가하기
        </button>

        <button className="btn-back" onClick={handleBack}>
          돌아가기
        </button>
      </div>
    </div>
  );
}

export default BattleLobby;

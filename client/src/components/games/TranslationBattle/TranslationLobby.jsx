import { useState, useEffect, useCallback } from 'react';
import useTranslationWS from '../../../hooks/useTranslationWS';
import TranslationBattle from './TranslationBattle';
import './TranslationLobby.css';

// Sanitize nickname input
const sanitizeNickname = (value) => {
  return value
    .slice(0, 10)
    .replace(/[<>"'&;]/g, '');
};

// Sanitize room code
const sanitizeRoomCode = (value) => {
  return value
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, '')
    .slice(0, 6);
};

function TranslationLobby({ onBack }) {
  const [nickname, setNickname] = useState('');
  const [roomCode, setRoomCode] = useState('');
  const [inputRoomCode, setInputRoomCode] = useState('');
  const [lobbyState, setLobbyState] = useState('input'); // input, waiting, ready, countdown, playing, result, finished
  const [opponentNickname, setOpponentNickname] = useState('');
  const [countdown, setCountdown] = useState(0);
  const [difficulty, setDifficulty] = useState('medium');
  const [error, setError] = useState('');
  const [gameData, setGameData] = useState(null);
  const [initialSentence, setInitialSentence] = useState('');
  const [rematchRequested, setRematchRequested] = useState(false);
  const [opponentReady, setOpponentReady] = useState(false);
  const [isHost, setIsHost] = useState(false);

  const { isConnected, connect, disconnect, sendMessage, onMessage } = useTranslationWS();

  // Setup message handlers
  useEffect(() => {
    const unsubscribers = [
      onMessage('t_room_created', (data) => {
        setRoomCode(data.roomCode);
        setLobbyState('waiting');
        setError('');
        setIsHost(true);
      }),

      onMessage('t_opponent_joined', (data) => {
        setOpponentNickname(data.nickname);
        setLobbyState('ready');  // Host/Guest 모두 ready 상태로 전환
      }),

      onMessage('t_game_start', (data) => {
        setDifficulty(data.difficulty);
        setLobbyState('countdown');
      }),

      onMessage('t_countdown', (data) => {
        setCountdown(data.count);
      }),

      onMessage('t_round_start', (data) => {
        setInitialSentence(data.sentence);
        setLobbyState('playing');
      }),

      onMessage('t_round_result', (data) => {
        setLobbyState('result');
        setGameData(prev => ({
          ...prev,
          roundResult: data,
          isGameOver: data.isGameOver
        }));
      }),

      onMessage('t_game_over', (data) => {
        setLobbyState('finished');
        setGameData(prev => ({ ...prev, finalResult: data }));
        setRematchRequested(false);
        setOpponentReady(false);
      }),

      onMessage('t_rematch_start', () => {
        setLobbyState('countdown');
        setGameData(null);
        setRematchRequested(false);
        setOpponentReady(false);
      }),

      onMessage('t_opponent_ready', () => {
        setOpponentReady(true);
      }),

      onMessage('t_opponent_left', () => {
        setError('상대방이 나갔습니다.');
        setLobbyState('input');
        setRoomCode('');
        setOpponentNickname('');
        setIsHost(false);
      }),

      onMessage('t_error', (data) => {
        setError(data.message);
      }),
    ];

    return () => {
      unsubscribers.forEach(unsub => unsub());
    };
  }, [onMessage, isHost]);

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
    setTimeout(() => {
      sendMessage({ type: 't_create', nickname: nickname.trim() });
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
    setIsHost(false);
    connect();
    setTimeout(() => {
      sendMessage({ 
        type: 't_join', 
        roomCode: inputRoomCode.toUpperCase(), 
        nickname: nickname.trim() 
      });
    }, 500);
  }, [nickname, inputRoomCode, connect, sendMessage]);

  const handleStartGame = useCallback(() => {
    sendMessage({ type: 't_start_game', difficulty });
  }, [sendMessage, difficulty]);

  const handleCancel = useCallback(() => {
    sendMessage({ type: 't_leave' });
    disconnect();
    setLobbyState('input');
    setRoomCode('');
    setOpponentNickname('');
    setError('');
    setIsHost(false);
  }, [sendMessage, disconnect]);

  const handleBack = useCallback(() => {
    disconnect();
    onBack();
  }, [disconnect, onBack]);

  const handleRematch = useCallback(() => {
    sendMessage({ type: 't_rematch' });
    setRematchRequested(true);
  }, [sendMessage]);

  const handleLeaveGame = useCallback(() => {
    sendMessage({ type: 't_leave' });
    disconnect();
    setLobbyState('input');
    setRoomCode('');
    setOpponentNickname('');
    setGameData(null);
    setIsHost(false);
  }, [sendMessage, disconnect]);

  // Render game screen
  if (lobbyState === 'playing' || lobbyState === 'result') {
    return (
      <TranslationBattle
        nickname={nickname}
        opponentNickname={opponentNickname}
        difficulty={difficulty}
        initialSentence={initialSentence}
        sendMessage={sendMessage}
        onMessage={onMessage}
        onFinished={() => setLobbyState('finished')}
      />
    );
  }

  // Render finished screen
  if (lobbyState === 'finished' && gameData?.finalResult) {
    const { finalResult } = gameData;
    return (
      <div className="translation-lobby">
        <div className="result-screen">
          <h2 className={`result-title ${finalResult.winner}`}>
            {finalResult.winner === 'me' && '승리!'}
            {finalResult.winner === 'opponent' && '패배...'}
            {finalResult.winner === 'draw' && '무승부!'}
          </h2>
          
          <div className="result-summary">
            <div className="result-player">
              <span className="result-nickname">{nickname}</span>
              <span className="result-wins">{finalResult.myWins}승</span>
              <span className="result-total">{finalResult.myTotalScore}점</span>
            </div>
            <span className="vs">vs</span>
            <div className="result-player">
              <span className="result-nickname">{opponentNickname}</span>
              <span className="result-wins">{finalResult.opponentWins}승</span>
              <span className="result-total">{finalResult.opponentTotal}점</span>
            </div>
          </div>

          {finalResult.winnerNickname && (
            <p className="winner-text">{finalResult.winnerNickname} 승리!</p>
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
      <div className="translation-lobby">
        <div className="countdown-screen">
          <p className="vs-text">vs {opponentNickname}</p>
          <div className="difficulty-badge">{difficulty === 'easy' ? '쉬움' : difficulty === 'hard' ? '어려움' : '보통'}</div>
          <div className="countdown-number">{countdown}</div>
        </div>
      </div>
    );
  }

  // Render ready screen (host selects difficulty)
  if (lobbyState === 'ready' && isHost) {
    return (
      <div className="translation-lobby">
        <div className="ready-screen">
          <h2>게임 준비</h2>
          <p className="opponent-info">상대: <strong>{opponentNickname}</strong></p>
          
          <div className="difficulty-selector">
            <label>난이도 선택</label>
            <div className="difficulty-options">
              <button 
                className={`difficulty-btn ${difficulty === 'easy' ? 'selected' : ''}`}
                onClick={() => setDifficulty('easy')}
              >
                쉬움
                <span className="difficulty-desc">기본 문장</span>
              </button>
              <button 
                className={`difficulty-btn ${difficulty === 'medium' ? 'selected' : ''}`}
                onClick={() => setDifficulty('medium')}
              >
                보통
                <span className="difficulty-desc">중급 문장</span>
              </button>
              <button 
                className={`difficulty-btn ${difficulty === 'hard' ? 'selected' : ''}`}
                onClick={() => setDifficulty('hard')}
              >
                어려움
                <span className="difficulty-desc">고급 문장</span>
              </button>
            </div>
          </div>

          <button className="btn-start" onClick={handleStartGame}>
            게임 시작
          </button>
          <button className="btn-cancel" onClick={handleCancel}>
            취소
          </button>
        </div>
      </div>
    );
  }

  // Render waiting for host screen (guest)
  if (lobbyState === 'ready' && !isHost) {
    return (
      <div className="translation-lobby">
        <div className="waiting-screen">
          <h2>대기 중</h2>
          <p className="opponent-info">상대: <strong>{opponentNickname}</strong></p>
          <p className="hint">방장이 게임을 시작할 때까지 기다려주세요</p>
          <div className="waiting-spinner"></div>
          <button className="btn-cancel" onClick={handleCancel}>
            취소
          </button>
        </div>
      </div>
    );
  }

  // Render waiting screen (waiting for opponent)
  if (lobbyState === 'waiting') {
    return (
      <div className="translation-lobby">
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
    <div className="translation-lobby">
      <div className="lobby-content">
        <h2 className="lobby-title">Translation Battle</h2>
        <p className="lobby-subtitle">AI와 함께하는 영어 번역 대결</p>

        {error && <div className="error-message">{error}</div>}

        <div className="input-group">
          <label>닉네임</label>
          <input
            type="text"
            value={nickname}
            onChange={(e) => setNickname(sanitizeNickname(e.target.value))}
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
            onChange={(e) => setInputRoomCode(sanitizeRoomCode(e.target.value))}
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

export default TranslationLobby;

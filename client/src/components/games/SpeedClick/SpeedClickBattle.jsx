import { useState, useEffect, useRef, useCallback } from 'react';
import './SpeedClickBattle.css';

const GAME_WIDTH = 1200;
const GAME_HEIGHT = 800;

function SpeedClickBattle({ nickname, opponentNickname, playerIndex, sendMessage, onMessage }) {
  const [myScore, setMyScore] = useState(0);
  const [opponentScore, setOpponentScore] = useState(0);
  const [timeLeft, setTimeLeft] = useState(10);
  const [ball, setBall] = useState(null);
  const [clickEffect, setClickEffect] = useState(null);
  const [lastResult, setLastResult] = useState(null);
  
  const gameAreaRef = useRef(null);
  const ballTimerRef = useRef(null);

  // Setup message handlers
  useEffect(() => {
    const unsubscribers = [
      onMessage('ball_spawn', (data) => {
        setBall({
          id: data.id,
          x: data.x,
          y: data.y,
          isRed: data.isRed,
          size: data.size,
          timeLimit: data.timeLimit,
          timeLeft: data.timeLimit,
        });
        setLastResult(null);
      }),

      onMessage('ball_result', (data) => {
        setBall(null);
        // playerIndex에 따라 점수 올바르게 설정
        setMyScore(data.scores[playerIndex]);
        setOpponentScore(data.scores[1 - playerIndex]);
        
        // Show result effect
        if (data.clickedBy !== 'none') {
          const isMe = (data.clickedBy === 'player1' && playerIndex === 0) || 
                       (data.clickedBy === 'player2' && playerIndex === 1);
          setLastResult({
            clickedBy: isMe ? nickname : opponentNickname,
            isMe,
          });
        } else {
          setLastResult({ clickedBy: 'none', isMe: false });
        }
      }),

      onMessage('time_update', (data) => {
        setTimeLeft(data.timeLeft);
      }),
    ];

    return () => {
      unsubscribers.forEach(unsub => unsub());
    };
  }, [onMessage, nickname, opponentNickname, playerIndex]);

  // Ball timer (visual countdown)
  useEffect(() => {
    if (!ball) return;

    ballTimerRef.current = setInterval(() => {
      setBall(prev => {
        if (!prev) return null;
        const newTimeLeft = prev.timeLeft - 0.016;
        if (newTimeLeft <= 0) return prev; // Let server handle timeout
        return { ...prev, timeLeft: newTimeLeft };
      });
    }, 16);

    return () => {
      if (ballTimerRef.current) {
        clearInterval(ballTimerRef.current);
      }
    };
  }, [ball?.id]);

  // Handle ball click
  const handleBallClick = useCallback((e) => {
    e.stopPropagation();
    if (!ball) return;

    // Show click effect
    setClickEffect({
      x: ball.x,
      y: ball.y,
      type: ball.isRed ? 'success' : 'wrong',
    });
    setTimeout(() => setClickEffect(null), 300);

    // Send click to server
    sendMessage({ type: 'click' });
  }, [ball, sendMessage]);

  // Timer bar ratio
  const timerRatio = ball ? ball.timeLeft / ball.timeLimit : 0;

  return (
    <div className="battle-container">
      {/* Score Board */}
      <div className="battle-score-board">
        <div className="player-info me">
          <span className="player-label">나</span>
          <span className="player-name">{nickname}</span>
          <span className="player-score">{myScore}</span>
        </div>
        
        <div className="time-display">
          <span className="time-value">{timeLeft.toFixed(1)}</span>
          <span className="time-label">초</span>
        </div>
        
        <div className="player-info opponent">
          <span className="player-label">상대</span>
          <span className="player-name">{opponentNickname}</span>
          <span className="player-score">{opponentScore}</span>
        </div>
      </div>

      {/* Game Area */}
      <div
        className="battle-game-area"
        ref={gameAreaRef}
        style={{ width: GAME_WIDTH, height: GAME_HEIGHT }}
      >
        {/* Timer Bar */}
        {ball && (
          <div className="timer-bar-container">
            <div
              className={`timer-bar ${timerRatio < 0.3 ? 'danger' : ''}`}
              style={{ width: `${timerRatio * 100}%` }}
            />
          </div>
        )}

        {/* Ball */}
        {ball && (
          <div
            className={`battle-ball ${ball.isRed ? 'red' : 'blue'}`}
            style={{
              left: ball.x - ball.size / 2,
              top: ball.y - ball.size / 2,
              width: ball.size,
              height: ball.size,
            }}
            onClick={handleBallClick}
          />
        )}

        {/* Click Effect */}
        {clickEffect && (
          <div
            className={`click-effect ${clickEffect.type}`}
            style={{ left: clickEffect.x, top: clickEffect.y }}
          >
            {clickEffect.type === 'success' ? '+1' : '-1'}
          </div>
        )}

        {/* Last Result */}
        {lastResult && (
          <div className="last-result">
            {lastResult.clickedBy === 'none' ? (
              <span className="result-text timeout">시간 초과!</span>
            ) : (
              <span className={`result-text ${lastResult.isMe ? 'me' : 'opponent'}`}>
                {lastResult.clickedBy} 클릭!
              </span>
            )}
          </div>
        )}

        {/* Waiting for ball */}
        {!ball && !lastResult && (
          <div className="waiting-ball">
            <span>준비...</span>
          </div>
        )}
      </div>

      {/* Instructions */}
      <div className="battle-instructions">
        <p><span className="ball-icon red">●</span> 빨간 공 = +1점</p>
        <p><span className="ball-icon blue">●</span> 파란 공 = -1점</p>
        <p>먼저 클릭한 사람이 점수를 가져갑니다!</p>
      </div>
    </div>
  );
}

export default SpeedClickBattle;

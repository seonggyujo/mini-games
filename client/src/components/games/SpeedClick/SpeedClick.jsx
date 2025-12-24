import { useState, useEffect, useCallback, useRef } from 'react';
import './SpeedClick.css';

const GAME_WIDTH = 1200;
const GAME_HEIGHT = 800;

// 레벨 설정
const LEVELS = [
  { level: 1, timeLimit: 1.00, ballSize: 80, blueChance: 0.10, requiredScore: 0 },
  { level: 2, timeLimit: 0.90, ballSize: 75, blueChance: 0.13, requiredScore: 5 },
  { level: 3, timeLimit: 0.80, ballSize: 70, blueChance: 0.16, requiredScore: 10 },
  { level: 4, timeLimit: 0.70, ballSize: 65, blueChance: 0.20, requiredScore: 15 },
  { level: 5, timeLimit: 0.60, ballSize: 60, blueChance: 0.23, requiredScore: 20 },
  { level: 6, timeLimit: 0.50, ballSize: 55, blueChance: 0.26, requiredScore: 25 },
  { level: 7, timeLimit: 0.40, ballSize: 50, blueChance: 0.30, requiredScore: 30 },
];

function SpeedClick() {
  const [gameState, setGameState] = useState('ready'); // ready, playing, gameover
  const [score, setScore] = useState(0);
  const [lives, setLives] = useState(3);
  const [currentLevel, setCurrentLevel] = useState(1);
  const [ball, setBall] = useState(null); // { x, y, isRed, timeLeft, maxTime }
  const [showNicknameModal, setShowNicknameModal] = useState(false);
  const [nickname, setNickname] = useState('');
  const [highScore, setHighScore] = useState(() => {
    return parseInt(localStorage.getItem('speedclick-highscore') || '0');
  });
  const [clickEffect, setClickEffect] = useState(null); // { x, y, type }

  const ballTimerRef = useRef(null);
  const gameAreaRef = useRef(null);

  // 현재 레벨 설정 가져오기
  const getLevelConfig = useCallback(() => {
    for (let i = LEVELS.length - 1; i >= 0; i--) {
      if (score >= LEVELS[i].requiredScore) {
        return LEVELS[i];
      }
    }
    return LEVELS[0];
  }, [score]);

  // 레벨 업데이트
  useEffect(() => {
    const config = getLevelConfig();
    if (config.level !== currentLevel) {
      setCurrentLevel(config.level);
    }
  }, [score, currentLevel, getLevelConfig]);

  // 새 공 생성
  const spawnBall = useCallback(() => {
    const config = getLevelConfig();
    const isRed = Math.random() > config.blueChance;
    const padding = config.ballSize;
    
    const x = padding + Math.random() * (GAME_WIDTH - padding * 2);
    const y = padding + Math.random() * (GAME_HEIGHT - padding * 2);

    setBall({
      x,
      y,
      isRed,
      timeLeft: config.timeLimit,
      maxTime: config.timeLimit,
      size: config.ballSize
    });
  }, [getLevelConfig]);

  // 공 타이머 (시간 감소)
  useEffect(() => {
    if (gameState !== 'playing' || !ball) return;

    ballTimerRef.current = setInterval(() => {
      setBall(prev => {
        if (!prev) return null;
        
        const newTimeLeft = prev.timeLeft - 0.016; // ~60fps
        
        if (newTimeLeft <= 0) {
          // 시간 초과
          if (prev.isRed) {
            // 빨간 공을 놓침 - 목숨 감소
            setLives(l => {
              const newLives = l - 1;
              if (newLives <= 0) {
                setGameState('gameover');
              }
              return newLives;
            });
            setClickEffect({ x: prev.x, y: prev.y, type: 'miss' });
            setTimeout(() => setClickEffect(null), 300);
          }
          // 파란 공을 놓침 - 안전
          return null;
        }
        
        return { ...prev, timeLeft: newTimeLeft };
      });
    }, 16);

    return () => clearInterval(ballTimerRef.current);
  }, [gameState, ball]);

  // 공이 없으면 새로 생성
  useEffect(() => {
    if (gameState !== 'playing') return;
    if (!ball) {
      const timeout = setTimeout(spawnBall, 300);
      return () => clearTimeout(timeout);
    }
  }, [gameState, ball, spawnBall]);

  // 공 클릭 처리
  const handleBallClick = (e) => {
    e.stopPropagation();
    if (gameState !== 'playing' || !ball) return;

    const config = getLevelConfig();

    if (ball.isRed) {
      // 빨간 공 클릭 - 점수 획득
      const timeRatio = ball.timeLeft / ball.maxTime;
      let points = config.level;
      
      // 시간 보너스
      if (timeRatio >= 0.75) {
        points += 2; // 25% 이내 클릭
      } else if (timeRatio >= 0.50) {
        points += 1; // 50% 이내 클릭
      }
      
      setScore(s => s + points);
      setClickEffect({ x: ball.x, y: ball.y, type: 'success', points });
    } else {
      // 파란 공 클릭 - 목숨 감소
      setLives(l => {
        const newLives = l - 1;
        if (newLives <= 0) {
          setGameState('gameover');
        }
        return newLives;
      });
      setClickEffect({ x: ball.x, y: ball.y, type: 'wrong' });
    }
    
    setTimeout(() => setClickEffect(null), 300);
    setBall(null);
  };

  // 게임 영역 클릭 (공 외 영역)
  const handleAreaClick = () => {
    if (gameState === 'ready') {
      startGame();
    } else if (gameState === 'gameover') {
      resetGame();
    }
  };

  // 게임 시작
  const startGame = () => {
    setGameState('playing');
    setScore(0);
    setLives(3);
    setCurrentLevel(1);
    setBall(null);
  };

  // 게임 리셋
  const resetGame = () => {
    setGameState('ready');
    setScore(0);
    setLives(3);
    setCurrentLevel(1);
    setBall(null);
  };

  // 게임 오버 시 최고 점수 체크
  useEffect(() => {
    if (gameState === 'gameover' && score > highScore) {
      setHighScore(score);
      localStorage.setItem('speedclick-highscore', score.toString());
      setShowNicknameModal(true);
    }
  }, [gameState, score, highScore]);

  // 점수 제출
  const submitScore = async () => {
    if (!nickname.trim()) return;
    
    try {
      await fetch('/api/scores', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nickname: nickname.trim(),
          game: 'speed-click',
          score: score
        })
      });
    } catch (error) {
      console.error('Failed to submit score:', error);
    }
    
    setShowNicknameModal(false);
    setNickname('');
  };

  // 타이머 바 비율
  const timerRatio = ball ? ball.timeLeft / ball.maxTime : 0;

  return (
    <div className="speed-click-container">
      <div className="score-board">
        <div className="lives-display">
          <span>LIVES</span>
          <span className="lives-value">
            {[...Array(3)].map((_, i) => (
              <span key={i} className={`heart ${i < lives ? 'active' : ''}`}>
                {i < lives ? '❤️' : '🖤'}
              </span>
            ))}
          </span>
        </div>
        <div className="current-score">
          <span>SCORE</span>
          <span className="score-value">{score}</span>
        </div>
        <div className="level-display">
          <span>LEVEL</span>
          <span className="level-value">{currentLevel}</span>
        </div>
        <div className="high-score">
          <span>BEST</span>
          <span className="best-value">{highScore}</span>
        </div>
      </div>

      <div 
        className="game-area"
        ref={gameAreaRef}
        style={{ width: GAME_WIDTH, height: GAME_HEIGHT }}
        onClick={handleAreaClick}
      >
        {/* 타이머 바 */}
        {ball && gameState === 'playing' && (
          <div className="timer-bar-container">
            <div 
              className={`timer-bar ${timerRatio < 0.3 ? 'danger' : ''}`}
              style={{ width: `${timerRatio * 100}%` }}
            />
          </div>
        )}

        {/* 공 */}
        {ball && gameState === 'playing' && (
          <div
            className={`ball ${ball.isRed ? 'red' : 'blue'}`}
            style={{
              left: ball.x - ball.size / 2,
              top: ball.y - ball.size / 2,
              width: ball.size,
              height: ball.size,
            }}
            onClick={handleBallClick}
          />
        )}

        {/* 클릭 효과 */}
        {clickEffect && (
          <div 
            className={`click-effect ${clickEffect.type}`}
            style={{ left: clickEffect.x, top: clickEffect.y }}
          >
            {clickEffect.type === 'success' && `+${clickEffect.points}`}
            {clickEffect.type === 'wrong' && '-1'}
            {clickEffect.type === 'miss' && 'MISS!'}
          </div>
        )}

        {/* 게임 상태 오버레이 */}
        {gameState === 'ready' && (
          <div className="game-overlay">
            <h2 className="pixel-font">SPEED CLICK</h2>
            <div className="rules">
              <p><span className="ball-icon red">●</span> 빨간 공 - 클릭하면 점수!</p>
              <p><span className="ball-icon blue">●</span> 파란 공 - 클릭하면 안돼!</p>
            </div>
            <p className="hint">클릭으로 시작</p>
          </div>
        )}

        {gameState === 'gameover' && (
          <div className="game-overlay gameover">
            <h2 className="pixel-font">GAME OVER</h2>
            <p className="final-score">최종 점수: {score}</p>
            <p className="final-level">도달 레벨: {currentLevel}</p>
            <p className="hint">클릭으로 재시작</p>
          </div>
        )}
      </div>

      <div className="game-instructions">
        <p><strong>규칙:</strong> 빨간 공은 빠르게 클릭! 파란 공은 무시!</p>
        <p><strong>팁:</strong> 빠르게 클릭할수록 보너스 점수!</p>
      </div>

      {/* 닉네임 모달 */}
      {showNicknameModal && (
        <div className="modal-overlay">
          <div className="modal">
            <h3 className="pixel-font">NEW HIGH SCORE!</h3>
            <p>점수: {score}</p>
            <input
              type="text"
              placeholder="닉네임 입력"
              value={nickname}
              onChange={(e) => setNickname(e.target.value.slice(0, 20))}
              onKeyDown={(e) => e.key === 'Enter' && submitScore()}
              autoFocus
            />
            <div className="modal-buttons">
              <button onClick={submitScore} className="submit-btn">등록</button>
              <button onClick={() => setShowNicknameModal(false)} className="cancel-btn">취소</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default SpeedClick;

import { useState, useEffect, useCallback, useRef } from 'react';
import NicknameModal from '../../common/NicknameModal';
import useHighScore from '../../../hooks/useHighScore';
import { useSpeedClickSession, generateBall, LEVELS } from '../../../hooks/useGameSession';
import BattleLobby from './BattleLobby';
import './SpeedClick.css';

const GAME_WIDTH = 1200;
const GAME_HEIGHT = 800;
const SPAWN_DELAY = 300;

function SpeedClick() {
  const [gameState, setGameState] = useState('ready'); // ready, playing, gameover, battle
  const [score, setScore] = useState(0);
  const [lives, setLives] = useState(3);
  const [currentLevel, setCurrentLevel] = useState(1);
  const [ball, setBall] = useState(null);
  const [showNicknameModal, setShowNicknameModal] = useState(false);
  const [highScore, , checkAndUpdateHighScore] = useHighScore('speedclick');
  const [clickEffect, setClickEffect] = useState(null);

  const ballTimerRef = useRef(null);
  const gameAreaRef = useRef(null);
  const ballIndexRef = useRef(0);
  const prevBallEndTimeRef = useRef(0);
  const gameStartTimeRef = useRef(null);

  // 서버 세션 훅
  const {
    sessionId,
    seed,
    startSession,
    reportClick,
    reportMiss,
    submitScore,
    resetSession,
  } = useSpeedClickSession();

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

  // 새 공 생성 (시드 기반)
  const spawnBall = useCallback(() => {
    if (!seed) return;

    const newBall = generateBall(seed, ballIndexRef.current, score, prevBallEndTimeRef.current);
    
    setBall({
      x: newBall.x,
      y: newBall.y,
      isRed: newBall.isRed,
      timeLeft: newBall.duration / 1000,
      maxTime: newBall.duration / 1000,
      size: newBall.size,
      index: newBall.index,
    });
  }, [seed, score]);

  // 공 타이머 (시간 감소)
  useEffect(() => {
    if (gameState !== 'playing' || !ball) return;

    ballTimerRef.current = setInterval(() => {
      setBall(prev => {
        if (!prev) return null;
        
        const newTimeLeft = prev.timeLeft - 0.016;
        
        if (newTimeLeft <= 0) {
          // 시간 초과 - 서버에 miss 보고
          if (prev.isRed) {
            // 빨간 공을 놓침
            reportMiss(prev.index).then(response => {
              if (response.valid) {
                setLives(response.lives);
                if (response.gameOver) {
                  setGameState('gameover');
                }
              }
            }).catch(console.error);

            setClickEffect({ x: prev.x, y: prev.y, type: 'miss' });
            setTimeout(() => setClickEffect(null), 300);
          } else {
            // 파란 공을 놓침 - 서버에도 알림 (인덱스 동기화)
            reportMiss(prev.index).catch(console.error);
          }

          // 다음 공 준비
          const currentBallEndTime = prevBallEndTimeRef.current + SPAWN_DELAY + prev.maxTime * 1000;
          prevBallEndTimeRef.current = currentBallEndTime;
          ballIndexRef.current++;

          return null;
        }
        
        return { ...prev, timeLeft: newTimeLeft };
      });
    }, 16);

    return () => clearInterval(ballTimerRef.current);
  }, [gameState, ball, reportMiss]);

  // 공이 없으면 새로 생성
  useEffect(() => {
    if (gameState !== 'playing') return;
    if (!ball && seed) {
      const timeout = setTimeout(spawnBall, 300);
      return () => clearTimeout(timeout);
    }
  }, [gameState, ball, seed, spawnBall]);

  // 공 클릭 처리
  const handleBallClick = async (e) => {
    e.stopPropagation();
    if (gameState !== 'playing' || !ball) return;

    try {
      // 서버에 클릭 보고
      const response = await reportClick(ball.index);

      if (response.valid) {
        if (ball.isRed) {
          // 빨간 공: 서버에서 받은 점수 사용
          setScore(response.score);
          setClickEffect({ x: ball.x, y: ball.y, type: 'success', points: response.points });
        } else {
          // 파란 공: 목숨 감소
          setLives(response.lives);
          setClickEffect({ x: ball.x, y: ball.y, type: 'wrong' });
        }

        if (response.gameOver) {
          setGameState('gameover');
        }
      }
    } catch (err) {
      console.error('Click report failed:', err);
    }

    // 다음 공 준비
    const currentBallEndTime = prevBallEndTimeRef.current + SPAWN_DELAY + ball.maxTime * 1000;
    prevBallEndTimeRef.current = currentBallEndTime;
    ballIndexRef.current++;

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

  // 대결 모드 시작
  const handleBattleMode = () => {
    setGameState('battle');
  };

  // 대결 모드에서 돌아오기
  const handleBackFromBattle = () => {
    setGameState('ready');
  };

  // 게임 시작
  const startGame = async () => {
    try {
      // 서버 세션 시작
      await startSession();
      
      setGameState('playing');
      setScore(0);
      setLives(3);
      setCurrentLevel(1);
      setBall(null);
      ballIndexRef.current = 0;
      prevBallEndTimeRef.current = 0;
      gameStartTimeRef.current = Date.now();
    } catch (err) {
      console.error('Failed to start game:', err);
      alert('게임 시작에 실패했습니다. 다시 시도해주세요.');
    }
  };

  // 게임 리셋
  const resetGame = () => {
    resetSession();
    setGameState('ready');
    setScore(0);
    setLives(3);
    setCurrentLevel(1);
    setBall(null);
    ballIndexRef.current = 0;
    prevBallEndTimeRef.current = 0;
  };

  // 게임 오버 시 최고 점수 체크
  useEffect(() => {
    if (gameState === 'gameover' && checkAndUpdateHighScore(score)) {
      setShowNicknameModal(true);
    }
  }, [gameState, score, checkAndUpdateHighScore]);

  // 닉네임 제출 핸들러
  const handleNicknameSubmit = async (nickname) => {
    try {
      const result = await submitScore(nickname);
      if (result.success) {
        setShowNicknameModal(false);
      } else {
        throw new Error('Score submission failed');
      }
    } catch (err) {
      console.error('Failed to submit score:', err);
      throw err;
    }
  };

  // 타이머 바 비율
  const timerRatio = ball ? ball.timeLeft / ball.maxTime : 0;

  // 대결 모드 화면
  if (gameState === 'battle') {
    return <BattleLobby onBack={handleBackFromBattle} />;
  }

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
            <button className="battle-mode-btn" onClick={handleBattleMode}>
              대결 모드
            </button>
            <div className="rules">
              <p><span className="ball-icon red">●</span> 빨간 공 - 클릭하면 점수!</p>
              <p><span className="ball-icon blue">●</span> 파란 공 - 클릭하면 안돼!</p>
            </div>
            <p className="hint">클릭으로 솔로 모드 시작</p>
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
        <NicknameModal
          score={score}
          gameName="speed-click"
          sessionId={sessionId}
          onSubmit={handleNicknameSubmit}
          onClose={() => setShowNicknameModal(false)}
        />
      )}
    </div>
  );
}

export default SpeedClick;

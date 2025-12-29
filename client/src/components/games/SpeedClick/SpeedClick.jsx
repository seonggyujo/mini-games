import { useState, useEffect, useCallback, useRef } from 'react';
import NicknameModal from '../../common/NicknameModal';
import useHighScore from '../../../hooks/useHighScore';
import { useSpeedClickSession, LEVELS } from '../../../hooks/useGameSession';
import BattleLobby from './BattleLobby';
import './SpeedClick.css';

const GAME_WIDTH = 1200;
const GAME_HEIGHT = 800;

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
  const spawnTimeoutRef = useRef(null); // 공 생성 타이머
  const gameStartTimeRef = useRef(null);

  // 서버 세션 훅
  const {
    sessionId,
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

  // 새 공 생성 (서버에서 받은 정보 사용)
  const spawnBall = useCallback((ballInfo) => {
    if (!ballInfo) return;

    setBall({
      x: ballInfo.x,
      y: ballInfo.y,
      isRed: ballInfo.isRed,
      timeLeft: ballInfo.duration / 1000,
      maxTime: ballInfo.duration / 1000,
      size: ballInfo.size,
      index: ballInfo.index,
    });
  }, []);

  // 300ms 후 공 생성 스케줄
  const scheduleNextBall = useCallback((ballInfo) => {
    // 기존 타이머 취소
    if (spawnTimeoutRef.current) {
      clearTimeout(spawnTimeoutRef.current);
    }
    
    spawnTimeoutRef.current = setTimeout(() => {
      spawnBall(ballInfo);
      spawnTimeoutRef.current = null;
    }, 300);
  }, [spawnBall]);

  // 공 타이머 (시간 감소 + 시간 초과 처리)
  useEffect(() => {
    if (gameState !== 'playing' || !ball) return;

    const timer = setInterval(() => {
      setBall(prev => {
        if (!prev) return null;
        
        const newTimeLeft = prev.timeLeft - 0.016; // 16ms마다 감소
        
        if (newTimeLeft <= 0) {
          // 시간 초과 - 서버에 miss 보고
          reportMiss(prev.index).then(response => {
            if (response.valid) {
              if (response.isRed) {
                setLives(response.lives);
                setClickEffect({ x: prev.x, y: prev.y, type: 'miss' });
                setTimeout(() => setClickEffect(null), 300);
              }
              if (response.gameOver) {
                setGameState('gameover');
              } else if (response.nextBall) {
                // 다음 공 스케줄
                scheduleNextBall(response.nextBall);
              }
            }
          }).catch(console.error);
          
          return null; // 공 제거
        }
        
        return { ...prev, timeLeft: newTimeLeft };
      });
    }, 16);

    return () => clearInterval(timer);
  }, [gameState, ball, reportMiss, scheduleNextBall]);

  // 공 클릭 처리
  const handleBallClick = async (e) => {
    e.stopPropagation();
    if (gameState !== 'playing' || !ball) return;

    const clickedBall = ball;
    setBall(null); // 즉시 공 제거
    
    try {
      // 서버에 클릭 보고
      const response = await reportClick(clickedBall.index);

      if (response.valid) {
        // 서버의 isRed 값을 신뢰
        if (response.isRed) {
          // 빨간 공: 서버에서 받은 점수 사용
          setScore(response.score);
          setClickEffect({ x: clickedBall.x, y: clickedBall.y, type: 'success', points: response.points });
        } else {
          // 파란 공: 목숨 감소
          setLives(response.lives);
          setClickEffect({ x: clickedBall.x, y: clickedBall.y, type: 'wrong' });
        }

        if (response.gameOver) {
          setGameState('gameover');
        } else if (response.nextBall) {
          // 다음 공 스케줄
          scheduleNextBall(response.nextBall);
        }
      }
    } catch (err) {
      console.error('Click report failed:', err);
    }

    setTimeout(() => setClickEffect(null), 300);
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
  const handleBattleMode = (e) => {
    e.stopPropagation();  // 이벤트 버블링 방지
    setGameState('battle');
  };

  // 대결 모드에서 돌아오기
  const handleBackFromBattle = () => {
    setGameState('ready');
  };

  // 게임 시작
  const startGame = async () => {
    try {
      // 서버 세션 시작 및 첫 번째 공 정보 받기
      const sessionData = await startSession();
      
      setGameState('playing');
      setScore(0);
      setLives(3);
      setCurrentLevel(1);
      setBall(null);
      gameStartTimeRef.current = Date.now();
      
      // 첫 번째 공 스케줄
      scheduleNextBall(sessionData.nextBall);
    } catch (err) {
      console.error('Failed to start game:', err);
      alert('게임 시작에 실패했습니다. 다시 시도해주세요.');
    }
  };

  // 게임 리셋
  const resetGame = () => {
    // 타이머 정리
    if (spawnTimeoutRef.current) {
      clearTimeout(spawnTimeoutRef.current);
      spawnTimeoutRef.current = null;
    }
    
    resetSession();
    setGameState('ready');
    setScore(0);
    setLives(3);
    setCurrentLevel(1);
    setBall(null);
  };

  // 컴포넌트 언마운트 시 타이머 정리
  useEffect(() => {
    return () => {
      if (spawnTimeoutRef.current) {
        clearTimeout(spawnTimeoutRef.current);
      }
    };
  }, []);

  // 게임 오버 시 닉네임 입력 모달 표시
  useEffect(() => {
    if (gameState === 'gameover' && score > 0) {
      checkAndUpdateHighScore(score); // 로컬 최고점수 업데이트
      setShowNicknameModal(true); // 항상 닉네임 입력 기회 제공
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

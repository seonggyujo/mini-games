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

  const gameAreaRef = useRef(null);
  const spawnTimeoutRef = useRef(null); // 공 생성 타이머
  const timerRef = useRef(null); // 공 카운트다운 타이머

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

  // 다음 공 스케줄 (300ms 후 생성)
  const scheduleNextBall = useCallback((ballInfo) => {
    console.log('[scheduleNextBall] called with:', ballInfo);
    if (!ballInfo) {
      console.log('[scheduleNextBall] ballInfo is null, returning');
      return;
    }
    
    // 기존 타이머 취소
    if (spawnTimeoutRef.current) {
      console.log('[scheduleNextBall] clearing existing timeout');
      clearTimeout(spawnTimeoutRef.current);
    }
    
    console.log('[scheduleNextBall] scheduling new ball in 300ms');
    spawnTimeoutRef.current = setTimeout(() => {
      console.log('[scheduleNextBall] timeout fired, setting ball index:', ballInfo.index);
      setBall({
        x: ballInfo.x,
        y: ballInfo.y,
        isRed: ballInfo.isRed,
        timeLeft: ballInfo.duration / 1000,
        maxTime: ballInfo.duration / 1000,
        size: ballInfo.size,
        index: ballInfo.index,
      });
      spawnTimeoutRef.current = null;
    }, 300);
  }, []);

  // miss 처리 함수 (ref에 저장하여 타이머에서 접근 가능하게)
  const processMissRef = useRef(null);
  
  processMissRef.current = useCallback((missedBall) => {
    console.log('[processMiss] processing missed ball index:', missedBall.index);
    
    reportMiss(missedBall.index).then(response => {
      console.log('[processMiss] reportMiss response:', response);
      if (response.valid) {
        if (response.isRed) {
          setLives(response.lives);
          setClickEffect({ x: missedBall.x, y: missedBall.y, type: 'miss' });
          setTimeout(() => setClickEffect(null), 300);
        }
        if (response.gameOver) {
          console.log('[processMiss] game over');
          setGameState('gameover');
        } else if (response.nextBall) {
          console.log('[processMiss] scheduling next ball');
          scheduleNextBall(response.nextBall);
        } else {
          console.log('[processMiss] NO nextBall in response!');
        }
      }
    }).catch(console.error);
  }, [reportMiss, scheduleNextBall]);

  // 공 타이머 - 시간 감소 및 만료 시 miss 처리 직접 호출
  useEffect(() => {
    if (gameState !== 'playing' || !ball) return;

    timerRef.current = setInterval(() => {
      setBall(prev => {
        if (!prev) return null;
        
        const newTimeLeft = prev.timeLeft - 0.016;
        
        if (newTimeLeft <= 0) {
          // miss 처리 직접 호출 (다음 틱에서 실행하여 state 업데이트 완료 후)
          const missedBall = prev;
          setTimeout(() => {
            if (processMissRef.current) {
              processMissRef.current(missedBall);
            }
          }, 0);
          return null; // 공 제거
        }
        
        return { ...prev, timeLeft: newTimeLeft };
      });
    }, 16);

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [gameState, ball]);

  // 공 클릭 처리
  const handleBallClick = async (e) => {
    e.stopPropagation();
    if (gameState !== 'playing' || !ball) return;

    const clickedBall = ball;
    console.log('[handleBallClick] clicked ball index:', clickedBall.index);
    setBall(null); // 즉시 공 제거
    
    try {
      console.log('[handleBallClick] calling reportClick...');
      const response = await reportClick(clickedBall.index);
      console.log('[handleBallClick] reportClick response:', response);

      if (response.valid) {
        if (response.isRed) {
          setScore(response.score);
          setClickEffect({ x: clickedBall.x, y: clickedBall.y, type: 'success', points: response.points });
        } else {
          setLives(response.lives);
          setClickEffect({ x: clickedBall.x, y: clickedBall.y, type: 'wrong' });
        }

        if (response.gameOver) {
          console.log('[handleBallClick] game over');
          setGameState('gameover');
        } else if (response.nextBall) {
          console.log('[handleBallClick] scheduling next ball');
          scheduleNextBall(response.nextBall);
        } else {
          console.log('[handleBallClick] NO nextBall in response!');
        }
      } else {
        // valid: false여도 게임 계속 진행 - 서버에 현재 상태 재요청
        console.log('[handleBallClick] response.valid is false, message:', response.message);
        // 다음 공이 있으면 스케줄
        if (response.nextBall) {
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
    e.stopPropagation();
    setGameState('battle');
  };

  // 대결 모드에서 돌아오기
  const handleBackFromBattle = () => {
    setGameState('ready');
  };

  // 게임 시작
  const startGame = async () => {
    try {
      console.log('[startGame] starting...');
      const sessionData = await startSession();
      console.log('[startGame] sessionData:', sessionData);
      
      setGameState('playing');
      setScore(0);
      setLives(3);
      setCurrentLevel(1);
      setBall(null);
      
      // 첫 번째 공 스케줄
      console.log('[startGame] scheduling first ball');
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
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
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
      if (spawnTimeoutRef.current) clearTimeout(spawnTimeoutRef.current);
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  // 게임 오버 시 닉네임 입력 모달 표시
  useEffect(() => {
    if (gameState === 'gameover' && score > 0) {
      checkAndUpdateHighScore(score);
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

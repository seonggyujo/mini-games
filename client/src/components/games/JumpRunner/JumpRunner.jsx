import { useState, useEffect, useCallback, useRef } from 'react';
import NicknameModal from '../../common/NicknameModal';
import useHighScore from '../../../hooks/useHighScore';
import './JumpRunner.css';

const GAME_WIDTH = 1200;
const GAME_HEIGHT = 600;
const GROUND_HEIGHT = 40;
const PLAYER_WIDTH = 70;
const PLAYER_HEIGHT = 70;
const OBSTACLE_WIDTH = 60;
const OBSTACLE_HEIGHT = 75;
const GRAVITY = 0.8;
const JUMP_FORCE = 18;
const INITIAL_SPEED = 6;
const SPEED_INCREMENT = 0.002;

function JumpRunner() {
  const [gameState, setGameState] = useState('ready'); // ready, playing, gameover
  const [score, setScore] = useState(0);
  const [highScore, , checkAndUpdateHighScore] = useHighScore('jumprunner');
  const [playerY, setPlayerY] = useState(0);
  const [obstacles, setObstacles] = useState([]);
  const [speed, setSpeed] = useState(INITIAL_SPEED);
  const [showNicknameModal, setShowNicknameModal] = useState(false);
  
  const playerVelocityRef = useRef(0);
  const isJumpingRef = useRef(false);
  const gameLoopRef = useRef(null);
  const obstacleTimerRef = useRef(0);
  const lastTimeRef = useRef(0);
  const speedRef = useRef(INITIAL_SPEED);
  const scoreRef = useRef(0);

  // Jump handler
  const jump = useCallback(() => {
    if (!isJumpingRef.current && gameState === 'playing') {
      playerVelocityRef.current = JUMP_FORCE;
      isJumpingRef.current = true;
    }
  }, [gameState]);

  // Keyboard event handler
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.code === 'Space' || e.code === 'ArrowUp') {
        e.preventDefault();
        if (gameState === 'ready') {
          startGame();
        } else if (gameState === 'playing') {
          jump();
        } else if (gameState === 'gameover') {
          resetGame();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [gameState, jump]);

  // Start game
  const startGame = () => {
    setGameState('playing');
    setScore(0);
    setPlayerY(0);
    setObstacles([]);
    setSpeed(INITIAL_SPEED);
    playerVelocityRef.current = 0;
    isJumpingRef.current = false;
    obstacleTimerRef.current = 0;
    lastTimeRef.current = 0;
    speedRef.current = INITIAL_SPEED;
    scoreRef.current = 0;
  };

  // Reset game
  const resetGame = () => {
    setGameState('ready');
    setScore(0);
    setPlayerY(0);
    setObstacles([]);
    setSpeed(INITIAL_SPEED);
    lastTimeRef.current = 0;
    speedRef.current = INITIAL_SPEED;
    scoreRef.current = 0;
  };

  // Game loop
  useEffect(() => {
    if (gameState !== 'playing') return;

    const gameLoop = (currentTime) => {
      // Delta time 계산 (60fps 기준 정규화)
      if (lastTimeRef.current === 0) {
        lastTimeRef.current = currentTime;
      }
      const deltaTime = Math.min((currentTime - lastTimeRef.current) / 16.67, 3); // 최대 3프레임으로 제한
      lastTimeRef.current = currentTime;

      // Update player position (gravity and jump)
      setPlayerY(prev => {
        let newVelocity = playerVelocityRef.current - (GRAVITY * deltaTime);
        playerVelocityRef.current = newVelocity;
        
        let newY = prev + (newVelocity * deltaTime);
        
        // 땅에 닿으면 멈춤
        if (newY <= 0) {
          newY = 0;
          playerVelocityRef.current = 0;
          isJumpingRef.current = false;
        }
        return newY;
      });

      // Update obstacles
      setObstacles(prev => {
        const updated = prev
          .map(obs => ({ ...obs, x: obs.x - (speedRef.current * deltaTime) }))
          .filter(obs => obs.x > -OBSTACLE_WIDTH);
        return updated;
      });

      // Spawn new obstacles
      obstacleTimerRef.current += deltaTime;
      if (obstacleTimerRef.current > 100 - Math.min(speedRef.current * 5, 50)) {
        if (Math.random() < 0.02 + speedRef.current * 0.005) {
          setObstacles(prev => [...prev, { 
            x: GAME_WIDTH, 
            id: Date.now(),
            height: OBSTACLE_HEIGHT + Math.random() * 15
          }]);
          obstacleTimerRef.current = 0;
        }
      }

      // Increase speed over time
      speedRef.current += SPEED_INCREMENT * deltaTime;
      setSpeed(speedRef.current);

      // Update score
      scoreRef.current += deltaTime;
      setScore(Math.floor(scoreRef.current));

      gameLoopRef.current = requestAnimationFrame(gameLoop);
    };

    gameLoopRef.current = requestAnimationFrame(gameLoop);
    return () => cancelAnimationFrame(gameLoopRef.current);
  }, [gameState]);

  // Collision detection
  useEffect(() => {
    if (gameState !== 'playing') return;

    const playerLeft = 50;
    const playerRight = playerLeft + PLAYER_WIDTH - 10;
    const playerBottom = GROUND_HEIGHT + playerY;
    const playerTop = playerBottom + PLAYER_HEIGHT;

    for (const obs of obstacles) {
      const obsLeft = obs.x + 5;
      const obsRight = obs.x + OBSTACLE_WIDTH - 5;
      const obsBottom = GROUND_HEIGHT;
      const obsTop = GROUND_HEIGHT + obs.height;

      if (
        playerRight > obsLeft &&
        playerLeft < obsRight &&
        playerBottom < obsTop &&
        playerTop > obsBottom
      ) {
        // Collision detected
        setGameState('gameover');
        if (checkAndUpdateHighScore(score)) {
          setShowNicknameModal(true);
        }
        break;
      }
    }
  }, [obstacles, playerY, gameState, score, checkAndUpdateHighScore]);

  // Calculate speed level (every 2 speed increase = 1 level)
  const speedLevel = Math.floor((speed - INITIAL_SPEED) / 2) + 1;

  // formatScore for display (score / 10)
  const formatScore = (s) => Math.floor(s / 10);

  return (
    <div className="jump-runner-container">
      <div className="score-board">
        <div className="current-score">
          <span>SCORE</span>
          <span className="score-value">{formatScore(score)}</span>
        </div>
        <div className="speed-display">
          <span>LEVEL</span>
          <span className="level-value">{speedLevel}</span>
          <span className="speed-value">SPD: {speed.toFixed(1)}</span>
        </div>
        <div className="high-score">
          <span>BEST</span>
          <span className="score-value">{formatScore(highScore)}</span>
        </div>
      </div>

      <div 
        className="game-canvas"
        style={{ width: GAME_WIDTH, height: GAME_HEIGHT }}
        onClick={() => {
          if (gameState === 'ready') startGame();
          else if (gameState === 'playing') jump();
          else if (gameState === 'gameover') resetGame();
        }}
      >
        {/* Ground */}
        <div className="ground" style={{ height: GROUND_HEIGHT }} />

        {/* Player (Mouse) */}
        <div 
          className={`player ${isJumpingRef.current ? 'jumping' : ''}`}
          style={{
            left: 50,
            bottom: GROUND_HEIGHT + playerY,
            width: PLAYER_WIDTH,
            height: PLAYER_HEIGHT
          }}
        >
          <div className="mouse-body">
            <div className="mouse-ear left"></div>
            <div className="mouse-ear right"></div>
            <div className="mouse-eye"></div>
            <div className="mouse-nose"></div>
          </div>
          <div className="mouse-tail"></div>
        </div>

        {/* Obstacles (Cats) */}
        {obstacles.map(obs => (
          <div
            key={obs.id}
            className="obstacle"
            style={{
              left: obs.x,
              bottom: GROUND_HEIGHT,
              width: OBSTACLE_WIDTH,
              height: obs.height
            }}
          >
            <div className="cat-body">
              <div className="cat-ear left"></div>
              <div className="cat-ear right"></div>
              <div className="cat-eyes">
                <div className="cat-eye"></div>
                <div className="cat-eye"></div>
              </div>
              <div className="cat-whiskers">
                <div className="whisker"></div>
                <div className="whisker"></div>
                <div className="whisker"></div>
              </div>
            </div>
          </div>
        ))}

        {/* Game state overlays */}
        {gameState === 'ready' && (
          <div className="game-overlay">
            <h2 className="pixel-font">JUMP RUNNER</h2>
            <p>고양이를 피해 달려라!</p>
            <p className="hint">SPACE 또는 클릭으로 시작</p>
          </div>
        )}

        {gameState === 'gameover' && (
          <div className="game-overlay gameover">
            <h2 className="pixel-font">GAME OVER</h2>
            <p>점수: {formatScore(score)}</p>
            <p className="hint">SPACE 또는 클릭으로 재시작</p>
          </div>
        )}
      </div>

      <div className="game-instructions">
        <p><strong>조작법:</strong> SPACE, ↑ 또는 클릭으로 점프</p>
      </div>

      {/* Nickname modal */}
      {showNicknameModal && (
        <NicknameModal
          score={score}
          gameName="jump-runner"
          formatScore={formatScore}
          onSubmit={() => setShowNicknameModal(false)}
          onClose={() => setShowNicknameModal(false)}
        />
      )}
    </div>
  );
}

export default JumpRunner;

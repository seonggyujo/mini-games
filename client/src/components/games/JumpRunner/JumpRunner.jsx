import { useState, useEffect, useCallback, useRef } from 'react';
import './JumpRunner.css';

const GAME_WIDTH = 1200;
const GAME_HEIGHT = 450;
const GROUND_HEIGHT = 30;
const PLAYER_WIDTH = 50;
const PLAYER_HEIGHT = 50;
const OBSTACLE_WIDTH = 45;
const OBSTACLE_HEIGHT = 55;
const GRAVITY = 0.6;
const JUMP_FORCE = 14;
const INITIAL_SPEED = 6;
const SPEED_INCREMENT = 0.002;

function JumpRunner() {
  const [gameState, setGameState] = useState('ready'); // ready, playing, gameover
  const [score, setScore] = useState(0);
  const [highScore, setHighScore] = useState(() => {
    return parseInt(localStorage.getItem('jumprunner-highscore') || '0');
  });
  const [playerY, setPlayerY] = useState(0);
  const [obstacles, setObstacles] = useState([]);
  const [speed, setSpeed] = useState(INITIAL_SPEED);
  const [showNicknameModal, setShowNicknameModal] = useState(false);
  const [nickname, setNickname] = useState('');
  
  const playerVelocityRef = useRef(0);
  const isJumpingRef = useRef(false);
  const gameLoopRef = useRef(null);
  const obstacleTimerRef = useRef(0);

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
  };

  // Reset game
  const resetGame = () => {
    setGameState('ready');
    setScore(0);
    setPlayerY(0);
    setObstacles([]);
    setSpeed(INITIAL_SPEED);
  };

  // Game loop
  useEffect(() => {
    if (gameState !== 'playing') return;

    const gameLoop = () => {
      // Update player position (gravity and jump)
      setPlayerY(prev => {
        let newVelocity = playerVelocityRef.current - GRAVITY;
        playerVelocityRef.current = newVelocity;
        
        let newY = prev + newVelocity;
        
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
          .map(obs => ({ ...obs, x: obs.x - speed }))
          .filter(obs => obs.x > -OBSTACLE_WIDTH);
        return updated;
      });

      // Spawn new obstacles
      obstacleTimerRef.current++;
      if (obstacleTimerRef.current > 100 - Math.min(speed * 5, 50)) {
        if (Math.random() < 0.02 + speed * 0.005) {
          setObstacles(prev => [...prev, { 
            x: GAME_WIDTH, 
            id: Date.now(),
            height: OBSTACLE_HEIGHT + Math.random() * 15
          }]);
          obstacleTimerRef.current = 0;
        }
      }

      // Increase speed over time
      setSpeed(prev => prev + SPEED_INCREMENT);

      // Update score
      setScore(prev => prev + 1);

      gameLoopRef.current = requestAnimationFrame(gameLoop);
    };

    gameLoopRef.current = requestAnimationFrame(gameLoop);
    return () => cancelAnimationFrame(gameLoopRef.current);
  }, [gameState, speed]);

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
        if (score > highScore) {
          setHighScore(score);
          localStorage.setItem('jumprunner-highscore', score.toString());
          setShowNicknameModal(true);
        }
        break;
      }
    }
  }, [obstacles, playerY, gameState, score, highScore]);

  // Submit score to server
  const submitScore = async () => {
    if (!nickname.trim()) return;
    
    try {
      await fetch('/api/scores', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          nickname: nickname.trim(),
          game: 'jump-runner',
          score: score
        })
      });
    } catch (error) {
      console.error('Failed to submit score:', error);
    }
    
    setShowNicknameModal(false);
    setNickname('');
  };

  // Calculate speed level (every 2 speed increase = 1 level)
  const speedLevel = Math.floor((speed - INITIAL_SPEED) / 2) + 1;

  return (
    <div className="jump-runner-container">
      <div className="score-board">
        <div className="current-score">
          <span>SCORE</span>
          <span className="score-value">{Math.floor(score / 10)}</span>
        </div>
        <div className="speed-display">
          <span>LEVEL</span>
          <span className="level-value">{speedLevel}</span>
          <span className="speed-value">SPD: {speed.toFixed(1)}</span>
        </div>
        <div className="high-score">
          <span>BEST</span>
          <span className="score-value">{Math.floor(highScore / 10)}</span>
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
            <p>점수: {Math.floor(score / 10)}</p>
            <p className="hint">SPACE 또는 클릭으로 재시작</p>
          </div>
        )}
      </div>

      <div className="game-instructions">
        <p><strong>조작법:</strong> SPACE, ↑ 또는 클릭으로 점프</p>
      </div>

      {/* Nickname modal */}
      {showNicknameModal && (
        <div className="modal-overlay">
          <div className="modal">
            <h3 className="pixel-font">NEW HIGH SCORE!</h3>
            <p>점수: {Math.floor(score / 10)}</p>
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

export default JumpRunner;

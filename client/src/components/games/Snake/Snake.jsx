import { useState, useEffect, useCallback, useRef } from 'react';
import NicknameModal from '../../common/NicknameModal';
import useHighScore from '../../../hooks/useHighScore';
import './Snake.css';

const GRID_SIZE = 20;
const CELL_SIZE = 50;
const GAME_WIDTH = GRID_SIZE * CELL_SIZE;
const GAME_HEIGHT = GRID_SIZE * CELL_SIZE;

// 레벨 설정
const LEVEL_CONFIG = {
  easy: { speed: 200, speedIncrease: 0, minSpeed: 200, label: 'EASY', color: '#00b894' },
  medium: { speed: 150, speedIncrease: 5, minSpeed: 80, label: 'MEDIUM', color: '#fdcb6e' },
  hard: { speed: 100, speedIncrease: 5, minSpeed: 50, label: 'HARD', color: '#d63031' }
};

const DIRECTIONS = {
  UP: { x: 0, y: -1 },
  DOWN: { x: 0, y: 1 },
  LEFT: { x: -1, y: 0 },
  RIGHT: { x: 1, y: 0 }
};

const OPPOSITE = {
  UP: 'DOWN',
  DOWN: 'UP',
  LEFT: 'RIGHT',
  RIGHT: 'LEFT'
};

function Snake() {
  const [gameState, setGameState] = useState('ready'); // ready, levelSelect, playing, gameover
  const [level, setLevel] = useState(null);
  const [snake, setSnake] = useState([{ x: 10, y: 10 }]);
  const [direction, setDirection] = useState('RIGHT');
  const [food, setFood] = useState({ x: 15, y: 10 });
  const [score, setScore] = useState(0);
  const [highScore, , checkAndUpdateHighScore] = useHighScore('snake');
  const [showNicknameModal, setShowNicknameModal] = useState(false);
  const [currentSpeed, setCurrentSpeed] = useState(150);

  const gameLoopRef = useRef(null);
  const directionRef = useRef(direction);
  const directionQueueRef = useRef([]);

  // 방향 ref 동기화
  useEffect(() => {
    directionRef.current = direction;
  }, [direction]);

  // 랜덤 음식 생성
  const generateFood = useCallback((currentSnake) => {
    let newFood;
    do {
      newFood = {
        x: Math.floor(Math.random() * GRID_SIZE),
        y: Math.floor(Math.random() * GRID_SIZE)
      };
    } while (currentSnake.some(segment => segment.x === newFood.x && segment.y === newFood.y));
    return newFood;
  }, []);

  // 게임 시작
  const startGame = useCallback((selectedLevel) => {
    const config = LEVEL_CONFIG[selectedLevel];
    setLevel(selectedLevel);
    setSnake([{ x: 10, y: 10 }]);
    setDirection('RIGHT');
    directionRef.current = 'RIGHT';
    directionQueueRef.current = [];
    setFood({ x: 15, y: 10 });
    setScore(0);
    setCurrentSpeed(config.speed);
    setGameState('playing');
  }, []);

  // 게임 리셋
  const resetGame = () => {
    setGameState('levelSelect');
  };

  // 게임 루프
  useEffect(() => {
    if (gameState !== 'playing' || !level) return;

    const config = LEVEL_CONFIG[level];

    const moveSnake = () => {
      // 방향 큐에서 다음 방향 가져오기
      if (directionQueueRef.current.length > 0) {
        const nextDir = directionQueueRef.current.shift();
        if (nextDir !== OPPOSITE[directionRef.current]) {
          directionRef.current = nextDir;
          setDirection(nextDir);
        }
      }

      setSnake(prevSnake => {
        const head = prevSnake[0];
        const dir = DIRECTIONS[directionRef.current];
        const newHead = {
          x: head.x + dir.x,
          y: head.y + dir.y
        };

        // 벽 충돌 체크
        if (newHead.x < 0 || newHead.x >= GRID_SIZE || 
            newHead.y < 0 || newHead.y >= GRID_SIZE) {
          setGameState('gameover');
          return prevSnake;
        }

        // 자기 몸 충돌 체크
        if (prevSnake.some(segment => segment.x === newHead.x && segment.y === newHead.y)) {
          setGameState('gameover');
          return prevSnake;
        }

        const newSnake = [newHead, ...prevSnake];

        // 음식 먹기 체크
        setFood(prevFood => {
          if (newHead.x === prevFood.x && newHead.y === prevFood.y) {
            setScore(s => {
              const newScore = s + 10;
              // 속도 증가 (레벨에 따라)
              if (config.speedIncrease > 0) {
                setCurrentSpeed(speed => Math.max(config.minSpeed, speed - config.speedIncrease));
              }
              return newScore;
            });
            return generateFood(newSnake);
          }
          return prevFood;
        });

        // 음식을 먹지 않았으면 꼬리 제거
        setFood(prevFood => {
          if (newHead.x !== prevFood.x || newHead.y !== prevFood.y) {
            newSnake.pop();
          }
          return prevFood;
        });

        return newSnake;
      });
    };

    gameLoopRef.current = setInterval(moveSnake, currentSpeed);

    return () => clearInterval(gameLoopRef.current);
  }, [gameState, level, currentSpeed, generateFood]);

  // 키보드 입력 처리
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (gameState !== 'playing') return;

      let newDirection = null;

      switch (e.key) {
        case 'ArrowUp':
        case 'w':
        case 'W':
          newDirection = 'UP';
          break;
        case 'ArrowDown':
        case 's':
        case 'S':
          newDirection = 'DOWN';
          break;
        case 'ArrowLeft':
        case 'a':
        case 'A':
          newDirection = 'LEFT';
          break;
        case 'ArrowRight':
        case 'd':
        case 'D':
          newDirection = 'RIGHT';
          break;
        default:
          return;
      }

      e.preventDefault();

      // 마지막 방향(큐의 마지막 또는 현재 방향)과 반대 방향이 아닌지 확인
      const lastDirection = directionQueueRef.current.length > 0 
        ? directionQueueRef.current[directionQueueRef.current.length - 1]
        : directionRef.current;

      if (newDirection !== OPPOSITE[lastDirection] && newDirection !== lastDirection) {
        directionQueueRef.current.push(newDirection);
        // 큐가 너무 길어지지 않도록 제한
        if (directionQueueRef.current.length > 2) {
          directionQueueRef.current.shift();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [gameState]);

  // 모바일 컨트롤
  const handleMobileControl = (newDirection) => {
    if (gameState !== 'playing') return;

    const lastDirection = directionQueueRef.current.length > 0 
      ? directionQueueRef.current[directionQueueRef.current.length - 1]
      : directionRef.current;

    if (newDirection !== OPPOSITE[lastDirection] && newDirection !== lastDirection) {
      directionQueueRef.current.push(newDirection);
      if (directionQueueRef.current.length > 2) {
        directionQueueRef.current.shift();
      }
    }
  };

  // 게임 오버 시 최고 점수 체크
  useEffect(() => {
    if (gameState === 'gameover' && checkAndUpdateHighScore(score)) {
      setShowNicknameModal(true);
    }
  }, [gameState, score, checkAndUpdateHighScore]);

  return (
    <div className="snake-container">
      <div className="score-board">
        <div className="current-score">
          <span>SCORE</span>
          <span className="score-value">{score}</span>
        </div>
        {level && (
          <div className="level-display">
            <span>LEVEL</span>
            <span className="level-value" style={{ color: LEVEL_CONFIG[level].color }}>
              {LEVEL_CONFIG[level].label}
            </span>
          </div>
        )}
        <div className="high-score">
          <span>BEST</span>
          <span className="best-value">{highScore}</span>
        </div>
      </div>

      <div 
        className="game-area"
        style={{ 
          width: GAME_WIDTH, 
          height: GAME_HEIGHT,
          '--cell-size': `${CELL_SIZE}px`
        }}
      >
        {/* 그리드 배경은 CSS로 처리 */}

        {/* 뱀 */}
        {gameState === 'playing' && snake.map((segment, index) => (
          <div
            key={index}
            className={`snake-segment ${index === 0 ? 'head' : ''}`}
            style={{
              left: segment.x * CELL_SIZE,
              top: segment.y * CELL_SIZE,
              width: CELL_SIZE - 2,
              height: CELL_SIZE - 2,
            }}
          />
        ))}

        {/* 음식 */}
        {gameState === 'playing' && (
          <div
            className="food"
            style={{
              left: food.x * CELL_SIZE,
              top: food.y * CELL_SIZE,
              width: CELL_SIZE - 2,
              height: CELL_SIZE - 2,
            }}
          >
            🍎
          </div>
        )}

        {/* 시작 화면 */}
        {gameState === 'ready' && (
          <div className="game-overlay">
            <h2 className="pixel-font">SNAKE</h2>
            <p className="subtitle">클래식 스네이크 게임</p>
            <button className="start-btn" onClick={() => setGameState('levelSelect')}>
              게임 시작
            </button>
          </div>
        )}

        {/* 레벨 선택 */}
        {gameState === 'levelSelect' && (
          <div className="game-overlay">
            <h2 className="pixel-font">SELECT LEVEL</h2>
            <div className="level-buttons">
              <button 
                className="level-btn easy"
                onClick={() => startGame('easy')}
              >
                <span className="level-name">EASY</span>
                <span className="level-desc">느린 속도, 속도 증가 없음</span>
              </button>
              <button 
                className="level-btn medium"
                onClick={() => startGame('medium')}
              >
                <span className="level-name">MEDIUM</span>
                <span className="level-desc">보통 속도, 점수에 따라 빨라짐</span>
              </button>
              <button 
                className="level-btn hard"
                onClick={() => startGame('hard')}
              >
                <span className="level-name">HARD</span>
                <span className="level-desc">빠른 속도, 급격히 빨라짐</span>
              </button>
            </div>
          </div>
        )}

        {/* 게임 오버 */}
        {gameState === 'gameover' && (
          <div className="game-overlay gameover">
            <h2 className="pixel-font">GAME OVER</h2>
            <p className="final-score">최종 점수: {score}</p>
            <p className="snake-length">뱀 길이: {snake.length}</p>
            <button className="restart-btn" onClick={resetGame}>
              다시 하기
            </button>
          </div>
        )}
      </div>

      {/* 모바일 컨트롤 */}
      {gameState === 'playing' && (
        <div className="mobile-controls">
          <div className="control-row">
            <button 
              className="control-btn up"
              onTouchStart={(e) => { e.preventDefault(); handleMobileControl('UP'); }}
              onClick={() => handleMobileControl('UP')}
            >
              ▲
            </button>
          </div>
          <div className="control-row">
            <button 
              className="control-btn left"
              onTouchStart={(e) => { e.preventDefault(); handleMobileControl('LEFT'); }}
              onClick={() => handleMobileControl('LEFT')}
            >
              ◀
            </button>
            <button 
              className="control-btn down"
              onTouchStart={(e) => { e.preventDefault(); handleMobileControl('DOWN'); }}
              onClick={() => handleMobileControl('DOWN')}
            >
              ▼
            </button>
            <button 
              className="control-btn right"
              onTouchStart={(e) => { e.preventDefault(); handleMobileControl('RIGHT'); }}
              onClick={() => handleMobileControl('RIGHT')}
            >
              ▶
            </button>
          </div>
        </div>
      )}

      <div className="game-instructions">
        <p><strong>조작:</strong> 방향키 또는 W A S D / 화면 버튼</p>
        <p><strong>규칙:</strong> 사과를 먹고, 벽과 자기 몸을 피하세요!</p>
      </div>

      {/* 닉네임 모달 */}
      {showNicknameModal && (
        <NicknameModal
          score={score}
          gameName="snake"
          onSubmit={() => setShowNicknameModal(false)}
          onClose={() => setShowNicknameModal(false)}
        />
      )}
    </div>
  );
}

export default Snake;

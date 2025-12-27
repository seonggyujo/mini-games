import { useState, useEffect, useCallback, useRef } from 'react';
import NicknameModal from '../../common/NicknameModal';
import useHighScore from '../../../hooks/useHighScore';
import './MemoryCard.css';

// 이모지 카드 세트
const CARD_EMOJIS = [
  '🍎', '🍊', '🍋', '🍇', '🍓', '🍒',
  '🐶', '🐱', '🐼', '🦊', '🐰', '🐻',
  '⭐', '🌙', '☀️', '🌈', '❤️', '💎',
  '🎮', '🎨'
];

// 레벨 설정
const LEVEL_CONFIG = {
  easy: { 
    cols: 4, 
    rows: 3, 
    pairs: 6, 
    timeLimit: 60, 
    bonus: 0,
    label: 'EASY',
    color: '#00b894'
  },
  medium: { 
    cols: 4, 
    rows: 4, 
    pairs: 8, 
    timeLimit: 90, 
    bonus: 100,
    label: 'MEDIUM',
    color: '#fdcb6e'
  },
  hard: { 
    cols: 5, 
    rows: 4, 
    pairs: 10, 
    timeLimit: 120, 
    bonus: 200,
    label: 'HARD',
    color: '#d63031'
  }
};

function MemoryCard() {
  const [gameState, setGameState] = useState('ready'); // ready, levelSelect, playing, gameover, win
  const [level, setLevel] = useState(null);
  const [cards, setCards] = useState([]);
  const [flipped, setFlipped] = useState([]);
  const [matched, setMatched] = useState([]);
  const [moves, setMoves] = useState(0);
  const [timeLeft, setTimeLeft] = useState(60);
  const [score, setScore] = useState(0);
  const [highScore, , checkAndUpdateHighScore] = useHighScore('memory-card');
  const [showNicknameModal, setShowNicknameModal] = useState(false);
  const [isChecking, setIsChecking] = useState(false);

  const timerRef = useRef(null);

  // 카드 셔플
  const shuffleCards = useCallback((pairCount) => {
    const selectedEmojis = CARD_EMOJIS.slice(0, pairCount);
    const cardPairs = [...selectedEmojis, ...selectedEmojis];
    
    // Fisher-Yates 셔플
    for (let i = cardPairs.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [cardPairs[i], cardPairs[j]] = [cardPairs[j], cardPairs[i]];
    }

    return cardPairs.map((emoji, index) => ({
      id: index,
      emoji,
      isFlipped: false,
      isMatched: false
    }));
  }, []);

  // 게임 시작
  const startGame = useCallback((selectedLevel) => {
    const config = LEVEL_CONFIG[selectedLevel];
    setLevel(selectedLevel);
    setCards(shuffleCards(config.pairs));
    setFlipped([]);
    setMatched([]);
    setMoves(0);
    setTimeLeft(config.timeLimit);
    setScore(0);
    setIsChecking(false);
    setGameState('playing');
  }, [shuffleCards]);

  // 게임 리셋
  const resetGame = () => {
    clearInterval(timerRef.current);
    setGameState('levelSelect');
  };

  // 타이머
  useEffect(() => {
    if (gameState !== 'playing') return;

    timerRef.current = setInterval(() => {
      setTimeLeft(prev => {
        if (prev <= 1) {
          setGameState('gameover');
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timerRef.current);
  }, [gameState]);

  // 카드 클릭 처리
  const handleCardClick = (index) => {
    if (gameState !== 'playing') return;
    if (isChecking) return;
    if (flipped.includes(index)) return;
    if (matched.includes(index)) return;
    if (flipped.length >= 2) return;

    const newFlipped = [...flipped, index];
    setFlipped(newFlipped);

    if (newFlipped.length === 2) {
      setMoves(m => m + 1);
      setIsChecking(true);

      const [first, second] = newFlipped;
      
      if (cards[first].emoji === cards[second].emoji) {
        // 매칭 성공
        setTimeout(() => {
          setMatched(prev => [...prev, first, second]);
          setFlipped([]);
          setIsChecking(false);
        }, 500);
      } else {
        // 매칭 실패
        setTimeout(() => {
          setFlipped([]);
          setIsChecking(false);
        }, 1000);
      }
    }
  };

  // 승리 체크
  useEffect(() => {
    if (gameState !== 'playing' || !level) return;

    const config = LEVEL_CONFIG[level];
    if (matched.length === config.pairs * 2) {
      clearInterval(timerRef.current);
      
      // 점수 계산: (남은 시간 × 10) + 레벨 보너스 - (이동 횟수 × 2)
      const finalScore = Math.max(0, (timeLeft * 10) + config.bonus - (moves * 2));
      setScore(finalScore);
      setGameState('win');
    }
  }, [matched, gameState, level, timeLeft, moves]);

  // 게임 종료 시 최고 점수 체크
  useEffect(() => {
    if ((gameState === 'win' || gameState === 'gameover') && score > 0 && checkAndUpdateHighScore(score)) {
      setShowNicknameModal(true);
    }
  }, [gameState, score, checkAndUpdateHighScore]);

  // 시간 포맷
  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const config = level ? LEVEL_CONFIG[level] : null;

  return (
    <div className="memory-card-container">
      <div className="score-board">
        <div className="time-display">
          <span>TIME</span>
          <span className={`time-value ${timeLeft <= 10 ? 'danger' : ''}`}>
            {formatTime(timeLeft)}
          </span>
        </div>
        <div className="moves-display">
          <span>MOVES</span>
          <span className="moves-value">{moves}</span>
        </div>
        {level && (
          <div className="level-display">
            <span>LEVEL</span>
            <span className="level-value" style={{ color: config.color }}>
              {config.label}
            </span>
          </div>
        )}
        <div className="high-score">
          <span>BEST</span>
          <span className="best-value">{highScore}</span>
        </div>
      </div>

      <div className="game-area">
        {/* 시작 화면 */}
        {gameState === 'ready' && (
          <div className="game-overlay">
            <h2 className="pixel-font">MEMORY CARD</h2>
            <p className="subtitle">카드 짝 맞추기 게임</p>
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
                <span className="level-desc">3×4 카드, 60초</span>
              </button>
              <button 
                className="level-btn medium"
                onClick={() => startGame('medium')}
              >
                <span className="level-name">MEDIUM</span>
                <span className="level-desc">4×4 카드, 90초, +100 보너스</span>
              </button>
              <button 
                className="level-btn hard"
                onClick={() => startGame('hard')}
              >
                <span className="level-name">HARD</span>
                <span className="level-desc">4×5 카드, 120초, +200 보너스</span>
              </button>
            </div>
          </div>
        )}

        {/* 게임 플레이 */}
        {(gameState === 'playing' || gameState === 'win' || gameState === 'gameover') && config && (
          <div 
            className="card-grid"
            style={{
              gridTemplateColumns: `repeat(${config.cols}, 1fr)`,
              gridTemplateRows: `repeat(${config.rows}, 1fr)`
            }}
          >
            {cards.map((card, index) => (
              <div
                key={card.id}
                className={`card ${flipped.includes(index) || matched.includes(index) ? 'flipped' : ''} ${matched.includes(index) ? 'matched' : ''}`}
                onClick={() => handleCardClick(index)}
              >
                <div className="card-inner">
                  <div className="card-front">
                    <span className="card-back-icon">?</span>
                  </div>
                  <div className="card-back">
                    <span className="card-emoji">{card.emoji}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* 승리 오버레이 */}
        {gameState === 'win' && (
          <div className="game-overlay result win">
            <h2 className="pixel-font">YOU WIN!</h2>
            <p className="final-score">점수: {score}</p>
            <p className="final-detail">남은 시간: {formatTime(timeLeft)} | 이동 횟수: {moves}</p>
            <button className="restart-btn" onClick={resetGame}>
              다시 하기
            </button>
          </div>
        )}

        {/* 게임 오버 */}
        {gameState === 'gameover' && (
          <div className="game-overlay result gameover">
            <h2 className="pixel-font">TIME OUT!</h2>
            <p className="final-score">매칭: {matched.length / 2} / {config?.pairs}</p>
            <button className="restart-btn" onClick={resetGame}>
              다시 하기
            </button>
          </div>
        )}
      </div>

      <div className="game-instructions">
        <p><strong>규칙:</strong> 같은 그림의 카드 2장을 찾아 짝을 맞추세요!</p>
        <p><strong>점수:</strong> (남은 시간 × 10) + 레벨 보너스 - (이동 횟수 × 2)</p>
      </div>

      {/* 닉네임 모달 */}
      {showNicknameModal && (
        <NicknameModal
          score={score}
          gameName="memory-card"
          onSubmit={() => setShowNicknameModal(false)}
          onClose={() => setShowNicknameModal(false)}
        />
      )}
    </div>
  );
}

export default MemoryCard;

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
  const [sessionId, setSessionId] = useState(null);
  const [finalScore, setFinalScore] = useState(0);

  const timerRef = useRef(null);

  // 게임 시작 - 서버에서 세션 및 카드 배치 발급
  const startGame = useCallback(async (selectedLevel) => {
    try {
      const response = await fetch('/api/game/memorycard/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ level: selectedLevel })
      });
      
      if (!response.ok) {
        console.error('Failed to start game session');
        return;
      }
      
      const data = await response.json();
      setSessionId(data.sessionId);
      
      // 서버에서 받은 카드 배치를 이모지로 변환
      const serverCards = data.cards.map((emojiIndex, index) => ({
        id: index,
        emoji: CARD_EMOJIS[emojiIndex],
        emojiIndex: emojiIndex,
        isFlipped: false,
        isMatched: false
      }));
      
      const config = LEVEL_CONFIG[selectedLevel];
      setLevel(selectedLevel);
      setCards(serverCards);
      setFlipped([]);
      setMatched([]);
      setMoves(0);
      setTimeLeft(data.timeLimit || config.timeLimit);
      setScore(0);
      setIsChecking(false);
      setGameState('playing');
    } catch (err) {
      console.error('Failed to start game:', err);
    }
  }, []);

  // 게임 리셋
  const resetGame = () => {
    clearInterval(timerRef.current);
    setGameState('levelSelect');
    setSessionId(null);
    setFinalScore(0);
  };

  // 타이머
  useEffect(() => {
    if (gameState !== 'playing') return;

    timerRef.current = setInterval(() => {
      setTimeLeft(prev => {
        if (prev <= 1) {
          setGameState('gameover');
          endGame();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timerRef.current);
  }, [gameState]);

  // 매칭 시도 - 서버에 보고
  const reportMatch = async (card1, card2) => {
    if (!sessionId) return null;
    
    try {
      const response = await fetch('/api/game/memorycard/match', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sessionId, card1, card2 })
      });
      
      if (response.ok) {
        return await response.json();
      }
    } catch (err) {
      console.error('Failed to report match:', err);
    }
    return null;
  };

  // 게임 종료 - 서버에 결과 전송
  const endGame = async () => {
    if (!sessionId) return;
    
    try {
      const response = await fetch('/api/game/memorycard/end', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sessionId })
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.valid) {
          setFinalScore(data.finalScore);
          if (data.canSubmit && checkAndUpdateHighScore(data.finalScore)) {
            setShowNicknameModal(true);
          }
        }
      }
    } catch (err) {
      console.error('Failed to end game:', err);
    }
  };

  // Submit score handler
  const handleSubmitScore = async (nickname) => {
    if (!sessionId) return;
    
    try {
      const response = await fetch('/api/game/memorycard/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sessionId,
          nickname
        })
      });
      
      if (!response.ok) {
        throw new Error('Failed to submit score');
      }
      
      setShowNicknameModal(false);
    } catch (err) {
      console.error('Failed to submit score:', err);
      throw err;
    }
  };

  // 카드 클릭 처리
  const handleCardClick = async (index) => {
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
      
      // 서버에 매칭 보고
      const result = await reportMatch(first, second);
      
      if (result && result.valid) {
        if (result.isMatch) {
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
      } else {
        // 서버 오류 시 클라이언트 기준으로 처리
        if (cards[first].emoji === cards[second].emoji) {
          setTimeout(() => {
            setMatched(prev => [...prev, first, second]);
            setFlipped([]);
            setIsChecking(false);
          }, 500);
        } else {
          setTimeout(() => {
            setFlipped([]);
            setIsChecking(false);
          }, 1000);
        }
      }
    }
  };

  // 승리 체크
  useEffect(() => {
    if (gameState !== 'playing' || !level) return;

    const config = LEVEL_CONFIG[level];
    if (matched.length === config.pairs * 2) {
      clearInterval(timerRef.current);
      setGameState('win');
      endGame();
    }
  }, [matched, gameState, level]);

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
            <p className="final-score">점수: {finalScore}</p>
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
          score={finalScore}
          gameName="memory-card"
          sessionId={sessionId}
          onSubmit={handleSubmitScore}
          onClose={() => setShowNicknameModal(false)}
        />
      )}
    </div>
  );
}

export default MemoryCard;

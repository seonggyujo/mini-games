import { useState, useEffect, useRef, useCallback } from 'react';
import useSoundEffects from '../../../hooks/useSoundEffects';
import './TranslationBattle.css';

function TranslationBattle({ nickname, opponentNickname, difficulty, initialSentence, sendMessage, onMessage, onFinished }) {
  const [phase, setPhase] = useState('playing'); // playing, evaluating, result
  const [round, setRound] = useState(1);
  const [sentence, setSentence] = useState(initialSentence || '');
  const [translation, setTranslation] = useState('');
  const [timeLeft, setTimeLeft] = useState(60);
  const [submitted, setSubmitted] = useState(false);
  const [opponentSubmitted, setOpponentSubmitted] = useState(false);
  const [myWins, setMyWins] = useState(0);
  const [opponentWins, setOpponentWins] = useState(0);
  const [roundResult, setRoundResult] = useState(null);
  const [isGameOver, setIsGameOver] = useState(false);
  const [nextRoundReady, setNextRoundReady] = useState(false);
  const [opponentNextReady, setOpponentNextReady] = useState(false);
  const textareaRef = useRef(null);
  const lastTickRef = useRef(null);
  
  const { 
    isMuted, 
    toggleMute, 
    playClick, 
    playAlert, 
    playVictory, 
    playDefeat, 
    playCelebration, 
    playTick,
    playDraw 
  } = useSoundEffects();

  // Setup message handlers
  useEffect(() => {
    const unsubscribers = [
      onMessage('t_round_start', (data) => {
        setPhase('playing');
        setRound(data.round);
        setSentence(data.sentence);
        setTimeLeft(data.timeLeft);
        setTranslation('');
        setSubmitted(false);
        setOpponentSubmitted(false);
        setRoundResult(null);
        setNextRoundReady(false);
        setOpponentNextReady(false);
        // Focus textarea
        setTimeout(() => {
          textareaRef.current?.focus();
        }, 100);
      }),

      onMessage('t_time_update', (data) => {
        setTimeLeft(data.timeLeft);
      }),

      onMessage('t_opponent_submitted', () => {
        setOpponentSubmitted(true);
        playAlert();
      }),

      onMessage('t_evaluating', () => {
        setPhase('evaluating');
      }),

      onMessage('t_round_result', (data) => {
        setPhase('result');
        setRoundResult(data);
        setMyWins(data.totalWins[0]);
        setOpponentWins(data.totalWins[1]);
        setIsGameOver(data.isGameOver);
        setNextRoundReady(false);
        setOpponentNextReady(false);
        
        // 라운드 결과 효과음
        if (data.isGameOver && data.roundWinner === 'me') {
          playCelebration();
        } else if (data.roundWinner === 'me') {
          playVictory();
        } else if (data.roundWinner === 'opponent') {
          playDefeat();
        } else {
          playDraw();
        }
      }),

      onMessage('t_opponent_next_ready', () => {
        setOpponentNextReady(true);
      }),

      onMessage('t_game_over', () => {
        // "최종 결과 보기" 버튼 클릭 시 onFinished() 호출
        // 여기서는 자동 전환하지 않음 - 사용자가 라운드 결과를 확인할 수 있도록
      }),
    ];

    return () => {
      unsubscribers.forEach(unsub => unsub());
    };
  }, [onMessage, onFinished, playAlert, playVictory, playDefeat, playCelebration, playDraw]);

  const handleSubmit = useCallback(() => {
    if (submitted || phase !== 'playing') return;
    sendMessage({ type: 't_submit', translation: translation.trim() });
    setSubmitted(true);
    playClick();
  }, [submitted, phase, translation, sendMessage, playClick]);

  const handleNextRound = useCallback(() => {
    if (nextRoundReady) return;
    sendMessage({ type: 't_next_round' });
    setNextRoundReady(true);
  }, [nextRoundReady, sendMessage]);

  // Keyboard shortcut for submit (Ctrl+Enter)
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.ctrlKey && e.key === 'Enter' && !submitted && phase === 'playing') {
        handleSubmit();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleSubmit, submitted, phase]);

  // Timer tick sound (10초 이하일 때)
  useEffect(() => {
    if (phase === 'playing' && timeLeft <= 10 && timeLeft > 0 && !submitted) {
      // 같은 초에 중복 재생 방지
      if (lastTickRef.current !== timeLeft) {
        lastTickRef.current = timeLeft;
        playTick();
      }
    }
  }, [timeLeft, phase, submitted, playTick]);

  // Format time
  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  // Get difficulty label
  const getDifficultyLabel = () => {
    switch (difficulty) {
      case 'easy': return '쉬움';
      case 'hard': return '어려움';
      default: return '보통';
    }
  };

  // Render result phase
  if (phase === 'result' && roundResult) {
    return (
      <div className="translation-battle">
        <div className="battle-header">
          <div className="round-info">
            <span className="round-number">Round {roundResult.round}/3</span>
            <span className="difficulty-label">{getDifficultyLabel()}</span>
          </div>
          <div className="score-display">
            <span className="my-score">{nickname}: {myWins}승</span>
            <span className="vs">vs</span>
            <span className="opponent-score">{opponentNickname}: {opponentWins}승</span>
            <button 
              className={`mute-button ${isMuted ? 'muted' : ''}`} 
              onClick={toggleMute}
              title={isMuted ? '소리 켜기' : '소리 끄기'}
            >
              {isMuted ? '🔇' : '🔊'}
            </button>
          </div>
        </div>

        <div className="result-content">
          <div className={`round-result-banner ${roundResult.roundWinner}`}>
            {roundResult.roundWinner === 'me' && 'Round 승리!'}
            {roundResult.roundWinner === 'opponent' && 'Round 패배'}
            {roundResult.roundWinner === 'draw' && '무승부'}
          </div>

          <div className="original-sentence">
            <label>원문</label>
            <p>{roundResult.sentence}</p>
          </div>

          <div className="translations-compare">
            <div className="translation-card my-card">
              <div className="card-header">
                <span className="player-name">{nickname}</span>
                <span className="total-score">{roundResult.myScore.total}점</span>
              </div>
              <p className="translation-text">{roundResult.myTranslation || '(미제출)'}</p>
              <div className="score-breakdown">
                <div className="score-item">
                  <span className="score-label">의미</span>
                  <span className="score-value">{roundResult.myScore.meaning}/40</span>
                </div>
                <div className="score-item">
                  <span className="score-label">문법</span>
                  <span className="score-value">{roundResult.myScore.grammar}/30</span>
                </div>
                <div className="score-item">
                  <span className="score-label">자연스러움</span>
                  <span className="score-value">{roundResult.myScore.naturalness}/30</span>
                </div>
              </div>
              <p className="feedback">{roundResult.myScore.feedback}</p>
            </div>

            <div className="translation-card opponent-card">
              <div className="card-header">
                <span className="player-name">{opponentNickname}</span>
                <span className="total-score">{roundResult.opponentScore.total}점</span>
              </div>
              <p className="translation-text">{roundResult.opponentTranslation || '(미제출)'}</p>
              <div className="score-breakdown">
                <div className="score-item">
                  <span className="score-label">의미</span>
                  <span className="score-value">{roundResult.opponentScore.meaning}/40</span>
                </div>
                <div className="score-item">
                  <span className="score-label">문법</span>
                  <span className="score-value">{roundResult.opponentScore.grammar}/30</span>
                </div>
                <div className="score-item">
                  <span className="score-label">자연스러움</span>
                  <span className="score-value">{roundResult.opponentScore.naturalness}/30</span>
                </div>
              </div>
              <p className="feedback">{roundResult.opponentScore.feedback}</p>
            </div>
          </div>

          <div className="model-answer">
            <label>모범 답안</label>
            <p>{roundResult.modelAnswer}</p>
          </div>

          {!isGameOver ? (
            <div className="next-round-section">
              <button 
                className={`btn-next-round ${nextRoundReady ? 'ready' : ''}`}
                onClick={handleNextRound}
                disabled={nextRoundReady}
              >
                {nextRoundReady ? '대기 중...' : '다음 라운드'}
              </button>
              
              {opponentNextReady && !nextRoundReady && (
                <p className="opponent-next-ready-text">상대가 준비됨!</p>
              )}
              {nextRoundReady && !opponentNextReady && (
                <p className="waiting-next-text">상대를 기다리는 중...</p>
              )}
              {nextRoundReady && opponentNextReady && (
                <p className="waiting-next-text">다음 라운드 시작 중...</p>
              )}
            </div>
          ) : (
            <div className="next-round-section">
              <button 
                className="btn-final-result"
                onClick={onFinished}
              >
                최종 결과 보기
              </button>
            </div>
          )}
        </div>
      </div>
    );
  }

  // Render evaluating phase
  if (phase === 'evaluating') {
    return (
      <div className="translation-battle">
        <div className="battle-header">
          <div className="round-info">
            <span className="round-number">Round {round}/3</span>
            <span className="difficulty-label">{getDifficultyLabel()}</span>
          </div>
          <div className="score-display">
            <span className="my-score">{nickname}: {myWins}승</span>
            <span className="vs">vs</span>
            <span className="opponent-score">{opponentNickname}: {opponentWins}승</span>
            <button 
              className={`mute-button ${isMuted ? 'muted' : ''}`} 
              onClick={toggleMute}
              title={isMuted ? '소리 켜기' : '소리 끄기'}
            >
              {isMuted ? '🔇' : '🔊'}
            </button>
          </div>
        </div>

        <div className="evaluating-content">
          <div className="evaluating-spinner"></div>
          <h3>AI가 평가 중...</h3>
          <p>두 번역을 비교 분석하고 있습니다</p>
        </div>
      </div>
    );
  }

  // Render playing phase
  return (
    <div className="translation-battle">
      <div className="battle-header">
        <div className="round-info">
          <span className="round-number">Round {round}/3</span>
          <span className="difficulty-label">{getDifficultyLabel()}</span>
        </div>
        <div className={`timer ${timeLeft <= 10 ? 'warning' : ''}`}>
          {formatTime(timeLeft)}
        </div>
        <div className="score-display">
          <span className="my-score">{nickname}: {myWins}승</span>
          <span className="vs">vs</span>
          <span className="opponent-score">{opponentNickname}: {opponentWins}승</span>
          <button 
            className={`mute-button ${isMuted ? 'muted' : ''}`} 
            onClick={toggleMute}
            title={isMuted ? '소리 켜기' : '소리 끄기'}
          >
            {isMuted ? '🔇' : '🔊'}
          </button>
        </div>
      </div>

      <div className="battle-content">
        <div className="sentence-box">
          <label>한국어 문장</label>
          <p className="korean-sentence">{sentence}</p>
        </div>

        <div className="translation-box">
          <label>영어 번역</label>
          <textarea
            ref={textareaRef}
            value={translation}
            onChange={(e) => setTranslation(e.target.value.slice(0, 500))}
            placeholder="영어로 번역하세요..."
            disabled={submitted}
            maxLength={500}
          />
          <div className="translation-footer">
            <span className="char-count">{translation.length}/500</span>
          </div>
        </div>

        <button 
          className={`btn-submit ${submitted ? 'submitted' : ''}`}
          onClick={handleSubmit}
          disabled={submitted}
        >
          {submitted ? '제출 완료' : '제출 (Ctrl+Enter)'}
        </button>

        <div className="submit-status">
          <span className={`status-item ${submitted ? 'submitted' : ''}`}>
            나: {submitted ? '제출 완료' : '응답 중...'}
          </span>
          <span className={`status-item ${opponentSubmitted ? 'submitted' : ''}`}>
            상대: {opponentSubmitted ? '제출 완료' : '응답 중...'}
          </span>
        </div>
      </div>
    </div>
  );
}

export default TranslationBattle;

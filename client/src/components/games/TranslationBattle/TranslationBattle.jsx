import { useState, useEffect, useRef, useCallback } from 'react';
import useSoundEffects from '../../../hooks/useSoundEffects';
import { ConnectionState } from '../../../hooks/useTranslationWS';
import './TranslationBattle.css';

function TranslationBattle({ nickname, opponentNickname, difficulty, initialSentence, initialTense, sendMessage, onMessage, connectionState, onFinished }) {
  const [phase, setPhase] = useState('playing'); // playing, evaluating, result
  const [round, setRound] = useState(1);
  const [sentence, setSentence] = useState(initialSentence || '');
  const [tense, setTense] = useState(initialTense || '현재형');
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
  const [finalResultData, setFinalResultData] = useState(null);
  const [copied, setCopied] = useState(false);
  const [showEmptyWarning, setShowEmptyWarning] = useState(false);
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
        setTense(data.tense || '현재형');
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

      onMessage('t_game_over', (data) => {
        // 최종 결과 데이터 저장 - "최종 결과 보기" 버튼 클릭 시 사용
        setFinalResultData(data);
      }),
    ];

    return () => {
      unsubscribers.forEach(unsub => unsub());
    };
  }, [onMessage, onFinished, playAlert, playVictory, playDefeat, playCelebration, playDraw]);

  const handleSubmit = useCallback(() => {
    if (submitted || phase !== 'playing') return;
    
    // 빈 번역 제출 방지 (시간 종료 시에는 허용)
    if (!translation.trim() && timeLeft > 1) {
      setShowEmptyWarning(true);
      setTimeout(() => setShowEmptyWarning(false), 2000);
      textareaRef.current?.focus();
      return;
    }
    
    sendMessage({ type: 't_submit', translation: translation.trim() });
    setSubmitted(true);
    playClick();
    
    // 진동 피드백 (지원하는 기기에서)
    if (navigator.vibrate) {
      navigator.vibrate(100);
    }
  }, [submitted, phase, translation, timeLeft, sendMessage, playClick]);

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

  // 시간 종료 시 자동 제출 (서버는 timeLeft <= 0이면 바로 평가 단계로 진입하므로 1초 남았을 때 제출)
  useEffect(() => {
    if (timeLeft <= 1 && !submitted && phase === 'playing') {
      handleSubmit();
    }
  }, [timeLeft, submitted, phase, handleSubmit]);

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

  // Get tense badge class
  const getTenseBadgeClass = (tenseValue) => {
    switch (tenseValue) {
      case '과거형': return 'tense-past';
      case '미래형': return 'tense-future';
      default: return 'tense-present';
    }
  };

  // Copy model answer to clipboard
  const handleCopyModelAnswer = useCallback(() => {
    if (roundResult?.modelAnswer) {
      navigator.clipboard.writeText(roundResult.modelAnswer).then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }).catch(() => {
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = roundResult.modelAnswer;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      });
    }
  }, [roundResult]);

  // Calculate score difference
  const getScoreDiff = () => {
    if (!roundResult) return 0;
    return roundResult.myScore.total - roundResult.opponentScore.total;
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
            {roundResult.roundWinner !== 'draw' && (
              <span className="score-diff">
                ({getScoreDiff() > 0 ? '+' : ''}{getScoreDiff()}점)
              </span>
            )}
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
            <div className="model-answer-header">
              <label>모범 답안</label>
              <button 
                className={`btn-copy-answer ${copied ? 'copied' : ''}`}
                onClick={handleCopyModelAnswer}
                title="클립보드에 복사"
              >
                {copied ? '복사됨!' : '복사'}
              </button>
            </div>
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
                onClick={() => onFinished(finalResultData)}
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
      {/* 연결 상태 오버레이 */}
      {connectionState === ConnectionState.RECONNECTING && (
        <div className="connection-overlay">
          <div className="connection-content">
            <div className="connection-spinner"></div>
            <p>서버에 재연결 중...</p>
          </div>
        </div>
      )}
      
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
          <div className="sentence-header">
            <label>한국어 문장</label>
            <span className={`tense-badge ${getTenseBadgeClass(tense)}`}>{tense}</span>
          </div>
          <p className="korean-sentence">{sentence}</p>
        </div>

        <div className="translation-box">
          <label>영어 번역</label>
          <textarea
            ref={textareaRef}
            value={translation}
            onChange={(e) => {
              setTranslation(e.target.value.slice(0, 500));
              setShowEmptyWarning(false);
            }}
            placeholder="영어로 번역하세요..."
            disabled={submitted}
            maxLength={500}
            className={showEmptyWarning ? 'shake' : ''}
          />
          <div className="translation-footer">
            {showEmptyWarning && (
              <span className="empty-warning">번역을 입력해주세요!</span>
            )}
            <span className="char-count">{translation.length}/500</span>
          </div>
        </div>

        <button 
          className={`btn-submit ${submitted ? 'submitted' : ''}`}
          onClick={handleSubmit}
          disabled={submitted}
        >
          {submitted ? '제출 완료!' : '제출 (Ctrl+Enter)'}
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

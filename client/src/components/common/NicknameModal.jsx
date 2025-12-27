import { useState } from 'react';
import './NicknameModal.css';

function NicknameModal({ score, gameName, sessionId, onSubmit, onClose, formatScore }) {
  const [nickname, setNickname] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const displayScore = formatScore ? formatScore(score) : score;

  const handleSubmit = async () => {
    if (!nickname.trim()) return;
    if (isSubmitting) return;

    setIsSubmitting(true);
    setError(null);

    try {
      // sessionId가 있으면 새 방식 (서버 측 점수), 없으면 기존 방식
      if (sessionId && typeof onSubmit === 'function') {
        // 새 방식: 부모 컴포넌트가 제출 처리
        await onSubmit(nickname.trim());
      } else {
        // 기존 방식: 직접 API 호출 (다른 게임 호환성)
        const response = await fetch('/api/scores', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            nickname: nickname.trim(),
            game: gameName,
            score: score
          })
        });
        
        if (!response.ok) {
          throw new Error('Failed to submit score');
        }
        
        onSubmit();
      }
    } catch (err) {
      console.error('Failed to submit score:', err);
      setError('점수 등록에 실패했습니다. 다시 시도해주세요.');
      setIsSubmitting(false);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') {
      handleSubmit();
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal">
        <h3 className="pixel-font">NEW HIGH SCORE!</h3>
        <p>점수: {displayScore}</p>
        {error && <p className="modal-error">{error}</p>}
        <input
          type="text"
          placeholder="닉네임 입력"
          value={nickname}
          onChange={(e) => setNickname(e.target.value.slice(0, 20))}
          onKeyDown={handleKeyDown}
          disabled={isSubmitting}
          autoFocus
        />
        <div className="modal-buttons">
          <button 
            onClick={handleSubmit} 
            className="submit-btn"
            disabled={isSubmitting || !nickname.trim()}
          >
            {isSubmitting ? '등록 중...' : '등록'}
          </button>
          <button 
            onClick={onClose} 
            className="cancel-btn"
            disabled={isSubmitting}
          >
            취소
          </button>
        </div>
      </div>
    </div>
  );
}

export default NicknameModal;

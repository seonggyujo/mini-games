import { useState, useCallback } from 'react';

/**
 * 로컬 스토리지 기반 하이스코어 관리 훅
 * @param {string} gameKey - 게임별 고유 키 (예: 'jump-runner', 'snake')
 * @returns {[number, function, function]} [highScore, updateHighScore, checkAndUpdateHighScore]
 */
function useHighScore(gameKey) {
  const storageKey = `${gameKey}-highscore`;
  
  const [highScore, setHighScore] = useState(() => {
    return parseInt(localStorage.getItem(storageKey) || '0', 10);
  });

  /**
   * 하이스코어 직접 업데이트
   */
  const updateHighScore = useCallback((newScore) => {
    setHighScore(newScore);
    localStorage.setItem(storageKey, newScore.toString());
  }, [storageKey]);

  /**
   * 현재 점수와 비교하여 하이스코어 업데이트
   * @returns {boolean} 새로운 하이스코어인지 여부
   */
  const checkAndUpdateHighScore = useCallback((currentScore) => {
    if (currentScore > highScore) {
      updateHighScore(currentScore);
      return true;
    }
    return false;
  }, [highScore, updateHighScore]);

  return [highScore, updateHighScore, checkAndUpdateHighScore];
}

export default useHighScore;

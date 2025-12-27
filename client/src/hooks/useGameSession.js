import { useState, useCallback, useRef } from 'react';

/**
 * Mulberry32 PRNG - must match server implementation exactly
 */
function createMulberry32(seed) {
  let state = seed >>> 0; // Convert to unsigned 32-bit
  
  return function() {
    state = (state + 0x6D2B79F5) >>> 0;
    let t = state;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t = (t ^ (t + Math.imul(t ^ (t >>> 7), t | 61))) >>> 0;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

// SpeedClick 레벨 설정 (서버와 동일)
const LEVELS = [
  { level: 1, timeLimit: 1.00, ballSize: 80, blueChance: 0.10, requiredScore: 0 },
  { level: 2, timeLimit: 0.90, ballSize: 75, blueChance: 0.13, requiredScore: 5 },
  { level: 3, timeLimit: 0.80, ballSize: 70, blueChance: 0.16, requiredScore: 10 },
  { level: 4, timeLimit: 0.70, ballSize: 65, blueChance: 0.20, requiredScore: 15 },
  { level: 5, timeLimit: 0.60, ballSize: 60, blueChance: 0.23, requiredScore: 20 },
  { level: 6, timeLimit: 0.50, ballSize: 55, blueChance: 0.26, requiredScore: 25 },
  { level: 7, timeLimit: 0.40, ballSize: 50, blueChance: 0.30, requiredScore: 30 },
];

const GAME_WIDTH = 1200;
const GAME_HEIGHT = 800;
const SPAWN_DELAY = 300;

function getLevelConfig(score) {
  for (let i = LEVELS.length - 1; i >= 0; i--) {
    if (score >= LEVELS[i].requiredScore) {
      return LEVELS[i];
    }
  }
  return LEVELS[0];
}

/**
 * Generate a ball deterministically based on seed and index
 * Must match server implementation exactly
 */
function generateBall(seed, index, score, prevBallEndTime) {
  const ballSeed = (seed + index * 12345) >>> 0;
  const rng = createMulberry32(ballSeed);
  
  const config = getLevelConfig(score);
  const padding = config.ballSize;
  
  const x = padding + rng() * (GAME_WIDTH - padding * 2);
  const y = padding + rng() * (GAME_HEIGHT - padding * 2);
  const isRed = rng() > config.blueChance;
  
  const spawnTime = prevBallEndTime + SPAWN_DELAY;
  
  return {
    index,
    x,
    y,
    isRed,
    spawnTime,
    duration: config.timeLimit * 1000,
    size: config.ballSize,
    level: config.level,
  };
}

/**
 * Custom hook for managing SpeedClick game sessions
 */
export function useSpeedClickSession() {
  const [sessionId, setSessionId] = useState(null);
  const [seed, setSeed] = useState(null);
  const [serverStartTime, setServerStartTime] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const clientStartTimeRef = useRef(null);

  /**
   * Start a new game session
   */
  const startSession = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch('/api/game/speedclick/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      
      if (!response.ok) {
        throw new Error('Failed to start session');
      }
      
      const data = await response.json();
      setSessionId(data.sessionId);
      setSeed(data.seed);
      setServerStartTime(data.startTime);
      clientStartTimeRef.current = Date.now();
      
      return {
        sessionId: data.sessionId,
        seed: data.seed,
        startTime: data.startTime,
      };
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  /**
   * Report a ball click to the server
   */
  const reportClick = useCallback(async (ballIndex) => {
    if (!sessionId) {
      throw new Error('No active session');
    }
    
    const clickTimeMs = Date.now() - clientStartTimeRef.current;
    
    try {
      const response = await fetch('/api/game/speedclick/click', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sessionId,
          ballIndex,
          clickTimeMs,
        }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to report click');
      }
      
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    }
  }, [sessionId]);

  /**
   * Report a missed ball to the server
   */
  const reportMiss = useCallback(async (ballIndex) => {
    if (!sessionId) {
      throw new Error('No active session');
    }
    
    try {
      const response = await fetch('/api/game/speedclick/miss', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sessionId,
          ballIndex,
        }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to report miss');
      }
      
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    }
  }, [sessionId]);

  /**
   * End the current game session
   */
  const endSession = useCallback(async () => {
    if (!sessionId) {
      throw new Error('No active session');
    }
    
    try {
      const response = await fetch('/api/game/speedclick/end', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sessionId }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to end session');
      }
      
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    }
  }, [sessionId]);

  /**
   * Submit the score with a nickname
   */
  const submitScore = useCallback(async (nickname) => {
    if (!sessionId) {
      throw new Error('No active session');
    }
    
    try {
      const response = await fetch('/api/game/speedclick/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sessionId,
          nickname,
        }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to submit score');
      }
      
      const result = await response.json();
      
      // Clear session after successful submission
      if (result.success) {
        setSessionId(null);
        setSeed(null);
        setServerStartTime(null);
      }
      
      return result;
    } catch (err) {
      setError(err.message);
      throw err;
    }
  }, [sessionId]);

  /**
   * Reset the session state
   */
  const resetSession = useCallback(() => {
    setSessionId(null);
    setSeed(null);
    setServerStartTime(null);
    setError(null);
    clientStartTimeRef.current = null;
  }, []);

  /**
   * Get elapsed time since game start
   */
  const getElapsedTime = useCallback(() => {
    if (!clientStartTimeRef.current) return 0;
    return Date.now() - clientStartTimeRef.current;
  }, []);

  return {
    sessionId,
    seed,
    serverStartTime,
    isLoading,
    error,
    startSession,
    reportClick,
    reportMiss,
    endSession,
    submitScore,
    resetSession,
    getElapsedTime,
    generateBall,
  };
}

export { generateBall, createMulberry32, getLevelConfig, LEVELS };

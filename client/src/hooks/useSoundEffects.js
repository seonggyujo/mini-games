import { useCallback, useState, useEffect, useRef } from 'react';

// Web Audio API 기반 효과음 훅
function useSoundEffects() {
  const [isMuted, setIsMuted] = useState(() => {
    const saved = localStorage.getItem('soundMuted');
    return saved === 'true';
  });
  
  const audioContextRef = useRef(null);

  // AudioContext 초기화 (lazy)
  const getAudioContext = useCallback(() => {
    if (!audioContextRef.current) {
      audioContextRef.current = new (window.AudioContext || window.webkitAudioContext)();
    }
    // Resume if suspended (브라우저 정책)
    if (audioContextRef.current.state === 'suspended') {
      audioContextRef.current.resume();
    }
    return audioContextRef.current;
  }, []);

  // 음소거 상태 저장
  useEffect(() => {
    localStorage.setItem('soundMuted', isMuted.toString());
  }, [isMuted]);

  // 음소거 토글
  const toggleMute = useCallback(() => {
    setIsMuted(prev => !prev);
  }, []);

  // 기본 비프음 생성
  const playTone = useCallback((frequency, duration, type = 'sine', volume = 0.3) => {
    if (isMuted) return;
    
    try {
      const ctx = getAudioContext();
      const oscillator = ctx.createOscillator();
      const gainNode = ctx.createGain();
      
      oscillator.connect(gainNode);
      gainNode.connect(ctx.destination);
      
      oscillator.type = type;
      oscillator.frequency.setValueAtTime(frequency, ctx.currentTime);
      
      gainNode.gain.setValueAtTime(volume, ctx.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + duration);
      
      oscillator.start(ctx.currentTime);
      oscillator.stop(ctx.currentTime + duration);
    } catch (e) {
      console.warn('Sound play failed:', e);
    }
  }, [isMuted, getAudioContext]);

  // 카운트다운 비프음 (3→2→1 음 높아짐)
  const playBeep = useCallback((count) => {
    const frequencies = {
      3: 440,  // A4
      2: 554,  // C#5
      1: 660,  // E5
    };
    const freq = frequencies[count] || 440;
    playTone(freq, 0.15, 'sine', 0.4);
  }, [playTone]);

  // 제출 완료 클릭음
  const playClick = useCallback(() => {
    playTone(800, 0.08, 'sine', 0.3);
  }, [playTone]);

  // 상대 제출 알림음
  const playAlert = useCallback(() => {
    if (isMuted) return;
    
    try {
      const ctx = getAudioContext();
      
      // 두 음 연속 재생
      [600, 800].forEach((freq, i) => {
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, ctx.currentTime);
        gain.gain.setValueAtTime(0.25, ctx.currentTime + i * 0.1);
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + i * 0.1 + 0.1);
        osc.start(ctx.currentTime + i * 0.1);
        osc.stop(ctx.currentTime + i * 0.1 + 0.1);
      });
    } catch (e) {
      console.warn('Sound play failed:', e);
    }
  }, [isMuted, getAudioContext]);

  // 라운드 승리음
  const playVictory = useCallback(() => {
    if (isMuted) return;
    
    try {
      const ctx = getAudioContext();
      const notes = [523, 659, 784, 1047]; // C5, E5, G5, C6
      
      notes.forEach((freq, i) => {
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, ctx.currentTime);
        gain.gain.setValueAtTime(0.3, ctx.currentTime + i * 0.12);
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + i * 0.12 + 0.2);
        osc.start(ctx.currentTime + i * 0.12);
        osc.stop(ctx.currentTime + i * 0.12 + 0.2);
      });
    } catch (e) {
      console.warn('Sound play failed:', e);
    }
  }, [isMuted, getAudioContext]);

  // 라운드 패배음
  const playDefeat = useCallback(() => {
    if (isMuted) return;
    
    try {
      const ctx = getAudioContext();
      const notes = [400, 350, 300]; // 하강하는 음
      
      notes.forEach((freq, i) => {
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, ctx.currentTime);
        gain.gain.setValueAtTime(0.25, ctx.currentTime + i * 0.15);
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + i * 0.15 + 0.2);
        osc.start(ctx.currentTime + i * 0.15);
        osc.stop(ctx.currentTime + i * 0.15 + 0.2);
      });
    } catch (e) {
      console.warn('Sound play failed:', e);
    }
  }, [isMuted, getAudioContext]);

  // 최종 승리 축하음
  const playCelebration = useCallback(() => {
    if (isMuted) return;
    
    try {
      const ctx = getAudioContext();
      // 팡파레 느낌의 멜로디
      const melody = [
        { freq: 523, time: 0 },      // C5
        { freq: 659, time: 0.1 },    // E5
        { freq: 784, time: 0.2 },    // G5
        { freq: 1047, time: 0.35 },  // C6
        { freq: 784, time: 0.5 },    // G5
        { freq: 1047, time: 0.65 },  // C6
      ];
      
      melody.forEach(({ freq, time }) => {
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, ctx.currentTime);
        gain.gain.setValueAtTime(0.35, ctx.currentTime + time);
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + time + 0.15);
        osc.start(ctx.currentTime + time);
        osc.stop(ctx.currentTime + time + 0.15);
      });
    } catch (e) {
      console.warn('Sound play failed:', e);
    }
  }, [isMuted, getAudioContext]);

  // 타이머 틱톡음 (10초 이하)
  const playTick = useCallback(() => {
    playTone(1000, 0.05, 'sine', 0.15);
  }, [playTone]);

  // 무승부음
  const playDraw = useCallback(() => {
    if (isMuted) return;
    
    try {
      const ctx = getAudioContext();
      const notes = [440, 440]; // 같은 음 두 번
      
      notes.forEach((freq, i) => {
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();
        osc.connect(gain);
        gain.connect(ctx.destination);
        osc.type = 'sine';
        osc.frequency.setValueAtTime(freq, ctx.currentTime);
        gain.gain.setValueAtTime(0.2, ctx.currentTime + i * 0.15);
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + i * 0.15 + 0.15);
        osc.start(ctx.currentTime + i * 0.15);
        osc.stop(ctx.currentTime + i * 0.15 + 0.15);
      });
    } catch (e) {
      console.warn('Sound play failed:', e);
    }
  }, [isMuted, getAudioContext]);

  return {
    isMuted,
    toggleMute,
    playBeep,
    playClick,
    playAlert,
    playVictory,
    playDefeat,
    playCelebration,
    playTick,
    playDraw,
  };
}

export default useSoundEffects;

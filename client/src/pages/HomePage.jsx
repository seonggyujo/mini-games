import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import './HomePage.css';

const games = [
  {
    id: 'jump-runner',
    title: 'Jump Runner',
    description: '쥐가 고양이를 피해 달리는 게임!',
    icon: '🐭',
    available: true
  },
  {
    id: 'speed-click',
    title: 'Speed Click',
    description: '빨간 공만 빠르게 클릭!',
    icon: '🔴',
    available: true
  },
  {
    id: 'snake',
    title: 'Snake',
    description: '클래식 뱀 게임! 사과를 먹고 성장하세요',
    icon: '🐍',
    available: true
  },
  {
    id: 'memory-card',
    title: 'Memory Card',
    description: '카드를 뒤집어 같은 그림을 찾으세요!',
    icon: '🃏',
    available: true
  },
  {
    id: 'translation-battle',
    title: 'Translation Battle',
    description: 'AI와 함께하는 영어 번역 대결!',
    icon: '🔤',
    available: true
  }
];

function HomePage() {
  const [rankings, setRankings] = useState({});
  const [loadingRankings, setLoadingRankings] = useState(true);

  useEffect(() => {
    let isMounted = true;

    const fetchRankings = async () => {
      try {
        const results = {};
        await Promise.all(
          games.map(async (game) => {
            try {
              const response = await fetch(`/api/ranking?game=${game.id}&limit=3`);
              if (!response.ok) {
                results[game.id] = [];
                return;
              }
              const data = await response.json();
              results[game.id] = data;
            } catch {
              results[game.id] = [];
            }
          })
        );
        if (isMounted) {
          setRankings(results);
          setLoadingRankings(false);
        }
      } catch (error) {
        console.error('Failed to fetch rankings:', error);
        if (isMounted) {
          setLoadingRankings(false);
        }
      }
    };
    fetchRankings();

    return () => { isMounted = false; };
  }, []);

  const formatScore = (score) => {
    return score.toLocaleString();
  };

  return (
    <div className="home-page">
      <header className="home-header">
        <h1 className="pixel-font">GAME ARCADE</h1>
        <p>재미있는 미니게임을 즐겨보세요!</p>
      </header>

      <div className="games-grid">
        {games.map(game => (
          <div key={game.id} className={`game-card ${!game.available ? 'disabled' : ''}`}>
            <div className="game-icon">{game.icon}</div>
            <h3 className="game-title">{game.title}</h3>
            <p className="game-description">{game.description}</p>
            {game.available ? (
              <Link to={`/game/${game.id}`} className="play-button">
                PLAY
              </Link>
            ) : (
              <span className="coming-soon">COMING SOON</span>
            )}
          </div>
        ))}
      </div>

      {/* 랭킹 미리보기 섹션 */}
      <section className="ranking-preview">
        <h2 className="pixel-font">TOP RANKINGS</h2>
        
        {loadingRankings ? (
          <div className="ranking-preview-grid">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="ranking-preview-card skeleton">
                <div className="ranking-preview-header skeleton-header"></div>
                <div className="ranking-preview-list">
                  <div className="skeleton-item"></div>
                  <div className="skeleton-item"></div>
                  <div className="skeleton-item"></div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="ranking-preview-grid">
            {games.map(game => (
              <div key={game.id} className="ranking-preview-card">
                <div className="ranking-preview-header">
                  <span className="preview-icon">{game.icon}</span>
                  <span className="preview-title">{game.title}</span>
                </div>
                <div className="ranking-preview-list">
                  {rankings[game.id]?.length > 0 ? (
                    rankings[game.id].slice(0, 3).map((record, idx) => (
                      <div key={record.id} className="ranking-preview-item">
                        <span className="medal">{['🥇', '🥈', '🥉'][idx]}</span>
                        <span className="nickname">{record.nickname}</span>
                        <span className="score">{formatScore(record.score)}</span>
                      </div>
                    ))
                  ) : (
                    <p className="no-record">아직 기록이 없습니다</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
        
        <Link to="/ranking" className="view-all-btn">
          전체 랭킹 보기
        </Link>
      </section>
    </div>
  );
}

export default HomePage;

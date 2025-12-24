import { useState, useEffect } from 'react';
import './RankingPage.css';

function RankingPage() {
  const [allRankings, setAllRankings] = useState({});
  const [loading, setLoading] = useState(true);

  const games = [
    { id: 'jump-runner', name: 'Jump Runner', icon: '🐭' },
    { id: 'speed-click', name: 'Speed Click', icon: '🔴' }
  ];

  useEffect(() => {
    fetchAllRankings();
  }, []);

  const fetchAllRankings = async () => {
    setLoading(true);
    try {
      const results = {};
      await Promise.all(
        games.map(async (game) => {
          const response = await fetch(`/api/ranking?game=${game.id}&limit=10`);
          const data = await response.json();
          results[game.id] = data;
        })
      );
      setAllRankings(results);
    } catch (error) {
      console.error('Failed to fetch rankings:', error);
    }
    setLoading(false);
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ko-KR', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div className="ranking-page">
      <header className="ranking-header">
        <h1 className="pixel-font">RANKING</h1>
        <p>전세계 플레이어들과 경쟁하세요!</p>
      </header>

      {loading ? (
        <div className="loading">Loading...</div>
      ) : (
        <div className="rankings-grid">
          {games.map(game => (
            <div key={game.id} className="ranking-card">
              <div className="ranking-card-header">
                <span className="game-icon">{game.icon}</span>
                <h2>{game.name}</h2>
              </div>
              <div className="ranking-card-body">
                {!allRankings[game.id] || allRankings[game.id].length === 0 ? (
                  <div className="no-data">
                    <p>아직 기록이 없습니다.</p>
                  </div>
                ) : (
                  <table className="ranking-table">
                    <thead>
                      <tr>
                        <th>#</th>
                        <th>닉네임</th>
                        <th>점수</th>
                      </tr>
                    </thead>
                    <tbody>
                      {allRankings[game.id].map((record, index) => (
                        <tr key={record.id} className={index < 3 ? `top-${index + 1}` : ''}>
                          <td className="rank">
                            {index === 0 && <span className="medal">🥇</span>}
                            {index === 1 && <span className="medal">🥈</span>}
                            {index === 2 && <span className="medal">🥉</span>}
                            {index > 2 && <span>{index + 1}</span>}
                          </td>
                          <td className="nickname">{record.nickname}</td>
                          <td className="score">{record.score}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default RankingPage;

import { useState, useEffect } from 'react';
import './RankingPage.css';

function RankingPage() {
  const [rankings, setRankings] = useState([]);
  const [selectedGame, setSelectedGame] = useState('jump-runner');
  const [loading, setLoading] = useState(true);

  const games = [
    { id: 'jump-runner', name: 'Jump Runner' },
    { id: 'speed-click', name: 'Speed Click' }
  ];

  useEffect(() => {
    fetchRankings();
  }, [selectedGame]);

  const fetchRankings = async () => {
    setLoading(true);
    try {
      const response = await fetch(`/api/ranking?game=${selectedGame}&limit=10`);
      const data = await response.json();
      setRankings(data);
    } catch (error) {
      console.error('Failed to fetch rankings:', error);
      setRankings([]);
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

      <div className="game-filter">
        {games.map(game => (
          <button
            key={game.id}
            className={`filter-btn ${selectedGame === game.id ? 'active' : ''}`}
            onClick={() => setSelectedGame(game.id)}
          >
            {game.name}
          </button>
        ))}
      </div>

      <div className="ranking-table-container">
        {loading ? (
          <div className="loading">Loading...</div>
        ) : rankings.length === 0 ? (
          <div className="no-data">
            <p>아직 기록이 없습니다.</p>
            <p>첫 번째 기록을 세워보세요!</p>
          </div>
        ) : (
          <table className="ranking-table">
            <thead>
              <tr>
                <th>순위</th>
                <th>닉네임</th>
                <th>점수</th>
                <th>날짜</th>
              </tr>
            </thead>
            <tbody>
              {rankings.map((record, index) => (
                <tr key={record.id} className={index < 3 ? `top-${index + 1}` : ''}>
                  <td className="rank">
                    {index === 0 && <span className="medal gold">🥇</span>}
                    {index === 1 && <span className="medal silver">🥈</span>}
                    {index === 2 && <span className="medal bronze">🥉</span>}
                    {index > 2 && <span>{index + 1}</span>}
                  </td>
                  <td className="nickname">{record.nickname}</td>
                  <td className="score">{record.score}</td>
                  <td className="date">{formatDate(record.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default RankingPage;

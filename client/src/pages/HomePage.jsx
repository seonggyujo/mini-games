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
  }
];

function HomePage() {
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
    </div>
  );
}

export default HomePage;

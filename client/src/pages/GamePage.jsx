import { useParams } from 'react-router-dom';
import JumpRunner from '../components/games/JumpRunner/JumpRunner';
import './GamePage.css';

const gameComponents = {
  'jump-runner': JumpRunner
};

function GamePage() {
  const { gameId } = useParams();
  const GameComponent = gameComponents[gameId];

  if (!GameComponent) {
    return (
      <div className="game-page">
        <div className="game-not-found">
          <h2>Game Not Found</h2>
          <p>이 게임은 아직 준비 중입니다.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="game-page">
      <GameComponent />
    </div>
  );
}

export default GamePage;

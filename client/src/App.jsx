import { BrowserRouter, Routes, Route } from 'react-router-dom';
import NavBar from './components/common/NavBar';
import HomePage from './pages/HomePage';
import GamePage from './pages/GamePage';
import RankingPage from './pages/RankingPage';

function App() {
  return (
    <BrowserRouter>
      <NavBar />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/game/:gameId" element={<GamePage />} />
        <Route path="/ranking" element={<RankingPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
